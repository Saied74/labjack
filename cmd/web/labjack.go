package main

// #cgo CFLAGS: -g -Wall
// #cgo amd64 386 CFLAGS: -DX86=1
// #cgo LDFLAGS: -llabjackusb
// #include <stdlib.h>
//#include <stdio.h>
//#include <errno.h>
//#include <../../pkg/labjackusb/labjackusb.h>
//#include <../../pkg/libusb/libusb.h>
// #include "greeter.h"
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	eight       = 8
	nine        = 9
	ten         = 10
	eleven      = 11
	twelve      = 12
	fourteen    = 14
	twentysix   = 26
	thirtyeight = 38
)

//This is a generic function for writing to the Labjack U3 and getting
//the results back
func (app *application) u3SendRec(op string, mask byte) {
	sendBuffer := make([]byte, app.srData[op].sendLength)
	recBuffer := make([]byte, app.srData[op].recLength)
	devHandle := C.LJUSB_OpenDevice(1, 0, C.U3_PRODUCT_ID)
	if devHandle == nil {
		app.u3.Message = fmt.Sprintf("Couldn't open U3. Please connect one and try again %v", devHandle)
		fmt.Println("1: ", app.u3.Message)
		return
	}

	app.srData[op].buildBytes(app.srData[op], sendBuffer, mask)

	// Write the command to the device.
	// LJUSB_Write( handle, sendBuffer, length of sendBuffer )
	sBuff := (*C.uchar)(unsafe.Pointer(&sendBuffer[0]))
	sBuffLength := C.ulong(app.srData[op].sendLength)

	r := C.LJUSB_Write(devHandle, sBuff, sBuffLength) //CONFIGU3_COMMAND_LENGTH)
	switch op {
	case configJack:
		sendBuffer = (*[C.ulong(twentysix)]byte)(unsafe.Pointer(sBuff))[:C.ulong(twentysix):C.ulong(twentysix)]
	case configIO:
		sendBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(sBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case ain:
		sendBuffer = (*[C.ulong(ten)]byte)(unsafe.Pointer(sBuff))[:C.ulong(ten):C.ulong(ten)]
	case led:
		sendBuffer = (*[C.ulong(nine)]byte)(unsafe.Pointer(sBuff))[:C.ulong(nine):C.ulong(nine)]
	case portStateRead:
		sendBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(sBuff))[:C.ulong(eight):C.ulong(eight)]
	case portStateWrite:
		sendBuffer = (*[C.ulong(fourteen)]byte)(unsafe.Pointer(sBuff))[:C.ulong(fourteen):C.ulong(fourteen)]
	case portDirRead:
		sendBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(sBuff))[:C.ulong(eight):C.ulong(eight)]
	case portDirWrite:
		sendBuffer = (*[C.ulong(fourteen)]byte)(unsafe.Pointer(sBuff))[:C.ulong(fourteen):C.ulong(fourteen)]
	case tempSense:
		sendBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(sBuff))[:C.ulong(eight):C.ulong(eight)]
	case vReg:
		sendBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(sBuff))[:C.ulong(eight):C.ulong(eight)]
	}

	// sendBuffer = (*[sBuffLength]byte)(unsafe.Pointer(sBuff))[:sBuffLength:sBuffLength]
	if r != sBuffLength {
		app.u3.Message = fmt.Sprintf("An error occurred when trying to write the buffer")
		fmt.Println("2: ", app.u3.Message)
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Read the result from the device.
	// LJUSB_Read( handle, recBuffer, number of bytes to read)
	rBuff := (*C.uchar)(unsafe.Pointer(&recBuffer[0]))
	rBuffLength := C.ulong(app.srData[op].recLength)
	r = C.LJUSB_Read(devHandle, rBuff, rBuffLength)

	switch op {
	case configJack:
		recBuffer = (*[C.ulong(thirtyeight)]byte)(unsafe.Pointer(rBuff))[:C.ulong(thirtyeight):C.ulong(thirtyeight)]
	case configIO:
		recBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(rBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case ain:
		recBuffer = (*[C.ulong(eleven)]byte)(unsafe.Pointer(rBuff))[:C.ulong(eleven):C.ulong(eleven)]
	case led:
		recBuffer = (*[C.ulong(nine)]byte)(unsafe.Pointer(rBuff))[:C.ulong(nine):C.ulong(nine)]
	case portStateRead:
		recBuffer = (*[C.ulong(ten)]byte)(unsafe.Pointer(rBuff))[:C.ulong(ten):C.ulong(ten)]
	case portStateWrite:
		recBuffer = (*[C.ulong(nine)]byte)(unsafe.Pointer(rBuff))[:C.ulong(nine):C.ulong(nine)]
	case portDirRead:
		recBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(rBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case portDirWrite:
		recBuffer = (*[C.ulong(fourteen)]byte)(unsafe.Pointer(rBuff))[:C.ulong(fourteen):C.ulong(fourteen)]
	case tempSense:
		recBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(rBuff))[:C.ulong(eight):C.ulong(eight)]
	case vReg:
		recBuffer = (*[C.ulong(eight)]byte)(unsafe.Pointer(rBuff))[:C.ulong(eight):C.ulong(eight)]
	}

	// recBuffer = (*[rBuffLength]byte)(unsafe.Pointer(rBuff))[:rBuffLength:rBuffLength]
	if r != rBuffLength {
		app.u3.Message = fmt.Sprintf("An error occurred when trying to read from the U3 r: %v, rBuffLength: %v", r, rBuffLength)
		fmt.Println("3: ", app.u3.Message, recBuffer)
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Check the command for errors
	if err := app.srData[op].checkReturn(app.srData[op], recBuffer); err != nil {
		// if err := app.checkResponse(op, recBuffer); err != nil {
		C.LJUSB_CloseDevice(devHandle)
		app.u3.Message = fmt.Sprintf("%v", err)
		fmt.Println("4: ", app.u3.Message)
		return
	}
	switch op {
	case configJack:
		app.u3.parseConfigU3Bytes(recBuffer)
	case configIO:
		app.u3.parseBitBytes(recBuffer)
	case portDirRead:
		app.u3.parseDirBits(recBuffer)
	}
	//Close the device.
	C.LJUSB_CloseDevice(devHandle)
	app.u3.Message = "No Message"

}
