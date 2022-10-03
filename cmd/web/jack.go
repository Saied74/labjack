package main

import "fmt"

//Jack file is a set of LabJack helper frunctions.

const (
	configJack     = "Config U3"
	configIO       = "Config IO"
	ain            = "AIN"
	led            = "LED"
	portStateRead  = "Port State Read"
	portStateWrite = "Port State Write"
	portDirRead    = "Port Direction Read"
	portDirWrite   = "Port Direction Write"
	tempSense      = "Temperature Sense"
	vReg           = "VReg"
)

/*
Pin is the model for each of the U3 device pins.  It can either be loaded
from the flash device memory or from the voltaile device memory.  It is not
instanciated directly.  It is a component of the U3 type.
*/
type Pin struct {
	AD            string //Analog or digital
	IO            string //Input or Output
	AnalogRead    uint16 //A/D convertor raw read
	AnalogVoltage string //Analog read convergted to voltage
	DigitalRead   int    //only one and zero allowed
	DigitalWrite  int    //only one and zero allowed
}

/*
U3 is a model of the U3 device.  The FIO, EIO, and CIO fields are described
on the home page and in the device documentation under low level function
reference.  The rest of the fields should be self explanatory from their names.
They are updated from the device flash memory.

U3 is also the template data that will processed by the template parser to
build the web pages.

This is a violation of the MVC design pattern of sorts, but one can argue That
the model is really the U3 device memory (sort of)
*/
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
	open              bool
}

/*
functions newPin and newU3 are constructed in the hope that refrence to the
results will not cause nil pointer refrence panic.
*/

//Builds a blank instance of the Pin type.
func newPin() *Pin {
	return &Pin{}
}

//builds a blank instance of the U3 type.
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

/*
u3srData type is the model for each individual U3 command.  The send and recieved
lengths for each command are different.  Also, the meaning of each byte is different
for each commmand and commmand type.  For this reason, they are documented inline
in the buildU3SRData function for each command.  The names of the commands are
the same as the names of the commands in the documentation.
*/
type u3srElement struct {
	sendLength  int
	recLength   int
	byte1       byte
	byte2       byte
	byte3       byte
	byte6       byte
	byte7       byte
	byte8       byte
	byte9       byte
	byte10      byte
	byte11      byte
	byte12      byte
	byte13      byte
	checkReturn func(*u3srElement, []byte) error
	buildBytes  func(*u3srElement, []byte, byte)
}

//u3srData type is the collection of all the commands available for the U3 device
type u3srData map[string]*u3srElement

/*

 */
func buildU3srData() u3srData {
	return u3srData{
		configJack: &u3srElement{ //ConfigU3 changed to configJack (older naming conflict)
			sendLength:  26,                  //to make the sendBuffer in u3SendRec function.
			recLength:   38,                  //to make the recBuffer in the u3SendRec function
			byte1:       0xF8,                //per device low level function reference
			byte2:       0x0A,                //per device low level function refrence
			byte3:       0x08,                //per device low level function refrence
			byte6:       0x00,                //per device low level function refrence
			byte7:       0x00,                //per device low level function refrence
			checkReturn: checkJack,           //call to check the validity of the returned bytes
			buildBytes:  buildJackSendBuffer, //call to form the send bytes.
		},
		//Analog or digital nature of pins is set with this command, it does not impact flash
		configIO: &u3srElement{ //all fields the same as configJack fields.
			sendLength:  12,
			recLength:   12,
			byte1:       0xF8,
			byte2:       0x03,
			byte3:       0x0B,
			byte6:       0x00,
			byte7:       0x00,
			checkReturn: checkIO,
			buildBytes:  buildJackSendBuffer,
		},
		//the following are all subcommands of the "feedback" command.
		ain: &u3srElement{ //read analog pin voltage
			sendLength:  10,
			recLength:   12,
			byte1:       0xF8, //per device low level function refrence
			byte2:       2,    //number of words (two byte pairs) startying with byte 6
			byte3:       0x00,
			byte6:       0,
			byte7:       1, //feedback subcommand
			checkReturn: checkFeedback,
			buildBytes:  buildAINReadBuffer,
		},
		led: &u3srElement{ //set led state (on or off)
			sendLength: 9,
			recLength:  9,
			byte1:      0xF8,
			byte2:      2, //number of words (two byte pairs) startying with byte 6
			byte3:      0x00,
			byte6:      0,
			byte7:      9, //feedback subcommand
		},
		portStateRead: &u3srElement{ //read the state of digital input pins, high or low
			sendLength:  8,
			recLength:   12,
			byte1:       0xF8,
			byte2:       1, //number of words (two byte pairs) startying with byte 6
			byte3:       0x00,
			byte6:       0,
			byte7:       26, //feedback subcommand
			checkReturn: checkFeedback,
			buildBytes:  buildPortStateReadBuffer,
		},
		portStateWrite: &u3srElement{ //write the state of digital output pins (high or low)
			sendLength:  14,
			recLength:   10,
			byte1:       0xF8,
			byte2:       4, //number of words (two byte pairs) startying with byte 6
			byte3:       0x00,
			byte6:       0,
			byte7:       27, //feedback subcommand
			checkReturn: checkFeedback,
			buildBytes:  buildPortStateWriteBuffer,
		},
		portDirRead: &u3srElement{ //read the direction of the digital pins, input or output
			sendLength:  8,
			recLength:   12,
			byte1:       0xF8,
			byte2:       1, //number of words (two byte pairs) startying with byte 6
			byte3:       0x00,
			byte6:       0,
			byte7:       28, //feedback subcommand
			checkReturn: checkFeedback,
			buildBytes:  buildPortDirReadBuffer,
		},
		portDirWrite: &u3srElement{ //write the direction of digital pins, input or output
			sendLength:  14,
			recLength:   10,
			byte1:       0xF8,
			byte2:       4, //number of words (two byte pairs) startying with byte 6
			byte3:       0x00,
			byte6:       0,
			byte7:       29, //feedback subcommand
			checkReturn: checkFeedback,
			buildBytes:  buildPortDirWriteBuffer,
		},
		tempSense: &u3srElement{ // read temperature
			sendLength: 8,
			recLength:  11,
			byte1:      0xF8,
			byte2:      1, //number of words (two byte pairs) startying with byte 6
			byte3:      0x00,
			byte6:      0,
			byte7:      30, //feedback subcommand
		},
		vReg: &u3srElement{ //I don't know what this reads, we will find out.
			sendLength: 8,
			recLength:  11,
			byte1:      0xF8,
			byte2:      1, //number of words (two byte pairs) startying with byte 6
			byte3:      0x00,
			byte6:      0,
			byte7:      31, //feedback subcommand
		},
	}
}

//<+++++++++++++++++++ Check Methods for Commands to the Device +++++++++++++++>

func checkJack(sr *u3srElement, recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
	} else {
		if recBuffer[1] != 0xF8 || recBuffer[2] != 0x10 || recBuffer[3] != 0x08 {
			// Make sure the command bytes match what we expect.
			return fmt.Errorf("Got the wrong command bytes back from the U3")
		}

		checksum16 := calculateChecksum16(recBuffer, sr.recLength)
		checksum8 := calculateChecksum8(recBuffer)
		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
		} else {
			if recBuffer[6] != 0 { // Check the error code in the packet. See section 5.3 of the U3
				return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
			}
			return nil
		}
	}
}

func checkIO(sr *u3srElement, recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
	} else {
		if recBuffer[1] != sr.byte1 || recBuffer[2] != sr.byte2 || recBuffer[3] != sr.byte3 {
			return fmt.Errorf("Got the wrong command bytes back from the U3")
		}

		checksum16 := calculateChecksum16(recBuffer, sr.recLength)
		checksum8 := calculateChecksum8(recBuffer)
		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
		} else {
			if recBuffer[6] != 0 { // Check the error code in the packet. See section 5.3 of the U3
				return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
			}
			return nil
		}
	}
}

func checkFeedback(sr *u3srElement, recBuffer []byte) error {
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
	}
	if recBuffer[1] != sr.byte1 {
		return fmt.Errorf("Got the wrong command bytes back from the U3")
	}
	checksum16 := calculateChecksum16(recBuffer, sr.recLength)
	checksum8 := calculateChecksum8(recBuffer)
	if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
		return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
	}
	if recBuffer[6] != 0 { // Check the error code in the packet. See section 5.3 of the U3
		return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
	}
	return nil
}

//<+++++++++++++ Build Methods for sendBuffer for the commands ++++++++++++++++>

//builds the configU3 command send buffer templated after the vendor C example
func buildJackSendBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	sendBuffer[6] = writeMask
	for i := 7; i < sr.sendLength; i++ {
		sendBuffer[i] = 0
	}
	sendBuffer[10] = sr.byte10
	sendBuffer[11] = sr.byte11
	addChecksum(sr, sendBuffer)
}

//templated after the configU3 buffer build provided by the vendor (in C)
func buildPortDirReadBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	sendBuffer[6] = writeMask
	sendBuffer[7] = sr.byte7
	addChecksum(sr, sendBuffer)
}

func buildPortStateReadBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	sendBuffer[6] = writeMask
	sendBuffer[7] = sr.byte7
	addChecksum(sr, sendBuffer)
}

func buildAINReadBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	sendBuffer[6] = writeMask
	sendBuffer[7] = sr.byte7
	sendBuffer[8] = sr.byte8
	sendBuffer[9] = 0x00
	addChecksum(sr, sendBuffer)
}

/*
The feedback functions all have a different send and recieve buffer templetaes.
Hence all send buffer builds will be different.
*/
func buildPortDirWriteBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	for i := 8; i < sr.sendLength; i++ {
		sendBuffer[i] = 0x00
	}
	sendBuffer[8] = 0xff
	sendBuffer[9] = 0xff
	sendBuffer[10] = 0xff
	sendBuffer[11] = sr.byte11
	sendBuffer[12] = sr.byte12
	sendBuffer[13] = sr.byte13
	addChecksum(sr, sendBuffer)
}

func buildPortStateWriteBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	for i := 8; i < sr.sendLength; i++ {
		sendBuffer[i] = 0x00
	}
	sendBuffer[8] = sr.byte8
	sendBuffer[9] = sr.byte9
	sendBuffer[10] = sr.byte10
	sendBuffer[11] = sr.byte11
	sendBuffer[12] = sr.byte12
	sendBuffer[13] = sr.byte13
	addChecksum(sr, sendBuffer)
}

//<++++++++++++++++++++++++  Helper Functions ++++++++++++++++++++++++++++++++>
//helper function for building sendBuffer
func copyHead(sr *u3srElement, sendBuffer []byte) {
	sendBuffer[1] = sr.byte1
	sendBuffer[2] = sr.byte2
	sendBuffer[3] = sr.byte3
	sendBuffer[6] = sr.byte6
	sendBuffer[7] = sr.byte7
}

//helper function for building sendbuffer
func addChecksum(sr *u3srElement, sendBuffer []byte) {
	checksum := 0
	checksum = calculateChecksum16(sendBuffer, sr.sendLength)
	sendBuffer[4] = byte(checksum & 0xff)
	sendBuffer[5] = byte((checksum / 256) & 0xff)
	sendBuffer[0] = calculateChecksum8(sendBuffer)
}

//helper function for building send and checking recieve buffers.
func calculateChecksum16(buffer []byte, len int) int {
	checksum := 0
	for i := 6; i < len; i++ {
		checksum += int(buffer[i])
	}

	return checksum
}

//helper function for building the send and checking the recive buffers.
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

// Takes a buffer and an offset, and turns into an 32-bit integer
func makeInt(buffer []byte, offset int) int {
	return int((buffer[offset+3] << 24) + (buffer[offset+2] << 16) + (buffer[offset+1] << 8) + buffer[offset])
}

// Takes a buffer and an offset, and turns into an 16-bit integer
func makeShort(buffer []byte, offset int) int {
	return int((buffer[offset+1] << 8) + buffer[offset])
}

//<++++++++  Functions for mapping the recieve buffer to app.u3 +++++++++++++++>

// Parses the ConfigU3 recBuffer and put them into app.u3.
func (u *U3) parseConfigU3Bytes(recBuffer []byte) {

	u.FirmwareVersion = fmt.Sprintf("%d.%02d", int(recBuffer[10]), int(recBuffer[9]))
	u.BootLoaderVersion = fmt.Sprintf("%d.%02d", recBuffer[12], recBuffer[11])
	u.HardwareVersion = fmt.Sprintf("%d.%02d", recBuffer[14], recBuffer[13])
	u.SerialNumber = fmt.Sprintf("%d", makeInt(recBuffer, 15))
	u.ProductID = fmt.Sprintf("%d", makeShort(recBuffer, 19))
	u.LocalID = fmt.Sprintf("%d", recBuffer[21])

	u.parseFlashBytes(recBuffer)

	// fmt.Printf("  TimerCounterMask = %d\n", recBuffer[22])
	// fmt.Printf("  DAC1Enable = %d\n", recBuffer[31])
	// fmt.Printf("  DAC0 = %d\n", recBuffer[32])
	// fmt.Printf("  DAC1 = %d\n", recBuffer[33])
	// fmt.Printf("  TimerClockConfig = %d\n", recBuffer[34])
	// fmt.Printf("  TimerClockDivisor = %d\n", recBuffer[35])
	// fmt.Printf("  CompatibilityOptions = %d\n", recBuffer[36])
	// fmt.Printf("  VersionInfo = %d\n", recBuffer[37])

	buf37 := int(recBuffer[37])
	switch buf37 {
	case 0:
		u.DeviceName = fmt.Sprintf("U3A")
	case 1:
		u.DeviceName = fmt.Sprintf("U3B")
	case 2:
		u.DeviceName = fmt.Sprintf("U3-LV")
	case 18:
		u.DeviceName = fmt.Sprintf("U3-HV")
	default:
		u.DeviceName = fmt.Sprintf("Not recognized")
	}

}

//parse the configIO recieve buffer and map into app.u3
func (u *U3) parseBitBytes(recBuffer []byte) {

	for i := 0; i < 8; i++ {
		u.FIO[i].AD = "Digital"
		if recBuffer[10]&(1<<i) != 0 {
			u.FIO[i].AD = "Analog"
		}
		u.EIO[i].AD = "Digital"
		if recBuffer[11]&(1<<i) != 0 {
			u.EIO[i].AD = "Analog"
		}
		if i < 4 {
			u.CIO[i].AD = "Digital"
		}
	}
}

//parse the portDirRead recBuffer and map into app.u3
func (u *U3) parseDirBits(recBuffer []byte) {
	for i := 0; i < 8; i++ {
		if i > 3 {
			u.FIO[i].IO = "Input"
			if recBuffer[9]&(1<<i) != 0 {
				u.FIO[i].IO = "Output"
			}
		}
		u.EIO[i].IO = "Input"
		if recBuffer[10]&(1<<i) != 0 {
			u.EIO[i].IO = "Output"
		}
		if i < 4 {
			u.CIO[i].IO = "Input"
			if recBuffer[11]&(1<<i) != 0 {
				u.CIO[i].IO = "Output"
			}
		}
	}
}

func (u *U3) parseStateBits(recBuffer []byte) {
	for i := 0; i < 8; i++ {
		if i > 3 {
			u.FIO[i].DigitalRead = 0
			if recBuffer[9]&(1<<i) != 0 {
				u.FIO[i].DigitalRead = 1
			}
		}
		u.EIO[i].DigitalRead = 0
		if recBuffer[10]&(1<<i) != 0 {
			u.EIO[i].DigitalRead = 1
		}
		if i < 4 {
			u.CIO[i].DigitalRead = 0
			if recBuffer[11]&(1<<i) != 0 {
				u.CIO[i].DigitalRead = 1
			}
		}
	}
}

const (
	max      = 65535.0
	hvSlope  = 10.3 / max
	slope    = 2.44 / max
	hvOffset = -5.0
	offset   = -0.527
)

func (u *U3) parseAINBits(b byte, recBuffer []byte) {
	ch := int(b & 0x1F)
	fmt.Println("Channel: ", ch)
	if ch < 8 {
		if u.FIO[ch].AD == "Analog" {
			read := uint16(recBuffer[9]) + uint16(recBuffer[10])*256
			u.FIO[ch].AnalogRead = read
			if ch < 4 {
				u.FIO[ch].AnalogVoltage = fmt.Sprintf("%0.3f",
					(float64(read)*hvSlope+hvOffset)*2)
				return
			}
			u.FIO[ch].AnalogVoltage = fmt.Sprintf("%0.3f",
				(float64(read)*slope+offset)*2)
			return
		}
	}
	ch -= 8
	if u.EIO[ch].AD == "Analog" {
		read := uint16(recBuffer[9]) + uint16(recBuffer[10])*256
		u.EIO[ch].AnalogRead = read
		u.EIO[ch].AnalogVoltage = fmt.Sprintf("%0.3f",
			(float64(read)*slope+offset)*2)
	}
}

//helper function for processing FIO, EIO, and CIO bits when reading from flash.
func (u *U3) parseFlashBytes(recBuffer []byte) {
	for i := 0; i < 8; i++ {
		u.FIO[i].AD = "Digital"
		if recBuffer[23]&(1<<i) != 0 {
			u.FIO[i].AD = "Analog"
		}
		u.EIO[i].AD = "Digital"
		if recBuffer[26]&(1<<i) != 0 {
			u.EIO[i].AD = "Analog"
		}
		if i < 4 {
			u.CIO[i].AD = "Digital"
		}
		if i > 3 {
			u.FIO[i].IO = "Input"
			if recBuffer[24]&(1<<i) != 0 {
				u.FIO[i].IO = "Output"
			}
		}
		u.EIO[i].IO = "Input"
		if recBuffer[27]&(1<<i) != 0 {
			u.FIO[i].IO = "Output"
		}
		if i < 4 {
			u.CIO[i].IO = "Input"
			if recBuffer[29]&(1<<i) != 0 {
				u.CIO[i].IO = "Output"
			}
		}
	}
}
