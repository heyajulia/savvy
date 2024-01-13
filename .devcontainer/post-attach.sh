#!/usr/bin/env bash

set -euxo pipefail

if [ -d "/home/vscode/.cache" ]; then
  # Thanks to https://golang.org/issue/42353#issuecomment-721913348
  sudo chown vscode:vscode /home/vscode/.cache
fi

if [ -d "/IdeaProjects/energieprijzen" ]; then
  sudo chown vscode:vscode /IdeaProjects/energieprijzen
fi

if [ -d "/IdeaProjects" ]; then
  git config --global --add safe.directory /IdeaProjects
fi
