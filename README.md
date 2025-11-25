# Phantom Flux (pflux)

**Secure Configuration Management for Kubernetes with SOPS and Age Encryption**

Phantom Flux is a command-line tool that streamlines the management of encrypted Kubernetes secrets in GitOps workflows. It integrates SOPS (Secrets OPerationS), age encryption, and Kubernetes to provide a seamless experience for editing and managing encrypted configuration files alongside your deployment manifests.

## 🚀 Features

- **🔐 Age Encryption Integration**: Uses age keys for fast, modern encryption
- **📝 SOPS Integration**: Full compatibility with SOPS for encrypted YAML/JSON files
- **⚡ Kubernetes Native**: Seamlessly integrates with your existing Kubernetes clusters
- **✏️ Interactive Editing**: Edit encrypted files directly with your preferred editor
- **🔑 Key Management**: Store and manage encryption keys in `~/.pFlux`
- **🎯 GitOps Ready**: Keep encrypted secrets in the same repository as your deployment files
- **🛡️ Secure Workflow**: Files are never stored unencrypted on disk during editing

## 📋 Prerequisites

Before using Phantom Flux, ensure you have:

- **Go 1.25.0+** (for building from source)
- **Kubernetes cluster access** with `kubectl` configured
- **SOPS** installed and available in your PATH
- **Age** keys configured in your cluster (typically in `flux-system` namespace)

## 📦 Installation

### Building from Source

```bash
git clone <repository-url>
cd phantom-flux
go build -o pflux .
sudo mv pflux /usr/local/bin/
```

### Verify Installation

```bash
pflux --help
```

## 🚀 Quick Start

### 1. Add Age Keys from Your Kubernetes Cluster

First, you need to extract the age keys from your Kubernetes cluster and store them locally:

```bash
# Add keys from your current kubectl context
pflux key add --from-current-context

# Or specify a specific cluster context
pflux key add --cluster=production

# List stored keys
pflux key list
```

### 2. Create a SOPS Configuration

Create a `.sops.yaml` file in your repository to define encryption rules:

```yaml
# .sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    age: age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d
```

### 3. Edit Encrypted Secrets

```bash
# Edit an encrypted secret file
pflux secret edit secrets.yaml --cluster=production

# Decrypt and view a secret (outputs to stdout)
pflux secret decrypt secrets.yaml --cluster=production
```

## 📖 Command Reference

### Global Flags

- `--cluster, -c`: Specify the Kubernetes cluster context to use for key operations

### Key Management Commands

#### `pflux key add`

Add encryption keys from a Kubernetes cluster to local storage.

```bash
pflux key add [flags]
```

**Flags:**
- `--from-current-context`: Use the current kubectl context instead of specifying `--cluster`
- `--namespace, -n`: The namespace where the secret is located (default: `flux-system`)
- `--secret, -s`: The name of the secret containing the SOPS key (default: `sops-age`)
- `--key, -k`: The key within the secret that holds the age key (default: `age.agekey`)

**Examples:**

```bash
# Add keys from current context
pflux key add --from-current-context

# Add keys from specific cluster and custom secret location
pflux key add --cluster=staging --namespace=encryption --secret=my-age-key

# Add keys with custom key name
pflux key add --cluster=production --key=private.key
```

#### `pflux key list`

List all age keys stored locally.

```bash
pflux key list
```

#### `pflux key remove`

Remove age keys from local storage.

```bash
pflux key remove [flags]
```

### Secret Management Commands

#### `pflux secret edit`

Edit encrypted secret files using your default editor with automatic encryption/decryption.

```bash
pflux secret edit [file] [flags]
```

**Flags:**
- `--decode, -d`: Edit a decoded secret property without manually encrypting the entire file
- `--k, -k`: Specify the key within the secret to decode and edit (used with `--decode`)
- `--env, -e`: Specify environment variable that holds the decoded value

**Examples:**

```bash
# Edit entire encrypted file
pflux secret edit secrets.yaml --cluster=production

# Edit a specific decoded property
pflux secret edit secrets.yaml --cluster=production --decode --k=database-password

# Edit single property (auto-detected if only one exists)
pflux secret edit secrets.yaml --cluster=production --decode
```

**Workflow:**
1. Decrypts the file using the cluster's private age key
2. Opens the decrypted content in your system's default editor
3. Re-encrypts the content after you save and close the editor
4. Atomically writes the encrypted content back to the original file

#### `pflux secret decrypt`

Decrypt a SOPS-encrypted file and output the result to stdout.

```bash
pflux secret decrypt <file> [flags]
```

**Examples:**

```bash
# Decrypt and view file contents
pflux secret decrypt secrets.yaml --cluster=production

# Decrypt and pipe to another command
pflux secret decrypt secrets.yaml --cluster=production | grep password
```

## ⚙️ Configuration

### Environment Variables

Phantom Flux respects standard environment variables for editor selection:

- `VISUAL`: Primary editor preference
- `EDITOR`: Secondary editor preference
- **Default**: `vi` on Unix systems, `notepad` on Windows

### Key Storage

Age keys are stored securely in `~/.pFlux/` directory with appropriate file permissions.

### SOPS Configuration

Create a `.sops.yaml` file in your project root to configure encryption rules:

```yaml
creation_rules:
  - path_regex: .*secrets.*\.yaml$
    encrypted_regex: ^(data|stringData)$
    age: age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d
  - path_regex: .*config.*\.yaml$
    encrypted_regex: ^(spec\.data)$
    age: age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d
```

## 📝 Common Workflows

### Setting Up a New Environment

1. **Configure age keys in Kubernetes:**
   ```bash
   # Generate age key
   age-keygen -o age.key
   
   # Create Kubernetes secret
   kubectl create secret generic sops-age \
     --from-file=age.agekey=age.key \
     --namespace=flux-system
   ```

2. **Add keys to Phantom Flux:**
   ```bash
   pflux key add --from-current-context
   ```

3. **Create SOPS config:**
   ```bash
   echo "creation_rules:
     - path_regex: .*.yaml
       encrypted_regex: ^(data|stringData)$
       age: $(age-keygen -y age.key)" > .sops.yaml
   ```

### Editing Database Credentials

```bash
# Create/edit encrypted secret file
pflux secret edit database-secret.yaml --cluster=production

# Edit just the password field
pflux secret edit database-secret.yaml --cluster=production --decode --k=password
```

### Rotating Secrets

```bash
# Decrypt current secrets
pflux secret decrypt app-secrets.yaml --cluster=production > temp-secrets.yaml

# Edit the decrypted file
$EDITOR temp-secrets.yaml

# Re-encrypt with SOPS
sops --encrypt --in-place temp-secrets.yaml
mv temp-secrets.yaml app-secrets.yaml

# Clean up
rm -f temp-secrets.yaml
```

## 🔍 Troubleshooting

### Key Not Found Error

```
Error: failed to retrieve key from cluster
```

**Solutions:**
- Verify kubectl context: `kubectl config current-context`
- Check secret exists: `kubectl get secret sops-age -n flux-system`
- Ensure you have proper permissions to read secrets

### Editor Not Opening

```
Error: failed to open editor
```

**Solutions:**
- Set `EDITOR` environment variable: `export EDITOR=nano`
- Ensure your editor is in PATH
- Try with a simple editor like `nano` or `vi`

### SOPS Decryption Failed

```
Error: failed to decrypt file
```

**Solutions:**
- Verify the age key is correctly added: `pflux key list`
- Check SOPS configuration in `.sops.yaml`
- Ensure the file was encrypted with the correct age key

### Permission Denied

```
Error: permission denied when accessing key storage
```

**Solutions:**
- Check permissions on `~/.pFlux/` directory
- Ensure the directory is owned by your user account
- Recreate the directory: `rm -rf ~/.pFlux && pflux key add --from-current-context`

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and enhancement requests.

### Development Setup

```bash
git clone <repository-url>
cd phantom-flux
go mod download
go build -o pflux .
```

### Running Tests

```bash
go test ./...
```

## 📄 License

[License information to be added]

## 🔗 Related Projects

- [SOPS](https://github.com/mozilla/sops) - Secrets OPerationS
- [Age](https://age-encryption.org/) - Simple, modern encryption
- [Flux](https://fluxcd.io/) - GitOps toolkit for Kubernetes

---

**Made with ❤️ for secure GitOps workflows**
