package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"

	"google.golang.org/appengine/aetest"
)

// retry for datastore's eventual consistency
func retryWith(max int, impl func() func()) {
	for i := 0; i < max+1; i++ {
		f := impl()
		if f == nil {
			return
		}
		if i == max {
			f()
		} else {
			// Exponential backoff
			d := time.Duration(math.Pow(2.0, float64(i)) * 5.0)
			time.Sleep(d * time.Millisecond)
		}
	}
}

type (
	TopicUrl struct {
		topic string
		url   string
	}

	dummyNotifier struct {
		updated []TopicUrl
		deleted []TopicUrl
	}
)

func (dn *dummyNotifier) Updated(ctx context.Context, topic, url string) error {
	dn.updated = append(dn.updated, TopicUrl{topic, url})
	return nil
}
func (dn *dummyNotifier) Deleted(ctx context.Context, topic, url string) error {
	dn.deleted = append(dn.deleted, TopicUrl{topic, url})
	return nil
}

func BuildData(bucket, path string) map[string]interface{} {
	return map[string]interface{}{
		"selfLink":       "https://www.googleapis.com/storage/v1/b/" + bucket + "/o/dir1%2Ftestfile-20170220-1038.yml",
		"generation":     "1487554916603322",
		"metageneration": 1,
		"contentType":    "binary/octet-stream",
		"storageClass":   "NEARLINE",
		"kind":           "storage#object",
		"id":             bucket + "/" + path + "/1487554916603322",
		"bucket":         bucket,
		"updated":        "2017-02-20T01:41:56.589Z",
		"owner": map[string]string{
			"entity":   "user-00b4903a97fb634e7bd281721e3fa9acb6fa30bfa0a060f59e28449208eb3669",
			"entityId": "00b4903a97fb634e7bd281721e3fa9acb6fa30bfa0a060f59e28449208eb3669",
		},
		"etag":        "CLqrjfPFndICEAE=",
		"timeCreated": "2017-02-20T01:41:56.589Z",
		"crc32c":      "xgUuHw==",
		"name":        path,
		"timeStorageClassUpdated": "2017-02-20T01:41:56.589Z",
		"size":      1660,
		"md5Hash":   "t+IZeQ/RK0l4d1qVv3fpUA==",
		"mediaLink": "https://www.googleapis.com/download/storage/v1/b/" + bucket + "/o/dir1%2Ftestfile-20170220-1038.yml?generation=1487554916603322&alt=media",
	}
}

func TestProcessorExecute(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	notifier := &dummyNotifier{
		updated: []TopicUrl{},
		deleted: []TopicUrl{},
	}
	processor := &DefaultProcessor{}

	bucket1 := "test-bucket01"
	path1 := "dir1/testfile-20170220-1038.yml"

	ClearDatastore(t, ctx, WATCH_KIND)
	service := &WatchService{ctx}
	watch := &Watch{
		Seq: 1,
		Pattern: `\Ags://` + bucket1 + `/`,
		Topic: "projects/dummy-proj-999/topics/foo",
	}
	err = service.Create(watch)
	assert.NoError(t, err)

	retryWith(10, func() func() {
		watches, err := service.All()
		if assert.NoError(t, err)	{
			if len(watches) == 0 {
				return func(){
					t.Fatalf("len(watches) expects %v but was %v\n", 1, len(watches))
				}
			} else {
				return nil // OK
			}
		} else {
			return nil // Ignore Error
		}
	})

	data := BuildData(bucket1, path1)
	byteData, err := json.Marshal(data)
	assert.NoError(t, err)

	reader := bytes.NewReader(byteData)
	err = processor.execute(ctx, notifier, "exists", ioutil.NopCloser(reader))
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(notifier.deleted))
		if assert.Equal(t, 1, len(notifier.updated)) {
			assert.Equal(t, "gs://"+bucket1+"/"+path1, notifier.updated[0].url)
			assert.Equal(t, watch.Topic, notifier.updated[0].topic)
		}
	}

	notifier.updated = []TopicUrl{}
	reader = bytes.NewReader(byteData)
	err = processor.execute(ctx, notifier, "not_exists", ioutil.NopCloser(reader))
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(notifier.updated))
		if assert.Equal(t, 1, len(notifier.deleted)) {
			assert.Equal(t, "gs://"+bucket1+"/"+path1, notifier.deleted[0].url)
			assert.Equal(t, watch.Topic, notifier.deleted[0].topic)
		}
	}

	notifier.deleted = []TopicUrl{}
	notifier.updated = []TopicUrl{}

	invalidData1 := map[string]interface{}{
		"bucket": 1,
		"name": path1,
	}
	byteData, err = json.Marshal(invalidData1)
	assert.NoError(t, err)
	reader = bytes.NewReader(byteData)
	err = processor.execute(ctx, notifier, "exists", ioutil.NopCloser(reader))
	assert.Error(t, err)
	assert.Regexp(t, "bucket must be a string", err.Error())
	assert.Equal(t, 0, len(notifier.updated))
	assert.Equal(t, 0, len(notifier.deleted))

	invalidData2 := map[string]interface{}{
		"bucket": bucket1,
		"name": 2,
	}
	byteData, err = json.Marshal(invalidData2)
	assert.NoError(t, err)
	reader = bytes.NewReader(byteData)
	err = processor.execute(ctx, notifier, "exists", ioutil.NopCloser(reader))
	assert.Error(t, err)
	assert.Regexp(t, "name must be a string", err.Error())
	assert.Equal(t, 0, len(notifier.updated))
	assert.Equal(t, 0, len(notifier.deleted))
}
