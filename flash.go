package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
)

type (
	FlashHandler struct {
		path string
		expire time.Duration
	}

	Flash struct {
		Alert  string
		Notice string
	}
)

func (fh *FlashHandler) set(c echo.Context, name, value string) {
	fh.setWithExpire(c, name, value, time.Now().Add(fh.expire))
}

func (fh *FlashHandler) setWithExpire(c echo.Context, name, value string, expire time.Time) {
	cookie := new(http.Cookie)
	cookie.Path = "/admin/"
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = expire
	c.SetCookie(cookie)
}

func (fh *FlashHandler) load(c echo.Context) *Flash {
	f := Flash{}
	cookie, err := c.Cookie("alert")
	if err == nil {
		f.Alert = cookie.Value
	}
	cookie, err = c.Cookie("notice")
	if err == nil {
		f.Notice = cookie.Value
	}
	return &f
}

func (fh *FlashHandler) clear(c echo.Context) {
	_, err := c.Cookie("alert")
	if err == nil {
		fh.setWithExpire(c, "alert", "", time.Now().AddDate(0, 0, 1))
	}
	_, err = c.Cookie("notice")
	if err == nil {
		fh.setWithExpire(c, "notice", "", time.Now().AddDate(0, 0, 1))
	}
}

func (fh *FlashHandler) with(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		f := fh.load(c)
		c.Set("flash", f)
		fh.clear(c)
		return impl(c)
	}
}
