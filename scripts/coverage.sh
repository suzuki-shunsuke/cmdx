ee() {
  echo "+ $@"
  eval "$@"
}

cd `dirname $0`/..

if [ "$1" = "" ]; then
  target=`find pkg -type d | fzf`
  if [ "$target" = "" ]; then
    exit 0
  fi
else
  target=$1
fi

if [ ! -d "$target" ]; then
  echo "$target is not found" >&2
  exit 1
fi

ee mkdir -p .coverage/$target || exit 1
ee go test ./$target -coverprofile=.coverage/$target/coverage.txt -covermode=atomic || exit 1
ee go tool cover -html=.coverage/$target/coverage.txt
