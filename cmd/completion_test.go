package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestGenerateCompletion_Bash tests generating bash completion script
func TestGenerateCompletion_Bash(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("bash")

	output := stdout.String()
	errOutput := stderr.String()

	// Check that output was generated
	if output == "" {
		t.Error("Expected bash completion output, got empty string")
	}

	// Check that no errors were written
	if errOutput != "" {
		t.Errorf("Expected no errors, got: %s", errOutput)
	}

	// Check for bash-specific completion markers
	if !strings.Contains(output, "bash") && !strings.Contains(output, "completion") {
		t.Errorf("Expected bash completion script markers in output")
	}
}

// TestGenerateCompletion_Zsh tests generating zsh completion script
func TestGenerateCompletion_Zsh(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("zsh")

	output := stdout.String()
	errOutput := stderr.String()

	// Check that output was generated
	if output == "" {
		t.Error("Expected zsh completion output, got empty string")
	}

	// Check that no errors were written
	if errOutput != "" {
		t.Errorf("Expected no errors, got: %s", errOutput)
	}

	// Check for zsh-specific completion markers
	if !strings.Contains(output, "#compdef") {
		t.Errorf("Expected zsh completion script to contain #compdef directive")
	}
}

// TestGenerateCompletion_Fish tests generating fish completion script
func TestGenerateCompletion_Fish(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("fish")

	output := stdout.String()
	errOutput := stderr.String()

	// Check that output was generated
	if output == "" {
		t.Error("Expected fish completion output, got empty string")
	}

	// Check that no errors were written
	if errOutput != "" {
		t.Errorf("Expected no errors, got: %s", errOutput)
	}

	// Check for fish-specific completion markers
	if !strings.Contains(output, "complete") || !strings.Contains(output, "did") {
		t.Errorf("Expected fish completion script to contain 'complete' command for 'did'")
	}
}

// TestGenerateCompletion_PowerShell tests generating powershell completion script
func TestGenerateCompletion_PowerShell(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("powershell")

	output := stdout.String()
	errOutput := stderr.String()

	// Check that output was generated
	if output == "" {
		t.Error("Expected powershell completion output, got empty string")
	}

	// Check that no errors were written
	if errOutput != "" {
		t.Errorf("Expected no errors, got: %s", errOutput)
	}

	// Check for powershell-specific completion markers
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("Expected powershell completion script to contain 'Register-ArgumentCompleter'")
	}
}

// TestGenerateCompletion_InvalidShell tests error handling for unsupported shell type
func TestGenerateCompletion_InvalidShell(t *testing.T) {
	exitCalled := false
	exitCode := 0
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit: func(code int) {
			exitCalled = true
			exitCode = code
		},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("invalidshell")

	// Check that exit was called with error code
	if !exitCalled {
		t.Error("Expected exit to be called for invalid shell type")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Check error message
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Unsupported shell") {
		t.Errorf("Expected 'Unsupported shell' error, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "invalidshell") {
		t.Errorf("Expected error to include invalid shell name 'invalidshell', got: %s", errOutput)
	}

	// Check that supported shells are listed
	if !strings.Contains(errOutput, "bash") {
		t.Errorf("Expected error to list 'bash' as supported shell, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "zsh") {
		t.Errorf("Expected error to list 'zsh' as supported shell, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "fish") {
		t.Errorf("Expected error to list 'fish' as supported shell, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "powershell") {
		t.Errorf("Expected error to list 'powershell' as supported shell, got: %s", errOutput)
	}

	// Check that no output was written to stdout
	if stdout.String() != "" {
		t.Errorf("Expected no stdout output for invalid shell, got: %s", stdout.String())
	}
}

// TestGenerateCompletion_EmptyShell tests error handling for empty shell string
func TestGenerateCompletion_EmptyShell(t *testing.T) {
	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit: func(code int) {
			exitCalled = true
		},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion("")

	// Check that exit was called
	if !exitCalled {
		t.Error("Expected exit to be called for empty shell type")
	}

	// Check error message
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Unsupported shell") {
		t.Errorf("Expected 'Unsupported shell' error, got: %s", errOutput)
	}
}

// TestGenerateCompletion_CaseSensitivity tests that shell types are case-sensitive
func TestGenerateCompletion_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name        string
		shell       string
		shouldError bool
	}{
		{"bash lowercase", "bash", false},
		{"bash uppercase", "BASH", true},
		{"bash mixed case", "Bash", true},
		{"zsh lowercase", "zsh", false},
		{"zsh uppercase", "ZSH", true},
		{"fish lowercase", "fish", false},
		{"fish uppercase", "FISH", true},
		{"powershell lowercase", "powershell", false},
		{"powershell uppercase", "POWERSHELL", true},
		{"powershell mixed case", "PowerShell", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit: func(code int) {
					exitCalled = true
				},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(tt.shell)

			if tt.shouldError {
				if !exitCalled {
					t.Errorf("Expected exit to be called for shell type %q", tt.shell)
				}
				if !strings.Contains(stderr.String(), "Unsupported shell") {
					t.Errorf("Expected error message for shell type %q, got: %s", tt.shell, stderr.String())
				}
			} else {
				if exitCalled {
					t.Errorf("Expected no error for shell type %q, but exit was called", tt.shell)
				}
				if stdout.String() == "" {
					t.Errorf("Expected completion output for shell type %q, got empty string", tt.shell)
				}
			}
		})
	}
}

// TestCompletionCmd_ValidArgs tests that the completion command has correct ValidArgs
func TestCompletionCmd_ValidArgs(t *testing.T) {
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}

	if len(completionCmd.ValidArgs) != len(expectedArgs) {
		t.Errorf("Expected %d ValidArgs, got %d", len(expectedArgs), len(completionCmd.ValidArgs))
	}

	// Check that all expected shells are present (order doesn't matter)
	for _, expected := range expectedArgs {
		found := false
		for _, actual := range completionCmd.ValidArgs {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected ValidArg %q not found in ValidArgs", expected)
		}
	}
}

// TestCompletionCmd_Run tests the completion command's Run function
func TestCompletionCmd_Run(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	// Call the completion command's Run function directly
	completionCmd.Run(completionCmd, []string{"bash"})

	output := stdout.String()

	// Check that output was generated
	if output == "" {
		t.Error("Expected completion output from Run function, got empty string")
	}

	// Check that no errors were written
	if stderr.String() != "" {
		t.Errorf("Expected no errors from Run function, got: %s", stderr.String())
	}
}

// TestCompletionCmd_HelpText tests that the completion command has proper documentation
func TestCompletionCmd_HelpText(t *testing.T) {
	// Check Short description
	if completionCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Check Long description contains usage examples
	if completionCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	// Check that Long description contains installation instructions for each shell
	longDesc := completionCmd.Long
	if !strings.Contains(longDesc, "bash") {
		t.Error("Expected Long description to contain bash installation instructions")
	}
	if !strings.Contains(longDesc, "zsh") {
		t.Error("Expected Long description to contain zsh installation instructions")
	}
	if !strings.Contains(longDesc, "fish") {
		t.Error("Expected Long description to contain fish installation instructions")
	}
	if !strings.Contains(longDesc, "powershell") {
		t.Error("Expected Long description to contain powershell installation instructions")
	}

	// Check that Long description contains usage examples
	if !strings.Contains(longDesc, "did completion bash") {
		t.Error("Expected Long description to contain usage example for bash")
	}
	if !strings.Contains(longDesc, "did completion zsh") {
		t.Error("Expected Long description to contain usage example for zsh")
	}
	if !strings.Contains(longDesc, "did completion fish") {
		t.Error("Expected Long description to contain usage example for fish")
	}
	if !strings.Contains(longDesc, "did completion powershell") {
		t.Error("Expected Long description to contain usage example for powershell")
	}
}

// TestGenerateCompletion_OutputSize tests that completion scripts are reasonably sized
func TestGenerateCompletion_OutputSize(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit:   func(code int) {},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(shell)

			output := stdout.String()

			// Check that output is non-empty and reasonably sized (at least 100 bytes)
			// Completion scripts should be substantial
			if len(output) < 100 {
				t.Errorf("Expected %s completion output to be at least 100 bytes, got %d bytes", shell, len(output))
			}

			// Check that output is not suspiciously large (more than 1MB would be unusual)
			if len(output) > 1024*1024 {
				t.Errorf("Expected %s completion output to be less than 1MB, got %d bytes", shell, len(output))
			}
		})
	}
}

// TestGenerateCompletion_ContainsDidCommand tests that completion scripts reference the 'did' command
func TestGenerateCompletion_ContainsDidCommand(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit:   func(code int) {},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(shell)

			output := stdout.String()

			// Check that output contains reference to 'did' command
			if !strings.Contains(output, "did") {
				t.Errorf("Expected %s completion output to contain 'did' command reference", shell)
			}
		})
	}
}

// TestGenerateCompletion_SpecialCharacters tests handling of shell names with special characters
func TestGenerateCompletion_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{"shell with space", "ba sh"},
		{"shell with dash", "ba-sh"},
		{"shell with underscore", "ba_sh"},
		{"shell with number", "bash5"},
		{"shell with dot", "bash.sh"},
		{"shell with slash", "bash/zsh"},
		{"shell with backslash", "bash\\zsh"},
		{"shell with tab", "bash\tzsh"},
		{"shell with newline", "bash\nzsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			exitCode := 0
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit: func(code int) {
					exitCalled = true
					exitCode = code
				},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(tt.shell)

			// All special character inputs should be treated as invalid
			if !exitCalled {
				t.Errorf("Expected exit to be called for shell %q", tt.shell)
			}
			if exitCode != 1 {
				t.Errorf("Expected exit code 1 for shell %q, got %d", tt.shell, exitCode)
			}

			errOutput := stderr.String()
			if !strings.Contains(errOutput, "Unsupported shell") {
				t.Errorf("Expected 'Unsupported shell' error for %q, got: %s", tt.shell, errOutput)
			}

			// No output should be written to stdout
			if stdout.String() != "" {
				t.Errorf("Expected no stdout for invalid shell %q, got: %s", tt.shell, stdout.String())
			}
		})
	}
}

// TestGenerateCompletion_UnicodeShellNames tests handling of unicode in shell names
func TestGenerateCompletion_UnicodeShellNames(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{"emoji", "bash🚀"},
		{"chinese characters", "中文"},
		{"arabic characters", "العربية"},
		{"russian characters", "русский"},
		{"mixed unicode", "bäsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit: func(code int) {
					exitCalled = true
				},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(tt.shell)

			// Unicode shell names should be treated as invalid
			if !exitCalled {
				t.Errorf("Expected exit to be called for unicode shell %q", tt.shell)
			}

			errOutput := stderr.String()
			if !strings.Contains(errOutput, "Unsupported shell") {
				t.Errorf("Expected 'Unsupported shell' error for %q, got: %s", tt.shell, errOutput)
			}
		})
	}
}

// TestGenerateCompletion_VeryLongShellName tests handling of excessively long shell names
func TestGenerateCompletion_VeryLongShellName(t *testing.T) {
	// Create a very long shell name (1000 characters)
	longShell := strings.Repeat("bash", 250)

	exitCalled := false
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit: func(code int) {
			exitCalled = true
		},
		StoragePath: func() (string, error) {
			return "", nil
		},
	}
	SetDeps(d)
	defer ResetDeps()

	generateCompletion(longShell)

	// Very long shell names should be treated as invalid
	if !exitCalled {
		t.Error("Expected exit to be called for very long shell name")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Unsupported shell") {
		t.Errorf("Expected 'Unsupported shell' error, got: %s", errOutput)
	}

	// No output should be written to stdout
	if stdout.String() != "" {
		t.Errorf("Expected no stdout for invalid shell, got: %s", stdout.String())
	}
}

// TestGenerateCompletion_WhitespaceOnly tests handling of whitespace-only shell names
func TestGenerateCompletion_WhitespaceOnly(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{"single space", " "},
		{"multiple spaces", "   "},
		{"tab", "\t"},
		{"multiple tabs", "\t\t\t"},
		{"mixed whitespace", " \t \n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit: func(code int) {
					exitCalled = true
				},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(tt.shell)

			// Whitespace-only input should be treated as invalid
			if !exitCalled {
				t.Errorf("Expected exit to be called for whitespace shell %q", tt.shell)
			}

			errOutput := stderr.String()
			if !strings.Contains(errOutput, "Unsupported shell") {
				t.Errorf("Expected 'Unsupported shell' error for %q, got: %s", tt.shell, errOutput)
			}
		})
	}
}

// TestGenerateCompletion_SimilarShellNames tests that only exact matches work
func TestGenerateCompletion_SimilarShellNames(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{"bash with trailing space", "bash "},
		{"bash with leading space", " bash"},
		{"bash with surrounding spaces", " bash "},
		{"zsh with tab", "zsh\t"},
		{"fish with newline", "fish\n"},
		{"bash-completion", "bash-completion"},
		{"zbash", "zbash"},
		{"bashrc", "bashrc"},
		{"fishes", "fishes"},
		{"zshell", "zshell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCalled := false
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  strings.NewReader(""),
				Exit: func(code int) {
					exitCalled = true
				},
				StoragePath: func() (string, error) {
					return "", nil
				},
			}
			SetDeps(d)
			defer ResetDeps()

			generateCompletion(tt.shell)

			// Only exact shell names should work
			if !exitCalled {
				t.Errorf("Expected exit to be called for near-match shell %q", tt.shell)
			}

			errOutput := stderr.String()
			if !strings.Contains(errOutput, "Unsupported shell") {
				t.Errorf("Expected 'Unsupported shell' error for %q, got: %s", tt.shell, errOutput)
			}

			// No output should be written to stdout
			if stdout.String() != "" {
				t.Errorf("Expected no stdout for invalid shell %q, got: %s", tt.shell, stdout.String())
			}
		})
	}
}
