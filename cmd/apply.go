package cmd

import (
	"Plow/plow/utility"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply changes to the configured target",
	Long:  `apply changes to the configured target`,

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

		ctx := context.Background()

		err = operation.ApplyChanges(ctx, changes)
		if err != nil {
			fmt.Println("Error occurred during application of changes")
			fmt.Println(fmt.Sprintf("Error: %s", err.Error()))
		}

		fmt.Println("Application Results.....")

		for _, b := range changes.Bundles {
			fmt.Println(fmt.Sprintf("Change Bundle:[%s]", b.Ref.Hash))
			if len(b.Items) == 0 {
				utility.TabbedPrintln(2, "No Changes identified within this bundle")
			} else {
				objOrder := operation.GetExecutionOrder()
				for _, otype := range objOrder {
					items, err := b.GetChangesOfType(otype)
					if err == nil {
						for _, c := range items {
							fmt.Println(fmt.Sprintf("\t[%s] %s ",
								c.ObjectType,
								c.Metadata.Name))
							utility.TabbedPrintln(2, "Validation Information:........................")
							for _, v := range c.Validation.Steps {
								var es string
								if v.Error != nil {
									es = v.Error.Error()
								}
								utility.TabbedPrintlnf(3, "validator: %s passed:[%t] %s", v.ValidatorName, v.Success, es)
							}

							utility.TabbedPrintln(2, "Application information:.......................")

							success, partial, err := c.ApplyInformation.IsSuccess()
							if success {
								utility.TabbedPrintln(3, "Success: object applied to target")

							} else {
								if c.ApplyInformation.Executed == false {
									utility.TabbedPrintln(3, "Skipped: true, this object was not executed due to error preceding it")
								} else {
									utility.TabbedPrintlnf(3, "Failed,  Object Partially Applied: %t", partial)
									if err != nil {
										utility.TabbedPrintlnf(3, "Error: %s", err.Error())
									}
									//print each scope and status to convey detail error context
									for _, scope := range c.ApplyInformation.GetScopes() {
										info := scope.GetEffectInfo()
										var msg string
										if info.Error != nil {
											msg = err.Error()
										}

										utility.TabbedPrintlnf(4, "Scope: %s, Executed: %t, Success: %t, Partial: %t, Error: %s",
											scope.Name,
											info.Executed,
											info.Success,
											info.Partial,
											msg)

										utility.TabbedPrintln(4, "---------command")
										for _, cmd := range scope.Commands {
											utility.TabbedPrintln(4, cmd)
										}
										utility.TabbedPrintln(4, "................")
									}
								}
							}
						}
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
