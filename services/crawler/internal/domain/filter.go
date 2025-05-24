package domain

import "context"

type Filter interface {
	AddCtx(ctx context.Context, data []byte) error
	ExistsCtx(ctx context.Context, data []byte) (bool, error)
}
