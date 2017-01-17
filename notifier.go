package main

import (
	"golang.org/x/net/context"
)

type Notifier interface {
	Created(ctx context.Context, url string)
	Updated(ctx context.Context, url string)
	Deleted(ctx context.Context, url string)
}
