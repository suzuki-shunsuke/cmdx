git fetch || exit 1
git tag || exit 1
t=`git tag | tail -n 1`
if [ -n "$t" ]; then
  git tag ${t}-alpha
else
  git tag v0.1.0-alpha
fi
