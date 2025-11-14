package actions

import (
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
)

type LoadAssets struct {
	endpoint registry.Endpoint
}

func NewLoadAssets(endpoint string) *LoadAssets {
	return &LoadAssets{
		endpoint: registry.Endpoint(endpoint),
	}
}

func (action *LoadAssets) Run(source string) error {
	slog.Debug("executing load assets.",
		"AetherAPI", action.endpoint,
		"AssetsSource", source,
	)

	return nil
}
