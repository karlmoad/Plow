package cmd

import (
	"Plow/plow"
	"Plow/plow/objects"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

var cfgFile string
var fullChangeSet bool
var fastForward bool
var commitId string
var environment string

var config objects.Configuration
var options objects.Options
var operation *plow.Operation

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "plow",
	Short: "Solution to push object definitions into database structures",
	Long:  `Solution to push object definitions into a database based on change management repository in git`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func initBase() error {
	err := initializeConfiguration()
	if err != nil {
		return err
	}

	err = initOptions()
	if err != nil {
		return err
	}

	err = initOperation()
	if err != nil {
		return err
	}

	return nil
}

func initializeConfiguration() error {
	var configPath string
	if len(strings.TrimSpace(cfgFile)) > 0 {
		configPath = cfgFile
	} else {
		str, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		configPath = fmt.Sprintf("%s/.plow", str)
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var sysConfig objects.SystemConfiguration
	err = yaml.Unmarshal(bytes, &sysConfig)
	if err != nil {
		return err
	}

	if cfg, ok := sysConfig.Environments[environment]; ok {
		fmt.Println(fmt.Sprintf("<< Using Environment: %s >>", environment))
		config = cfg
	} else {
		return fmt.Errorf("error: unknown environment [%s]", environment)
	}
	return nil
}

func initOptions() error {
	if fastForward {
		options.OptionFlags.Set(objects.FastForwardSetting)
	}

	fmt.Println(fmt.Sprintf("Fast Forward set: %t", options.OptionFlags.Has(objects.FastForwardSetting)))

	if len(strings.TrimSpace(commitId)) > 0 {
		options.CommitId = &commitId
	}

	return nil
}

func initOperation() error {
	ctx, err := objects.NewPlowContext(context.Background(), config, options)
	if err != nil {
		return err
	}

	op, err := plow.NewOperation(ctx)
	if err != nil {
		log.Fatal(err)
	}
	operation = op
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.plow")
	rootCmd.PersistentFlags().StringVarP(&environment, "env", "e", "DEFAULT", "Environment name. Required, must match a specified name within config")
	rootCmd.PersistentFlags().BoolVar(&fullChangeSet, "full", false, "apply all files, not just changes")
	rootCmd.PersistentFlags().BoolVar(&fastForward, "fast-forward", false, "advance to commit ignoring history, if commit is not supplied HEAD will be assumed ")
	rootCmd.PersistentFlags().StringVar(&commitId, "commit", "", "commit id to process up to and including")
}
