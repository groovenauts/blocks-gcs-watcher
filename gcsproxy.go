package main

import (
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Publisher interface {
	Publish(ctx context.Context, msgs ...*pubsub.Message) ([]string, error)
}

type GCSProxyNotifier struct {
	Topic Publisher
}

func NewGCSProxyNotifier(ctx context.Context, c *Watch) GCSProxyNotifier {
	// Creates a pubsubClient
	pubsubClient, err := pubsub.NewClient(ctx, c.ProjectID)
	if err != nil {
		log.Errorf(ctx, "Failed to create pubsubClient for %s: %v\n", c.ProjectID, err)
	}
	notifier := GCSProxyNotifier{pubsubClient.Topic(c.TopicName)}
	return notifier
}

func (n *GCSProxyNotifier) Created(ctx context.Context, url string) {
	log.Debugf(ctx, "GCSProxyNotifier#Created url: %v\n", url)

	attrs := map[string]string{
		"download_files": url,
	}
	log.Debugf(ctx, "GCSProxyNotifier#Created before Publish %v to %v\n", attrs, n.Topic)
	msgIDs, err := n.Topic.Publish(ctx, &pubsub.Message{
		Data:       []byte(""),
		Attributes: attrs,
	})
	log.Debugf(ctx, "GCSProxyNotifier#Created after Publish err: %v\n", err)
	if err != nil {
		log.Errorf(ctx, "Failed to publish of insertion of %v cause of %v\n", url, err)
	} else {
		log.Infof(ctx, "Message[ %v ] is published successfully\n", msgIDs)
	}
}

func (n *GCSProxyNotifier) Updated(ctx context.Context, url string) {
	// Do nothing
}

func (n *GCSProxyNotifier) Deleted(ctx context.Context, url string) {
	// Do nothing
}
