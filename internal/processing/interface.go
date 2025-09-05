package processing

import (
	"context"
	"server/internal/models"
)

type Reader interface {
	ReadStream(ctx context.Context) ([]models.Records, error)
	ReadLatest(ctx context.Context) (models.Records, error)
}

type DBWriter interface {
	WriteToDB(ctx context.Context, records *[]models.ResponsePayload) error
}


