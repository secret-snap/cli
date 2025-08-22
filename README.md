# Secrets Snapshot CLI

**Secrets Snapshot** — Encrypt a `.env` into a single bundle, share it safely, and unbundle on machines or CI.

- **Free**: local-only encrypt/decrypt/run
- **Paid**: login via license, cloud push/pull, share, simple audit

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

### Free Mode (Local Only)

```bash
# Initialize configuration
secretsnap init

# Encrypt your .env file
secretsnap bundle .env --out secrets.envsnap

# Decrypt to a new .env file
secretsnap unbundle secrets.envsnap --out .env

# Run a command with environment variables
secretsnap run secrets.envsnap -- npm start
```

### Paid Mode (Cloud Features)

```bash
# Login with your license key
secretsnap login --license your-license-key

# Create a project
secretsnap project create "My App"

# Push bundle to cloud
secretsnap bundle .env --push --project "My App"

# Pull latest bundle
secretsnap pull --project "My App" --out .env

# Share with team member
secretsnap share --project "My App" --user alice@example.com --role member

# View audit logs
secretsnap audit --project "My App"
```

## 📋 Commands

### Free Commands

| Command                   | Description                            |
| ------------------------- | -------------------------------------- |
| `init`                    | Initialize local configuration         |
| `bundle <file>`           | Encrypt .env file with passphrase      |
| `unbundle <file>`         | Decrypt bundle to .env file            |
| `run <file> -- <command>` | Run command with environment variables |

### Paid Commands

| Command                 | Description                    |
| ----------------------- | ------------------------------ |
| `login`                 | Login with license key         |
| `project create <name>` | Create new project             |
| `bundle --push`         | Push bundle to cloud           |
| `pull`                  | Pull latest bundle from cloud  |
| `share`                 | Share project with team member |
| `audit`                 | View project audit logs        |

## 🔧 Configuration

Configuration is stored in `~/.secretsnap/config.json`:

```json
{
  "mode": "local",
  "project": "local"
}
```

For cloud mode:

```json
{
  "mode": "cloud",
  "project": "your-project-id"
}
```

## 🔐 Security Model

### Free Mode

- Uses age encryption with passphrase
- All encryption/decryption happens locally
- No data sent to servers

### Paid Mode

- CLI generates 32-byte random data key
- Encrypts .env with age symmetric encryption using data key
- Data key is encrypted with AWS KMS and stored in cloud
- Server never sees plaintext secrets

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
      - run: secretsnap pull --project ${{ secrets.SECRETSNAP_PROJECT_ID }} --out .env
      - run: make test
```

### Environment Variables

Set these secrets in your CI environment:

- `SECRETSNAP_LICENSE`: Your license key
- `SECRETSNAP_PROJECT_ID`: Your project ID

## 💰 Pricing

| Plan        | Price      | Features                         |
| ----------- | ---------- | -------------------------------- |
| **Free**    | $0         | Local encryption only            |
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

**"Failed to decrypt"**

- Check that you're using the correct passphrase
- Ensure the bundle file wasn't corrupted

**"Not logged in"**

- Run `secretsnap login --license your-key`
- Check that your license is valid

**"Access denied to project"**

- Verify you're a member of the project
- Check that the project ID is correct

**"Failed to connect to API"**

- Ensure the API server is running
- Check your network connection
- Verify the API URL in your configuration

### Getting Help

- 🐛 [Report Issues](https://github.com/secret-snap/cli/issues)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

---

**Made with ❤️ by the Secrets Snapshot team**
