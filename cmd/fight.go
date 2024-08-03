package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/spf13/cobra"
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

			sec := resp.GetRemainingCooldown()
			slog.Info("fight results",
				"results", resp.FightResponse,
				"cooldown", sec,
			)
			time.Sleep(time.Duration(sec) * time.Second)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(fightCmd)
}
