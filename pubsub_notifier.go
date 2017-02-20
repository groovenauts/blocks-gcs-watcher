package main

import (
	"os"

	pubsub "google.golang.org/api/pubsub/v1"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/log"
)

type PubsubNotifier struct {
	service *pubsub.Service
}

func NewPubsubNotifier(ctx context.Context) (Notifier, error) {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Errorf(ctx, "Failed to create DrfaultClient\n")
		return nil, err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Errorf(ctx, "Failed to create pubsub.Service with %v: %v\n", client)
		return nil, err
	}

	notifier := PubsubNotifier{service}
	return &notifier, nil
}

func (n *PubsubNotifier) Updated(ctx context.Context, url string) error {
	log.Debugf(ctx, "PubsubNotifier#Updated url: %v\n", url)
	topic := os.Getenv("PUBSUB_TOPIC")

	// https://github.com/google/google-api-go-client/blob/master/examples/pubsub.go#L236-L244
	msg := &pubsub.PubsubMessage{
		Attributes: map[string]string{
			"download_files": url,
		},
	}
	req := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{msg},
	}
	log.Debugf(ctx, "PubsubNotifier#Updated before Publish %v to %v\n", msg, topic)
	if _, err := n.service.Projects.Topics.Publish(topic, req).Do(); err != nil {
		log.Errorf(ctx, "Failed to publish the update message of %v cause of %v\n", url, err)
		return err
	}

	return nil
}

func (n *PubsubNotifier) Deleted(ctx context.Context, url string) error {
	return nil
}
