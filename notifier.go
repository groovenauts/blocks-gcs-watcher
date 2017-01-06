package main

import (
	"golang.org/x/net/context"
)

type Notifier interface {
	Created(ctx context.Context, f *UploadedFile)
	Updated(ctx context.Context, f *UploadedFile)
	Deleted(ctx context.Context, f *UploadedFile)
}
