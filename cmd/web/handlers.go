package main

import (
	"net/http"
)

type templateData struct {
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	//	testConfig()
	app.render(w, r, "home.page.html", nil)
}

func (app *application) notImplemented(w http.ResponseWriter, r *http.Request) {
	//	testConfig()
	app.render(w, r, "notimplemented.page.html", nil)
}
