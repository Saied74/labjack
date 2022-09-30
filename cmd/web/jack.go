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

type u3srData map[string]*u3srElement

func buildU3srData() u3srData {
	return u3srData{
		configJack: &u3srElement{
			sendLength:  26,
			recLength:   38,
			byte1:       0xF8,
			byte2:       0x0A,
			byte3:       0x08,
			byte6:       0x00,
			byte7:       0x00,
			checkReturn: checkJack,
			buildBytes:  buildJackSendBuffer,
		},
		configIO: &u3srElement{
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
		ain: &u3srElement{
			sendLength: 10,
			recLength:  11,
			byte1:      0xF8,
			byte2:      3,
			byte3:      0x00,
			byte6:      0,
			byte7:      1,
		},
		led: &u3srElement{
			sendLength: 9,
			recLength:  9,
			byte1:      0xF8,
			byte2:      2,
			byte3:      0x00,
			byte6:      0,
			byte7:      9,
		},
		portStateRead: &u3srElement{
			sendLength: 8,
			recLength:  10,
			byte1:      0xF8,
			byte2:      2,
			byte3:      0x00,
			byte6:      0,
			byte7:      10,
		},
		portStateWrite: &u3srElement{
			sendLength: 14,
			recLength:  11,
			byte1:      0xF8,
			byte2:      7,
			byte3:      0x00,
			byte6:      0,
			byte7:      27,
		},
		portDirRead: &u3srElement{
			sendLength:  8,
			recLength:   12,
			byte1:       0xF8,
			byte2:       1,
			byte3:       0x00,
			byte6:       0,
			byte7:       28,
			checkReturn: checkFeedback,
			buildBytes:  buildPortDirReadBuffer,
		},
		portDirWrite: &u3srElement{
			sendLength:  14,
			recLength:   10,
			byte1:       0xF8,
			byte2:       4,
			byte3:       0x00,
			byte6:       0,
			byte7:       29,
			checkReturn: checkFeedback,
			buildBytes:  buildPortDirWriteBuffer,
		},
		tempSense: &u3srElement{
			sendLength: 8,
			recLength:  11,
			byte1:      0xF8,
			byte2:      1,
			byte3:      0x00,
			byte6:      0,
			byte7:      30,
		},
		vReg: &u3srElement{
			sendLength: 8,
			recLength:  11,
			byte1:      0xF8,
			byte2:      1,
			byte3:      0x00,
			byte6:      0,
			byte7:      31,
		},
	}
}

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

// func checkFeedback(sr *u3srElement, recBuffer []byte) error {
// 	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
// 		return fmt.Errorf("The U3 detected a bad checksum. Double check your checksum calculations and try again")
// 	} else {
// 		if recBuffer[1] != sr.byte1 || recBuffer[2] != sr.byte2 || recBuffer[3] != sr.byte3 {
// 			return fmt.Errorf("Got the wrong command bytes back from the U3")
// 		}
//
// 		checksum16 := calculateChecksum16(recBuffer, sr.recLength)
// 		checksum8 := calculateChecksum8(recBuffer)
// 		if checksum8 != recBuffer[0] || int(recBuffer[4]) != checksum16&0xff || int(recBuffer[5]) != ((checksum16/256)&0xff) {
// 			return fmt.Errorf("Response had invalid checksum.\n%d != %d, %d != %d, %d != %d", checksum8, recBuffer[0], checksum16&0xff, recBuffer[4], ((checksum16 / 256) & 0xff), recBuffer[5])
// 		} else {
// 			if recBuffer[6] != 0 { // Check the error code in the packet. See section 5.3 of the U3
// 				return fmt.Errorf("Command returned with an errorcode = %d", recBuffer[6])
// 			}
// 			return nil
// 		}
// 	}
// }

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

func buildPortDirReadBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	sendBuffer[6] = writeMask
	sendBuffer[7] = sr.byte7
	addChecksum(sr, sendBuffer)
}

func buildPortDirWriteBuffer(sr *u3srElement, sendBuffer []byte, writeMask byte) {
	copyHead(sr, sendBuffer)
	for i := 8; i < sr.sendLength; i++ {
		sendBuffer[i] = 0x00
	}
	sendBuffer[8] = 0xf0
	sendBuffer[9] = 0xff
	sendBuffer[10] = 0x0f
	sendBuffer[11] = sr.byte11
	sendBuffer[12] = sr.byte12
	sendBuffer[13] = sr.byte13
	addChecksum(sr, sendBuffer)
}

func copyHead(sr *u3srElement, sendBuffer []byte) {
	sendBuffer[1] = sr.byte1
	sendBuffer[2] = sr.byte2
	sendBuffer[3] = sr.byte3
	sendBuffer[6] = sr.byte6
	sendBuffer[7] = sr.byte7
}

func addChecksum(sr *u3srElement, sendBuffer []byte) {
	checksum := 0
	checksum = calculateChecksum16(sendBuffer, sr.sendLength)
	sendBuffer[4] = byte(checksum & 0xff)
	sendBuffer[5] = byte((checksum / 256) & 0xff)
	sendBuffer[0] = calculateChecksum8(sendBuffer)
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
	open              bool
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

// Parses the ConfigU3 packet into something useful.
func (u *U3) parseConfigU3Bytes(recBuffer []byte) {

	u.FirmwareVersion = fmt.Sprintf("%d.%02d", int(recBuffer[10]), int(recBuffer[9]))
	u.BootLoaderVersion = fmt.Sprintf("%d.%02d", recBuffer[12], recBuffer[11])
	u.HardwareVersion = fmt.Sprintf("%d.%02d", recBuffer[14], recBuffer[13])
	u.SerialNumber = fmt.Sprintf("%d", makeInt(recBuffer, 15))
	u.ProductID = fmt.Sprintf("%d", makeShort(recBuffer, 19))
	u.LocalID = fmt.Sprintf("%d", recBuffer[21])

	u.parseFlashBytes(recBuffer)

	// fmt.Printf("  TimerCounterMask = %d\n", recBuffer[22])
	// fmt.Printf("  FIOAnalog = %d\n", recBuffer[23])
	// fmt.Printf("  FIODireciton = %d\n", recBuffer[24])
	// fmt.Printf("  FIOState = %d\n", recBuffer[25])
	// fmt.Printf("  EIOAnalog = %d\n", recBuffer[26])
	// fmt.Printf("  EIODirection = %d\n", recBuffer[27])
	// fmt.Printf("  EIOState = %d\n", recBuffer[28])
	// fmt.Printf("  CIODirection = %d\n", recBuffer[29])
	// fmt.Printf("  CIOState = %d\n", recBuffer[30])
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

// Takes a buffer and an offset, and turns into an 32-bit integer
func makeInt(buffer []byte, offset int) int {
	return int((buffer[offset+3] << 24) + (buffer[offset+2] << 16) + (buffer[offset+1] << 8) + buffer[offset])
}

// Takes a buffer and an offset, and turns into an 16-bit integer
func makeShort(buffer []byte, offset int) int {
	return int((buffer[offset+1] << 8) + buffer[offset])
}
