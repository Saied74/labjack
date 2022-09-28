package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"path/filepath"
	"runtime/debug"
)

// <+++++++++++++++++++++++ Template Processing +++++++++++++++++++++++++++>

//This is straight out of Alex Edward's Let's Go book
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.html"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.html"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

//This is straight out of Alex Edward's Let's Go book
func (app *application) render(w http.ResponseWriter, r *http.Request,
	name string, u *U3) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist",
			name))
		return
	}
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, u)
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

//<++++++++++++++++   centralized error handling   +++++++++++++++++++>

//This is straight out of Alex Edward's Let's Go book
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace) //to not get the helper file...
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

//This is straight out of Alex Edward's Let's Go book
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

//This is straight out of Alex Edward's Let's Go book
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

//<++++++++++++++++   extracting option settings   ++++++++++++++++++++++++++++>

func (u *U3) pullAD(r url.Values) error {
	cio := []string{"cioAD0", "cioAD1", "cioAD2", "cioAD3"}
	for i := range cio {
		u.CIO[i].AD = "Digital"
	}

	eio := []string{"eioAD0", "eioAD1", "eioAD2", "eioAD3", "eioAD4", "eioAD5", "eioAD6", "eioAD7"}
	for i, c := range eio {
		val, ok := r[c]
		if ok {
			u.EIO[i].AD = "Digital"
			if val[0] == "2" {
				u.EIO[i].AD = "Analog"
			}
		}
	}

	fio := []string{"fioAD0", "fioAD1", "fioAD2", "fioAD3", "fioAD4", "fioAD5", "fioAD6", "fioAD7"}
	for i, c := range fio {
		val, ok := r[c]
		if ok {
			u.FIO[i].AD = "Digital"
			if val[0] == "2" {
				u.FIO[i].AD = "Analog"
			}
		}
	}
	return nil
}

func (u *U3) pullIO(r url.Values) error {
	cio := []string{"cioIO0", "cioIO1", "cioIO2", "cioIO3"}
	for i, c := range cio {
		val, ok := r[c]
		if ok {
			u.CIO[i].IO = "Input"
			if val[0] == "2" {
				u.CIO[i].IO = "Output"
			}
		}
	}

	eio := []string{"eioIO0", "eioIO1", "eioIO2", "eioIO3", "eioIO4", "eioIO5", "eioIO6", "eioIO7"}
	for i, c := range eio {
		val, ok := r[c]
		if ok {
			u.EIO[i].IO = "Input"
			if val[0] == "2" {
				u.EIO[i].IO = "Output"
			}
		}
	}

	fio := []string{"fioIO0", "fioIO1", "fioIO2", "fioIO3", "fioIO4", "fioIO5", "fioIO6", "fioIO7"}
	for i, c := range fio {
		val, ok := r[c]
		if ok {
			u.FIO[i].IO = "Input"
			if val[0] == "2" {
				u.FIO[i].IO = "Output"
			}
		}
	}
	return nil
}
