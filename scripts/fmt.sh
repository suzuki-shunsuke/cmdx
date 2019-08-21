find . -type d -name node_modules -prune -o \
  -type d -name .git -prune -o \
  -type d -name vendor -prune -o \
  -type f -name "*.go" -print \
  | xargs gofmt -l -s -w
