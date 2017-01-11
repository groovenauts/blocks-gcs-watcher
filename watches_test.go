package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	// "google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "appengine_internal"
)

const (
	test_proj     = "proj-123"
	test_bucket   = "bucket1"
	test_topic    = "topic1"
	test_watch_id = "WID1"
)

func setupEnv() {
	os.Setenv("PROJECT", test_proj)
	os.Setenv("BUCKET", test_bucket)
	os.Setenv("TOPIC", test_topic)
	os.Setenv("WATCH_ID", test_watch_id)
}

func TestShowConfig(t *testing.T) {
	setupEnv()

	inst, req := newInstanceAndReq(t, func(i aetest.Instance) (*http.Request, error) {
		return i.NewRequest("GET", "/watches", nil)
	})
	defer inst.Close()

	// Setup
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/watches")
	h := &watcherHandler{}

	// Assertions
	expected := `{"ProjectID":"` + test_proj + `","BucketName":"` + test_bucket + `","TopicName":"` + test_topic + `","WatchID":"` + test_watch_id + `"}`
	if assert.NoError(t, h.ShowConfig(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expected, rec.Body.String())
	}
}

type dummyWatcher struct {
	givenWatch *Watch
	processed  bool
}

func (dw *dummyWatcher) setup(ctx context.Context, w *Watch) {
	dw.givenWatch = w
}

func (dw *dummyWatcher) process(ctx context.Context) {
	dw.processed = true
}

func TestRunWatcher(t *testing.T) {
	setupEnv()

	inst, req := newInstanceAndReq(t, func(i aetest.Instance) (*http.Request, error) {
		return i.NewRequest("POST", "/watches/run", nil)
	})
	defer inst.Close()

	// Setup
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/watches/run")
	dw := &dummyWatcher{}
	h := &watcherHandler{dw}

	if assert.NoError(t, h.RunWatcher(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	patterns := [][]string{
		{"ProjectID", test_proj, dw.givenWatch.ProjectID},
		{"BucketName", test_bucket, dw.givenWatch.BucketName},
		{"TopicName", test_topic, dw.givenWatch.TopicName},
		{"WatchID", test_watch_id, dw.givenWatch.WatchID},
	}
	for _, pattern := range patterns {
		if pattern[1] != pattern[2] {
			t.Fatalf("%v expected % but was %v\n", pattern[0], pattern[1], pattern[2])
		}
	}

	if !dw.processed {
		t.Fatalf("dummyWatcher wasn't processed")
	}
}

func newInstanceAndReq(t *testing.T, f func(aetest.Instance) (*http.Request, error)) (aetest.Instance, *http.Request) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	//defer inst.Close()

	req, err := f(inst)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	return inst, req
}
