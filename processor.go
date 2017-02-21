package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
)

type (
	Processor interface {
		Run(ctx context.Context, state string, body io.ReadCloser) error
	}

	DefaultProcessor struct{}
)

func (dp *DefaultProcessor) Run(ctx context.Context, state string, body io.ReadCloser) error {
	notifier, err := NewPubsubNotifier(ctx)
	if err != nil {
		return err
	}
	return dp.execute(ctx, notifier, state, body)
}

func (dp *DefaultProcessor) execute(ctx context.Context, notifier Notifier, state string, body io.ReadCloser) error {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	var obj map[string]interface{}
	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}
	log.Infof(ctx, "%v\n", obj)

	bucket, ok := obj["bucket"].(string)
	if !ok {
		return fmt.Errorf("bucket must be a string but it was an %T (%v)", obj["bucket"], obj["bucket"])
	}
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("name must be a string but it was an %T (%v)", obj["name"], obj["name"])
	}
	url := "gs://" + bucket + "/" + name

	switch state {
	case "exists":
		err = notifier.Updated(ctx, url)
	case "not_exists":
		err = notifier.Deleted(ctx, url)
	default:
		err = fmt.Errorf("Unknown state %v is given", state)
	}
	return err
}
