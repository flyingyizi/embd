// +build ignore

package generic

/*
#include <linux/i2c.h>
#include <linux/i2c-dev.h>
#include <linux/spi/spidev.h>
*/
import "C"

// for I2c

const (
	slaveCmd = C.I2C_SLAVE
	rdrwCmd  = C.I2C_RDWR
	rd       = C.I2C_M_RD

	I2cSMBus      = C.I2C_SMBUS
	I2cSlaveForce = C.I2C_SLAVE_FORCE

	I2cSMBusRead  = C.I2C_SMBUS_READ
	I2cSMBusWrite = C.I2C_SMBUS_WRITE

	I2cSMBusQuick        = C.I2C_SMBUS_QUICK
	I2cSMBusByte         = C.I2C_SMBUS_BYTE
	I2cSMBusByteData     = C.I2C_SMBUS_BYTE_DATA
	I2cSMBusWordData     = C.I2C_SMBUS_WORD_DATA
	I2cSMBusProcCall     = C.I2C_SMBUS_PROC_CALL
	I2cSMBusBlockData    = C.I2C_SMBUS_BLOCK_DATA
	I2cSMBusI2cBlockData = C.I2C_SMBUS_I2C_BLOCK_DATA
)

const (
	I2cSmBusBlockMax    = C.I2C_SMBUS_BLOCK_MAX /* As specified in SMBus standard */
	I2cSmBusI2cBlockMax = I2cSmBusBlockMax      /* Not specified but we use same structure */
)

type i2c_msg C.struct_i2c_msg
type i2c_rdwr_ioctl_data C.struct_i2c_rdwr_ioctl_data

type i2c_smbus_ioctl_data C.struct_i2c_smbus_ioctl_data

type spiIOCTransfer C.struct_spi_ioc_transfer

const (
	SizeofspiIOCTransfer       = C.sizeof_struct_spi_ioc_transfer
	Sizeofi2c_msg              = C.sizeof_struct_i2c_msg
	Sizeofi2c_smbus_ioctl_data = C.sizeof_struct_i2c_smbus_ioctl_data
	Sizeofi2c_rdwr_ioctl_data  = C.sizeof_struct_i2c_rdwr_ioctl_data
)

/*
#define SPI_MSGSIZE(N) \
	((((N)*(sizeof (struct spi_ioc_transfer))) < (1 << _IOC_SIZEBITS)) \
		? ((N)*(sizeof (struct spi_ioc_transfer))) : 0)
#define SPI_IOC_MESSAGE(N) _IOW(SPI_IOC_MAGIC, 0, char[SPI_MSGSIZE(N)])

*/

// for SPI
/*
// for SPI
const (
	spiIOCWrMode        = C.SPI_IOC_WR_MODE
	spiIOCWrBitsPerWord = C.SPI_IOC_WR_BITS_PER_WORD
	spiIOCWrMaxSpeedHz  = C.SPI_IOC_WR_MAX_SPEED_HZ
	spiIOCRdMode        = C.SPI_IOC_RD_MODE
	spiIOCRdBitsPerWord = C.SPI_IOC_RD_BITS_PER_WORD
	spiIOCRdMaxSpeedHz  = C.SPI_IOC_RD_MAX_SPEED_HZ
)
*/
