package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/guion-opensource/onnx-transcriber/internal/config"
	"github.com/guion-opensource/onnx-transcriber/internal/doctor"
	"github.com/guion-opensource/onnx-transcriber/internal/install"
	"github.com/guion-opensource/onnx-transcriber/internal/manifest"
	"github.com/guion-opensource/onnx-transcriber/internal/paragraph"
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
	modelName := fs.String("model", "seaco-paraformer-trilingual", "ASR model to install")
	dataDir := fs.String("data-dir", "", "override data directory")
	fromDir := fs.String("from-dir", "", "install from local downloads directory")
	_ = fs.Bool("with-ffmpeg", false, "reserved for future ffmpeg bundling")
	if err := fs.Parse(args); err != nil {
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
	fmt.Fprintln(stdout, "installing sherpa-onnx runtime")
	if err := inst.InstallRuntime("sherpa-onnx", *fromDir); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "installing ASR model:", *modelName)
	if err := inst.InstallModel(*modelName, *fromDir); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "installing punctuation model: ct-transformer-zh-en")
	if err := inst.InstallModel("ct-transformer-zh-en", *fromDir); err != nil {
		return err
	}
	r := doctor.Run(home)
	doctor.Write(stdout, r)
	fmt.Fprintln(stdout, "example: onnx-transcribe input.mp4 --hotwords hotwords.txt --out transcript.md")
	return nil
}

func doctorCmd(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stdout)
	dataDir := fs.String("data-dir", "", "override data directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	r := doctor.Run(home)
	doctor.Write(stdout, r)
	if !r.OK {
		return errors.New("setup is incomplete")
	}
	return nil
}

func models(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "list" {
		for name, model := range manifest.Default().Models {
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
	modelName := fs.String("model", "seaco-paraformer-trilingual", "model registry name")
	hotwords := fs.String("hotwords", "", "hotwords file")
	out := fs.String("out", "transcript.md", "markdown output path")
	dataDir := fs.String("data-dir", "", "override data directory")
	hotwordsScore := fs.Float64("hotwords-score", 1.5, "hotwords score")
	keepTemp := fs.Bool("keep-temp", false, "keep temporary working directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: onnx-transcribe input.mp4 --hotwords hotwords.txt --out transcript.md")
	}
	home, err := resolveHome(*dataDir)
	if err != nil {
		return err
	}
	r := runner{home: home, stderr: stderr}
	md, err := r.transcribe(fs.Arg(0), *modelName, *hotwords, *hotwordsScore, *keepTemp)
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
	home   string
	stderr io.Writer
}

func (r runner) transcribe(input, modelName, hotwords string, hotwordsScore float64, keepTemp bool) (string, error) {
	workDir, err := os.MkdirTemp("", "onnx-transcribe-*")
	if err != nil {
		return "", err
	}
	if !keepTemp {
		defer os.RemoveAll(workDir)
	}

	wav := filepath.Join(workDir, "audio.wav")
	if err := runCommand(r.stderr, "ffmpeg", "-y", "-i", input, "-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wav); err != nil {
		return "", err
	}

	asrBin, err := r.binary("sherpa-onnx-offline")
	if err != nil {
		return "", err
	}
	asrModelDir := config.ModelDir(r.home, modelName)
	model, err := findFirst(asrModelDir, "model.int8.onnx", "model.onnx")
	if err != nil {
		return "", fmt.Errorf("ASR model files missing under %s: %w", asrModelDir, err)
	}
	tokens, err := findFirst(asrModelDir, "tokens.txt")
	if err != nil {
		return "", fmt.Errorf("tokens missing under %s: %w", asrModelDir, err)
	}

	asrArgs := []string{"--tokens", tokens, "--paraformer", model, "--decoding-method", "modified_beam_search"}
	if hotwords != "" {
		asrArgs = append(asrArgs, "--hotwords-file", hotwords, "--hotwords-score", strconv.FormatFloat(hotwordsScore, 'f', -1, 64))
	}
	asrArgs = append(asrArgs, wav)
	raw, err := outputCommand(r.stderr, asrBin, asrArgs...)
	if err != nil {
		return "", err
	}

	punctuated := strings.TrimSpace(raw)
	if punctBin, err := r.binary("sherpa-onnx-offline-punctuation"); err == nil {
		if punct, err := r.punctuate(punctBin, punctuated); err == nil && strings.TrimSpace(punct) != "" {
			punctuated = punct
		}
	}

	body := paragraph.Format(punctuated, paragraph.Options{SentencesPerParagraph: 3, MaxChars: 500})
	return "# Transcript\n\n" + body + "\n", nil
}

func (r runner) punctuate(bin, text string) (string, error) {
	modelDir := config.ModelDir(r.home, "ct-transformer-zh-en")
	model, err := findFirst(modelDir, "model.int8.onnx", "model.onnx")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(bin, "--model", model)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stderr = r.stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	return out.String(), cmd.Run()
}

func (r runner) binary(name string) (string, error) {
	if filepath.Separator == '\\' {
		name += ".exe"
	}
	local := filepath.Join(config.RuntimeBinDir(r.home, config.CurrentPlatformKey()), name)
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("%s not found; run onnx-transcribe setup", name)
}

func runCommand(stderr io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stderr
	cmd.Stderr = stderr
	return cmd.Run()
}

func outputCommand(stderr io.Writer, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	return out.String(), cmd.Run()
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

func resolveHome(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return config.HomeDir()
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "onnx-transcribe input.mp4 --hotwords hotwords.txt --out transcript.md")
	fmt.Fprintln(w, "onnx-transcribe setup [--model seaco-paraformer-trilingual]")
	fmt.Fprintln(w, "onnx-transcribe models list")
	fmt.Fprintln(w, "onnx-transcribe models install <name>")
	fmt.Fprintln(w, "onnx-transcribe doctor")
}
