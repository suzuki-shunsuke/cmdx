#!/usr/bin/env sh

find . \
  -type d -name .git -prune -o \
  -type f -name "*.go" -print0 \
  | xargs gofmt -l -s -w
