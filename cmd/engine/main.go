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
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		log.Fatal(err)
	}
}

const (
	configFlag   = "config"
	tokenFlag    = "token"
	logLevelFlag = "log_level"
)

type Config struct {
	Token      string             `mapstructure:"token"`
	LogLevel   int                `mapstructure:"log_level"`
	Characters []Character        `mapstructure:"characters"`
	Orders     models.SimpleItems `mapstructure:"orders"`
}

type Character struct {
	Name    string   `mapstructure:"name"`
	Actions []string `mapstructure:"actions"`
}

func main() {
	v := initializeFlags()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

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

	wg := &sync.WaitGroup{}
	for _, c := range cfg.Characters {
		wg.Add(1)

		charCtx := logging.ContextWithLogger(ctx, slog.With("character", c.Name))
		l := logging.Get(charCtx)
		l.Info("starting execute engine", "actions", c.Actions)

		go func(charCtx context.Context) {
			defer wg.Done()
			err = blockInitialAction(charCtx, r, c.Name)
			if err != nil {
				log.Fatal(err)
			}
			err = engine.Execute(charCtx, r, c.Name, c.Actions, cfg.Orders)
			if err != nil {
				log.Fatal(err)
			}
		}(charCtx)
	}

	slog.Info("waiting for processes to complete")
	wg.Wait()
}

func initializeFlags() *viper.Viper {
	config := flag.String(configFlag, "", "path to config file")
	_ = flag.String(tokenFlag, "", "API token")
	_ = flag.Int(logLevelFlag, int(slog.LevelInfo), "log level")
	flag.Parse()

	err := initViper(*config)
	if err != nil {
		log.Fatal(err)
	}

	err = bindFlags([]string{configFlag, logLevelFlag, tokenFlag})
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
	l := logging.Get(ctx)
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	d, err := c.GetCooldownDuration()
	if err != nil {
		return fmt.Errorf("failed to get cooldown: %w", err)
	}

	if d > 0 {
		l.Info("character on cooldown waiting...", "character", character, "duration", d)
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
