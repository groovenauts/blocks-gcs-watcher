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
