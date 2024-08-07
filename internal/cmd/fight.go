package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
)

var fightCmd = &cobra.Command{
	Use:   "fight",
	Short: "fight something",
	RunE: func(cmd *cobra.Command, args []string) error {
		character := viper.GetViper().GetString("character")
		if character == "" {
			return fmt.Errorf("you must specify a character")
		}
		r := cmd.Context().Value(runnerKey).(*actions.Runner)
		for {
			slog.Info("about to fight")

			resp, err := r.Fight(cmd.Context(), character)
			if err != nil {
				slog.Error("failed to fight", "error", err.Error())
				return fmt.Errorf("failed to fight: %w", err)
			}

			cooldown := resp.GetCooldownDuration()
			slog.Info("fight results",
				"results", resp.FightResponse,
				"cooldown", cooldown,
			)
			time.Sleep(cooldown)
		}
	},
}

func init() {
	rootCmd.AddCommand(fightCmd)
}
