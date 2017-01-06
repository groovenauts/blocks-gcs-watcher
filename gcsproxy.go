package main

import (
	"log"
	"golang.org/x/net/context"
	"cloud.google.com/go/pubsub"
)

type GCSProxyNotifier struct {
	Topic *pubsub.Topic
}

func NewGCSProxyNotifier(ctx context.Context, c *Watch) *GCSProxyNotifier {
	// Creates a pubsubClient
	pubsubClient, err := pubsub.NewClient(ctx, c.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsubClient for %s: %v", c.ProjectID, err)
	}
	notifier := &GCSProxyNotifier{ pubsubClient.Topic(c.TopicName) }
	return notifier
}

func (n *GCSProxyNotifier) Created(ctx context.Context, uf *UploadedFile) {
	attrs := map[string]string {
		"download_files": uf.Url,
	}
	msgIDs, err := n.Topic.Publish(ctx, &pubsub.Message{
		Attributes: attrs,
	})
	if err != nil {
		log.Fatalln("Failed to publish of insertion of ", uf.Url, " cause of ", err)
	} else {
		log.Println("Message[", msgIDs, "] is published successfully")
	}
}

func (n *GCSProxyNotifier) Updated(ctx context.Context, f *UploadedFile) {
	// Do nothing
}

func (n *GCSProxyNotifier) Deleted(ctx context.Context, f *UploadedFile) {
	// Do nothing
}
