package main

import "fmt"

//Jack file is a set of LabJack helper frunctions.

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

// Takes a buffer and an offset, and turns into an 32-bit integer
func makeInt(buffer []byte, offset int) int {
	return int((buffer[offset+3] << 24) + (buffer[offset+2] << 16) + (buffer[offset+1] << 8) + buffer[offset])
}

// Takes a buffer and an offset, and turns into an 16-bit integer
func makeShort(buffer []byte, offset int) int {
	return int((buffer[offset+1] << 8) + buffer[offset])
}
