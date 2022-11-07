package cmd

import (
	"Plow/plow"
	"Plow/plow/utility"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List commands for the git repository",
	Long:  `List commands for the git repository`,
}

var listBranchCmd = &cobra.Command{
	Use:   "branches",
	Short: "Lists the branches the git repository",
	Long:  `Lists the branches of the git repository`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initBase()
		if err != nil {
			log.Fatal(err)
		}

		repo := operation.Repository()

		branches, err := repo.Branches()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Repository Branch List:")
		for _, b := range branches {
			fmt.Println(fmt.Sprintf("\t%s", b.Name()))
		}
	},
}

var listCommitCmd = &cobra.Command{
	Use:   "commits",
	Short: "Lists the change[commit] history of the git repository",
	Long:  `Lists the change[commit] history of the git repository`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initBase()
		if err != nil {
			log.Fatal(err)
		}

		repo := operation.Repository()

		commits, err := repo.GetCommitHistory(nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(fmt.Sprintf("Commits for branch [%s]", config.GitConfig.Branch))
		for _, c := range commits {
			fmt.Println(c)
		}
	},
}

var listChangesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Lists the files[changes] from the git repository",
	Long:  `Lists the files[changes] from the git repository`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initBase()
		if err != nil {
			log.Fatal(err)
		}

		changes, err := operation.GenerateChangeLog()
		if (changes == nil || len(changes.Bundles) == 0) && err == plow.ErrNoCommitsToProcess {
			fmt.Println("No changes identified, target at same commit level as repository, please confirm with log")
			return
		}

		if err != nil { // if err  is not no commits then something more significant is wrong
			log.Fatal(err)
		}

		for _, b := range changes.Bundles {
			fmt.Println(fmt.Sprintf("Change Bundle:[%s]", b.Ref.Hash))
			if len(b.Items) == 0 {
				utility.TabbedPrintln(2, "No Changes identified within this bundle")
			} else {
				for _, c := range b.Items {
					fmt.Println(fmt.Sprintf("\t[%s] : %s", c.ObjectType, c.Metadata.Name))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listBranchCmd)
	listCmd.AddCommand(listCommitCmd)
	listCmd.AddCommand(listChangesCmd)
}
