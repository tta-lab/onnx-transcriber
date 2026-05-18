package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/guion-opensource/onnx-transcriber/internal/config"
	"github.com/guion-opensource/onnx-transcriber/internal/doctor"
	"github.com/guion-opensource/onnx-transcriber/internal/install"
	"github.com/guion-opensource/onnx-transcriber/internal/manifest"
	"github.com/guion-opensource/onnx-transcriber/internal/paragraph"
)

const (
	runtimeCPU  = "cpu"
	runtimeCUDA = "cuda"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stdout)
		return nil
	}
	switch args[0] {
	case "setup":
		return setup(args[1:], stdout)
	case "doctor":
		return doctorCmd(args[1:], stdout)
	case "models":
		return models(args[1:], stdout)
	case "help", "-h", "--help":
		usage(stdout)
		return nil
	default:
		return transcribe(args, stdout, stderr)
	}
}

func setup(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stdout)
	modelName := fs.String("model", "funasr-nano-int8", "ASR model to install")
	dataDir := fs.String("data-dir", "", "override data directory")
	fromDir := fs.String("from-dir", "", "install from local downloads directory")
	runtimeName := fs.String("runtime", runtimeCPU, "runtime to install: cpu or cuda")
	_ = fs.Bool("with-ffmpeg", false, "reserved for future ffmpeg bundling")
	if err := fs.Parse(args); err != nil {
		return err
	}
	runtimePlatform, err := runtimePlatformKey(config.CurrentPlatformKey(), *runtimeName)
	if err != nil {
		return err
	}

	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	inst := install.Installer{Home: home, Manifest: manifest.Default(), Stdout: stdout}
	if err := inst.EnsureDirs(); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "installing sherpa-onnx runtime:", *runtimeName)
	if err := inst.InstallRuntimePlatform("sherpa-onnx", runtimePlatform, *fromDir); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "installing ASR model:", *modelName)
	if err := inst.InstallModel(*modelName, *fromDir); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "installing VAD model: silero-vad")
	if err := inst.InstallModel("silero-vad", *fromDir); err != nil {
		return err
	}
	r := doctor.Run(home, doctor.Options{ModelName: *modelName, UseVAD: true, Runtime: *runtimeName})
	doctor.Write(stdout, r)
	fmt.Fprintln(stdout, "example: onnx-transcribe input.mp4 --threads 8 --out transcript.md")
	return nil
}

func doctorCmd(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stdout)
	modelName := fs.String("model", "funasr-nano-int8", "ASR model to check")
	dataDir := fs.String("data-dir", "", "override data directory")
	runtimeName := fs.String("runtime", runtimeCPU, "runtime to check: cpu or cuda")
	noVAD := fs.Bool("no-vad", false, "disable VAD model check")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if _, err := runtimePlatformKey(config.CurrentPlatformKey(), *runtimeName); err != nil {
		return err
	}
	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	r := doctor.Run(home, doctor.Options{ModelName: *modelName, UseVAD: !*noVAD, Runtime: *runtimeName})
	doctor.Write(stdout, r)
	if !r.OK {
		return errors.New("setup is incomplete")
	}
	return nil
}

func models(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "list" {
		for name, model := range manifest.Default().Models {
			if model.Backend != "" {
				fmt.Fprintf(stdout, "%s\t%s\t%s\n", name, model.Type, model.Backend)
				continue
			}
			fmt.Fprintf(stdout, "%s\t%s\n", name, model.Type)
		}
		return nil
	}
	if args[0] != "install" {
		return fmt.Errorf("unknown models command %q", args[0])
	}
	fs := flag.NewFlagSet("models install", flag.ContinueOnError)
	fs.SetOutput(stdout)
	dataDir := fs.String("data-dir", "", "override data directory")
	fromDir := fs.String("from-dir", "", "install from local downloads directory")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: onnx-transcribe models install <name>")
	}
	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	inst := install.Installer{Home: home, Manifest: manifest.Default(), Stdout: stdout}
	if err := inst.EnsureDirs(); err != nil {
		return err
	}
	return inst.InstallModel(fs.Arg(0), *fromDir)
}

func transcribe(args []string, stdout, stderr io.Writer) error {
	args = normalizeTranscribeArgs(args)
	fs := flag.NewFlagSet("transcribe", flag.ContinueOnError)
	fs.SetOutput(stderr)
	modelName := fs.String("model", "funasr-nano-int8", "model registry name")
	hotwords := fs.String("hotwords", "", "hotwords file")
	out := fs.String("out", "transcript.md", "markdown output path")
	dataDir := fs.String("data-dir", "", "override data directory")
	runtimeName := fs.String("runtime", runtimeCPU, "runtime to use: cpu or cuda")
	hotwordsScore := fs.Float64("hotwords-score", 1.5, "hotwords score")
	keepTemp := fs.Bool("keep-temp", false, "keep temporary working directory")
	threads := fs.Int("threads", runtime.NumCPU(), "ASR CPU threads")
	noVAD := fs.Bool("no-vad", false, "disable Silero VAD before offline ASR")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: onnx-transcribe input.mp4 --threads 8 --out transcript.md")
	}
	runtimePlatform, err := runtimePlatformKey(config.CurrentPlatformKey(), *runtimeName)
	if err != nil {
		return err
	}
	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	r := runner{home: home, runtimeName: *runtimeName, runtimePlatform: runtimePlatform, stderr: stderr}
	md, err := r.transcribe(fs.Arg(0), *modelName, *hotwords, *hotwordsScore, *keepTemp, *threads, !*noVAD)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(stdout, *out)
	return nil
}

func normalizeTranscribeArgs(args []string) []string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return args
	}
	normalized := make([]string, 0, len(args))
	normalized = append(normalized, args[1:]...)
	normalized = append(normalized, args[0])
	return normalized
}

type runner struct {
	home            string
	runtimeName     string
	runtimePlatform string
	stderr          io.Writer
}

type asrConfig struct {
	Encoder     string
	Embedding   string
	LLM         string
	Tokenizer   string
	VADModel    string
	Hotwords    []string
	HotwordFile string
	Threads     int
	UseVAD      bool
}

func defaultModelForBackend(backend string) (string, error) {
	switch backend {
	case "nano":
		return "funasr-nano-int8", nil
	default:
		return "", fmt.Errorf("unknown backend %q", backend)
	}
}

func (r runner) transcribe(input, modelName, hotwords string, hotwordsScore float64, keepTemp bool, threads int, useVAD bool) (string, error) {
	workDir, err := os.MkdirTemp("", "onnx-transcribe-*")
	if err != nil {
		return "", err
	}
	if !keepTemp {
		defer os.RemoveAll(workDir)
	}

	wav := filepath.Join(workDir, "audio.wav")
	if err := runCommand(r.stderr, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", input, "-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wav); err != nil {
		return "", err
	}

	cfg, err := r.asrConfig(modelName, hotwords, threads, useVAD)
	if err != nil {
		return "", err
	}

	rawText, err := r.runASR(wav, cfg)
	if err != nil {
		return "", err
	}
	if rawText == "" {
		return "", errors.New("ASR returned no text")
	}

	body := paragraph.Format(rawText, paragraph.Options{SentencesPerParagraph: 3, MaxChars: 500})
	return "# Transcript\n\n" + body + "\n", nil
}

func (r runner) asrConfig(modelName, hotwordsPath string, threads int, useVAD bool) (asrConfig, error) {
	cfg := asrConfig{Threads: threads, UseVAD: useVAD, HotwordFile: hotwordsPath}
	asrModelDir := config.ModelDir(r.home, modelName)

	if useVAD {
		vadModelDir := config.ModelDir(r.home, "silero-vad")
		vadModel, err := findFirst(vadModelDir, "silero_vad.onnx")
		if err != nil {
			return cfg, fmt.Errorf("VAD model missing under %s: %w", vadModelDir, err)
		}
		cfg.VADModel = vadModel
	}

	encoder, err := findFirst(asrModelDir, "encoder_adaptor.int8.onnx", "encoder_adaptor.onnx")
	if err != nil {
		return cfg, fmt.Errorf("FunASR-Nano encoder adaptor missing under %s: %w", asrModelDir, err)
	}
	embedding, err := findFirst(asrModelDir, "embedding.int8.onnx", "embedding.onnx")
	if err != nil {
		return cfg, fmt.Errorf("FunASR-Nano embedding missing under %s: %w", asrModelDir, err)
	}
	llm, err := findFirst(asrModelDir, "llm.int8.onnx", "llm.fp16.onnx", "llm.onnx")
	if err != nil {
		return cfg, fmt.Errorf("FunASR-Nano LLM missing under %s: %w", asrModelDir, err)
	}
	tokenizer, err := findDir(asrModelDir, "Qwen3-0.6B")
	if err != nil {
		return cfg, fmt.Errorf("FunASR-Nano tokenizer missing under %s: %w", asrModelDir, err)
	}
	cfg.Encoder = encoder
	cfg.Embedding = embedding
	cfg.LLM = llm
	cfg.Tokenizer = tokenizer
	cfg.Hotwords, err = readHotwords(hotwordsPath)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func readHotwords(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var words []string
	for _, line := range strings.Split(string(content), "\n") {
		word := strings.TrimSpace(line)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
	}
	return words, nil
}

func (r runner) runASR(wav string, cfg asrConfig) (string, error) {
	if cfg.UseVAD {
		asrBin, err := r.binary("sherpa-onnx-vad-with-offline-asr")
		if err != nil {
			return "", err
		}
		raw, err := outputCommand(r.stderr, asrBin, buildASRArgs(cfg, wav)...)
		if err != nil {
			return "", r.runtimeCommandError(err)
		}
		return extractVADText(raw), nil
	}

	asrBin, err := r.binary("sherpa-onnx-offline")
	if err != nil {
		return "", err
	}
	raw, err := outputCommand(r.stderr, asrBin, buildASRArgs(cfg, wav)...)
	if err != nil {
		return "", r.runtimeCommandError(err)
	}
	return extractASRText(raw), nil
}

func (r runner) runtimeCommandError(err error) error {
	if r.runtimeName != runtimeCUDA {
		return err
	}
	return fmt.Errorf("CUDA runtime failed; run onnx-transcribe doctor --runtime cuda: %w", err)
}

func buildASRArgs(cfg asrConfig, wav string) []string {
	args := make([]string, 0, 12)
	if cfg.UseVAD {
		args = append(args, "--silero-vad-model="+cfg.VADModel)
	}

	args = append(args,
		"--funasr-nano-encoder-adaptor="+cfg.Encoder,
		"--funasr-nano-embedding="+cfg.Embedding,
		"--funasr-nano-llm="+cfg.LLM,
		"--funasr-nano-tokenizer="+cfg.Tokenizer,
		"--funasr-nano-language=中文",
		"--funasr-nano-itn=true",
	)
	if len(cfg.Hotwords) > 0 {
		args = append(args, "--funasr-nano-hotwords="+strings.Join(cfg.Hotwords, ","))
	}

	args = append(args,
		"--print-args=false",
		"--num-threads="+strconv.Itoa(cfg.Threads),
		wav,
	)
	return args
}

func (r runner) binary(name string) (string, error) {
	if strings.HasPrefix(r.runtimePlatform, "windows-") {
		name += ".exe"
	}
	platform := r.runtimePlatform
	if platform == "" {
		platform = config.CurrentPlatformKey()
	}
	root := config.RuntimeBinDir(r.home, platform)
	if local, err := findRuntimeBinary(root, name); err == nil {
		return local, nil
	}
	if r.runtimeName == runtimeCUDA {
		return "", errors.New("CUDA runtime not installed; run onnx-transcribe setup --runtime cuda")
	}
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("%s not found; run onnx-transcribe setup", name)
}

func runtimePlatformKey(basePlatform, runtimeName string) (string, error) {
	switch runtimeName {
	case "", runtimeCPU:
		return basePlatform, nil
	case runtimeCUDA:
		return basePlatform + "-cuda", nil
	default:
		return "", fmt.Errorf("unknown runtime %q; use cpu or cuda", runtimeName)
	}
}

func findRuntimeBinary(root, name string) (string, error) {
	var fallback string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(path) != name {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) == "bin" {
			fallback = path
			return filepath.SkipAll
		}
		if fallback == "" {
			fallback = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if fallback == "" {
		return "", os.ErrNotExist
	}
	return fallback, nil
}

func runCommand(stderr io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stderr
	cmd.Stderr = stderr
	return cmd.Run()
}

func outputCommand(stderr io.Writer, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil && errOut.Len() > 0 {
		_, _ = io.Copy(stderr, &errOut)
	}
	return out.String(), err
}

func extractASRText(output string) string {
	type result struct {
		Text string `json:"text"`
	}
	var r result
	if err := json.Unmarshal([]byte(output), &r); err == nil {
		return strings.TrimSpace(r.Text)
	}
	return strings.TrimSpace(output)
}

var vadLinePattern = regexp.MustCompile(`^\s*\d+(?:\.\d+)?\s+--\s+\d+(?:\.\d+)?:\s*(.+?)\s*$`)

func extractVADText(output string) string {
	var parts []string
	for _, line := range strings.Split(output, "\n") {
		match := vadLinePattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		text := strings.TrimSpace(match[1])
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func findFirst(root string, names ...string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base := filepath.Base(path)
		for _, name := range names {
			if base == name {
				found = path
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", os.ErrNotExist
	}
	return found, nil
}

func findDir(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if filepath.Base(path) == name {
			found = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", os.ErrNotExist
	}
	return found, nil
}

func resolveHome(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return config.HomeDir()
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "onnx-transcribe input.mp4 --threads 8 --out transcript.md")
	fmt.Fprintln(w, "onnx-transcribe setup [--runtime cpu|cuda]")
	fmt.Fprintln(w, "onnx-transcribe models list")
	fmt.Fprintln(w, "onnx-transcribe models install <name>")
	fmt.Fprintln(w, "onnx-transcribe doctor [--runtime cpu|cuda]")
}
