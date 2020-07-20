#!/usr/bin/env bash

set -eu
set -o pipefail

git ls-files | grep -E "\.go$" |
  xargs gofumpt -l -s -w
