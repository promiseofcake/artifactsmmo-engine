package main

import (
	"context"
	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/engine"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"os"
)

func init() {
	os.Setenv("TZ", "UTC")
}

func main() {
	slog.Info("starting artifacts-mmo game engine")
	var config = flag.String("config", "", "path to config file")
	flag.Parse()

	err := initViper(*config)
	if err != nil {
		log.Fatal(err)
	}

	v := viper.GetViper()

	r, err := actions.NewDefaultRunner(v.GetString("token"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	character := v.GetString("character")

	err = engine.Fight(ctx, r, character)
	if err != nil {
		log.Fatal(err)
	}

	//err = engine.Deposit(ctx, r, character)
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func initViper(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".artifactsmmo-engine")
	}

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("using config file:", viper.ConfigFileUsed())
	} else {
		return err
	}
	return nil
}
