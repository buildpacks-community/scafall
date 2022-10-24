package cmd

import (
	"github.com/spf13/cobra"

	scafall "github.com/AidanDelaney/scafall/pkg"
)

const (
	outputFolderFlag = "path"
	argumentsFlag    = "arg"
	subPath          = "sub-path"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scafall gitRepository",
		Short: "A project generation tool",
		Long:  `Scafall creates new project from project templates.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			s, err := scafall.NewScafall(url)
			if err != nil {
				return err
			}
			outputDirVal, err := cmd.Flags().GetString(outputFolderFlag)
			if err == nil {
				scafall.WithOutputFolder(outputDirVal)(&s)
			}
			argumentsVal, err := cmd.Flags().GetStringToString(argumentsFlag)
			if err == nil {
				scafall.WithArguments(argumentsVal)(&s)
			}
			subPathVal, err := cmd.Flags().GetString(subPath)
			if err == nil {
				scafall.WithSubPath(subPathVal)(&s)
			}

			return s.Scaffold()
		},
	}
)

func init() {
	rootCmd.AddCommand(argsCmd)
	rootCmd.Flags().StringP(outputFolderFlag, "p", ".", "scaffold project in the provided output directory")
	rootCmd.Flags().StringToStringP(argumentsFlag, "o", map[string]string{}, "provide overrides as key-value pairs")
	rootCmd.Flags().StringP(subPath, "s", "", "use sub directory in template project to scaffold project")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
