package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/spf13/cobra"
)

var (
	craftCode string
	craftQty  int
)

// craftCmd represents the gather command
var craftCmd = &cobra.Command{
	Use:   "craft",
	Short: "Start a craft loop in your current location",
	RunE: func(cmd *cobra.Command, args []string) error {
		character := viper.GetViper().GetString("character")
		if character == "" {
			return fmt.Errorf("you must specify a character")
		}
		r := cmd.Context().Value(runnerKey).(*actions.Runner)
		for {
			slog.Info("about to craft", "code", craftCode, "quantity", craftQty)

			resp, err := r.Craft(cmd.Context(), character, craftCode, craftQty)
			if err != nil {
				slog.Error("failed to craft", "error", err.Error())
				return fmt.Errorf("failed to craft: %w", err)
			}

			sec := resp.GetRemainingCooldown()
			slog.Info("craft results",
				"xp", resp.SkillInfo.Xp,
				"items", resp.SkillInfo.Items,
				"cooldown", sec,
			)
			time.Sleep(time.Duration(sec) * time.Second)
		}
	},
}

func init() {
	craftCmd.Flags().StringVar(&craftCode, "code", "", "The code of your item to craft")
	craftCmd.Flags().IntVar(&craftQty, "qty", 0, "The quantity to craft")
	rootCmd.AddCommand(craftCmd)
}