package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"github.com/spf13/cobra"
)

var stateContent string

// stateCmd represents the state command
var stateCmd = &cobra.Command{
	Use: "state",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := cmd.Context().Value(runnerKey).(*actions.Runner)

		if stateContent == "" {
			return fmt.Errorf("no map content provided")
		}

		mc, err := r.GetMaps(cmd.Context(), client.GetAllMapsMapsGetParamsContentType(stateContent))
		if err != nil {
			return fmt.Errorf("failed to get maps: %w", err)
		}

		for _, m := range mc {
			bts, _ := json.Marshal(m)
			fmt.Println(string(bts))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.Flags().StringVar(&stateContent, "content", "", "Type of map content to fetch")
}
