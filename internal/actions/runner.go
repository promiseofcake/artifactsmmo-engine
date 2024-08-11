package actions

import (
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Runner is an executor for various Actions (Character / State)
type Runner struct {
	Client *client.ClientWithResponses
}

// NewDefaultRunner returns a new Actions command runner with a default client
func NewDefaultRunner(token string) (*Runner, error) {
	rClient := retryablehttp.NewClient()
	c, err := client.NewClientWithResponses(
		"https://api.artifactsmmo.com",
		client.WithRequestEditorFn(client.NewBearerAuthorizationRequestFunc(token)),
		client.WithHTTPClient(rClient.StandardClient()),
	)
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
