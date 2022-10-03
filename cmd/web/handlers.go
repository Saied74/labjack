package main

import (
	"fmt"
	"net/http"
)

type templateData struct {
}

//home page contains very basic documentation.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", nil)
}

//target for reatures not implemented yet
func (app *application) notImplemented(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "notimplemented.page.html", nil)
}

//reads results from the device flash and displays them on the configuraton page.
//It needs to be invked at the start of operation.
func (app *application) flash(w http.ResponseWriter, r *http.Request) {
	//u3SendRec is the generic function for accessing all U3 commands.
	//the command name is passed on to the functon to choose the command.
	//configJack reads all data from the device flash memory
	//writeMask is set to zero to avoid aging the flash memory
	app.u3SendRec(configJack, 0x00)
	app.render(w, r, "configure.page.html", app.u3)
}

//reads the results from the device voltaile memory
func (app *application) getConfig(w http.ResponseWriter, r *http.Request) {
	app.u3SendRec(configIO, 0x00)    //reads the Anolog, Digital setting.
	app.u3SendRec(portDirRead, 0x00) //Reads the Input/Output setting for digital pins.
	app.render(w, r, "configure.page.html", app.u3)
}

//Writes the Analog/Digital and Inupt/Output (for digital pins) in device
//volatile memory and populates app.u3
func (app *application) configure(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	//pulls the Analog/Digital settings from the web form and populates app.u3
	err = app.u3.pullAD(r.PostForm)
	if err != nil {
		fmt.Println("pullAD returned error", err)
	}
	//pulls the Input/Output settings from the web form and populates app.u3
	//and app.srData.  Note that app.u3 is fit for the web page and app.srData
	//is fit for the U3 device itself.
	err = app.u3.pullIO(r.PostForm)
	if err != nil {
		fmt.Println("pullIO returned error", err)
	}
	//copy Analog/Digital setting from app.u3 to app.srData
	app.copyToWriteJack(configIO)
	writeMask := byte(0x0C)
	//write Analog/Digital setting to the device volatile memory
	app.u3SendRec(configIO, writeMask)
	//copy Input/Output setting for digital pins from app.u3 to app.srData
	app.copyToWriteDirection(portDirWrite)
	writeMask = byte(0x01) //just to satisfy function signature
	//write the Input/Output setting for digital pins to the device volatile memory.
	app.u3SendRec(portDirWrite, writeMask)
	app.render(w, r, "configure.page.html", app.u3)
}

func (app *application) measure(w http.ResponseWriter, r *http.Request) {

	app.u3SendRec(portStateRead, 0x00)
	for i, pin := range app.u3.FIO {
		app.srData[ain].byte8 = 0x00
		if pin.AD == "Analog" {
			app.srData[ain].byte8 = byte(i) | 0x40 //for long settling
			app.u3SendRec(ain, 0x00)
		}
	}
	for i, pin := range app.u3.EIO {
		app.srData[ain].byte8 = 0x00
		if pin.AD == "Analog" {
			app.srData[ain].byte8 = byte(i+8) | 0x40 //for long settling
			app.u3SendRec(ain, 0x00)
		}
	}
	app.render(w, r, "measure.page.html", app.u3)
}

func (app *application) updateDigital(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	fmt.Println("postform", r.PostForm)
	//pulls the digitalWrite settings from the web form and populates app.u3
	err = app.u3.pullDigitalOutput(r.PostForm)
	if err != nil {
		fmt.Println("pullDigitalOutput returned error", err)
	}
	app.copyToWirteDigitalOutput(portStateWrite)
	writeMask := byte(0x01)
	app.u3SendRec(portStateWrite, writeMask)
	app.render(w, r, "measure.page.html", app.u3)
}
