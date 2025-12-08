# Editor Package

A simple Go package for launching editors from CLI applications, inspired by kubectl's editor implementation. Perfect for creating commands like `sopsctl edit secret` that need to pause execution while the user edits a file.

## Features

- **Environment-aware**: Automatically detects editor from `EDITOR` and `VISUAL` environment variables
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Flexible**: Support for simple commands, commands with arguments, and complex shell commands
- **Temporary file editing**: Create temporary files for editing sensitive content
- **Stream editing**: Edit content from any `io.Reader`
- **Automatic cleanup**: Built-in cleanup for temporary files

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-org/sopsctl/pkg/editor"
)

func main() {
    // Create default editor (respects EDITOR/VISUAL env vars)
    ed := editor.NewDefaultEditor()
    
    // Edit existing file - program pauses until editor closes
    content, err := ed.EditFile("./secret.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("File edited! New content: %d bytes\n", len(content))
}
```

### SOPS Secret Editing Example

```go
func editSOPSSecret(encryptedFile string) error {
    // 1. Decrypt SOPS file (using your SOPS integration)
    decryptedContent, err := sops.Decrypt(encryptedFile)
    if err != nil {
        return err
    }
    
    // 2. Edit decrypted content in temporary file
    ed := editor.NewDefaultEditor()
    modifiedContent, cleanup, err := ed.EditTempFile("secret-", ".yaml", decryptedContent)
    if err != nil {
        return err
    }
    defer cleanup() // Clean up temp file
    
    // 3. Re-encrypt and save
    return sops.EncryptAndSave(encryptedFile, modifiedContent)
}
```

### Custom Editor

```go
// Use specific editor
ed := editor.NewEditor("code", "--wait")           // VS Code
ed := editor.NewEditor("nano")                     // nano
ed := editor.NewEditor("emacs", "-nw")             // emacs (terminal mode)
```

## API Reference

### Editor Creation

- `NewDefaultEditor()` - Creates editor from environment variables or defaults
- `NewEditor(command, args...)` - Creates editor with specific command and arguments

### Editing Methods

- `EditFile(filename)` - Edit existing file, returns modified content
- `EditTempFile(prefix, suffix, content)` - Edit content in temporary file
- `EditStream(prefix, suffix, reader)` - Edit content from io.Reader

### Environment Variables

The package respects standard Unix editor environment variables:

1. `VISUAL` - Takes precedence if set
2. `EDITOR` - Used if VISUAL is not set
3. Default fallback: `vi` on Unix systems, `notepad` on Windows

### Supported Editor Types

- **Simple commands**: `nano`, `vi`, `emacs`
- **Commands with arguments**: `code --wait`, `emacs -nw`
- **Complex shell commands**: `emacs -nw --eval "(setq backup-inhibited t)"`

## Integration with sopsctl

For your `sopsctl edit secret` command:

```go
func editSecretCommand(cmd *cobra.Command, args []string) error {
    if len(args) != 1 {
        return fmt.Errorf("usage: sopsctl edit secret <file>")
    }
    
    filename := args[0]
    
    // Create editor
    ed := editor.NewDefaultEditor()
    
    // Option 1: Edit file directly (if already decrypted)
    if !isSOPSEncrypted(filename) {
        _, err := ed.EditFile(filename)
        return err
    }
    
    // Option 2: SOPS workflow
    return editSOPSSecret(filename, ed)
}

func editSOPSSecret(filename string, ed *editor.Editor) error {
    // Decrypt
    decrypted, err := sopsDecrypt(filename)
    if err != nil {
        return err
    }
    
    // Edit in temp file
    modified, cleanup, err := ed.EditTempFile("sopsctl-secret-", ".yaml", decrypted)
    if err != nil {
        return err
    }
    defer cleanup()
    
    // Re-encrypt and save
    return sopsEncryptToFile(filename, modified)
}
```

## Error Handling

The package provides clear error messages for common issues:

- Editor not found
- File doesn't exist
- Permission issues
- Editor launch failures

## Testing

Run the tests:

```bash
go test ./pkg/editor
```

The tests use `cat` as a no-op editor for reliable testing without requiring interactive input.

## Platform Support

- **Linux/macOS**: Full support with shell command handling
- **Windows**: Basic support with cmd.exe fallback

## Thread Safety

The editor operations are not thread-safe by design, as they involve terminal interaction. Use appropriate synchronization if needed in concurrent applications.
