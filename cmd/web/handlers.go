package main

import (
	"fmt"
	"net/http"
)

type templateData struct {
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", nil)
}

func (app *application) notImplemented(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "notimplemented.page.html", nil)
}

func (app *application) flash(w http.ResponseWriter, r *http.Request) {
	app.u3SendRec(configJack, 0x00)
	app.render(w, r, "configure.page.html", app.u3)
}

func (app *application) getConfig(w http.ResponseWriter, r *http.Request) {
	app.u3SendRec(configIO, 0x00)
	app.u3SendRec(portDirRead, 0x00)
	app.render(w, r, "configure.page.html", app.u3)
}

func (app *application) configure(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.u3.pullAD(r.PostForm)
	if err != nil {
		fmt.Println("pullAD returned error", err)
	}
	err = app.u3.pullIO(r.PostForm)
	if err != nil {
		fmt.Println("pullIO returned error", err)
	}
	app.copyToWriteJack(configIO)
	writeMask := byte(0x0C)
	app.u3SendRec(configIO, writeMask)
	app.copyToWriteDirection(portDirWrite)
	writeMask = byte(0x01) //in this case, it does nothing.  Just to satisfy signature.
	app.u3SendRec(portDirWrite, writeMask)
	app.render(w, r, "configure.page.html", app.u3)
}
