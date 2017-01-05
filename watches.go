package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"appengine"
	//"appengine/log"
	"appengine/datastore"
)

func init() {
	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	g := e.Group("/watches")
	g.Use(middleware.CORS())

	g.POST("", createWatch)
	g.GET("", getWatches)
}

func createWatch(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	w := Watch{}
	if err := c.Bind(&w); err != nil {
		//log.Errorf(ctx, "Failed to Bind: %v %v cause of %v\n", c, w, err)
		return err
	}
	key := datastore.NewIncompleteKey(ctx, "Watches", nil)
	if _, err := datastore.Put(ctx, key, &w); err != nil {
		//log.Errorf(ctx, "Failed to Put: %v %v cause of %v \n", key, w, err)
		return err
	}
	return c.JSON(http.StatusCreated, w)
}

func getWatches(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	q := datastore.NewQuery("Watches")
	var watches []Watch
	if _, err := q.GetAll(ctx, &watches); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, watches)
}
