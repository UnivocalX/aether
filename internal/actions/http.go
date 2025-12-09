package actions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UnivocalX/aether/pkg/registry"
)

const (
	DEFAULT_HTTP_TIMEOUT = 30 * time.Second
)

type HTTPAction struct {
	Action
	endpoint registry.Endpoint
	client   http.Client
}

func (act *HTTPAction) Validate() error {
	if act.endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	return nil
}

func NewHTTPAction(ctx context.Context, endpoint string) *HTTPAction {
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return &HTTPAction{
		Action:   *NewAction(ctx),
		endpoint: registry.Endpoint(endpoint),
		client:   http.Client{Timeout: DEFAULT_HTTP_TIMEOUT},
	}
}
