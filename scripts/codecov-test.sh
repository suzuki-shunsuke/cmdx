#!/usr/bin/env bash
# https://github.com/codecov/example-go#caveat-multiple-files

echo "" > coverage.txt

for d in go list ./...; do
  go test -race -coverprofile=profile.out -covermode=atomic $d || exit 1
  if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done
