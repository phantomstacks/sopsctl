package editor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sopsctl/pkg/domain"
	"strings"
)

const (
	defaultEditor = "nano"
	windowsEditor = "notepad"
)

// editor represents an editor configuration
type editor struct {
	Args  []string
	Shell bool
}

func (e *editor) EditFileWithPostEditCallback(filename string, postEditCallback func([]byte) ([]byte, error)) ([]byte, error) {
	editedContent, err := e.EditFile(filename)
	if err != nil {
		return nil, err
	}
	return postEditCallback(editedContent)
}

// NewDefaultEditor creates an editor using environment variables EDITOR, VISUAL, or defaults
func NewDefaultEditor() domain.UserEditorService {
	args, shell := getEditorFromEnv()
	return &editor{
		Args:  args,
		Shell: shell,
	}
}

// NewEditor creates an editor with specific command and arguments
func NewEditor(command string, args ...string) domain.UserEditorService {
	allArgs := append([]string{command}, args...)
	return &editor{
		Args:  allArgs,
		Shell: false,
	}
}

func getEditorFromEnv() ([]string, bool) {
	for _, env := range []string{"SOPSCTL_EDITOR"} {
		if editor := os.Getenv(env); editor != "" {
			if !strings.Contains(editor, " ") {
				return []string{editor}, false
			}
			if strings.ContainsAny(editor, "\"'\\") {
				shell := "/bin/bash"
				flag := "-c"
				if runtime.GOOS == "windows" {
					shell = "cmd"
					flag = "/C"
				}
				return []string{shell, flag, editor}, true
			}
			return strings.Split(editor, " "), false
		}
	}

	// Default editor
	defaultCmd := defaultEditor
	if runtime.GOOS == "windows" {
		defaultCmd = windowsEditor
	}
	return []string{defaultCmd}, false
}

// EditFile opens an existing file for editing and returns the modified content
func (e *editor) EditFile(filename string) ([]byte, error) {
	if len(e.Args) == 0 {
		return nil, fmt.Errorf("no editor defined")
	}

	// Get absolute path
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", abs)
	}

	// Launch editor
	if err := e.launch(abs); err != nil {
		return nil, err
	}

	// Read the modified file
	content, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to read file after editing: %w", err)
	}

	return content, nil
}

// EditTempFile creates a temporary file with the provided content, opens it for editing,
// and returns the modified content along with cleanup function
func (e *editor) EditTempFile(prefix, suffix string, content []byte) ([]byte, func(), error) {
	if len(e.Args) == 0 {
		return nil, nil, fmt.Errorf("no editor defined")
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()

	// Write initial content
	if _, err := tmpFile.Write(content); err != nil {
		err := tmpFile.Close()
		if err != nil {
			return nil, nil, err
		}
		err = os.Remove(tmpPath)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to write to temp file: %w", err)
	}
	err = tmpFile.Close()
	if err != nil {
		return nil, nil, err
	}

	// Create cleanup function
	cleanup := func() {
		err := os.Remove(tmpPath)
		if err != nil {
			panic(err)
		}
	}

	// Launch editor
	if err := e.launch(tmpPath); err != nil {
		cleanup()
		return nil, nil, err
	}

	// Read the modified content
	modifiedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to read temp file after editing: %w", err)
	}

	return modifiedContent, cleanup, nil
}

// EditStream reads from a stream, opens it for editing, and returns the modified content
func (e *editor) EditStream(prefix, suffix string, r io.Reader) ([]byte, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}

	modifiedContent, cleanup, err := e.EditTempFile(prefix, suffix, content)
	if cleanup != nil {
		defer cleanup()
	}
	return modifiedContent, err
}

// launch starts the editor with the given file
func (e *editor) launch(path string) error {
	args := e.buildArgs(path)
	cmd := exec.Command(args[0], args[1:]...)

	// Connect to current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildArgs constructs the command line arguments for the editor
func (e *editor) buildArgs(path string) []string {
	args := make([]string, len(e.Args))
	copy(args, e.Args)

	if e.Shell {
		// For shell execution, append the path to the last argument
		last := args[len(args)-1]
		args[len(args)-1] = fmt.Sprintf("%s %q", last, path)
	} else {
		// For direct execution, add path as separate argument
		args = append(args, path)
	}

	return args
}
