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

### 3. Create and Manage Encrypted Secrets

```bash
# Create an encrypted secret from literal values
pflux secret create my-secret \
  --from-literal=username=admin \
  --from-literal=password=secret123 \
  --cluster=production > secret.yaml

# Create an encrypted secret from an environment file
pflux secret create db-credentials \
  --from-env-file=.env \
  --cluster=production > db-secret.yaml

# Edit an encrypted secret file
pflux secret edit secrets.yaml --cluster=production

# Decrypt and view a secret (outputs to stdout)
pflux secret decrypt secrets.yaml --cluster=production
```

## 📖 Command Reference

### Global Flags

All commands support the following global flag:
- `--cluster, -c`: Specify the Kubernetes cluster context to use for key operations

### Key Management Commands

#### `pflux key add`

Add encryption keys from a Kubernetes cluster to local storage. Retrieves age keys from a Kubernetes secret and stores them locally in `~/.pFlux/` for use with SOPS encryption/decryption.

```bash
pflux key add [flags]
```

**Flags:**
- `--from-current-context`: Use the current kubectl context instead of specifying `--cluster`
- `--namespace, -n`: The namespace where the secret is located (default: `flux-system`)
- `--secret, -s`: The name of the secret containing the SOPS key (default: `sops-age`)
- `--key, -k`: The key within the secret that holds the age key (default: `age.agekey`)

**Note:** Either `--from-current-context` or `--cluster` must be specified.

**Examples:**

```bash
# Add keys from current context
pflux key add --from-current-context

# Add keys from specific cluster
pflux key add --cluster=production

# Add keys from specific cluster and custom secret location
pflux key add --cluster=staging --namespace=encryption --secret=my-age-key

# Add keys with custom key name
pflux key add --cluster=production --key=private.key
```

#### `pflux key list`

List all age keys stored locally in `~/.pFlux/`. Shows the cluster name and public key for each stored key.

```bash
pflux key list [flags]
```

**Flags:**
- `--show-sensitive`: Show private keys in the list output (use with caution)

**Examples:**

```bash
# List all stored keys
pflux key list

# List keys including private keys
pflux key list --show-sensitive
```

#### `pflux key remove`

Remove age keys from local storage. Can remove keys for a specific cluster or all stored keys.

```bash
pflux key remove [cluster-name] [flags]
```

**Flags:**
- `--all`: Remove all SOPS keys from local storage

**Examples:**

```bash
# Remove keys for specific cluster
pflux key remove production

# Remove all stored keys
pflux key remove --all
```

#### `pflux key storage-mode`

View and manage SOPS key storage modes. Controls how and where encryption keys are stored.

```bash
pflux key storage-mode [flags]
```

**Flags:**
- `--set-storage-mode, -s`: Set storage mode for SOPS keys (options: `local`, `cluster`)

**Examples:**

```bash
# View current storage mode
pflux key storage-mode

# Set storage mode to local
pflux key storage-mode --set-storage-mode=local

# Set storage mode to cluster
pflux key storage-mode --set-storage-mode=cluster
```

### Secret Management Commands

#### `pflux secret create`

Create an encrypted Kubernetes secret from local files, directories, or literal values. This command mimics `kubectl create secret generic` but automatically encrypts the output using SOPS and age encryption. The secret is output as encrypted YAML that can be committed to version control.

```bash
pflux secret create NAME [flags]
```

**Flags:**
- `--from-file strings`: Create secret from files or directories. Can specify `key=filepath` to set custom keys
- `--from-literal stringArray`: Create secret from literal key=value pairs (e.g., `username=admin`)
- `--from-env-file strings`: Create secret from environment files containing `KEY=value` lines
- `--type string`: The type of secret to create (default: `Opaque`)
- `--namespace, -n string`: Namespace for the secret (default: `default`)
- `--append-hash`: Append a hash of the secret data to its name

**Examples:**

```bash
# Create secret from literal values
pflux secret create my-secret --from-literal=username=admin --from-literal=password=secret123

# Create secret from files
pflux secret create my-secret --from-file=ssh-privatekey=~/.ssh/id_rsa --from-file=ssh-publickey=~/.ssh/id_rsa.pub

# Create secret from a directory (uses filenames as keys)
pflux secret create my-secret --from-file=./config/

# Create secret from environment file
pflux secret create my-secret --from-env-file=.env

# Create secret with custom namespace and type
pflux secret create my-secret --from-literal=token=abc123 --namespace=production --type=kubernetes.io/service-account-token

# Create secret with hash appended to name
pflux secret create my-secret --from-literal=data=value --append-hash
```

**Notes:**
- The `--from-env-file` flag cannot be combined with `--from-file` or `--from-literal`
- Output is encrypted SOPS YAML that can be saved to a file: `pflux secret create my-secret --from-literal=key=value > secret.yaml`
- Secret data is base64-encoded and then encrypted with SOPS

#### `pflux secret edit`

Edit encrypted secret files using your default editor with automatic encryption/decryption. Provides a secure workflow where the file is temporarily decrypted, opened in an editor, then re-encrypted when you save.

```bash
pflux secret edit [file] [flags]
```

**Flags:**
- `--decode, -d`: Edit a decoded secret property without manually encrypting the entire file
- `--k, -k string`: Specify the key within the secret to decode and edit (used with `--decode`)
- `--env, -e`: Specify environment variable that holds the decoded value

**Examples:**

```bash
# Edit entire encrypted file
pflux secret edit secrets.yaml --cluster=production

# Edit a specific decoded property
pflux secret edit secrets.yaml --cluster=production --decode --k=database-password

# Edit single property (auto-detected if only one exists)
pflux secret edit secrets.yaml --cluster=production --decode

# Edit with environment variable
pflux secret edit secrets.yaml --cluster=production --env
```

**Workflow:**
1. Decrypts the file using the cluster's private age key
2. Opens the decrypted content in your system's default editor
3. After you save and close the editor, re-encrypts the content with the cluster's public key
4. Atomically writes the encrypted content back to the original file

**Editor Selection:**
The command respects the following environment variables (in order of precedence):
1. `VISUAL`
2. `EDITOR`
3. Default: `vi` (Unix) or `notepad` (Windows)

#### `pflux secret decrypt`

Decrypt a SOPS-encrypted file and output the plaintext result to stdout. Useful for viewing encrypted files, piping to other commands, or extracting specific values.

```bash
pflux secret decrypt <file> [flags]
```

**Examples:**

```bash
# Decrypt and view file contents
pflux secret decrypt secrets.yaml --cluster=production

# Decrypt and pipe to another command
pflux secret decrypt secrets.yaml --cluster=production | grep password

# Decrypt and save to a file
pflux secret decrypt secrets.yaml --cluster=production > decrypted.yaml

# Use with yq to extract specific values
pflux secret decrypt secrets.yaml --cluster=production | yq .data.password
```

**Security Note:** Be careful when decrypting files as the plaintext output may be sensitive. Avoid saving decrypted content to disk unnecessarily.

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
