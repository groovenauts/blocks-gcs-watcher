package main

import (
	"strconv"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
)

func init() {
	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	g := e.Group("/watches")
	g.Use(middleware.CORS())

	g.POST("", createWatch)
	g.GET("", getWatches)
	g.GET("/:id/refresh", refresh)
	g.POST("/:id/run", runWatcher)
}

func createWatch(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	w := Watch{}
	if err := c.Bind(&w); err != nil {
		log.Errorf(ctx, "Failed to Bind: %v %v cause of %v\n", c, w, err)
		return err
	}
	key := datastore.NewIncompleteKey(ctx, "Watches", nil)
	if _, err := datastore.Put(ctx, key, &w); err != nil {
		log.Errorf(ctx, "Failed to Put: %v %v cause of %v \n", key, w, err)
		return err
	}
	return c.JSON(http.StatusCreated, w)
}

func getWatches(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	q := datastore.NewQuery("Watches")
	var watches []map[string]string
	for t := q.Run(ctx); ; {
		var w Watch
		key, err := t.Next(&w)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		v := make(map[string]string)
		v["ID"] = strconv.FormatInt(key.IntID(), 10)
		v["Project"] = w.ProjectID
		v["Bucket"] = w.BucketName
		v["Topic"] = w.TopicName
		watches = append(watches, v)
	}
	return c.JSON(http.StatusOK, watches)
}

func refresh(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Debugf(ctx, "/refresh\n")
	cron, ok := req.Header["X-Appengine-Cron"]
	if !ok || cron[0] != "true" {
		return c.JSON(http.StatusForbidden, map[string]string{ "message": "error" })
	}
	t := taskqueue.NewPOSTTask("/watches/" + c.Param("id")  + "/run", map[string][]string{  })
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{ "id": c.Param("id") })
}

func runWatcher(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Debugf(ctx, "/run\n")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	log.Debugf(ctx, "/run id=%v\n", id)
	key := datastore.NewKey(ctx, "Watches", "", id, nil)
	log.Debugf(ctx, "/run id=%v key=%v\n", id, key)
	w := Watch{}
	if err := datastore.Get(ctx, key, &w); err != nil {
		return err
	}
	log.Debugf(ctx, "Watcher is running for %v\n", w)
	watcher := &Watcher{}
	watcher.config = &w
	watcher.watchKey = key
	watcher.process(ctx)
	return c.JSON(http.StatusOK, w)
}
