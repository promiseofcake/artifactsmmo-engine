package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/engine"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		log.Fatal(err)
	}
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

	err = blockInitialAction(ctx, r, character)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("starting BuildInventory engine")
	err = engine.BuildInventory(ctx, r, character)
	if err != nil {
		log.Fatal(err)
	}
}

func blockInitialAction(ctx context.Context, r *actions.Runner, character string) error {
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	d, err := c.GetCooldownDuration()
	if err != nil {
		return fmt.Errorf("failed to get cooldown: %w", err)
	}

	if d > 0 {
		slog.Info("character on cooldown waiting...", "duration", d)
		time.Sleep(d)
	}
	return nil
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
		slog.Debug("using config file:", "file", viper.ConfigFileUsed())
	} else {
		return err
	}
	return nil
}
