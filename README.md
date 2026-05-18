# onnx-transcriber

Local-first MP4 to Markdown transcription using Go orchestration and sherpa-onnx runtime binaries.

```bash
go run ./cmd/onnx-transcribe --help
go run ./cmd/onnx-transcribe setup
go run ./cmd/onnx-transcribe input.mp4 --threads 8 --out transcript.md
go run ./cmd/onnx-transcribe input.mp4 --backend nano --hotwords testdata/hotwords.txt --out transcript.md
```

The binary name is `onnx-transcribe`.

## Scope

P0 is mac-first and shells out to:

- `ffmpeg`
- `sherpa-onnx-offline`
- `sherpa-onnx-offline-punctuation`
- `sherpa-onnx-vad-with-offline-asr`

The Go code owns CLI flags, asset locations, setup, doctor checks, command orchestration, paragraph formatting, and Markdown writing.

## Data Directory

Default install paths:

- macOS: `~/Library/Application Support/onnx-transcriber`
- Linux: `~/.local/share/onnx-transcriber`
- Windows: `%LOCALAPPDATA%\onnx-transcriber`

Override with:

```bash
ONNX_TRANSCRIBER_HOME=/path/to/data onnx-transcribe doctor
onnx-transcribe setup --data-dir ./vendor/onnx-transcriber
```

## Current Caveats

- `sensevoice` is the default backend. `nano` is available for experimental FunASR-Nano transcription and passes `--hotwords` as prompt hotwords.
- The SenseVoice and Nano ASR archives do not publish checked-in GitHub asset digests, so the downloader cannot verify them yet without separate known SHA256 values.
- The first smoke test for a new runtime should verify backend flags with `onnx-transcribe doctor --backend <name>` and compare quality on the same Mandarin lecture clip.

## Local Smoke Fixture

Large media files are ignored. For local testing, this repo expects the copied MP4 at:

```text
testdata/media/space-illusion-30s.mp4
```

Run:

```bash
sh scripts/smoke-mp4.sh
```
