package fetcher

import (
	"context"
	"encoding/json"

	redis "github.com/redis/go-redis/v9"

	"server/internal/models"
)




type StreamWriter struct {
	writer *redis.Client
	stream string
}

func NewStreamWriter(client *redis.Client, stream string) *StreamWriter {
	return &StreamWriter{
		writer: client,
		stream: stream,
	}
}

func (sw *StreamWriter) Write(ctx context.Context, records models.Records) error {
	jsonPayload, err := json.Marshal(records)
	if err != nil {
		return err
	}

	return sw.writer.XAdd(ctx, &redis.XAddArgs{
		Stream: sw.stream,
		Values: map[string]any{
			"data": string(jsonPayload),
		},
	}).Err()

}

func (sw *StreamWriter) Delete(ctx context.Context) error {
	return sw.writer.Del(ctx, sw.stream).Err()
}

