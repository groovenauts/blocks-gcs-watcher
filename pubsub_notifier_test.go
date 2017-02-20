package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/aetest"
)

type (
	dummyPublisher struct {
		messages []*pubsub.PubsubMessage
	}
)

func (dp *dummyPublisher) Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error) {
	dp.messages = append(dp.messages, msg)
	return &pubsub.PublishResponse{}, nil
}

func TestNotifierFileUpdated(t *testing.T) {
	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	publisher := &dummyPublisher{[]*pubsub.PubsubMessage{}}
	notifier := &PubsubNotifier{publisher}

	url := "gs://test-bucket01/path/to/file"
	err = notifier.Updated(ctx, url)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(publisher.messages))

	msg := publisher.messages[0]
	assert.Equal(t, map[string]string{
		"download_files": url,
	}, msg.Attributes)
}
