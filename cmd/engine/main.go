package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/lmittmann/tint"
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
	configFlag     = "config"
	tokenFlag      = "token"
	logLevelFlag   = "log_level"
	charactersFlag = "characters"
)

func main() {
	v := initializeFlags()

	w := os.Stdout
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.Level(viper.GetInt(logLevelFlag)),
			TimeFormat: time.Kitchen,
		}),
	))

	slog.Info("starting artifacts-mmo game engine")

	r, err := actions.NewDefaultRunner(v.GetString("token"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	characters := v.GetStringSlice(charactersFlag)

	wg := &sync.WaitGroup{}
	for _, c := range characters {
		wg.Add(1)
		slog.Info("starting BuildInventory engine", "character", c)
		go func(ctx context.Context) {
			defer wg.Done()
			err = blockInitialAction(ctx, r, c)
			if err != nil {
				log.Fatal(err)
			}
			//err = engine.CookAll(ctx, r, c)
			err = engine.BuildInventory(ctx, r, c)
			if err != nil {
				log.Fatal(err)
			}
		}(ctx)
	}

	slog.Info("waiting for processes to complete")
	wg.Wait()
}

func initializeFlags() *viper.Viper {
	config := flag.String(configFlag, "", "path to config file")
	_ = flag.String(tokenFlag, "", "API token")
	_ = flag.StringSlice(charactersFlag, []string{}, "list of characters")
	_ = flag.Int(logLevelFlag, int(slog.LevelInfo), "log level")
	flag.Parse()

	err := initViper(*config)
	if err != nil {
		log.Fatal(err)
	}

	err = bindFlags([]string{configFlag, logLevelFlag, tokenFlag, charactersFlag})
	if err != nil {
		log.Fatal(err)
	}

	return viper.GetViper()
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
		slog.Info("character on cooldown waiting...", "character", character, "duration", d)
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
