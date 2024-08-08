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

const (
	configFlag    = "config"
	tokenFlag     = "token"
	characterFlag = "character"
)

func main() {
	slog.Info("starting artifacts-mmo game engine")
	config := flag.String(configFlag, "", "path to config file")
	_ = flag.String(tokenFlag, "", "API token")
	_ = flag.String(characterFlag, "", "character name")
	flag.Parse()

	err := initViper(*config)
	if err != nil {
		log.Fatal(err)
	}

	err = bindFlags([]string{configFlag, tokenFlag, characterFlag})
	if err != nil {
		log.Fatal(err)
	}

	v := viper.GetViper()

	r, err := actions.NewDefaultRunner(v.GetString("token"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	c := v.GetString(characterFlag)

	err = blockInitialAction(ctx, r, c)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("starting BuildInventory engine")
	err = engine.BuildInventory(ctx, r, c)
	if err != nil {
		log.Fatal(err)
	}
}

func bindFlags(flags []string) error {
	for _, f := range flags {
		err := viper.BindPFlag(f, flag.Lookup(f))
		if err != nil {
			return fmt.Errorf("failed to bind flag: %w", err)
		}
	}
	return nil
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
