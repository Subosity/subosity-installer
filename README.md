# 🚀 Subosity Installer

A turnkey, production-ready deployment tool for self-hosting the Subosity application. Eliminates the complexity of manual Docker, Supabase, and SSL configuration with a single command that handles everything from system dependencies to service management.

## ⚡ Quick Start

### Binary Installation (Recommended)

```bash
# Download and verify latest release
curl -fsSL https://github.com/subosity/subosity-installer/releases/latest/download/subosity-installer-linux-amd64 -o subosity-installer
curl -fsSL https://github.com/subosity/subosity-installer/releases/latest/download/subosity-installer-linux-amd64.sha256 -o subosity-installer.sha256
sha256sum -c subosity-installer.sha256

# Make executable and install
chmod +x subosity-installer
./subosity-installer setup --env prod --domain mysubosity.com --email admin@example.com
```

### Container-Based Installation

```bash
docker run --rm \
  -v /opt/subosity:/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  subosity/installer:latest setup \
    --env prod \
    --domain mysubosity.com \
    --email admin@example.com
```

### One-Line Installation (Advanced Users)

```bash
# Direct execution (downloads, verifies, and runs)
curl -fsSL https://install.subosity.com | bash -s -- --env prod --domain mysubosity.com --email admin@example.com
```

## 🎯 What It Does

- **🐳 Docker Management**: Installs and configures Docker CE + Docker Compose v2
- **🗄️ Supabase Setup**: Complete self-hosted Supabase platform (database, auth, storage, edge functions)
- **⚛️ Frontend Deployment**: Builds and deploys the React frontend in production mode
- **🔒 SSL Configuration**: Automatic HTTPS with Let's Encrypt or self-signed certificates
- **🛠️ Service Management**: Creates systemd services for automatic startup and management
- **📦 Backup System**: Automated database backups with configurable retention
- **🔄 Update Management**: Zero-downtime updates with automatic rollback on failure
- **📊 Health Monitoring**: Comprehensive health checks and status reporting

## 📋 System Requirements

**Minimum:**
- Ubuntu 20.04+ or Debian 11+ (x64/ARM64)
- 2GB RAM, 10GB disk space
- Ports 80, 443, 5432, 8000, 3000 available
- Internet connectivity

**Recommended for Production:**
- 4+ CPU cores, 8GB+ RAM, 100GB+ SSD
- Static IP address and configured domain

## 📚 Documentation

| Document | Purpose | Audience |
|----------|---------|----------|
| **[📋 PRD.md](docs/PRD.md)** | Complete product requirements and specifications | Product managers, stakeholders |
| **[🏗️ ARCHITECTURE.md](docs/ARCHITECTURE.md)** | System architecture and design patterns | Developers, architects |
| **[📏 STYLE_GUIDE.md](docs/STYLE_GUIDE.md)** | Coding standards and conventions | Developers, contributors |
| **[🛡️ SECURITY.md](docs/SECURITY.md)** | Security guidelines and threat model | Security engineers, DevOps |
| **[🔧 TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** | Common issues and solutions | Users, support teams |

## 🛠️ Available Commands

### Installation & Setup
```bash
# Fresh installation
subosity-installer setup --env prod --domain example.com

# Development installation
subosity-installer setup --env dev --domain localhost.local --ssl-provider self-signed
```

### Management Operations
```bash
# Check status
subosity-installer status

# Update to latest version
subosity-installer update

# Create backup
subosity-installer backup --retention 30d

# Restore from backup
subosity-installer restore --backup 2025-06-20T10:30:00Z

# View logs
subosity-installer logs --tail 100
```

## 🔍 Understanding the Installation Process

### What the Installer Actually Does (Not Just Docker Commands)

The installer is **much more comprehensive** than a simple Docker wrapper:

**1. System Validation & Preparation:**
```bash
# ❌ You can't just run: docker run setup
# ✅ The installer first does:
- OS detection and compatibility checking
- Resource validation (RAM, disk, ports)
- Dependency installation (Docker, if missing)
- User permission validation
- Network connectivity testing
```

**2. Complex System Integration:**
```bash
# The installer handles:
- Package repository setup (varies by distro)
- Docker daemon configuration and security
- User group management (docker group)
- Systemd service creation and management
- SSL certificate generation/management
- Firewall configuration (if requested)
- Log rotation setup
- Backup scheduling via cron
```

**3. Supabase Platform Orchestration:**
```bash
# Not just Docker - also:
- Supabase CLI installation and verification
- Database initialization and migrations
- Authentication provider configuration
- Edge functions setup
- Storage backend configuration
- API gateway configuration
```

**4. Application Deployment & Configuration:**
```bash
# Beyond Docker Compose:
- React frontend building and optimization
- Environment-specific configuration generation
- SSL/TLS certificate integration
- Reverse proxy configuration (Nginx)
- Health monitoring setup
- Backup system initialization
```

### Why You Can't Just Use Docker Commands

**Manual Docker approach would require:**
```bash
# You'd need to manually handle all of this:
1. Install Docker (varies by OS)
2. Configure Docker daemon securely
3. Download and verify Supabase CLI
4. Generate secure passwords and keys
5. Create proper directory structure with permissions
6. Generate SSL certificates
7. Configure reverse proxy with security headers
8. Set up systemd services for auto-start
9. Configure log rotation and retention
10. Set up automated backups
11. Initialize database with proper schema
12. Configure authentication providers
13. Build and optimize React frontend
14. Set up health monitoring
15. Configure proper networking and firewall rules
```

**The installer does ALL of this automatically and safely.**

### What Each Method Actually Does

**Method 1: Direct Binary (Recommended)**
```bash
# Binary validates environment, installs Docker, delegates to container
./subosity-installer setup --env prod --domain example.com
# Internally runs:
# docker run --rm -v /opt/subosity:/app subosity/installer:latest setup ...
```

**Method 2: Container-Based (Direct)**
```bash
# Skip the binary, run container directly (requires Docker pre-installed)
docker run --rm \
  -v /opt/subosity:/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  subosity/installer:latest setup --env prod --domain example.com
```

**Method 3: One-Line Convenience**
```bash
# Convenience wrapper that downloads binary and runs Method 1
curl -fsSL https://install.subosity.com | bash -s -- --env prod --domain example.com
```

### What Gets Installed (Same Result All Methods)

All methods result in the **same comprehensive installation**:
```
/opt/subosity/
├── data/              # Application data and database
├── backups/           # Automated backup storage  
├── logs/              # Centralized logging
├── configs/           # Generated configurations
├── docker/            # Docker Compose and related files
└── certs/             # SSL certificates

/etc/systemd/system/
└── subosity.service   # Auto-start service

System Integration:
- Docker CE installed and configured
- Supabase CLI installed
- User added to docker group
- Firewall configured (optional)
- Log rotation configured
- Backup cron jobs scheduled
- Health monitoring active
```

### Key Insight: Container Does the Heavy Lifting

The **subosity/installer:latest** container image contains all the complex logic:
- ✅ **Supabase platform setup** with database, auth, storage, edge functions
- ✅ **Application deployment** with React frontend building and optimization  
- ✅ **SSL/TLS management** with Let's Encrypt or self-signed certificates
- ✅ **System integration** with service creation and configuration
- ✅ **Operational features** like backups, monitoring, and health checks

The binary is just a **convenience wrapper** that ensures Docker is available and delegates to the container.

## 🏗️ Repository Structure

```
subosity-installer/
├── cmd/                    # CLI entry points and command implementations
├── internal/              # Private application code
│   ├── domain/            # Business logic and entities
│   ├── ports/             # Interface definitions
│   ├── adapters/          # External integrations (Docker, Supabase, etc.)
│   └── services/          # Application services
├── pkg/                   # Public library code
│   ├── config/            # Configuration management
│   ├── logger/            # Logging utilities
│   └── errors/            # Error types and handling
├── templates/             # Embedded configuration templates
├── configs/               # Configuration schemas
├── docs/                  # Documentation (see table above)
├── tests/                 # Test suites and integration tests
└── build/                 # Build scripts and CI/CD configuration
```

## 🚦 Environment Types

| Environment | Use Case | SSL | Logging | Backups | Resource Limits |
|-------------|----------|-----|---------|---------|-----------------|
| **dev** | Local development | Self-signed | DEBUG | Daily | Minimal |
| **staging** | Pre-production testing | Let's Encrypt | INFO | Daily | Production-like |
| **prod** | Production deployment | Let's Encrypt | WARN | Hourly | Optimized |

## 🔄 Development Workflow

### Prerequisites
- Go 1.21+
- Docker for testing
- golangci-lint for code quality

### Building from Source
```bash
# Clone repository
git clone https://github.com/subosity/subosity-installer.git
cd subosity-installer

# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Cross-compile for all platforms
make build-all
```

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests (requires Docker)
go test -tags=integration ./...

# Test coverage
make coverage
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch from `develop`
3. **Follow** the coding standards in [STYLE_GUIDE.md](docs/STYLE_GUIDE.md)
4. **Add tests** for new functionality
5. **Submit** a pull request with clear description

### Code Quality Requirements
- ✅ All tests must pass (unit, integration, security)
- ✅ 85%+ code coverage with critical path coverage ≥95%
- ✅ golangci-lint passes with zero issues
- ✅ Security scan passes (gosec, nancy, govulncheck)
- ✅ Performance benchmarks within acceptable ranges
- ✅ Documentation updated for new features
- ✅ No increase in technical debt (SonarQube quality gate)

## 📊 Exit Codes

| Code | Meaning | Action Required |
|------|---------|-----------------|
| **0** | Success | None |
| **1** | Warning/Recoverable | Review logs, may require manual intervention |
| **2** | Fatal Error | Check system requirements and configuration |
| **3** | Configuration Error | Fix configuration parameters |
| **4** | Permission Error | Run with appropriate privileges (sudo) |
| **5** | Network Error | Check internet connectivity |
| **6** | Resource Error | Free up disk space or memory |

## 🆘 Getting Help

**Common Issues:**
- Check [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for solutions
- Verify system requirements are met
- Ensure required ports are available

**Support Channels:**
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/subosity/subosity-installer/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/subosity/subosity-installer/discussions)
- 📖 **Documentation**: [docs/](docs/) directory

## 📜 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🌟 Features Roadmap

- [ ] **Multi-cloud Support**: AWS, Azure, GCP deployment options
- [ ] **Kubernetes Deployment**: Helm charts and operator support
- [ ] **High Availability**: Multi-node clustering and load balancing
- [ ] **Monitoring Integration**: Prometheus, Grafana, AlertManager
- [ ] **Configuration Management**: Ansible, Terraform integration
- [ ] **Air-gapped Installation**: Offline installation support

---

**Made with ❤️ by the Subosity team**

## 🏛️ Architecture Decision: Container-First Design

This installer uses a **container-first architecture** that separates concerns cleanly:

### Why Container-First?

**The Problem with Monolithic Binaries:**
- Large binary size with embedded dependencies
- Complex cross-platform compatibility issues
- Difficult to update installation logic
- Host environment variability causes issues

**Our Solution: Thin Client + Smart Container**

```
┌─────────────────────────────────────────┐
│     subosity-installer Binary          │
│            (Thin Client)                │
├─────────────────────────────────────────┤
│ • Environment detection & validation   │
│ • Docker installation (if missing)     │
│ • Container image management           │
│ • Volume/mount setup                   │
│ • Argument forwarding                  │
└─────────────────┬───────────────────────┘
                  │ delegates to
┌─────────────────▼───────────────────────┐
│   subosity/installer:latest Container   │
│         (All Installation Logic)        │
├─────────────────────────────────────────┤
│ • Complete Supabase platform setup     │
│ • Application deployment & config      │
│ • SSL/TLS certificate management       │
│ • System service configuration         │
│ • Backup & monitoring setup            │
│ • Error handling & recovery            │
│ • State management & persistence       │
│ • Health checks & validation           │
└─────────────────────────────────────────┘
```

### Division of Responsibilities

**Thin Binary (subosity-installer) Handles:**
```go
// ✅ Host environment preparation
- OS detection and compatibility checking
- Docker installation and configuration  
- Container image pulling and management
- Volume mounting and permission setup
- Argument parsing and forwarding

// ✅ Container orchestration
docker run --rm \
  -v /opt/subosity:/app/data \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /etc:/app/host-etc:ro \
  subosity/installer:latest setup --env prod --domain example.com
```

**Container (subosity/installer:latest) Handles:**
```yaml
# ✅ All complex installation logic (95% of functionality)
Installation_Logic:
  - Supabase CLI installation and setup
  - Database initialization and migrations
  - Frontend building and deployment
  - SSL certificate generation/management
  - Reverse proxy configuration
  - System service creation
  - Backup system initialization
  - Health monitoring setup
  - Configuration template generation
  - Error recovery and rollback
```

### Benefits of This Architecture

**For Users:**
- ✅ **Smallest possible binary download** (~5-10MB vs 50-100MB)
- ✅ **Always up-to-date logic** (container image updated independently)
- ✅ **Consistent behavior** regardless of host OS quirks
- ✅ **Better error isolation** (container failures don't affect host)

**For Developers:**
- ✅ **Simplified testing** (test container in isolation)
- ✅ **Easier updates** (update container image, binary stays same)
- ✅ **Better dependency management** (all deps in container)
- ✅ **Reduced complexity** (thin client is much simpler to maintain)

**For Operations:**
- ✅ **Predictable environment** (installation runs in known container)
- ✅ **Better security** (installation logic isolated in container)
- ✅ **Easier troubleshooting** (container logs are standardized)
- ✅ **Resource management** (container limits and monitoring)
