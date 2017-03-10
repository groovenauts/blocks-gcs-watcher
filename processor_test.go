package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"regexp"
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

	ext1 := ".dat"
	bucket1 := "test-bucket01"
	bucket2 := "test-bucket02"
	dir1 := "dir1"
	dir2 := "dir2"
	dir3 := "dir3"
	path1 := dir1 + "/testfile-20170220-1038.yml"
	path2 := dir2 + "/testfile-20170220-1038.yml"
	path3 := dir3 + "/testfile-20170220-1038.yml"
	path4 := dir2 + "/testfile-20170220-1038" + ext1

	topic1 := "projects/dummy-proj-999/topics/topic1"
	topic2 := "projects/dummy-proj-999/topics/topic2"
	topic3 := "projects/dummy-proj-999/topics/topic3"

	ClearDatastore(t, ctx, WATCH_KIND)
	service := &WatchService{ctx}
	watches := []*Watch{
		&Watch{
			Seq: 1,
			Pattern: `\Ags://` + bucket1 + `/` + dir1,
			Topic: topic1,
		},
		&Watch{
			Seq: 2,
			Pattern: regexp.QuoteMeta(ext1) + `\z`,
			Topic: topic3,
		},
		&Watch{
			Seq: 3,
			Pattern: `\Ags://` + bucket1 + `/` + dir2,
			Topic: topic2,
		},
	}
	for _, watch := range watches {
		err = service.Create(watch)
		assert.NoError(t, err)
	}

	retryWith(10, func() func() {
		r, err := service.All()
		if assert.NoError(t, err)	{
			if len(r) == len(watches) {
				return nil // OK
			} else {
				return func(){
					t.Fatalf("len(watches) expects %v but was %v\n", 1, len(watches))
				}
			}
		} else {
			return nil // Ignore Error
		}
	})

	type Pattern struct {
		bucket string
		path string
		topics []string
	}

	patterns := []Pattern{
		{bucket1, path1, []string{topic1}},
		{bucket1, path2, []string{topic2}},
		{bucket1, path3, []string{}},
		{bucket1, path4, []string{topic3}},
		{bucket2, path4, []string{topic3}},
	}

	for _, pattern := range patterns {
		notifier.deleted = []TopicUrl{}
		notifier.updated = []TopicUrl{}
		byteData, err := json.Marshal(BuildData(pattern.bucket, pattern.path))
		assert.NoError(t, err)
		reader := bytes.NewReader(byteData)
		err = processor.execute(ctx, notifier, "exists", ioutil.NopCloser(reader))
		if assert.NoError(t, err) {
			assert.Equal(t, 0, len(notifier.deleted))
			if assert.Equal(t, len(pattern.topics), len(notifier.updated)) {
				if len(pattern.topics) > 0 {
					assert.Equal(t, "gs://"+pattern.bucket+"/"+pattern.path, notifier.updated[0].url)
					for _, topic := range pattern.topics {
						assert.Equal(t, topic, notifier.updated[0].topic)
					}
				}
			}
		}
	}

	// Invalid data
	notifier.deleted = []TopicUrl{}
	notifier.updated = []TopicUrl{}

	invalidData1 := map[string]interface{}{
		"bucket": 1,
		"name": path1,
	}
	byteData, err := json.Marshal(invalidData1)
	assert.NoError(t, err)
	reader := bytes.NewReader(byteData)
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
