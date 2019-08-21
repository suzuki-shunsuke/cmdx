find . \
  -type d -name .git -prune -o \
  -type f -print | \
  durl check || exit 1
