package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/cmd/completion"
	"github.com/obay/hevycli/cmd/config"
	"github.com/obay/hevycli/cmd/exercise"
	"github.com/obay/hevycli/cmd/folder"
	"github.com/obay/hevycli/cmd/routine"
	"github.com/obay/hevycli/cmd/workout"
	internalConfig "github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	// Global flags
	cfgFile   string
	outputFmt string
	noColor   bool
	quiet     bool
	verbose   bool

	// Global state
	cfg       *internalConfig.Config
	formatter output.Formatter
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hevycli",
	Short: "CLI for the Hevy fitness tracking platform",
	Long: `hevycli is a command-line interface for the Hevy fitness tracking platform.

It provides both human-friendly and machine-readable interfaces for managing
workouts, routines, and exercise data. Designed for power users, developers,
and AI agents that need structured access to fitness data.

Get started:
  hevycli config init     Initialize configuration with your API key
  hevycli config show     Display current configuration
  hevycli workout list    List your workouts

Get your API key at: https://hevy.com/settings?developer
(Requires Hevy Pro subscription)`,
	PersistentPreRunE: initializeApp,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Format error based on output format
		if formatter != nil {
			output.PrintError(formatter, err)
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.hevycli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "",
		"output format: json, table, plain (default: table)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false,
		"disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"enable verbose/debug output")

	// Add subcommands
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(workout.Cmd)
	rootCmd.AddCommand(routine.Cmd)
	rootCmd.AddCommand(exercise.Cmd)
	rootCmd.AddCommand(folder.Cmd)
	rootCmd.AddCommand(completion.Cmd)
	rootCmd.AddCommand(versionCmd)
}

// initializeApp loads configuration and sets up the formatter
func initializeApp(cmd *cobra.Command, args []string) error {
	// Skip initialization for config init command (chicken-egg problem)
	if cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
		return nil
	}

	// Load configuration
	var err error
	cfg, err = internalConfig.Load(cfgFile)
	if err != nil {
		// Use defaults if config doesn't exist
		cfg = internalConfig.DefaultConfig()
	}

	// Override config with flags
	if cmd.Flags().Changed("output") {
		cfg.Display.OutputFormat = outputFmt
	} else if outputFmt == "" {
		outputFmt = cfg.Display.OutputFormat
	}

	if cmd.Flags().Changed("no-color") {
		cfg.Display.Color = !noColor
	}

	// Check HEVYCLI_NO_COLOR environment variable
	if os.Getenv("HEVYCLI_NO_COLOR") != "" {
		cfg.Display.Color = false
		noColor = true
	}

	// Check NO_COLOR environment variable (standard)
	if os.Getenv("NO_COLOR") != "" {
		cfg.Display.Color = false
		noColor = true
	}

	// Initialize formatter
	formatter = output.NewFormatter(output.Options{
		Format:  output.FormatType(cfg.Display.OutputFormat),
		NoColor: !cfg.Display.Color,
		Quiet:   quiet,
		Verbose: verbose,
		Writer:  os.Stdout,
	})

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *internalConfig.Config {
	return cfg
}

// GetFormatter returns the current formatter
func GetFormatter() output.Formatter {
	return formatter
}

// version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hevycli v0.1.0")
	},
}
