package cmd

import (
	scafall "github.com/AidanDelaney/scafall/pkg"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scafall url output_dir",
		Short: "A project generation tool",
		Long:  `Scafall creates new project from project templates.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := map[string]interface{}{}
			s := scafall.New(vars, []string{})
			return s.Scaffold(args[0], args[1])
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
