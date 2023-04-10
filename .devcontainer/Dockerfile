FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

ARG deno_version=1.32.1
ARG jsonnet_language_server_version=0.11.0

RUN cd /usr/local/bin \
        && wget https://github.com/denoland/deno/releases/download/v$deno_version/deno-x86_64-unknown-linux-gnu.zip \
        && wget -O jsonnet-language-server https://github.com/grafana/jsonnet-language-server/releases/download/v${jsonnet_language_server_version}/jsonnet-language-server_${jsonnet_language_server_version}_linux_amd64 \
        && unzip deno-x86_64-unknown-linux-gnu.zip \
        && chmod +x deno \
        && chmod +x jsonnet-language-server \
        && cd /tmp \
        && git clone --depth 1 https://github.com/google/jsonnet \
        && cd jsonnet \
        && make -j \
        && cp jsonnet jsonnetfmt /usr/local/bin \
        && rm -rf /tmp/jsonnet

USER vscode
