package main

import (
	"golang.org/x/net/context"
)

type Notifier interface {
	Updated(ctx context.Context, topic, url string) error
	Deleted(ctx context.Context, topic, url string) error
}
