package actions

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Runner is an executor for various Actions (Character / State)
type Runner struct {
	Client *client.ClientWithResponses
}

type retryLogger struct {
	*slog.Logger
}

func newRetryLogger() *retryLogger {
	return &retryLogger{
		slog.Default(),
	}
}

func (h *retryLogger) Error(msg string, keysAndValues ...interface{}) {
	h.Logger.Error(msg, keysAndValues...)
}
func (h *retryLogger) Info(msg string, keysAndValues ...interface{}) {
	h.Logger.Info(msg, keysAndValues...)
}
func (h *retryLogger) Debug(msg string, keysAndValues ...interface{}) {
	h.Logger.Debug(msg, keysAndValues...)
}
func (h *retryLogger) Warn(msg string, keysAndValues ...interface{}) {
	h.Logger.Warn(msg, keysAndValues...)
}

// NewDefaultRunner returns a new Actions command runner with a default client
func NewDefaultRunner(token string) (*Runner, error) {
	rClient := retryablehttp.NewClient()
	rClient.Logger = newRetryLogger()

	// setup retries for contention
	rClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		switch resp.StatusCode {
		case 461, 486, 499:
			return true, nil
		default:
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
	}

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
