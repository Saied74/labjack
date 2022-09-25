package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

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
	name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist",
			name))
		return
	}
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, td)
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
