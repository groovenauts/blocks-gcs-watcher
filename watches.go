package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func init() {
	// hook into the echo instance to create an endpoint group
	// and add specific middleware to it plus handlers
	g := e.Group("/watches")
	g.Use(middleware.CORS())

	h := &watcherHandler{&Watcher{}}
	g.GET("", h.ShowConfig)
	g.GET("/refresh", h.Refresh)
	g.POST("/run", h.RunWatcher)
}

type Processor interface {
	setup(ctx context.Context, w *Watch)
	process(ctx context.Context)
}

type watcherHandler struct {
	processor Processor
}

func (h *watcherHandler) ShowConfig(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "/showConfig\n")
	return c.JSON(http.StatusOK, NewWatch(ctx))
}

func (h *watcherHandler) Refresh(c echo.Context) error {
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

func (h *watcherHandler) RunWatcher(c echo.Context) error {
	req := c.Request()
	ctx := appengine.NewContext(req)
	w := NewWatch(ctx)
	log.Debugf(ctx, "/watches/run %v\n", w)
	defer h.setWatcher()
	h.processor.setup(ctx, w)
	h.processor.process(ctx)
	return c.JSON(http.StatusOK, w)
}

func (h *watcherHandler) setWatcher() {
	h.processor = &Watcher{}
}
