package processing

import (
	"context"
	"encoding/json"

	redis "github.com/redis/go-redis/v9"

	"server/internal/models"
)

type StreamReader struct {
	reader *redis.Client
	stream string
}

func NewStreamReader(client *redis.Client, stream string) *StreamReader {
	return &StreamReader{
		reader: client,
		stream: stream,
	}
}

func (r *StreamReader) ReadStream(ctx context.Context) ([]models.Records, error) {
	var records []models.Records

	stream, err := r.reader.XRange(ctx, r.stream, "-", "+").Result()
	if err != nil {
		return records, err
	}

	for _, msg := range stream {
		if data, ok := msg.Values["data"].(string); ok {
			var record models.Records
			if err := json.Unmarshal([]byte(data), &record); err != nil {
				return nil, err
			}
			records = append(records, record)
		}
	}

	return records, nil

}

func (r *StreamReader) ReadLatest(ctx context.Context) (models.Records, error) {
	var record models.Records

	stream, err := r.reader.XRevRangeN(ctx, r.stream, "+", "-", 1).Result()
	if err != nil {
		return record, err
	}

	if len(stream) == 0 {
		return record, nil
	}  

	if data, ok := stream[0].Values["data"].(string); ok {
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			return record, err
		}
	}
	return record, nil

}
