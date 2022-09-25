package main

import (
	//	"flag"

	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

//The design of this program is along the lines of Alex Edward's
//Let's Go except since it is a single user local program, it
//ignore the rules for a shared over the internet application

//for injecting data into handlers
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	debugOption   bool
	templateCache map[string]*template.Template
}

func main() {
	var err error

	optionDebug := flag.Bool("d", false, "true turns on debug option")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.LUTC)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile)

	//note, this requires the run command be issues from the project base
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		debugOption:   *optionDebug,
		templateCache: templateCache,
	}

	mux := app.routes()
	srv := &http.Server{
		Addr:     ":4000",
		ErrorLog: errorLog,
		Handler:  mux,
	}
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	infoLog.Printf("starting server on :4000")
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/u3config", app.notImplemented)
	mux.HandleFunc("/configure", app.notImplemented)
	mux.HandleFunc("/measure", app.notImplemented)
	mux.HandleFunc("/adjustments", app.notImplemented)
	mux.HandleFunc("/readjust", app.notImplemented)
	return mux
}
