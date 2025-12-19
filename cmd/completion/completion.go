package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/tui/prompt"
)

// Cmd is the completion command
var Cmd = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for hevycli.

Supported shells: bash, zsh, fish, powershell

To load completions:

Bash:
  $ source <(hevycli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ hevycli completion bash > /etc/bash_completion.d/hevycli
  # macOS:
  $ hevycli completion bash > $(brew --prefix)/etc/bash_completion.d/hevycli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ hevycli completion zsh > "${fpath[1]}/_hevycli"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ hevycli completion fish | source

  # To load completions for each session, execute once:
  $ hevycli completion fish > ~/.config/fish/completions/hevycli.fish

PowerShell:
  PS> hevycli completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> hevycli completion powershell > hevycli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Allow interactive mode to handle it
			if cmdutil.IsInteractive() {
				return nil
			}
			return fmt.Errorf("please specify a shell: bash, zsh, fish, or powershell\n\nExample: hevycli completion bash")
		}
		if len(args) > 1 {
			return fmt.Errorf("only one shell can be specified at a time")
		}
		for _, valid := range []string{"bash", "zsh", "fish", "powershell"} {
			if args[0] == valid {
				return nil
			}
		}
		return fmt.Errorf("unknown shell %q\n\nSupported shells: bash, zsh, fish, powershell", args[0])
	},
	RunE: runCompletion,
}

func runCompletion(cmd *cobra.Command, args []string) error {
	var shell string

	if len(args) == 0 {
		// Interactive mode - prompt for shell selection
		options := []prompt.SelectOption{
			{ID: "bash", Title: "Bash", Description: "GNU Bourne-Again SHell"},
			{ID: "zsh", Title: "Zsh", Description: "Z shell (macOS default)"},
			{ID: "fish", Title: "Fish", Description: "Friendly Interactive SHell"},
			{ID: "powershell", Title: "PowerShell", Description: "Microsoft PowerShell"},
		}

		selected, err := prompt.Select("Select your shell", options, "Choose the shell for completion scripts")
		if err != nil {
			return err
		}
		shell = selected.ID
	} else {
		shell = args[0]
	}

	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
	return nil
}
