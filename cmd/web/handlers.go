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

func (app *application) getConfig(w http.ResponseWriter, r *http.Request) {
	app.u3.getU3Config()
	app.u3.getSetPins(true)
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
	// fmt.Println(r.PostForm)
	err = app.u3.pullIO(r.PostForm)
	if err != nil {
		fmt.Println("pullIO returned error", err)
	}
	app.render(w, r, "configure.page.html", app.u3)
}
