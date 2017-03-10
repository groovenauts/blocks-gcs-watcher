package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func ClearDatastore(t *testing.T, ctx context.Context, kind string) {
	q := datastore.NewQuery(kind).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = datastore.DeleteMulti(ctx, keys); err != nil {
		t.Fatal(err)
	}
}

func TestWatchCreate(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	ClearDatastore(t, ctx, WATCH_KIND)

	service := &WatchService{ctx}
	watch1 := &Watch{
		Seq: 1,
		Pattern: `\Ags://bucket1/dir1/`,
		Topic: "projects/dummy-proj-999/topics/foo",
	}
	// Valid Pattern
	err = service.Create(watch1)
	assert.NoError(t, err)
	assert.NotEmpty(t, watch1.ID)

	// Invalid pattern
	watch2 := &Watch{
		Seq: 2,
		Pattern: `\Ags://bucket1/(?dir1)/`,
		Topic: "projects/dummy-proj-999/topics/foo",
	}
	err = service.Create(watch2)
	if assert.Error(t, err) {
		assert.Regexp(t, `Invalid pattern`, err.Error())
		assert.Empty(t, watch2.ID)
	}

	// Invalid Topic
	watch3 := &Watch{
		Seq: 3,
		Pattern: `\Ags://bucket1/dir1/`,
		Topic: "topic-only",
	}
	err = service.Create(watch3)
	if assert.Error(t, err) {
		assert.Regexp(t, `Invalid topic`, err.Error())
		assert.Empty(t, watch3.ID)
	}

}
