FROM mcr.microsoft.com/devcontainers/go:1-1.21-bookworm

RUN DIR=$(mktemp -d) && \
    cd $DIR && \
    wget https://github.com/a-h/templ/releases/download/v0.2.663/templ_Linux_x86_64.tar.gz && \
    tar xf templ_Linux_x86_64.tar.gz && \
    mv templ /usr/local/bin && \
    rm -r $DIR && \
    DIR=$(mktemp -d) && \
    cd $DIR && \
    wget https://github.com/charmbracelet/gum/releases/download/v0.13.0/gum_0.13.0_Linux_x86_64.tar.gz && \
    tar xf gum_0.13.0_Linux_x86_64.tar.gz && \
    mv gum /usr/local/bin && \
    rm -r $DIR
