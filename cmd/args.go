package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	scafall "github.com/buildpacks-community/scafall/pkg"
)

var (
	argsCmd = &cobra.Command{
		Use:   "args gitRepository",
		Short: "list arguments defined in a template",
		Long:  `Given gitRepository containing a template, list the arguments supported by the template.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			s, err := scafall.NewScafall(url)
			if err != nil {
				return err
			}
			subPathVal, err := cmd.Flags().GetString(subPath)
			if err == nil {
				scafall.WithSubPath(subPathVal)(&s)
			}

			description, sArgs, _ := s.TemplateArguments()
			fmt.Println(description)
			for _, a := range sArgs {
				fmt.Printf("\t%s\n", a)
			}
			return nil
		},
	}
)

func init() {
	argsCmd.Flags().StringP(subPath, "s", "", "use sub directory in template project to scaffold project")
}
