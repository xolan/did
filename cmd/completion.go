package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for did.

The completion command allows you to generate shell completion scripts for
bash, zsh, fish, and powershell. This enables tab-completion for commands,
flags, and arguments in your shell.

Usage:
  did completion bash       Generate bash completion script
  did completion zsh        Generate zsh completion script
  did completion fish       Generate fish completion script
  did completion powershell Generate powershell completion script

Installation Instructions:

Bash:
  # Load completion temporarily (current session only):
  source <(did completion bash)

  # Install completion permanently:
  # Linux:
  did completion bash > ~/.local/share/bash-completion/completions/did

  # macOS (requires bash-completion from Homebrew):
  did completion bash > $(brew --prefix)/etc/bash_completion.d/did

Zsh:
  # Load completion temporarily (current session only):
  source <(did completion zsh)

  # Install completion permanently:
  # Add to ~/.zshrc:
  echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
  echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

  # Generate completion file:
  mkdir -p ~/.zsh/completion
  did completion zsh > ~/.zsh/completion/_did

  # Then restart your shell

Fish:
  # Install completion permanently:
  did completion fish > ~/.config/fish/completions/did.fish

PowerShell:
  # Open your PowerShell profile:
  notepad $PROFILE

  # Add this line to your profile:
  did completion powershell | Out-String | Invoke-Expression

  # Save and restart PowerShell`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generateCompletion(args[0])
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

// generateCompletion generates the appropriate completion script based on shell type
func generateCompletion(shell string) {
	var err error

	switch shell {
	case "bash":
		err = rootCmd.GenBashCompletion(deps.Stdout)
	case "zsh":
		err = rootCmd.GenZshCompletion(deps.Stdout)
	case "fish":
		err = rootCmd.GenFishCompletion(deps.Stdout, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(deps.Stdout)
	default:
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Unsupported shell '%s'\n", shell)
		_, _ = fmt.Fprintln(deps.Stderr, "Supported shells: bash, zsh, fish, powershell")
		deps.Exit(1)
		return
	}

	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to generate %s completion: %v\n", shell, err)
		deps.Exit(1)
		return
	}
}
