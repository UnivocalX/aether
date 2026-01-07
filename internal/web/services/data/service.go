package data

import (
	"github.com/UnivocalX/aether/pkg/registry"
)

type Service struct {
	engine *registry.Engine
}

func NewService(engine *registry.Engine) *Service {
	return &Service{
		engine: engine,
	}
}
