package main

import (
	"testing"

	"cloud.google.com/go/pubsub"

	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
)

type DummyPublisher struct {
	messages []*pubsub.Message
}

func (dp *DummyPublisher) Publish(ctx context.Context, msgs ...*pubsub.Message) ([]string, error) {
	dp.messages = msgs
	return []string{"dummyMessageID"}, nil
}

func TestGcsProxyNotifier(t *testing.T) {
	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	url := "gs://bucket1/path/to/foo.txt"
	dummyPublisher := DummyPublisher{}
	notifier := GCSProxyNotifier{Topic: &dummyPublisher}
	notifier.Created(ctx, url)
	if len(dummyPublisher.messages) != 1 {
		t.Fatalf("messages has just 1 message but it was %v\n", dummyPublisher.messages)
	}
	msg := dummyPublisher.messages[0]
	v, ok := msg.Attributes["download_files"]
	if !ok {
		t.Fatalf("message has no download_files attribute\n", msg.Attributes)
	}
	if url != v {
		t.Fatalf("download_files expected %v but was %v\n", url, v)
	}
}
