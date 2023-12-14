FROM mcr.microsoft.com/devcontainers/go:1-1.21-bookworm

RUN cd $(mktemp -d) && \
    wget https://github.com/a-h/templ/releases/download/v0.2.476/templ_Linux_x86_64.tar.gz && \
    tar xf templ_Linux_x86_64.tar.gz && \
    mv templ /usr/local/bin
