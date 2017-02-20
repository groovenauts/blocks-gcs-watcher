package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"

	"google.golang.org/appengine/aetest"
)

type (
	dummyNotifier struct {
		updatedUrls []string
		deletedUrls []string
	}
)

func (dn *dummyNotifier) Updated(ctx context.Context, url string) error {
	dn.updatedUrls = append(dn.updatedUrls, url)
	return nil
}
func (dn *dummyNotifier) Deleted(ctx context.Context, url string) error {
	dn.deletedUrls = append(dn.deletedUrls, url)
	return nil
}

func TestProcessorExecute(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	notifier := &dummyNotifier{
		updatedUrls: []string{},
		deletedUrls: []string{},
	}
	processor := &DefaultProcessor{}

	bucket1 := "test-bucket01"
	path1 := "dir1/testfile-20170220-1038.yml"

	data := map[string]interface{}{
		"selfLink":       "https://www.googleapis.com/storage/v1/b/" + bucket1 + "/o/dir1%2Ftestfile-20170220-1038.yml",
		"generation":     "1487554916603322",
		"metageneration": 1,
		"contentType":    "binary/octet-stream",
		"storageClass":   "NEARLINE",
		"kind":           "storage#object",
		"id":             bucket1 + "/" + path1 + "/1487554916603322",
		"bucket":         bucket1,
		"updated":        "2017-02-20T01:41:56.589Z",
		"owner": map[string]string{
			"entity":   "user-00b4903a97fb634e7bd281721e3fa9acb6fa30bfa0a060f59e28449208eb3669",
			"entityId": "00b4903a97fb634e7bd281721e3fa9acb6fa30bfa0a060f59e28449208eb3669",
		},
		"etag":        "CLqrjfPFndICEAE=",
		"timeCreated": "2017-02-20T01:41:56.589Z",
		"crc32c":      "xgUuHw==",
		"name":        path1,
		"timeStorageClassUpdated": "2017-02-20T01:41:56.589Z",
		"size":      1660,
		"md5Hash":   "t+IZeQ/RK0l4d1qVv3fpUA==",
		"mediaLink": "https://www.googleapis.com/download/storage/v1/b/" + bucket1 + "/o/dir1%2Ftestfile-20170220-1038.yml?generation=1487554916603322&alt=media",
	}
	byteData, err := json.Marshal(data)
	assert.NoError(t, err)

	reader := bytes.NewReader(byteData)
	processor.execute(ctx, notifier, "exists", ioutil.NopCloser(reader))
	assert.Equal(t, 1, len(notifier.updatedUrls))
	assert.Equal(t, 0, len(notifier.deletedUrls))
	assert.Equal(t, "gs://"+bucket1+"/"+path1, notifier.updatedUrls[0])

	notifier.updatedUrls = []string{}
	reader = bytes.NewReader(byteData)
	processor.execute(ctx, notifier, "not_exists", ioutil.NopCloser(reader))
	assert.Equal(t, 0, len(notifier.updatedUrls))
	assert.Equal(t, 1, len(notifier.deletedUrls))
	assert.Equal(t, "gs://"+bucket1+"/"+path1, notifier.deletedUrls[0])
}
