package main

import (
	"sort"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const layout = "2006-01-02 15:04:05 MST"

func parse(s string) time.Time {
	t, _ := time.Parse(layout, s)
	return t
}

func TestWatcherCalcDifferences(t *testing.T) {
	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	w := &Watcher{}

	files1 := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 11:11:11 JST"),
		"gs://bucket1/path/to/bar.txt": parse("2017-02-22 22:22:22 JST"),
	}

	files2 := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 11:11:11 JST"), // not changed
		"gs://bucket1/path/to/bar.txt": parse("2017-02-22 22:33:33 JST"), // updated
		"gs://bucket1/path/to/baz.txt": parse("2017-03-03 03:33:33 JST"), // created
	}

	files3 := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 11:11:11 JST"), // not changed
	}

	check := func(patterns [][]string, diffs *differences) {
		actuals := [][]string{diffs.created, diffs.updated, diffs.deleted}
		for i, pattern := range patterns {
			sort.Strings(pattern)
			sort.Strings(actuals[i])
			if !sameStrings(pattern, actuals[i]) {
				t.Fatalf("Expected %v but was %v", pattern, actuals[i])
			}
		}
	}

	check([][]string{
		{"gs://bucket1/path/to/baz.txt"},
		{"gs://bucket1/path/to/bar.txt"},
		{},
	}, w.calcDifferences(ctx, files1, files2))

	check([][]string{
		{},
		{},
		{"gs://bucket1/path/to/bar.txt", "gs://bucket1/path/to/baz.txt"},
	}, w.calcDifferences(ctx, files2, files3))
}

func sameStrings(strs1, strs2 []string) bool {
	if len(strs1) != len(strs2) {
		return false
	}
	for i, v := range strs1 {
		if v != strs2[i] {
			return false
		}
	}
	return true
}

type dummyNotifier struct {
	created []string
	updated []string
	deleted []string
}

func (n *dummyNotifier) Created(ctx context.Context, url string) {
	n.created = append(n.created, url)
}

func (n *dummyNotifier) Updated(ctx context.Context, url string) {
	n.updated = append(n.updated, url)
}

func (n *dummyNotifier) Deleted(ctx context.Context, url string) {
	n.deleted = append(n.deleted, url)
}

func TestWatcherStoreAndNotify(t *testing.T) {
	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	w := &Watcher{}
	w.watchKey = datastore.NewKey(ctx, "TestWatches", "TestWatchId", 0, nil)

	keys := func(m map[string]time.Time) []string {
		r := []string{}
		for k, _ := range m {
			r = append(r, k)
		}
		return r
	}

	fileBases := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 12:00:00 JST"),
		"gs://bucket1/path/to/bar.txt": parse("2017-01-11 15:00:00 JST"),
		"gs://bucket1/path/to/baz.txt": parse("2017-01-11 14:00:00 JST"),
	}
	files := keys(fileBases)
	sort.Strings(files)

	check := func(diffs differences) {
		notifier := dummyNotifier{
			created: make([]string, 0, 1),
			updated: make([]string, 0, 1),
			deleted: make([]string, 0, 1),
		}
		w.notifier = &notifier
		w.storeAndNotify(ctx, &diffs, fileBases)
		if !sameStrings(diffs.created, notifier.created) {
			t.Fatalf("Expected %v but was %v", diffs.created, notifier.created)
		}
		if !sameStrings(diffs.updated, notifier.updated) {
			t.Fatalf("Expected %v but was %v", diffs.updated, notifier.updated)
		}
		if !sameStrings(diffs.deleted, notifier.deleted) {
			t.Fatalf("Expected %v but was %v", diffs.deleted, notifier.deleted)
		}
	}

	deleteAll(t, ctx, w.watchKey)

	check(differences{created: files, updated: []string{}, deleted: []string{}})

	updated1 := fetchUrls(t, ctx, w.watchKey)
	if !sameStrings(files, updated1) {
		t.Fatalf("Expected %v but was %v\n", files, updated1)
	}

	check(differences{created: []string{}, updated: files, deleted: []string{}})

	updated2 := fetchUrls(t, ctx, w.watchKey)
	if !sameStrings(files, updated2) {
		t.Fatalf("Expected %v but was %v\n", files, updated2)
	}

	check(differences{created: []string{}, updated: []string{}, deleted: files})

	updated3 := fetchUrls(t, ctx, w.watchKey)
	if len(updated3) > 0 {
		t.Fatalf("Expected no URL found but found %v\n", updated3)
	}
}

func deleteAll(t *testing.T, ctx context.Context, key *datastore.Key) {
	q := datastore.NewQuery("UploadedFiles").Ancestor(key).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = datastore.DeleteMulti(ctx, keys); err != nil {
		t.Fatal(err)
	}
}

func fetchUrls(t *testing.T, ctx context.Context, key *datastore.Key) []string {
	q := datastore.NewQuery("UploadedFiles").Ancestor(key)
	var ufs []UploadedFile
	if _, err := q.GetAll(ctx, &ufs); err != nil {
		t.Fatal(err)
	}
	log.Infof(ctx, "Fetched Files: %v\n", ufs)
	r := []string{}
	for _, uf := range ufs {
		r = append(r, uf.Url)
	}
	sort.Strings(r)
	return r
}
