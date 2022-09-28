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
	u3            *U3
}

type Pin struct {
	AD            string  //Analog or digital
	IO            string  //Input or Output
	AnalogRead    int     //A/D convertor raw read
	AnalogVoltage float64 //Analog read convergted to voltage
	DigitalRead   int     //only one and zero allowed
	DigitalWrite  int     //only one and zero allowed
}

type U3 struct {
	FIO               []*Pin
	EIO               []*Pin
	CIO               []*Pin
	FirmwareVersion   string
	BootLoaderVersion string
	HardwareVersion   string
	SerialNumber      string
	ProductID         string
	LocalID           string
	DeviceName        string
	Message           string
}

//This is a terrible practice, but I will do it instead of building a DB.
//Anything else like a cloture is fraught with all sorts of side effects.

// var u3 = newU3()

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
	mux.HandleFunc("/home", app.home)
	mux.HandleFunc("/getConfig", app.getConfig)
	mux.HandleFunc("/configure", app.configure)
	mux.HandleFunc("/measure", app.notImplemented)
	mux.HandleFunc("/adjustments", app.notImplemented)
	mux.HandleFunc("/readjust", app.notImplemented)
	return mux
}

func newPin() *Pin {
	return &Pin{}
}

func newU3() *U3 {
	u3 := U3{}
	for i := 0; i < 8; i++ {
		u3.EIO = append(u3.EIO, newPin())
		u3.FIO = append(u3.FIO, newPin())
		u3.CIO = append(u3.CIO, newPin())
	}
	u3.Message = "No Message"
	return &u3
}
