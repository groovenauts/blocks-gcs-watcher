package main

import (
	"testing"
	"time"

	// "golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
)

const layout = "2006-01-02 15:04:05 MST"

func parse(s string) time.Time {
	t, _ := time.Parse(layout, s)
	return t
}

func TestWatcherCalcDifferences(t *testing.T) {
	stored := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 11:11:11 JST"),
		"gs://bucket1/path/to/bar.txt": parse("2017-02-22 22:22:22 JST"),
	}
	found := map[string]time.Time{
		"gs://bucket1/path/to/foo.txt": parse("2017-01-11 11:11:11 JST"), // not changed
		"gs://bucket1/path/to/bar.txt": parse("2017-02-22 22:33:33 JST"), // updated
		"gs://bucket1/path/to/baz.txt": parse("2017-03-03 03:33:33 JST"), // created
	}

	// ctx := context.Background()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	w := &Watcher{}
	r := w.calcDifferences(ctx, stored, found)
	patterns := [][][]string{
		{r.created, {"gs://bucket1/path/to/baz.txt"}},
		{r.updated, {"gs://bucket1/path/to/bar.txt"}},
		{r.deleted, {}},
	}
	for _, pattern := range patterns {
		if !sameStrings(pattern[1], pattern[0]) {
			t.Fatalf("Expected %v but was %v", pattern[1], pattern[0])
		}
	}
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
