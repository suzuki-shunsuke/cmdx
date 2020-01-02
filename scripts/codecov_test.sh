#!/usr/bin/env bash
# https://github.com/codecov/example-go#caveat-multiple-files

ee() {
  echo "+ $*"
  eval "$@"
}

echo "" > coverage.txt

for d in $(go list ./...); do
  echo "$d"
  ee go test -race -coverprofile=profile.out -covermode=atomic "$d" || exit 1
  if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done
