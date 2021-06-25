#!/bin/bash

if ! [ -x "$(command -v "go")" ]; then
  echo -e "ERROR: golang is not available. Please install the latest version of Go from https://golang.org/dl/!"
  exit 1
fi

echo "Running rendertemplates pre-commit hook..."

cd "$(git rev-parse --show-toplevel)"
  if files=`go run development/tools/cmd/rendertemplates/main.go -config templates/config.yaml -show-output-dir=true`; then
    modified=`git diff --exit-code --name-only $files` && echo "No changes made. Continuing..." || echo "Templates have been regenerated and automatically added to your commit.

Modified files:
$modified"
  else
    echo "Failed to render templates. Commit aborted."
    exit 1
  fi
  git add -u $files
cd - &> /dev/null
