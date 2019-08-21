git fetch
git tag
t=`git tag | tail -n 1`
if [ -n "$t" ]; then
  git tag ${t}-alpha || exit 1
else
  git tag v0.1.0-alpha || exit 1
fi
