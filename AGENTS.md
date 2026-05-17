# Repository Guidelines

## Project Structure & Module Organization

This is a Go CLI for local MP4-to-Markdown transcription. The entry point is `cmd/onnx-transcribe/`. Packages under `internal/`:

- `internal/config`: data directory and path resolution.
- `internal/doctor`: environment and dependency checks.
- `internal/install`: setup/download workflow.
- `internal/manifest`: runtime asset metadata.
- `internal/paragraph`: transcript paragraph formatting.

Guides live in `docs/`, helper scripts in `scripts/`, and fixtures in `testdata/`. Large media should stay out of git; the smoke test expects a local file at `testdata/media/space-illusion-30s.mp4`.

## Build, Test, and Development Commands

- `make build`: builds `./onnx-transcribe` from `./cmd/onnx-transcribe`.
- `make run ARGS='doctor'`: builds, then runs the CLI with arguments.
- `make setup`: installs the binary and prints local setup steps.
- `make test`: runs `go test -v -count=1 ./...`.
- `make fmt`: runs `gofmt -w -s .`.
- `make lint`: runs `golangci-lint run ./...`; requires `golangci-lint` locally.
- `make ci`: runs lint, tests, and build.
- `sh scripts/smoke-mp4.sh`: runs the local MP4 smoke test when the fixture exists.

## Coding Style & Naming Conventions

Use standard Go style. Format with `gofmt -s`; do not hand-align. Keep package names short and lowercase. Prefer guard clauses and early error returns. Keep command-line behavior in `cmd/onnx-transcribe` and reusable logic in `internal/`.

Use names that match the domain: `manifest`, `paths`, `doctor`, `paragraph`. Avoid broad utility packages unless shared behavior is stable.

## Testing Guidelines

Tests use the standard Go `testing` package. Place tests beside the code as `*_test.go`, as in `internal/paragraph/paragraph_test.go`. Prefer table-driven tests for parsing, path, and formatting logic. Run `make test` before submitting changes; run `make ci` when touching build, install, or CLI behavior.

Smoke tests may require local runtime binaries and the ignored MP4 fixture. Do not commit generated audio, video, or model files.

## Commit & Pull Request Guidelines

Use scoped Conventional Commit messages, matching the history:

```text
feat(cli): default to vad transcription
fix(cli): verify short seaco pipeline
chore(build): add make workflow
docs(guide): tidy fixture section markup
```

PRs should describe the behavior change, list verification commands, and note any fixture or external binary requirements. Include screenshots only for `docs/*.html` visual changes. Do not include generated binaries, downloaded models, or local data directories.

## Security & Configuration Tips

The CLI uses local data paths by default and supports `ONNX_TRANSCRIBER_HOME` or `--data-dir` overrides. Keep downloaded model archives, runtime binaries, transcripts with private content, and local smoke media out of git.
