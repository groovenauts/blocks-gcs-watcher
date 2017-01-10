package main

import (
	"os"
)

type Watch struct {
	ProjectID string
	BucketName string
	TopicName string
	WatchID string
}

func NewWatch(ctx context.Context) *Watch {
	return &Watch{
		ProjectID:  os.Getenv("PROJECT"),
		BucketName: os.Getenv("BUCKET"),
		TopicName: os.Getenv("TOPIC"),
		WatchID: os.Getenv("WATCH_ID"),
	}
}
