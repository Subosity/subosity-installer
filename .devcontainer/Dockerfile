FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/New_York
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # Essential tools
    bash \
    curl \
    wget \
    git \
    make \
    sudo \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    tzdata \
    # Build tools
    build-essential \
    gcc \
    g++ \
    # Development tools
    vim \
    nano \
    tree \
    htop \
    jq \
    unzip \
    zip \
    tar \
    gzip \
    # GVM requirements (all packages gvm needs)
    bsdmainutils \
    util-linux \
    binutils \
    bison \
    mercurial \
    pkg-config \
    libc6-dev \
    gcc \
    g++ \
    make \
    git \
    curl \
    # Network tools
    openssh-client \
    net-tools \
    iputils-ping \
    # Database client
    postgresql-client \
    # Node.js and npm
    nodejs \
    npm \
    # Clean up
    && rm -rf /var/lib/apt/lists/*

# Configure timezone
RUN ln -sf /usr/share/zoneinfo/America/New_York /etc/localtime && \
    echo "America/New_York" > /etc/timezone

# Install Docker CE
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# Install Supabase CLI for development/testing
RUN ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = "amd64" ]; then \
    wget -qO- https://github.com/supabase/cli/releases/latest/download/supabase_linux_amd64.tar.gz | tar -xz -C /usr/local/bin; \
    elif [ "$ARCH" = "arm64" ]; then \
    wget -qO- https://github.com/supabase/cli/releases/latest/download/supabase_linux_arm64.tar.gz | tar -xz -C /usr/local/bin; \
    fi

# Create non-root user with proper shell and groups FIRST
RUN groupadd --gid 1000 vscode \
    && useradd --uid 1000 --gid vscode --shell /bin/bash --create-home vscode \
    && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers \
    && usermod -aG docker vscode \
    && mkdir -p /go \
    && chown -R vscode:vscode /go

# Install GVM (Go Version Manager) for vscode user
USER vscode
RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash

# Switch back to root for final setup, then back to vscode
USER root

# Set up Go paths (will be updated by post-setup script)
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Switch to vscode user for final setup
USER vscode

# Set working directory
WORKDIR /workspaces/subosity-installer
