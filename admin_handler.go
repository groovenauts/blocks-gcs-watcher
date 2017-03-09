package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo"
	// "github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

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
	g.GET("", h.wrap(h.index))
	g.POST("", h.wrap(h.create))
	g.GET("/:id/edit", h.withId(h.edit))
	g.POST("/:id/update", h.withId(h.update))
	g.GET("/:id/delete", h.withId(h.delete))
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
	ctx := c.Get("aecontext").(context.Context)
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
	log.Debugf(ctx, "indexPage r: %v\n", r)
	return c.Render(http.StatusOK, "index", &r)
}

func (h *adminHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	watch := Watch{}
	c.Bind(&watch)
	log.Debugf(ctx, "Binded Watch: %v\n", watch)
	service := &WatchService{ctx}
	err := service.Create(&watch)
	if err != nil {
		h.flash.set(c, "alert", err.Error())
	} else {
		h.flash.set(c, "notice", "Watch is created successfully")
	}
	return c.Redirect(http.StatusFound, "/admin/watches")
}

func (h *adminHandler) delete(c echo.Context, w *Watch) error {
	ctx := c.Get("aecontext").(context.Context)
	service := &WatchService{ctx}
	err := service.Delete(w.ID)
	if err != nil {
		h.flash.set(c, "alert", fmt.Sprintf("Failed to destroy watch. id: %v error: ", w.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	h.flash.set(c, "notice", fmt.Sprintf("The Watch is deleted successfully. id: %v", w.ID))
	return c.Redirect(http.StatusFound, "/admin/watches")
}


type EditRes struct {
	Flash *Flash
	Watches Watches
	Target string
}

func (h *adminHandler) edit(c echo.Context, w *Watch) error {
	ctx := c.Get("aecontext").(context.Context)
	log.Debugf(ctx, "edit1: %v\n", w)
	service := &WatchService{ctx}
	watches, err := service.All()
	log.Debugf(ctx, "edit2: %v\n", w)
	if err != nil {
		log.Errorf(ctx, "edit: %v, [%T]%v\n", w, err, err)
		return err
	}
	log.Debugf(ctx, "edit3: %v\n", w)
	r := EditRes{
		Flash: c.Get("flash").(*Flash),
		Watches: watches,
		Target: w.ID,
	}
	log.Debugf(ctx, "edit4: %q\n", r.Target)
	return c.Render(http.StatusOK, "edit", &r)
}

func (h *adminHandler) update(c echo.Context, w *Watch) error {
	ctx := c.Get("aecontext").(context.Context)
	c.Bind(w)
	service := &WatchService{ctx}
	log.Debugf(ctx, "update: %v\n", w)
	err := service.Update(w)
	if err != nil {
		h.flash.set(c, "alert", err.Error())
	} else {
		h.flash.set(c, "notice", "Watch is created successfully")
	}
	return c.Redirect(http.StatusFound, "/admin/watches")
}




func (h *adminHandler) wrap(f func(c echo.Context) error) func(c echo.Context) error {
	return h.flash.with(h.withAEContext(f))
}

func (h *adminHandler) withAEContext(f func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := appengine.NewContext(c.Request())
		c.Set("aecontext", ctx)
		return f(c)
	}
}

func (h *adminHandler) withId(f func (c echo.Context, w *Watch) error) func (c echo.Context) error {
	return h.wrap(func(c echo.Context) error{
		ctx := c.Get("aecontext").(context.Context)
		service := &WatchService{ctx}
		w, err := service.Find(c.Param("id"))
		if err != nil {
			switch err.(type) {
			case *EntityNotFound:
				h.flash.set(c, "alert", fmt.Sprintf("Watch not found for id: %v", c.Param("id")))
				return c.Redirect(http.StatusFound, "/admin/watches")
			default:
				h.flash.set(c, "alert", fmt.Sprintf("Failed to find watch for id: %v error: ", c.Param("id"), err))
				return c.Redirect(http.StatusFound, "/admin/watches")
			}
		}
		return f(c, w)
	})
}
