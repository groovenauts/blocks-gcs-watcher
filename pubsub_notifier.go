package main

import (
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/log"
)

type (
	Publisher interface {
		Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error)
	}

	pubsubPublisher struct {
		topicsService *pubsub.ProjectsTopicsService
	}

	PubsubNotifier struct {
		publisher Publisher
	}
)

func (pp *pubsubPublisher) Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error) {
	req := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{msg},
	}
	return pp.topicsService.Publish(topic, req).Do()
}

func NewPubsubNotifier(ctx context.Context) (Notifier, error) {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Errorf(ctx, "Failed to create DefaultClient\n")
		return nil, err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Errorf(ctx, "Failed to create pubsub.Service with %v: %v\n", client)
		return nil, err
	}

	notifier := PubsubNotifier{&pubsubPublisher{service.Projects.Topics}}
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
	log.Debugf(ctx, "PubsubNotifier#Updated before Publish %v to %v\n", msg, topic)
	if _, err := n.publisher.Publish(topic, msg); err != nil {
		log.Errorf(ctx, "Failed to publish the update message of %v cause of %v\n", url, err)
		return err
	}

	return nil
}

func (n *PubsubNotifier) Deleted(ctx context.Context, url string) error {
	return nil
}
