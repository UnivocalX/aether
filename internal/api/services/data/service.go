package data

import (
	"github.com/UnivocalX/aether/pkg/registry"
)

type Service struct {
	registry *registry.Engine
}

func NewService(registry *registry.Engine) *Service {
	return &Service{
		registry: registry,
	}
}
