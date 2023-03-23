FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

ARG deno_version=1.32.0

RUN cd /usr/local/bin \
    && wget https://github.com/denoland/deno/releases/download/v$deno_version/deno-x86_64-unknown-linux-gnu.zip \
    && unzip deno-x86_64-unknown-linux-gnu.zip \
    && chmod +x deno

USER vscode
