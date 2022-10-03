package main

// #cgo CFLAGS: -g -Wall
// #cgo amd64 386 CFLAGS: -DX86=1
// #cgo LDFLAGS: -llabjackusb
// #include <stdlib.h>
//#include <stdio.h>
//#include <errno.h>
//#include <../../pkg/labjackusb/labjackusb.h>
//#include <../../pkg/libusb/libusb.h>
import "C"
import (
	"fmt"
	"unsafe"
)

//All C dependencies are confined to this file.

/*
These constants are defined for constructing buffers in the C envirnment.
C does not deal with variables for array construction.
*/
const (
	eight       = 8
	nine        = 9
	ten         = 10
	eleven      = 11
	twelve      = 12
	fourteen    = 14
	fifteen = 15
	twentysix   = 26
	thirtyeight = 38
)

//This is a generic function for writing to the Labjack U3 and getting
//the results back
func (app *application) u3SendRec(op string, mask byte) {
	sendBuffer := make([]byte, app.srData[op].sendLength)
	recBuffer := make([]byte, app.srData[op].recLength)
	//see labjackusb.h for documentation.
	devHandle := C.LJUSB_OpenDevice(1, 0, C.U3_PRODUCT_ID)
	if devHandle == nil {
		app.u3.Message = fmt.Sprintf("Couldn't open U3. Please connect one and try again %v", devHandle)
		fmt.Println("1: ", app.u3.Message)
		return
	}

	app.srData[op].buildBytes(app.srData[op], sendBuffer, mask)

	// Write the command to the device.
	// LJUSB_Write( handle, sendBuffer, length of sendBuffer )

	//pointer to the first byte of the sendBuffer, the way that C likes it.
	sBuff := (*C.uchar)(unsafe.Pointer(&sendBuffer[0]))
	//cast go int to C unsighed long
	sBuffLength := C.ulong(app.srData[op].sendLength)
	//write to the device.
	r := C.LJUSB_Write(devHandle, sBuff, sBuffLength) //CONFIGU3_COMMAND_LENGTH)

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
		recBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(rBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case led:
		recBuffer = (*[C.ulong(nine)]byte)(unsafe.Pointer(rBuff))[:C.ulong(nine):C.ulong(nine)]
	case portStateRead:
		recBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(rBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case portStateWrite:
		recBuffer = (*[C.ulong(ten)]byte)(unsafe.Pointer(rBuff))[:C.ulong(ten):C.ulong(ten)]
	case portDirRead:
		recBuffer = (*[C.ulong(twelve)]byte)(unsafe.Pointer(rBuff))[:C.ulong(twelve):C.ulong(twelve)]
	case portDirWrite:
		recBuffer = (*[C.ulong(ten)]byte)(unsafe.Pointer(rBuff))[:C.ulong(ten):C.ulong(ten)]
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
	fmt.Printf("Send Buffer (op: %s): %v\n", op, sendBuffer)
	fmt.Printf("Rec Buffer (op: %s): %v\n", op, recBuffer)
	// Check the command for errors
	if err := app.srData[op].checkReturn(app.srData[op], recBuffer); err != nil {
		// if err := app.checkResponse(op, recBuffer); err != nil {
		C.LJUSB_CloseDevice(devHandle)
		app.u3.Message = fmt.Sprintf("%v", err)
		fmt.Println("4: ", app.u3.Message)
		return
	}
	/*
		Parsing the return bytes and putting the results into the U3 structure were
		build as methods on U3.  That limits their utility in being called from
		inside this function.  For now, the switch function will do.  Later, I will
		refactor them and pass a pointer to the instance of U3 as the first Parameter
		(just like invoking the method on the structure does)
	*/

	switch op {
	case configJack:
		app.u3.parseConfigU3Bytes(recBuffer)
	case configIO:
		app.u3.parseBitBytes(recBuffer)
	case portDirRead:
		app.u3.parseDirBits(recBuffer)
	case portStateRead:
		app.u3.parseStateBits(recBuffer)
	case ain:
		app.u3.parseAINBits(sendBuffer[8], recBuffer)
	}
	//Close the device.
	C.LJUSB_CloseDevice(devHandle)
	app.u3.Message = "No Message"

}
