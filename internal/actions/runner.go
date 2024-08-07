package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Runner is an executor for various Actions (Character / State)
type Runner struct {
	Client *client.ClientWithResponses
}

// NewDefaultRunner returns a new Actions command runner with a default client
func NewDefaultRunner(token string) (*Runner, error) {
	c, err := client.NewClientWithResponses("https://api.artifactsmmo.com", client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to init new client: %w", err)
	}
	return &Runner{
		Client: c,
	}, nil
}

// NewRunnerWithClient returns a new Actions command runner with a pre-configured client
func NewRunnerWithClient(client *client.ClientWithResponses) *Runner {
	return &Runner{
		Client: client,
	}
}
