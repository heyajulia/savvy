FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

ARG deno_version=1.32.1
ARG dhall_haskell_version=1.41.2
ARG dhall_json_version=1.7.11
ARG dhall_lsp_server_version=1.1.2

RUN cd /usr/local/bin \
        && wget https://github.com/denoland/deno/releases/download/v$deno_version/deno-x86_64-unknown-linux-gnu.zip \
        https://github.com/dhall-lang/dhall-haskell/releases/download/$dhall_haskell_version/dhall-$dhall_haskell_version-x86_64-Linux.tar.bz2 \
        https://github.com/dhall-lang/dhall-haskell/releases/download/$dhall_haskell_version/dhall-json-$dhall_json_version-x86_64-Linux.tar.bz2 \
        https://github.com/dhall-lang/dhall-haskell/releases/download/$dhall_haskell_version/dhall-lsp-server-$dhall_lsp_server_version-x86_64-Linux.tar.bz2 \
        && unzip deno-x86_64-unknown-linux-gnu.zip \
        && chmod +x deno \
        && tar xf dhall-$dhall_haskell_version-x86_64-Linux.tar.bz2 \
        && mv bin/* . \
        && tar xf dhall-json-$dhall_json_version-x86_64-Linux.tar.bz2 \
        && mv bin/* . \
        && tar xf dhall-lsp-server-$dhall_lsp_server_version-x86_64-Linux.tar.bz2 \
        && mv bin/* . \
        && rm -r bin deno-x86_64-unknown-linux-gnu.zip dhall-$dhall_haskell_version-x86_64-Linux.tar.bz2 dhall-json-$dhall_json_version-x86_64-Linux.tar.bz2 dhall-lsp-server-$dhall_lsp_server_version-x86_64-Linux.tar.bz2

USER vscode
