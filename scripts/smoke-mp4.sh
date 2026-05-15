#!/usr/bin/env sh
set -eu

fixture="testdata/media/space-illusion-720p-h265.mp4"

if [ ! -f "$fixture" ]; then
  echo "missing fixture: $fixture" >&2
  echo "copy an MP4 into testdata/media/ or restore it from ~/Downloads" >&2
  exit 1
fi

go run ./cmd/onnx-transcribe "$fixture" \
  --hotwords testdata/hotwords.txt \
  --out testdata/transcript.smoke.md \
  "$@"
