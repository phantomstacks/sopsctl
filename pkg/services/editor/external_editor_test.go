package editor

import (
	"bytes"
	"os"
	"phantom-flux/pkg/domain"
	"strings"
	"testing"
)

func editorForTest() *editor {
	args, shell := getEditorFromEnv()
	return &editor{
		Args:  args,
		Shell: shell,
	}
}

func newTestEditor(command string, args ...string) *editor {
	allArgs := append([]string{command}, args...)
	return &editor{
		Args:  allArgs,
		Shell: false,
	}
}

func TestNewDefaultEditor(t *testing.T) {
	// Save original env vars
	originalSopsctlEditor := os.Getenv(domain.EditorEnvName)

	// Restore env vars after test
	defer func() {
		if originalSopsctlEditor == "" {
			os.Unsetenv(domain.EditorEnvName)
		} else {
			os.Setenv(domain.EditorEnvName, originalSopsctlEditor)
		}
	}()

	// Test with SOPSCTL_EDITOR set to simple command
	os.Setenv(domain.EditorEnvName, "nano")

	editorService := editorForTest()
	if len(editorService.Args) != 1 || editorService.Args[0] != "nano" {
		t.Errorf("Expected editorService args [nano], got %v", editorService.Args)
	}
	if editorService.Shell {
		t.Errorf("Expected Shell to be false, got true")
	}

	// Test with command with spaces (should split on spaces)
	os.Setenv(domain.EditorEnvName, "code --wait")

	editorService = editorForTest()
	if len(editorService.Args) != 2 || editorService.Args[0] != "code" || editorService.Args[1] != "--wait" {
		t.Errorf("Expected editorService args [code, --wait], got %v", editorService.Args)
	}

	// Test with complex editorService command that requires shell (contains quotes)
	os.Setenv(domain.EditorEnvName, `emacs -nw --eval "(setq backup-inhibited t)"`)

	editorService = editorForTest()
	if !editorService.Shell {
		t.Errorf("Expected Shell to be true for complex command")
	}
}

func TestNewEditor(t *testing.T) {
	newEditor := newTestEditor("nano")
	if len(newEditor.Args) != 1 || newEditor.Args[0] != "nano" {
		t.Errorf("Expected args [nano], got %v", newEditor.Args)
	}
	if newEditor.Shell {
		t.Errorf("Expected Shell to be false")
	}

	newEditor = newTestEditor("code", "--wait", "--new-window")
	expected := []string{"code", "--wait", "--new-window"}
	if len(newEditor.Args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(newEditor.Args))
	}
	for i, arg := range expected {
		if i >= len(newEditor.Args) || newEditor.Args[i] != arg {
			t.Errorf("Expected args %v, got %v", expected, newEditor.Args)
			break
		}
	}
}

func TestBuildArgs(t *testing.T) {
	testCases := []struct {
		name     string
		editor   *editor
		path     string
		expected []string
	}{
		{
			name:     "simple command",
			editor:   newTestEditor("nano"),
			path:     "/tmp/test.txt",
			expected: []string{"nano", "/tmp/test.txt"},
		},
		{
			name:     "command with args",
			editor:   newTestEditor("code", "--wait"),
			path:     "/tmp/test.txt",
			expected: []string{"code", "--wait", "/tmp/test.txt"},
		},
		{
			name: "shell command",
			editor: &editor{
				Args:  []string{"/bin/bash", "-c", "emacs -nw"},
				Shell: true,
			},
			path:     "/tmp/test.txt",
			expected: []string{"/bin/bash", "-c", `emacs -nw "/tmp/test.txt"`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.editor.buildArgs(tc.path)
			if len(args) != len(tc.expected) {
				t.Errorf("Expected %d args, got %d", len(tc.expected), len(args))
				return
			}
			for i, expected := range tc.expected {
				if args[i] != expected {
					t.Errorf("Expected args[%d] = %q, got %q", i, expected, args[i])
				}
			}
		})
	}
}

func TestEditTempFile(t *testing.T) {
	// Use cat as a "no-op editor" that just reads and outputs the file
	editor := NewEditor("cat")

	originalContent := []byte("Hello, World!\nThis is a test.")

	modifiedContent, cleanup, err := editor.EditTempFile("test-", ".txt", originalContent)
	if err != nil {
		t.Fatalf("EditTempFile failed: %v", err)
	}
	defer cleanup()

	if !bytes.Equal(originalContent, modifiedContent) {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", originalContent, modifiedContent)
	}
}

func TestEditStream(t *testing.T) {
	// Use cat as a "no-op editor"
	editor := NewEditor("cat")

	originalContent := "Hello from stream!\nLine 2"
	reader := strings.NewReader(originalContent)

	modifiedContent, err := editor.EditStream("stream-", ".txt", reader)
	if err != nil {
		t.Fatalf("EditStream failed: %v", err)
	}

	if string(modifiedContent) != originalContent {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", originalContent, string(modifiedContent))
	}
}
