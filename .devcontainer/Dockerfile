FROM docker:24.0.7-dind

# Install dependencies (excluding Go - will be managed by gvm)
RUN apk add --no-cache \
    bash \
    curl \
    git \
    make \
    nodejs \
    npm \
    postgresql-client \
    libc6-compat \
    sudo \
    shadow \
    ca-certificates \
    wget \
    tar \
    gzip \
    openssh-client \
    build-base \
    binutils \
    mercurial \
    gcc \
    musl-dev

# Install Go Version Manager (gvm)
RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash

# Set up environment for gvm
ENV GVM_ROOT=/root/.gvm
ENV PATH=$GVM_ROOT/bin:$PATH

# Set up Go paths (will be updated by post-setup script)
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Install Supabase CLI for development/testing
RUN wget -qO- https://github.com/supabase/cli/releases/latest/download/supabase_linux_amd64.tar.gz | tar -xz -C /usr/local/bin

# Create non-root user and add to sudoers + docker group
RUN addgroup -g 1000 vscode && \
    adduser -u 1000 -G vscode -s /bin/bash -D vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    usermod -aG docker vscode && \
    mkdir -p /go && \
    chown -R vscode:vscode /go

USER vscode

# Set working directory
WORKDIR /workspaces/subosity-installer
