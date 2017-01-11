package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	// "golang.org/x/net/context"
	// "google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "appengine_internal"
)

func TestShowConfig(t *testing.T) {
	test_proj := "proj-123"
	test_bucket := "bucket1"
	test_topic := "topic1"
	test_watch_id := "WID1"
	os.Setenv("PROJECT", test_proj)
	os.Setenv("BUCKET", test_bucket)
	os.Setenv("TOPIC", test_topic)
	os.Setenv("WATCH_ID", test_watch_id)

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
