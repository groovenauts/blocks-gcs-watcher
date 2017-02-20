package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo"
	// "github.com/labstack/echo/middleware"

	// "golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	// "google.golang.org/appengine/taskqueue"
)

func init() {
	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	h := &watcherHandler{}
	e.GET("/", h.get)
	e.POST("/", h.post)
}

type watcherHandler struct {
}

func (h *watcherHandler) get(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "GET request to notification page.\n")
	verification := os.Getenv("GOOGLE_SITE_VERIFICATION")
	res := `<html><head>` +
        `<meta name="google-site-verification" content="` + verification + `" />` +
				`<title>blocks-gcs-watcher</title>` +
				`</head>` +
				`<body>` +
	      `</body></html>`
	return c.HTML(http.StatusOK, res)
}

func (h *watcherHandler) post(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "Processing OCN POST request\nHeader: %v\n", req.Header)
	resource_state := req.Header.Get("X-Goog-Resource-State")
	if resource_state == "" {
		log.Infof(ctx, "Unkown message received.\n")
	} else if resource_state == "sync" {
		log.Infof(ctx, "Sync message received.\n")
	} else {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error to read body")
    }
		var obj interface{}
		json.Unmarshal(body, &obj)
		log.Infof(ctx, "%v\n", obj)
	}
	return c.String(http.StatusOK, "OK")
}
