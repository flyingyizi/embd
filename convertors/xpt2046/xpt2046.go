// Package mcp3008 allows interfacing with the mcp3008 8-channel, 10-bit ADC through SPI protocol.
package xpt2046

import (
	"fmt"
	"math"

	"sort"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

// XPT2046 represents a xpt2046 SAR DAC.
type XPT2046 struct {
	PenIrq embd.DigitalPin
	Bus    embd.SPIBus
	XY     chan Coordinate
}

type Coordinate struct {
	X, Y int
}

func NewPENIRQ(penIrq interface{}) embd.DigitalPin {
	if penIrq == nil {
		return nil
	}
	var digitalPin embd.DigitalPin

	if pin, ok := penIrq.(embd.DigitalPin); ok {
		digitalPin = pin
	} else {
		var err error
		digitalPin, err = embd.NewDigitalPin(penIrq)
		if err != nil {
			glog.V(1).Infof("GPIO: error creating digital pin %+v: %s", penIrq, err)
			return nil
		}
	}
	if err := digitalPin.SetDirection(embd.In); err != nil {
		glog.Errorf("GPIO: error setting pin %+v to out direction: %s", penIrq, err)
		digitalPin.Close()
		return nil
	}
	if err := digitalPin.ActiveLow(false); err != nil {
		glog.Errorf("GPIO: error setting pin %+v to low active : %s", penIrq, err)
		digitalPin.Close()
		return nil
	}

	return digitalPin
}

// New creates a representation of the mcp3008 convertor
//penIrq   connect to penirq gpio
func New(bus embd.SPIBus, irqpin embd.DigitalPin) *XPT2046 {
	chx := make(chan Coordinate)
	return &XPT2046{Bus: bus, PenIrq: irqpin, XY: chx}
}

//Watch register PENIRQ
func (hd *XPT2046) Watch() error {
	if hd.PenIrq == nil {
		return fmt.Errorf("irq pin is nil, can not watch")
	}

	//EdgeNone    Edge = "none"
	//EdgeRising  Edge = "rising"
	//EdgeFalling Edge = "falling"
	//EdgeBoth    Edge = "EdgeBoth"
	//
	err := hd.PenIrq.Watch(embd.EdgeFalling, func(p embd.DigitalPin) {
		x, _ := hd.ReadX()
		y, _ := hd.ReadY()
		v, _ := p.Read()
		fmt.Println("triggered", "x:", x, "y:", y, "irqpin:", v)
		//if x, y, err := hd.TOUCH_XPT_ReadXY(); err == nil {
		if x == 0 && y == 0 {
			return
		}
		hd.XY <- Coordinate{X: x, Y: y}
		//}
	})
	return err
}
func (hd *XPT2046) StopWatching() error {
	if hd.PenIrq == nil {
		return fmt.Errorf("irq pin is nil, can not stopwatch")
	}
	err := hd.PenIrq.StopWatching()
	return err
}

func (hd *XPT2046) ReadX() (int, error) {
	return hd.readADCValue(ads12Bit, ChannelXPosition)
}
func (hd *XPT2046) ReadY() (int, error) {
	return hd.readADCValue(ads12Bit, ChannelYPosition)
}
func (hd *XPT2046) ReadZ1() (int, error) {
	return hd.readADCValue(ads12Bit, ChannelZ1Position)
}
func (hd *XPT2046) ReadZ2() (int, error) {
	return hd.readADCValue(ads12Bit, ChannelZ2Position)
}
func (hd *XPT2046) ReadTouchPressure() (int, error) {
	//# Formula (option 1) according to the datasheet (12bit conversion)
	//# RTouch = RX-Plate.(XPosition/4096).((Z1/Z2)-1)
	//# Not sure of the correct value of RX-Plate.
	//# Assuming the ratio is sufficient.
	//# Empirically this function seems to yield a values in the range of 0.4
	//# for a firm touch, and 1.75 for a light touch.

	x, _ := hd.ReadX()
	z1, _ := hd.ReadZ1()
	z2, _ := hd.ReadZ2()

	//# Avoid division by zero exception
	if z1 == 0 {
		z1 = 1
	}
	xDivisor := 4096
	//if hd.Conversion == conversion8Bit {
	//	xDivisor = 256
	//}
	result := (x / xDivisor) * ((z2 / z1) - 1)
	return result, nil
}

func (hd *XPT2046) ReadTemp0() (int, error) {
	return hd.readADCValue(ads8Bit, ChannelTemp0)
}
func (hd *XPT2046) ReadTemp1() (int, error) {
	return hd.readADCValue(ads8Bit, ChannelTemp1)
}
func (hd *XPT2046) ReadAux() (int, error) {
	return hd.readADCValue(ads8Bit, ChannelAuxiliary)
}
func (hd *XPT2046) ReadBatteryVoltage() (int, error) {
	return hd.readADCValue(ads8Bit, ChannelBatteryVoltage)
}

const (
	WINDOWSWIDTH = 4
)

func (hd *XPT2046) TOUCH_XPT_ReadXY() (x int, y int, err error) {

	//---分别读两次X值和Y值, 交叉着读可以提高一些读取精度---//
	var x1, x2, y1, y2 int
	if x1, err = hd.readFilterValue(ads12Bit, ChannelXPosition, WINDOWSWIDTH); err != nil {
		return 0, 0, err
	}
	if y1, err = hd.readFilterValue(ads12Bit, ChannelYPosition, WINDOWSWIDTH); err != nil {
		return 0, 0, err
	}
	if x2, err = hd.readFilterValue(ads12Bit, ChannelXPosition, WINDOWSWIDTH); err != nil {
		return 0, 0, err
	}
	if y2, err = hd.readFilterValue(ads12Bit, ChannelYPosition, WINDOWSWIDTH); err != nil {
		return 0, 0, err
	}

	//---求取X,y值的差值---//
	deltax := math.Abs(float64(x1 - x2))
	deltay := math.Abs(float64(y1 - y2))
	if (deltax > 50) || (deltay > 50) {

		s := fmt.Sprintf("deltax: %+v ;deltay: %+v\n", deltax, deltay)
		return 0, 0, fmt.Errorf(s)
	}

	x = (x1 + x2) / 2
	y = (y1 + y2) / 2
	err = nil
	return
}

func (hd *XPT2046) makeControlByte(conv ConversionSelect, chl ChannelSelect) byte {
	return (startBit | byte(chl) | byte(conv))
}

// readADCValue returns the  value at the given channel of the convertor.
func (hd *XPT2046) readADCValue(conv ConversionSelect, chl ChannelSelect) (int, error) {
	controlByte := hd.makeControlByte(conv, chl)

	var data [3]uint8
	data[0] = controlByte
	data[1] = 0
	data[2] = 0

	glog.V(2).Infof("xpt2046: sendingdata buffer %v", data)
	if err := hd.Bus.TransferAndReceiveData(data[:]); err != nil {
		return 0, err
	}

	resp := int(uint16(data[0])<<8 | uint16(data[1]))
	//resp = resp >> 4
	return resp, nil
}

//winWidth:  filter windows width, suggest value is 10
func (hd *XPT2046) readFilterValue(conv ConversionSelect, chl ChannelSelect, winWidth int) (int, error) {

	if winWidth < 3 {
		winWidth = 3
	}

	sli := make([]int, 0, winWidth)
	for i := 0; i < winWidth; i++ { //读取XY_READ_TIMS次结果
		if val, err := hd.readADCValue(conv, chl); err == nil {
			sli = append(sli, val)
		}
	}
	//---software filer
	sort.Ints(sli)
	//fmt.Println("sli:   ", sli)
	length := len(sli)

	var endValue int

	for _, v := range sli {
		endValue += v
	}
	endValue /= length

	return endValue, nil
}

//	class ChannelSelect(object):
type ChannelSelect byte

const (
	ChannelXPosition      ChannelSelect = 0x50 // 0b0101 0000  (5<<4)
	ChannelYPosition      ChannelSelect = 0x10 // 0b0001 0000  (1<<4)
	ChannelZ1Position     ChannelSelect = 0x30 // 0b0011 0000  (3<<4)
	ChannelZ2Position     ChannelSelect = 0x40 // 0b0100 0000  (4<<4)
	ChannelTemp0          ChannelSelect = 0x00 // 0b0000 0000  (0<<4)
	ChannelTemp1          ChannelSelect = 0x70 // 0b0111 0000  (7<<4)
	ChannelBatteryVoltage ChannelSelect = 0x20 // 0b0010 0000  (2<<4)
	ChannelAuxiliary      ChannelSelect = 0x60 // 0b0110 0000  (6<<4)
)

const (
	startBit byte = 0x80 // 0b1000 0000   (1<<7)
)

type ConversionSelect byte

const (
	ads8Bit  ConversionSelect = 0x08 //0b0000 1000 (1<<3)
	ads12Bit ConversionSelect = 0x00 //0b0000 0000 (0<<3)
)

const (
	// SingleMode represents the single-ended mode for the mcp3008.
	SingleMode = 1

	// DifferenceMode represents the diffenrential mode for the mcp3008.
	DifferenceMode = 0
)

/*
command format
7   6   5   4    3      2         1     0
S   A2  A3  A0   MODE   SER/DFR   PD1   PD0

mode:0(12 bit)  1(8bit)
*/
/*
A2   A1   A0
0    0    0    temp0
1    1    1    temp1
0    1    0    Vbat
1    1    0     AUXILIARY

1    0    0     z2 position
0    0    1     y position
0    1    1     z1 position
1    0    1     x position
*/
///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
