package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clones the configured github repository to the current directory",
	Long:  `clones the configured github repository to the current directory`,
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
