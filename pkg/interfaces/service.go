package interfaces

import (
	"context"
)

type CloudService interface {
	Name() string
	RegisterTools(server any) error
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
