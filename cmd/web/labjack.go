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
	CONFIGU3_COMMAND_LENGTH      = 26 // Defines how long the command is
	CONFIGU3_RESPONSE_LENGTH     = 38 // Defines how long the response is
	BIT_CONFIGU3_COMMAND_LENGTH  = 12 // Defines how long the command is
	BIT_CONFIGU3_RESPONSE_LENGTH = 12 // Defines how long the response is
)

//Specifically not using this function with the Write Mask set to 1 because
//updating flash ages it.  I am also not using it for reading the IO Pin
//status because I want to have a function that can read and write the
//state of the IO pins without updating the flash.
func (u *U3) getU3Config() {
	sendBuffer := make([]byte, CONFIGU3_COMMAND_LENGTH)
	recBuffer := make([]byte, CONFIGU3_RESPONSE_LENGTH)

	devHandle := C.LJUSB_OpenDevice(1, 0, C.U3_PRODUCT_ID)
	if devHandle == nil {
		u.Message = fmt.Sprintf("Couldn't open U3. Please connect one and try again")
		return
	}

	buildConfigU3Bytes(sendBuffer)

	// Write the command to the device.
	// LJUSB_Write( handle, sendBuffer, length of sendBuffer )
	sBuff := (*C.uchar)(unsafe.Pointer(&sendBuffer[0]))
	x := C.ulong(26)

	r := C.LJUSB_Write(devHandle, sBuff, x) //CONFIGU3_COMMAND_LENGTH)
	sendBuffer = (*[CONFIGU3_COMMAND_LENGTH]byte)(unsafe.Pointer(sBuff))[:CONFIGU3_COMMAND_LENGTH:CONFIGU3_COMMAND_LENGTH]
	if r != CONFIGU3_COMMAND_LENGTH {
		u.Message = fmt.Sprintf("An error occurred when trying to write the buffer")
		// *Always* close the device when you error out.
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Read the result from the device.
	// LJUSB_Read( handle, recBuffer, number of bytes to read)
	rBuff := (*C.uchar)(unsafe.Pointer(&recBuffer[0]))
	r = C.LJUSB_Read(devHandle, rBuff, CONFIGU3_RESPONSE_LENGTH)
	recBuffer = (*[CONFIGU3_RESPONSE_LENGTH]byte)(unsafe.Pointer(rBuff))[:CONFIGU3_RESPONSE_LENGTH:CONFIGU3_RESPONSE_LENGTH]
	if r != CONFIGU3_RESPONSE_LENGTH {
		u.Message = fmt.Sprintf("An error occurred when trying to read from the U3")
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Check the command for errors
	if err := u.checkResponseForErrors(recBuffer); err != nil {
		C.LJUSB_CloseDevice(devHandle)
		u.Message = fmt.Sprintf("%v", err)
		return
	}

	// Parse the response into something useful
	u.parseConfigU3Bytes(recBuffer)

	//Close the device.
	C.LJUSB_CloseDevice(devHandle)
	u.Message = "No Message"

}

func buildConfigU3Bytes(sendBuffer []byte) {

	checksum := 0

	// Build up the bytes
	//sendBuffer[0] = Checksum8
	sendBuffer[1] = 0xF8
	sendBuffer[2] = 0x0A
	sendBuffer[3] = 0x08
	//sendBuffer[4] = Checksum16 (LSB)
	//sendBuffer[5] = Checksum16 (MSB)

	// We just want to read, so we set the WriteMask to zero, and zero out the
	// rest of the bytes.
	sendBuffer[6] = 0
	for i := 7; i < CONFIGU3_COMMAND_LENGTH; i++ {
		sendBuffer[i] = 0
	}

	// Calculate and set the checksum16
	checksum = calculateChecksum16(sendBuffer, CONFIGU3_COMMAND_LENGTH)
	sendBuffer[4] = byte(checksum & 0xff)
	sendBuffer[5] = byte((checksum / 256) & 0xff)

	// Calculate and set the checksum8
	sendBuffer[0] = calculateChecksum8(sendBuffer)

	// The bytes have been set, and the checksum calculated. We are ready to
	// write to the U3.
}

// Checks the response for any errors.
func (u *U3) checkResponseForErrors(recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		// If the packet is [ 0xB8, 0xB8 ], that's a bad checksum.
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
	} else {
		if recBuffer[1] != 0xF8 || recBuffer[2] != 0x10 || recBuffer[3] != 0x08 {
			// Make sure the command bytes match what we expect.
			return fmt.Errorf("Got the wrong command bytes back from the U3")
		}

		// Calculate the checksums.
		checksum16 := calculateChecksum16(recBuffer, CONFIGU3_RESPONSE_LENGTH)
		checksum8 := calculateChecksum8(recBuffer)

		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
			// Check the checksum
			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
		} else {
			if recBuffer[6] != 0 {
				// Check the error code in the packet. See section 5.3 of the U3
				// User's Guide for errorcode descriptions.
				return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
			}

			return nil

		}
	}
}

func (u *U3) getSetPins(set bool) {
	sendBuffer := make([]byte, BIT_CONFIGU3_COMMAND_LENGTH)
	recBuffer := make([]byte, BIT_CONFIGU3_RESPONSE_LENGTH)

	devHandle := C.LJUSB_OpenDevice(1, 0, C.U3_PRODUCT_ID)
	if devHandle == nil {
		u.Message = fmt.Sprintf("Couldn't open U3. Please connect one and try again")
		C.LJUSB_CloseDevice(devHandle)
		return
	}

	buildGetSetPinBytes(sendBuffer)

	// Write the command to the device.
	// LJUSB_Write( handle, sendBuffer, length of sendBuffer )
	sBuff := (*C.uchar)(unsafe.Pointer(&sendBuffer[0]))
	r := C.LJUSB_Write(devHandle, sBuff, BIT_CONFIGU3_COMMAND_LENGTH)
	sendBuffer = (*[BIT_CONFIGU3_COMMAND_LENGTH]byte)(unsafe.Pointer(sBuff))[:BIT_CONFIGU3_COMMAND_LENGTH:BIT_CONFIGU3_COMMAND_LENGTH]
	if r != BIT_CONFIGU3_COMMAND_LENGTH {
		u.Message = fmt.Sprintf("An error occurred when trying to write the buffer")
		// *Always* close the device when you error out.
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Read the result from the device
	rBuff := (*C.uchar)(unsafe.Pointer(&recBuffer[0]))
	r = C.LJUSB_Read(devHandle, rBuff, BIT_CONFIGU3_RESPONSE_LENGTH)
	recBuffer = (*[CONFIGU3_RESPONSE_LENGTH]byte)(unsafe.Pointer(rBuff))[:BIT_CONFIGU3_RESPONSE_LENGTH:BIT_CONFIGU3_RESPONSE_LENGTH]
	if r != BIT_CONFIGU3_RESPONSE_LENGTH {
		u.Message = fmt.Sprintf("An error occurred when trying to read from the U3")
		C.LJUSB_CloseDevice(devHandle)
		return
	}
	// Check the command for errors
	if err := u.checkResponseForBitErrors(recBuffer); err != nil {
		u.Message = fmt.Sprintf("%v", err)
		C.LJUSB_CloseDevice(devHandle)
		return
	}

	// Parse the response into something useful
	u.parseBitBytes(recBuffer)
	u.Message = "No Message"
}

func buildGetSetPinBytes(sendBuffer []byte) {

	checksum := 0

	// Build up the bytes
	//sendBuffer[0] = Checksum8
	sendBuffer[1] = 0xF8
	sendBuffer[2] = 0x03
	sendBuffer[3] = 0x0B
	//sendBuffer[4] = Checksum16 (LSB)
	//sendBuffer[5] = Checksum16 (MSB)

	// We just want to read, so we set the WriteMask to zero, and zero out the
	// rest of the bytes.
	sendBuffer[6] = 0
	for i := 7; i < BIT_CONFIGU3_COMMAND_LENGTH; i++ {
		sendBuffer[i] = 0
	}

	// Calculate and set the checksum16
	checksum = calculateChecksum16(sendBuffer, BIT_CONFIGU3_COMMAND_LENGTH)
	sendBuffer[4] = byte(checksum & 0xff)
	sendBuffer[5] = byte((checksum / 256) & 0xff)

	// Calculate and set the checksum8
	sendBuffer[0] = calculateChecksum8(sendBuffer)

	// The bytes have been set, and the checksum calculated. We are ready to
	// write to the U3.
}

func (u *U3) checkResponseForBitErrors(recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		// If the packet is [ 0xB8, 0xB8 ], that's a bad checksum.
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
	} else {
		if recBuffer[1] != 0xF8 || recBuffer[2] != 0x03 || recBuffer[3] != 0x0B {
			// Make sure the command bytes match what we expect.
			return fmt.Errorf("Got the wrong command bytes back from the U3")
		}

		// Calculate the checksums.
		checksum16 := calculateChecksum16(recBuffer, BIT_CONFIGU3_RESPONSE_LENGTH)
		checksum8 := calculateChecksum8(recBuffer)

		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
			// Check the checksum
			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
		} else {
			if recBuffer[6] != 0 {
				// Check the error code in the packet. See section 5.3 of the U3
				// User's Guide for errorcode descriptions.
				return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
			}

			return nil

		}
	}
}
