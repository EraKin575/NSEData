package fetcher

import (
	"context"
	"server/internal/models"
)

type Writer interface {
	Write(ctx context.Context, data models.Records) error
	Delete(ctx context.Context) error
}


