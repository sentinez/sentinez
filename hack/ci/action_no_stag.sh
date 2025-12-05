#!/usr/bin/env bash
set -euo pipefail

NAME="${1:-core}"
PREFIX="${NAME}"
BRANCH=$(git rev-parse --abbrev-ref HEAD)
REMOTE="https://kyerans:${SENZ_GITHUB_TOKEN}@github.com/sentinez/${NAME}.git"
REMOTE_SAFE="https://github.com/sentinez/${NAME}.git"

git config --global pull.rebase false
git config --global rebase.autoStash false
git config --global merge.strategy ours
git config user.name "Duc-Hung Ho"
git config user.email "iduchungho@gmail.com"

echo "MODULE: $NAME"
echo "PREFIX: $PREFIX"
echo "BRANCH: $BRANCH"

SUBTREE_SHA=$(git subtree split --prefix="$PREFIX" "$BRANCH")
git config --unset-all http.https://github.com/.extraheader || true
git fetch "$REMOTE" "$BRANCH":"tmp_remote" || true

if git rev-parse --verify tmp_remote >/dev/null 2>&1; then
    git merge -s ours tmp_remote \
        -m "Merge remote branch $BRANCH (history only, keep subtree)" \
        --allow-unrelated-histories
fi

git push "$REMOTE" "$SUBTREE_SHA:$BRANCH" --force
git branch -D tmp_remote || true
echo "Subtree sync complete."
