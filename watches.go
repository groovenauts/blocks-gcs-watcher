package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func init() {
	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	g := e.Group("/watches")
	g.Use(middleware.CORS())

	g.GET("", showConfig)
	g.GET("/refresh", refresh)
	g.POST("/run", runWatcher)
}

func showConfig(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "/showConfig\n")
	return c.JSON(http.StatusOK, NewWatch(ctx))
}

func refresh(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Debugf(ctx, "/refresh\n")
	// cron, ok := req.Header["X-Appengine-Cron"]
	// if !ok || cron[0] != "true" {
	// 	return c.JSON(http.StatusForbidden, map[string]string{ "message": "error" })
	// }
	t := taskqueue.NewPOSTTask("/watches/run", map[string][]string{})
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "OK"})
}

func runWatcher(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	w := NewWatch(ctx)
	log.Debugf(ctx, "/watches/run %v\n", w)
	watcher := &Watcher{}
	watcher.process(ctx, w)
	return c.JSON(http.StatusOK, w)
}
