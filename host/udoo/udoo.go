/*
	Package rpi provides Raspberry Pi (including A+/B+) support.
	The following features are supported on Linux kernel 3.8+

	GPIO (digital (rw))
	I²C
	LED
*/
package udoo

import (
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/host/generic"
)

var spiDeviceMinor = 0

//# GPIO1_IO_25 means BANK=1 and ID=25
//# GPIO_NUMBER = ((1 - 1) * 32 ) + 25 = 25
//echo 25 > /sys/class/gpio/export
var neorev1Pins = embd.PinMap{
	&embd.PinDesc{ID: "J7_47", Aliases: []string{"4", "GPIO_4", "TXD", "UART1_TXD"}, Caps: embd.CapDigital | embd.CapUART, DigitalLogical: 4},
	&embd.PinDesc{ID: "J7_46", Aliases: []string{"5", "GPIO_5", "RXD", "UART1_RXD"}, Caps: embd.CapDigital | embd.CapUART, DigitalLogical: 5},
	&embd.PinDesc{ID: "J7_45", Aliases: []string{"6", "GPIO_6", "TXD", "UART2_TXD"}, Caps: embd.CapDigital | embd.CapUART, DigitalLogical: 6},
	&embd.PinDesc{ID: "J7_44", Aliases: []string{"7", "GPIO_7", "RXD", "UART2_RXD"}, Caps: embd.CapDigital | embd.CapUART, DigitalLogical: 7},
	&embd.PinDesc{ID: "J7_43", Aliases: []string{"116", "GPIO_116", "MOSI", "SPI5_MOSI"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 116},
	&embd.PinDesc{ID: "J7_42", Aliases: []string{"127", "GPIO_127", "SCLK", "SPI5_SCLK"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 127},
	&embd.PinDesc{ID: "J7_41", Aliases: []string{"124", "GPIO_124", "SS0", "SPI5_SS0"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 124},
	&embd.PinDesc{ID: "J7_40", Aliases: []string{"119", "GPIO_119", "MISO", "SPI5_MISO"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 119},

	&embd.PinDesc{ID: "J5_39", Aliases: []string{"174", "GPIO_174", "SS0", "SPI2_SS0"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 174},
	&embd.PinDesc{ID: "J5_38", Aliases: []string{"175", "GPIO_175", "SCLK", "SPI2_SCLK"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 175},
	&embd.PinDesc{ID: "J5_37", Aliases: []string{"176", "GPIO_176", "SDA", "I2C2_SDA"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 176},
	&embd.PinDesc{ID: "J5_36", Aliases: []string{"177", "GPIO_177", "SCL", "I2C2_SCL"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 177},

	&embd.PinDesc{ID: "J6_33", Aliases: []string{"21", "GPIO_21", "SDA", "I2C4_SDA"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 21},
	&embd.PinDesc{ID: "J6_32", Aliases: []string{"20", "GPIO_20", "SCL", "I2C4_SCL"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 20},
	&embd.PinDesc{ID: "J6_31", Aliases: []string{"19", "GPIO_19", "PWM", "PWM_6"}, Caps: embd.CapDigital, DigitalLogical: 19},
	&embd.PinDesc{ID: "J6_30", Aliases: []string{"18", "GPIO_18", "PWM", "PWM_5"}, Caps: embd.CapDigital, DigitalLogical: 18},
	&embd.PinDesc{ID: "J6_29", Aliases: []string{"17", "GPIO_17"}, Caps: embd.CapDigital, DigitalLogical: 17},
	&embd.PinDesc{ID: "J6_28", Aliases: []string{"16", "GPIO_16"}, Caps: embd.CapDigital, DigitalLogical: 16},
	&embd.PinDesc{ID: "J6_27", Aliases: []string{"15", "GPIO_15", "SDA", "I2C1_SDA"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 15},
	&embd.PinDesc{ID: "J6_26", Aliases: []string{"14", "GPIO_14", "SCL", "I2C1_SCL"}, Caps: embd.CapDigital | embd.CapI2C, DigitalLogical: 14},
	&embd.PinDesc{ID: "J6_25", Aliases: []string{"22", "GPIO_22"}, Caps: embd.CapDigital, DigitalLogical: 22},
	&embd.PinDesc{ID: "J6_24", Aliases: []string{"25", "GPIO_25"}, Caps: embd.CapDigital, DigitalLogical: 25},

	&embd.PinDesc{ID: "J4_23", Aliases: []string{"124", "GPIO_124"}, Caps: embd.CapDigital, DigitalLogical: 124},
	&embd.PinDesc{ID: "J4_22", Aliases: []string{"182", "GPIO_182"}, Caps: embd.CapDigital, DigitalLogical: 182},
	&embd.PinDesc{ID: "J4_21", Aliases: []string{"173", "GPIO_173", "MOSI", "SPI2_MOSI"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 173},
	&embd.PinDesc{ID: "J4_20", Aliases: []string{"172", "GPIO_172", "MISO", "SPI2_MISO"}, Caps: embd.CapDigital | embd.CapSPI, DigitalLogical: 172},
	&embd.PinDesc{ID: "J4_19", Aliases: []string{"181", "GPIO_181"}, Caps: embd.CapDigital, DigitalLogical: 181},
	&embd.PinDesc{ID: "J4_18", Aliases: []string{"180", "GPIO_180"}, Caps: embd.CapDigital, DigitalLogical: 180},
	&embd.PinDesc{ID: "J4_17", Aliases: []string{"107", "GPIO_107", "PWM", "PWM_4"}, Caps: embd.CapDigital, DigitalLogical: 107},
	&embd.PinDesc{ID: "J4_16", Aliases: []string{"106", "GPIO_106", "PWM", "PWM_3"}, Caps: embd.CapDigital, DigitalLogical: 106},
}

//var ledMap = embd.LEDMap{
//	"led0": []string{"0", "led0", "LED0"},
//}

func init() {
	embd.Register(embd.HostUdoo, func(rev int) *embd.Descriptor {
		pins := neorev1Pins
		return &embd.Descriptor{
			GPIODriver: func() embd.GPIODriver {
				return embd.NewGPIODriver(pins, generic.NewDigitalPin, nil, nil)
			},
			I2CDriver: func() embd.I2CDriver {
				return embd.NewI2CDriver(generic.NewI2CBus)
			},
			//LEDDriver: func() embd.LEDDriver {
			//	return embd.NewLEDDriver(ledMap, generic.NewLED)
			//},
			SPIDriver: func() embd.SPIDriver {
				return embd.NewSPIDriver(spiDeviceMinor, generic.NewSPIBus, nil)
			},
		}
	})
}
