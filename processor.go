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

	DefaultProcessor struct {
		notifier Notifier
	}
)

func (dp *DefaultProcessor) Run(ctx context.Context, state string, body io.ReadCloser) error {
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

	bucket := obj["bucket"].(string)
	name := obj["name"].(string)
	url := "gs://" + bucket + "/" + name

	if dp.notifier == nil {
		notifier, err := NewPubsubNotifier(ctx)
		if err != nil {
			return err
		}
		dp.notifier = notifier
	}

	switch state {
	case "exists":
		err = dp.notifier.Updated(ctx, url)
	case "not_exists":
		err = dp.notifier.Deleted(ctx, url)
	default:
		err = fmt.Errorf("Unknown state %v is given", state)
	}
	return err
}
