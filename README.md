# Secrets Snapshot CLI

**Secrets Snapshot** — Encrypt `.env` files into secure bundles with zero-friction workflows for teams and developers.

- **Free**: Local encryption with cached keys (zero prompts)
- **Paid**: Cloud storage, team sharing, audit logs

## 🚀 Quick Start

### Installation

```bash
# Install via curl (recommended)
curl -sSL https://get.secretsnap.sh | bash

# Or build from source
git clone https://github.com/secretsnap/cli.git
cd cli
make build
sudo make install
```

### Zero-Prompt Workflow (Free)

```bash
# Initialize project (creates cached key automatically)
secretsnap init

# Encrypt your .env file (uses cached key, no prompts)
secretsnap bundle .env

# Decrypt to .env file (uses cached key, no prompts)
secretsnap unbundle secrets.envsnap

# Run commands with environment variables
secretsnap run secrets.envsnap -- npm start
```

### Passphrase Mode (Extra Security)

```bash
# Encrypt with passphrase (prompts once)
secretsnap bundle .env --pass-mode --out secrets.envsnap

# Decrypt with passphrase (prompts once)
secretsnap unbundle secrets.envsnap --pass-mode --out .env

# Run with passphrase (prompts once)
secretsnap run secrets.envsnap --pass-mode -- npm start
```

### Key Sharing (Free)

```bash
# Export your project key for teammates
secretsnap key export

# Teammate imports the key and can use zero-prompt workflow
# (Key sharing happens outside of secretsnap)
```

### Cloud Features (Paid)

```bash
# Login with your license key
secretsnap login --license your-license-key

# Create a project
secretsnap project create "My App"

# Push bundle to cloud
secretsnap bundle .env --push

# Pull latest bundle
secretsnap pull --out .env

# Share with team member
secretsnap share --user alice@example.com --role read

# View audit logs
secretsnap audit --limit 50
```

## 📋 Commands

### Core Commands

| Command                   | Description                               |
| ------------------------- | ----------------------------------------- |
| `init`                    | Initialize project with cached key        |
| `bundle <file>`           | Encrypt .env file (local mode by default) |
| `unbundle <file>`         | Decrypt bundle to .env file               |
| `run <file> -- <command>` | Run command with environment variables    |
| `key export`              | Export project key for team sharing       |

### Security Modes

| Flag              | Description                        |
| ----------------- | ---------------------------------- |
| `--pass-mode`     | Use passphrase (prompts for input) |
| `--pass <phrase>` | Use specific passphrase            |
| `--pass-file <f>` | Read passphrase from file          |

### Cloud Commands (Paid)

| Command                 | Description                    |
| ----------------------- | ------------------------------ |
| `login --license <key>` | Login with license key         |
| `project create <name>` | Create new project             |
| `bundle --push`         | Push bundle to cloud           |
| `pull [--out .env]`     | Pull latest bundle from cloud  |
| `share --user <email>`  | Share project with team member |
| `audit [--limit 50]`    | View project audit logs        |

## 🔧 Configuration

### Project Configuration (`.secretsnap.json`)

```json
{
  "project_name": "my-app",
  "project_id": "local",
  "mode": "local",
  "bundle_path": "secrets.envsnap"
}
```

### Global Key Cache (`~/.secretsnap/keys.json`)

```json
{
  "projects": {
    "my-app": {
      "key_id": "S+OI6LVBJwCuUITHN89zIQ==",
      "alg": "age-symmetric-v1",
      "key_b64": "ZBnqDGfuMNKFSU9Cjm+lxVdw5pujoC40ZpC3fv8bXy0=",
      "created_at": "2025-08-22T16:04:12.695702-06:00"
    }
  }
}
```

## 🔐 Security Model

### Local Mode (Default)

- **Zero prompts**: Uses cached 32-byte project key
- **Age encryption**: Symmetric encryption with project key
- **Secure storage**: Keys stored with 0600 permissions
- **No cloud dependency**: All operations local

### Passphrase Mode

- **Explicit security**: Prompts for passphrase each time
- **Age encryption**: Uses scrypt-derived key from passphrase
- **No key cache**: Never touches cached project keys
- **CI friendly**: Supports `--pass-file` for automation

### Cloud Mode (Paid)

- **Fresh data keys**: 32-byte key per bundle
- **KMS encryption**: Data keys wrapped with AWS KMS
- **Zero-knowledge**: Server never sees plaintext secrets
- **Team sharing**: Built-in access control and audit

## 🏗️ CI/CD Integration

### GitHub Actions

```yaml
name: CI
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: curl -sSL https://get.secretsnap.sh | bash

      # Option 1: Local mode with shared key
      - run: echo "${{ secrets.PROJECT_KEY }}" > project.key
      - run: secretsnap unbundle secrets.envsnap --out .env

      # Option 2: Cloud mode
      - run: secretsnap login --license ${{ secrets.SECRETSNAP_LICENSE }}
      - run: secretsnap pull --out .env

      - run: make test
```

### Environment Variables

For local mode:

- `PROJECT_KEY`: Base64-encoded project key (from `secretsnap key export`)

For cloud mode:

- `SECRETSNAP_LICENSE`: Your license key
- `SECRETSNAP_PROJECT_ID`: Your project ID

## 💰 Pricing

| Plan        | Price      | Features                         |
| ----------- | ---------- | -------------------------------- |
| **Free**    | $0         | Local encryption                 |
| **Indie**   | $9/month   | Cloud storage, team sharing      |
| **Startup** | $49/month  | Advanced audit, priority support |
| **Team**    | $199/month | SSO, advanced security features  |

## 🛠️ Development

### Building from Source

```bash
git clone https://github.com/secretsnap/cli.git
cd cli

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Install
make install
```

### Local Development

```bash
# Build and run in development mode
make dev

# Format code
make fmt

# Run linter
make lint
```

## 🔍 Troubleshooting

### Common Issues

**"No local project key found"**

```bash
# Option 1: Get key from teammate
secretsnap key export --project my-app

# Option 2: Use passphrase mode
secretsnap unbundle secrets.envsnap --pass-mode

# Option 3: Use cloud mode (paid)
secretsnap login --license your-key
secretsnap pull --out .env
```

**"Refusing to overwrite .env"**

```bash
# Use --force flag
secretsnap unbundle secrets.envsnap --force
```

**"Cloud sync is Pro"**

```bash
# Login with license key
secretsnap login --license your-license-key

# Or use local mode (no --push flag)
secretsnap bundle .env
```

**"Failed to decrypt"**

- Check that you're using the correct mode (local vs passphrase)
- Verify the bundle file wasn't corrupted
- For passphrase mode: ensure correct passphrase

### Getting Help

- 🐛 [Report Issues](https://github.com/secret-snap/cli/issues)
- 📖 [Documentation](https://docs.secretsnap.sh)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

---

**Made with ❤️ by the Secrets Snapshot team**
