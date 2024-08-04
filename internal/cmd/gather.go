package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/spf13/cobra"
)

// gatherCmd represents the gather command
var gatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "Start a gather loop in your current location",
	RunE: func(cmd *cobra.Command, args []string) error {
		character := viper.GetViper().GetString("character")
		if character == "" {
			return fmt.Errorf("you must specify a character")
		}

		r := cmd.Context().Value(runnerKey).(*actions.Runner)
		for {
			slog.Info("about to gather")

			resp, err := r.Gather(cmd.Context(), character)
			if err != nil {
				slog.Error("failed to gather", "error", err.Error())
				return fmt.Errorf("failed to gather: %w", err)
			}

			sec := resp.GetRemainingCooldown()
			slog.Info("gather results",
				"xp", resp.SkillInfo.Xp,
				"items", resp.SkillInfo.Items,
				"cooldown", sec,
			)
			time.Sleep(time.Duration(sec) * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(gatherCmd)
}
