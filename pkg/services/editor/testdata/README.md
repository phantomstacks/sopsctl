# Editor Service Test Data

This directory contains test fixtures for integration testing of the editor service with SOPS/Age encryption.

## Files

### age.key
The Age private key used for encrypting and decrypting test secrets.

**Format**: Age secret key format
**Public Key**: `age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d`
**Private Key**: `AGE-SECRET-KEY-13ZLWP4WFHQ6VHC2J5YYEUCFKGLZTD3SXQQPEGK3WU2M8FKYC238S7ZKNSV`

⚠️ **Note**: This key is for testing purposes only and should never be used for production secrets.

### .sops.yaml
SOPS configuration file that defines encryption rules for YAML files.

**Configuration**:
- Encrypts all `.yaml` files
- Uses regex pattern `^(data|stringData)$` to encrypt only Kubernetes secret data fields
- Uses the Age public key for encryption

### dec.yaml
A decrypted Kubernetes Secret manifest containing plain text values.

**Content**:
- API Version: v1
- Kind: Secret
- Name: big-3-frontend-development
- Namespace: big
- Data: Contains a plain text secret value "SecretThings"

This file represents what a user would see and edit when working with secrets.

### enc.yaml
The same Kubernetes Secret as `dec.yaml` but encrypted using SOPS with Age encryption.

**Encryption Details**:
- Encrypted using Age recipient: `age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d`
- Only the `data` field is encrypted (as per `.sops.yaml` configuration)
- Contains SOPS metadata including:
  - Age recipient information
  - Encrypted data key
  - MAC for integrity verification
  - Last modified timestamp

## Usage in Tests

These files are used in integration tests to verify the complete workflow:

1. **Decryption**: Read `enc.yaml` and decrypt it using the `age.key` private key
2. **Editing**: Simulate user editing of the decrypted content
3. **Re-encryption**: Encrypt the edited content back using the Age public key
4. **Verification**: Decrypt again to verify the changes were preserved correctly

### Example Test Flow

```go
// 1. Decrypt encrypted secret
encryptionSvc := encryption.NewSopsAgeDecryptStrategy()
decrypted, err := encryptionSvc.Decrypt("testdata/enc.yaml", testAgeKey)

// 2. Edit the decrypted content (using editor service)
editor := NewEditor("vim")
modified, cleanup, err := editor.EditTempFile("secret-", ".yaml", decrypted)
defer cleanup()

// 3. Re-encrypt the modified content
reencrypted, err := encryptionSvc.EncryptData(modified, testAgePublicKey)

// 4. Verify by decrypting again
final, err := encryptionSvc.DecryptData(reencrypted, testAgeKey)
// final should contain the edited values
```

## Generating New Test Files

If you need to regenerate these test files:

### 1. Generate a new Age key pair
```bash
age-keygen -o age.key
```

### 2. Create a plain secret (dec.yaml)
```yaml
apiVersion: v1
data:
  DATA: SecretThings
kind: Secret
metadata:
  name: big-3-frontend-development
  namespace: big
type: Opaque
```

### 3. Create .sops.yaml configuration
```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    age: <your-age-public-key>
```

### 4. Encrypt the secret
```bash
export SOPS_AGE_KEY_FILE=./age.key
sops -e dec.yaml > enc.yaml
```

### 5. Verify encryption/decryption
```bash
sops -d enc.yaml
```

## Security Considerations

- ⚠️ **Never commit real secrets or production keys to version control**
- These test keys are intentionally weak and for testing only
- In production, use proper key management solutions (Azure Key Vault, AWS KMS, etc.)
- Age keys should be stored securely and never exposed in logs or error messages

## Related Tests

- `external_editor_test.go`: Unit tests for basic editor functionality
- `integration_test.go`: Integration tests using these test data files
- `../../encryption/sops_age_strategy_test.go`: Tests for SOPS encryption/decryption

## References

- [SOPS Documentation](https://github.com/getsops/sops)
- [Age Encryption Tool](https://github.com/FiloSottile/age)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)

