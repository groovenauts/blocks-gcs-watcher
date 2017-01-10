package main

import (
	"os"
	"log"
	"testing"

	// "golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
)

func TestNewWatch(t *testing.T) {
	test_proj := "proj-123"
	test_bucket := "bucket1"
	test_topic := "topic1"
	test_watch_id := "WID1"
	os.Setenv("PROJECT", test_proj)
	os.Setenv("BUCKET", test_bucket)
	os.Setenv("TOPIC", test_topic)
	os.Setenv("WATCH_ID", test_watch_id)

	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	log.Printf("ctx: %v\n", ctx)
	w := NewWatch(ctx)

	if test_proj != w.ProjectID {
		t.Fatalf("Expected %v but was %v", test_proj, w.ProjectID)
	}
	if test_bucket != w.BucketName {
		t.Fatalf("Expected %v but was %v", test_bucket, w.BucketName)
	}
	if test_topic != w.TopicName {
		t.Fatalf("Expected %v but was %v", test_topic, w.TopicName)
	}
	if test_watch_id != w.WatchID {
		t.Fatalf("Expected %v but was %v", test_watch_id, w.WatchID)
	}
}
