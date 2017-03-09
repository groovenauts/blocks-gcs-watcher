package main

import (
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo"
	// "github.com/labstack/echo/middleware"

	// "golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	// "google.golang.org/appengine/taskqueue"
)

type adminHandler struct{
	flash *FlashHandler
}

func init() {
	h := &adminHandler{
		flash: &FlashHandler{
			path: "/admin/",
			expire: 10*time.Minute,
		},
	}

	t := &Template{
		templates: template.Must(template.ParseGlob("admin/*.html")),
	}
	e.Renderer = t

	g := e.Group("/admin/watches")
	g.GET("", h.flash.with(h.index))
	// g.POST("", h.flash.with(h.create))
	// g.GET("/:id/edit", h.withId(h.edit))
	// g.POST("/:id/update", h.withId(h.update))
	// g.POST("/:id/delete", h.withId(h.destroy))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type IndexRes struct {
	Flash *Flash
	Watches Watches
	NewSeq int
}

func (h *adminHandler) index(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	service := &WatchService{ctx}
	log.Debugf(ctx, "index\n")
	watches, err := service.All()
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	maxSeq := 0
	for _, w := range watches {
		if maxSeq < w.Seq {
			maxSeq = w.Seq
		}
	}
	log.Debugf(ctx, "indexPage watches: %v\n", watches)
	r := IndexRes{
		Flash: c.Get("flash").(*Flash),
		Watches: watches,
		NewSeq: maxSeq + 1,
	}
	return c.Render(http.StatusOK, "index", &r)
}
