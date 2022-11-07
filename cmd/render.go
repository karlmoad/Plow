package cmd

import (
	"Plow/plow/utility"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Renders the change object(s) to std.out only, does not apply change to target",
	Long:  `Renders the change object(s) to std.out only, does not apply change to target`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initBase()
		if err != nil {
			log.Fatal(err)
		}

		changes, err := operation.GenerateChangeLog()
		if err != nil {
			log.Fatal(err)
		}

		if changes == nil || len(changes.Bundles) == 0 {
			fmt.Println("No changes identified, target at same commit level as repository, please confirm with log")
			return
		}

		rendered, err := operation.RenderChanges(changes)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Validation Results.....")
		for _, b := range changes.Bundles {
			if len(b.Items) == 0 {
				utility.TabbedPrintln(2, "No Changes identified within this bundle")
			} else {
				fmt.Println(fmt.Sprintf("Change Bundle:[%s]", b.Ref.Hash))
				for _, c := range b.Items {
					fmt.Println(fmt.Sprintf("\tCritical(%d), Warning(%d), Success(%d) [%s] %s ",
						c.Validation.Critical,
						c.Validation.Warning,
						c.Validation.Success,
						c.ObjectType,
						c.Metadata.Name))
					for _, v := range c.Validation.Steps {
						var es string
						if v.Error != nil {
							es = v.Error.Error()
						}
						msg := fmt.Sprintf("\t\t validator: %s passed:[%t] %s", v.ValidatorName, v.Success, es)
						fmt.Println(msg)
					}
				}
			}
		}

		fmt.Println("Rendered Commands......")
		for _, rc := range rendered {
			fmt.Println(fmt.Sprintf("-- %s [%s]", rc.Item().Metadata.Name, rc.Item().Bundle.Ref.Hash))
			for _, scope := range rc.Item().ApplyInformation.GetScopes() {
				fmt.Println(fmt.Sprintf("-- scope:%s", scope.Name))
				for _, cmd := range scope.Commands {
					fmt.Println(cmd)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(renderCmd)
}
