package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	runnerKey = "runner"
)

var (
	cfgFile       string
	characterName string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "artifactsmmo-engine",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		token := viper.GetViper().GetString("token")
		if token == "" {
			return errors.New("token required")
		}
		r, err := actions.NewDefaultRunner(token)
		if err != nil {
			return err
		}
		ctx := context.WithValue(cmd.Context(), runnerKey, r)
		cmd.SetContext(ctx)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.artifactsmmo-engine.yaml)")
	rootCmd.PersistentFlags().StringVar(&characterName, "character", "", "The name of your character")
	viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".artifactsmmo-engine" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".artifactsmmo-engine")
	}

	viper.SetEnvPrefix("mmo")
	viper.BindEnv("token")
	viper.BindEnv("character")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
