package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "costaffective",
	Version: version,
	Short:  "Code Intelligence Research Platform",
	Long: `costaffective is a Code Intelligence Research Platform providing
MCP (Model Context Protocol) server for AI coding clients.

It provides:
  - Repository-aware retrieval pipeline
  - MCP server for AI coding assistants
  - Multi-client installation and configuration
  - Comprehensive diagnostics`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mycli.yaml)")
	// Cobra also supports local flags, which will only run when this command
	// is called directly,
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}