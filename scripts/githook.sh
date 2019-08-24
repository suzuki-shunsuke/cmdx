echoEval() {
  echo "+ $@"
  eval "$@"
}

cd `dirname $0`/..
if [ ! -f .git/hooks/pre-commit ]; then
  echoEval ln -s ../../githook/pre-commit.sh .git/hooks/pre-commit || exit 1
fi
echoEval chmod a+x githook/*
