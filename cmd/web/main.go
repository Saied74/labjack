package main

/*
This is a demo program for accessing LabJack U3-HV IO Pins, setting the Analog
and digital options and on digital pins, setting the input or the output direction.
It also reads the digital (when it is an input) and analog (always and input )
pin states. It also sets the state of the digital pins when it is an output.
All Analog/Digital and for digital pins, all direction and state setting and
reading is done in mass using the port rather than the pin commands (since it
is a demo program).  Analog pins are read one at a time since that is the only
thing the command set allowes.

The design of this program is along the lines of Alex Edward's
Let's Go except since it is a single user local program, it
ignore the rules for a shared over the internet application
*/

import (
	//	"flag"

	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

/*
application structure is for injecting fixed data into the handler functions.
It also has an additional function for holding context (bad idea, but since
this is a single user demo application, we can get away with it.)  Element

u3 holds context for the state of the device.  It can be updated either from
the device flash memory using the "Flash Setting" link or from the device memory
using the Config U3 setting.

srData (short for send/recieve data) Is the collection of the models for the
device commands.  Additionally, it holds fields for forming and testing the sent
and recieved byte slices for each individual command.

u3 and srData fields are described in the jack.go file.
*/

//for injecting data into handlers
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	debugOption   bool
	templateCache map[string]*template.Template
	u3            *U3
	srData        u3srData
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
		u3:            newU3(),
		srData:        buildU3srData(),
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

/*
clicking on each link on a web page invokes the corresponding function
The functions are located in the handlers file.
*/
func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/home", app.home)
	mux.HandleFunc("/flash", app.flash)
	mux.HandleFunc("/getConfig", app.getConfig)
	mux.HandleFunc("/configure", app.configure)
	mux.HandleFunc("/measure", app.measure)
	mux.HandleFunc("/updateDigital", app.updateDigital)
	mux.HandleFunc("/adjustments", app.notImplemented)
	mux.HandleFunc("/readjust", app.notImplemented)
	return mux
}
