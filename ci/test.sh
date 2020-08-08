#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

mkdir -p bin
curl -L -o bin/cc-test-reporter https://codeclimate.com/downloads/test-reporter/test-reporter-0.6.3-linux-amd64
chmod a+x bin/cc-test-reporter
export PATH="$PWD/bin:$PATH"
bash scripts/test-code-climate.sh
