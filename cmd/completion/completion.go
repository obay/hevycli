package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// Cmd is the completion command
var Cmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for hevycli.

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
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
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
	},
}
