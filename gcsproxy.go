package main

import (
	"golang.org/x/net/context"
	"cloud.google.com/go/pubsub"
	"google.golang.org/appengine/log"
)

type GCSProxyNotifier struct {
	Topic *pubsub.Topic
}

func NewGCSProxyNotifier(ctx context.Context, c *Watch) GCSProxyNotifier {
	// Creates a pubsubClient
	pubsubClient, err := pubsub.NewClient(ctx, c.ProjectID)
	if err != nil {
		log.Errorf(ctx, "Failed to create pubsubClient for %s: %v\n", c.ProjectID, err)
	}
	notifier := GCSProxyNotifier{ pubsubClient.Topic(c.TopicName) }
	return notifier
}

func (n *GCSProxyNotifier) Created(ctx context.Context, uf *UploadedFile) {
	log.Debugf(ctx, "GCSProxyNotifier#Created uf: %v\n", uf)

	attrs := map[string]string {
		"download_files": uf.Url,
	}
	log.Debugf(ctx, "GCSProxyNotifier#Created before Publish %v to %v\n", attrs, n.Topic)
	msgIDs, err := n.Topic.Publish(ctx, &pubsub.Message{
		Data: []byte(""),
		Attributes: attrs,
	})
	log.Debugf(ctx, "GCSProxyNotifier#Created after Publish err: %v\n", err)
	if err != nil {
		log.Errorf(ctx, "Failed to publish of insertion of %v cause of %v\n", uf.Url, err)
	} else {
		log.Infof(ctx, "Message[ %v ] is published successfully\n", msgIDs)
	}
}

func (n *GCSProxyNotifier) Updated(ctx context.Context, f *UploadedFile) {
	// Do nothing
}

func (n *GCSProxyNotifier) Deleted(ctx context.Context, f *UploadedFile) {
	// Do nothing
}
