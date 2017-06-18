// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs syscall_ignore.go

package generic

const (
	slaveCmd = 0x703
	rdrwCmd  = 0x707
	rd       = 0x1

	I2cSMBus      = 0x720
	I2cSlaveForce = 0x706

	I2cSMBusRead  = 0x1
	I2cSMBusWrite = 0x0

	I2cSMBusQuick        = 0x0
	I2cSMBusByte         = 0x1
	I2cSMBusByteData     = 0x2
	I2cSMBusWordData     = 0x3
	I2cSMBusProcCall     = 0x4
	I2cSMBusBlockData    = 0x5
	I2cSMBusI2cBlockData = 0x8
)

const (
	I2cSmBusBlockMax    = 0x20
	I2cSmBusI2cBlockMax = I2cSmBusBlockMax
)

type i2c_msg struct {
	Addr      uint16
	Flags     uint16
	Len       uint16
	Pad_cgo_0 [2]byte
	Buf       *uint8
}
type i2c_rdwr_ioctl_data struct {
	Msgs      *i2c_msg
	Nmsgs     uint32
	Pad_cgo_0 [4]byte
}

type i2c_smbus_ioctl_data struct {
	Write     uint8
	Command   uint8
	Pad_cgo_0 [2]byte
	Size      uint32
	Data      *[34]byte
}
