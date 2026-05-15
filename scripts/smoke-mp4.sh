#!/usr/bin/env sh
set -eu

fixture="testdata/media/131202 №30431 时空的虚幻 H265-高清-720P.mp4"

if [ ! -f "$fixture" ]; then
  echo "missing fixture: $fixture" >&2
  echo "copy an MP4 into testdata/media/ or restore it from ~/Downloads" >&2
  exit 1
fi

go run ./cmd/onnx-transcribe "$fixture" \
  --hotwords testdata/hotwords.txt \
  --out testdata/transcript.smoke.md \
  "$@"
