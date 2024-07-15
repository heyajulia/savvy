#!/usr/bin/env bash

set -euxo pipefail

ensure_vscode_ownership() {
  if [ -d "$1" ]; then
    sudo chown vscode:vscode "$1"
  fi
}

configure_git_safe_directory() {
  if [ -d "$1" ]; then
    git config --global --add safe.directory "$1"
  fi
}

ensure_vscode_ownership "/home/vscode/.cache" # Thanks to https://go.dev/issue/42353#issuecomment-721913348
ensure_vscode_ownership "/IdeaProjects/energieprijzen"

configure_git_safe_directory "/IdeaProjects"
configure_git_safe_directory "/workspaces/energieprijzen"
