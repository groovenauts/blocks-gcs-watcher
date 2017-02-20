package main

import (
	"golang.org/x/net/context"
)

type Notifier interface {
	Updated(ctx context.Context, url string) error
	Deleted(ctx context.Context, url string) error
}
