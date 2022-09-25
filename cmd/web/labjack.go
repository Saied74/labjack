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
	"os"
	"unsafe"
)

const (
	CONFIGU3_COMMAND_LENGTH  = 26 // Defines how long the command is
	CONFIGU3_RESPONSE_LENGTH = 38 // Defines how long the response is

)

func testConfig() {
	//<------------------------Basic learning try------------------------->
	// name := C.CString("Gopher")
	// defer C.free(unsafe.Pointer(name))
	// year := C.int(2018)
	// ptr := C.malloc(C.sizeof_char * 1024)
	// defer C.free(unsafe.Pointer(ptr))
	// size := C.greet(name, year, (*C.char)(ptr))
	// b := C.GoBytes(ptr, size)
	// fmt.Println("output: ", string(b))

	//<-----------------------End basic learning--------------------------->

	//<---------------------------Wrapper---------------------------------->
	sendBuffer := make([]byte, CONFIGU3_COMMAND_LENGTH)
	recBuffer := make([]byte, CONFIGU3_RESPONSE_LENGTH)

	devHandle := C.LJUSB_OpenDevice(1, 0, C.U3_PRODUCT_ID)
	if devHandle == nil {
		fmt.Printf("Couldn't open U3. Please connect one and try again.\n")
		os.Exit(-1)
	}

	buildConfigU3Bytes(sendBuffer)

	// Write the command to the device.
	// LJUSB_Write( handle, sendBuffer, length of sendBuffer )
	sBuff := (*C.uchar)(unsafe.Pointer(&sendBuffer[0]))

	r := C.LJUSB_Write(devHandle, sBuff, CONFIGU3_COMMAND_LENGTH)
	sendBuffer = (*[CONFIGU3_COMMAND_LENGTH]byte)(unsafe.Pointer(sBuff))[:CONFIGU3_COMMAND_LENGTH:CONFIGU3_COMMAND_LENGTH]
	if r != CONFIGU3_COMMAND_LENGTH {
		fmt.Printf("An error occurred when trying to write the buffer\n")
		// *Always* close the device when you error out.
		C.LJUSB_CloseDevice(devHandle)
		os.Exit(-1)
	}
	// Read the result from the device.
	// LJUSB_Read( handle, recBuffer, number of bytes to read)
	rBuff := (*C.uchar)(unsafe.Pointer(&recBuffer[0]))
	r = C.LJUSB_Read(devHandle, rBuff, CONFIGU3_RESPONSE_LENGTH)
	recBuffer = (*[CONFIGU3_RESPONSE_LENGTH]byte)(unsafe.Pointer(rBuff))[:CONFIGU3_RESPONSE_LENGTH:CONFIGU3_RESPONSE_LENGTH]
	if r != CONFIGU3_RESPONSE_LENGTH {
		fmt.Printf("An error occurred when trying to read from the U3\n")
		C.LJUSB_CloseDevice(devHandle)
		os.Exit(-1)
	}
	// fmt.Println(sendBuffer)
	// fmt.Println(recBuffer)
	// Check the command for errors
	if checkResponseForErrors(recBuffer) != nil {
		C.LJUSB_CloseDevice(devHandle)
		os.Exit(-1)
	}

	// Parse the response into something useful
	parseConfigU3Bytes(recBuffer)

	//Close the device.
	C.LJUSB_CloseDevice(devHandle)

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

// Calculates the checksum16
func calculateChecksum16(buffer []byte, len int) int {
	checksum := 0

	for i := 6; i < len; i++ {
		checksum += int(buffer[i])
	}

	return checksum
}

func calculateChecksum8(buffer []byte) byte {
	var temp int // For holding a value while we working.
	checksum := 0

	for i := 1; i < 6; i++ {
		checksum += int(buffer[i])
	}

	temp = checksum / 256
	checksum = (checksum - 256*temp) + temp
	temp = checksum / 256

	return byte((checksum - 256*temp) + temp)
}

// Checks the response for any errors.
func checkResponseForErrors(recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		// If the packet is [ 0xB8, 0xB8 ], that's a bad checksum.
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again.\n")
	} else {
		if recBuffer[1] != 0xF8 || recBuffer[2] != 0x10 || recBuffer[3] != 0x08 {
			// Make sure the command bytes match what we expect.
			return fmt.Errorf("Got the wrong command bytes back from the U3.\n")
		}

		// Calculate the checksums.
		checksum16 := calculateChecksum16(recBuffer, CONFIGU3_RESPONSE_LENGTH)
		checksum8 := calculateChecksum8(recBuffer)

		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
			// Check the checksum
			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d\n", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
		} else {
			if recBuffer[6] != 0 {
				// Check the error code in the packet. See section 5.3 of the U3
				// User's Guide for errorcode descriptions.
				return fmt.Errorf("Command returned with an errorcode = %d\n", recBuffer[6])
			}

			return nil

		}
	}
}

// Parses the ConfigU3 packet into something useful.
func parseConfigU3Bytes(recBuffer []byte) {
	fmt.Println("Parse Config called", recBuffer)
	fmt.Printf("Results of ConfigU3:\n")
	fmt.Printf("  FirmwareVersion = %d.%02d\n", int(recBuffer[10]), int(recBuffer[9]))
	fmt.Printf("  BootloaderVersion = %d.%02d\n", recBuffer[12], recBuffer[11])
	fmt.Printf("  HardwareVersion = %d.%02d\n", recBuffer[14], recBuffer[13])
	fmt.Printf("  SerialNumber = %d\n", makeInt(recBuffer, 15))
	fmt.Printf("  ProductID = %d\n", makeShort(recBuffer, 19))
	fmt.Printf("  LocalID = %d\n", recBuffer[21])
	fmt.Printf("  TimerCounterMask = %d\n", recBuffer[22])
	fmt.Printf("  FIOAnalog = %d\n", recBuffer[23])
	fmt.Printf("  FIODireciton = %d\n", recBuffer[24])
	fmt.Printf("  FIOState = %d\n", recBuffer[25])
	fmt.Printf("  EIOAnalog = %d\n", recBuffer[26])
	fmt.Printf("  EIODirection = %d\n", recBuffer[27])
	fmt.Printf("  EIOState = %d\n", recBuffer[28])
	fmt.Printf("  CIODirection = %d\n", recBuffer[29])
	fmt.Printf("  CIOState = %d\n", recBuffer[30])
	fmt.Printf("  DAC1Enable = %d\n", recBuffer[31])
	fmt.Printf("  DAC0 = %d\n", recBuffer[32])
	fmt.Printf("  DAC1 = %d\n", recBuffer[33])
	fmt.Printf("  TimerClockConfig = %d\n", recBuffer[34])
	fmt.Printf("  TimerClockDivisor = %d\n", recBuffer[35])
	fmt.Printf("  CompatibilityOptions = %d\n", recBuffer[36])
	fmt.Printf("  VersionInfo = %d\n", recBuffer[37])

	buf37 := int(recBuffer[37])
	switch buf37 {
	case 0:
		fmt.Printf("  DeviceName = U3A\n")
	case 1:
		fmt.Printf("  DeviceName = U3B\n")
	case 2:
		fmt.Printf("  DeviceName = U3-LV\n")
	case 18:
		fmt.Printf("  DeviceName = U3-HV\n")
	default:
		fmt.Printf("  None of the above\n")
	}

}

// Takes a buffer and an offset, and turns into an 32-bit integer
func makeInt(buffer []byte, offset int) int {
	return int((buffer[offset+3] << 24) + (buffer[offset+2] << 16) + (buffer[offset+1] << 8) + buffer[offset])
}

// Takes a buffer and an offset, and turns into an 16-bit integer
func makeShort(buffer []byte, offset int) int {
	return int((buffer[offset+1] << 8) + buffer[offset])
}
