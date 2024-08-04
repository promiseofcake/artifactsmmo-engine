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
	x int
	y int
)

// craftCmd represents the gather command
var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "move position",
	RunE: func(cmd *cobra.Command, args []string) error {
		character := viper.GetViper().GetString("character")
		if character == "" {
			return fmt.Errorf("you must specify a character")
		}
		r := cmd.Context().Value(runnerKey).(*actions.Runner)

		resp, err := r.Move(cmd.Context(), character, x, y)
		if err != nil {
			slog.Error("failed to move", "error", err.Error())
			return fmt.Errorf("failed to move: %w", err)
		}

		sec := resp.GetRemainingCooldown()
		slog.Info("move results",
			"cooldown", sec,
		)
		time.Sleep(time.Duration(sec) * time.Second)
		return nil
	},
}

func init() {
	moveCmd.Flags().IntVar(&x, "x", 0, "The x position")
	moveCmd.Flags().IntVar(&y, "y", 0, "The y position")
	rootCmd.AddCommand(moveCmd)
}
