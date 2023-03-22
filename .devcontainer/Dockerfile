FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

USER vscode

RUN curl -fsSL https://deno.land/x/install/install.sh | sh

ENV PATH=/home/vscode/.deno/bin:$PATH
