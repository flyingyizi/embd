// Package mcp3008 allows interfacing with the mcp3008 8-channel, 10-bit ADC through SPI protocol.
package xpt2046

import (
	"fmt"
	"math"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

// XPT2046 represents a xpt2046 SAR DAC.
type XPT2046 struct {
	Conversion ConversionSelect

	Bus embd.SPIBus
}

// New creates a representation of the mcp3008 convertor
func New(conversion ConversionSelect, bus embd.SPIBus) *XPT2046 {
	return &XPT2046{Conversion: conversion, Bus: bus}
}

//SetMode   set to 8bit or 12bit conversion
func (hd *XPT2046) SetMode(c ConversionSelect) {
	hd.Conversion = c
}

func (hd *XPT2046) ReadX() (int, error) {
	return hd.readValue(ChannelXPosition)
}
func (hd *XPT2046) ReadY() (int, error) {
	return hd.readValue(ChannelYPosition)
}
func (hd *XPT2046) ReadZ1() (int, error) {
	return hd.readValue(ChannelZ1Position)
}
func (hd *XPT2046) ReadZ2() (int, error) {
	return hd.readValue(ChannelZ2Position)
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
	if hd.Conversion == conversion8Bit {
		xDivisor = 256
	}
	result := (x / xDivisor) * ((z2 / z1) - 1)
	return result, nil
}

func (hd *XPT2046) TOUCH_XPT_ReadXY() (x int, y int, err error) {

	//---分别读两次X值和Y值, 交叉着读可以提高一些读取精度---//
	var x1, x2, y1, y2 int
	if x1, err = hd.readFilterValue(ChannelXPosition); err != nil {
		return 0, 0, err
	}
	if y1, err = hd.readFilterValue(ChannelYPosition); err != nil {
		return 0, 0, err
	}
	if x2, err = hd.readFilterValue(ChannelXPosition); err != nil {
		return 0, 0, err
	}
	if y2, err = hd.readFilterValue(ChannelYPosition); err != nil {
		return 0, 0, err
	}

	//---求取X,y值的差值---//
	deltax := math.Abs(float64(x1 - x2))
	deltay := math.Abs(float64(y1 - y2))
	if (deltax > 50) || (deltay > 50) {
		return 0, 0, fmt.Errorf("da")
	}

	//---求取两次读取值的平均数作为读取到的XY值---//
	x = (x1 + x2) / 2
	y = (y1 + y2) / 2

	//x &= 0xFFF0 //去掉低四位
	//y &= 0xFFF0

	//---确定XY值的范围，用在触摸屏大于TFT时---//
	if (x < 100) || (y > 4000) {
		return 0, 0, fmt.Errorf("da")
	}
	return
}

func (hd *XPT2046) makeControlByte(chl ChannelSelect) byte {
	return (startBit | byte(chl) | byte(hd.Conversion))
}

// readValue returns the  value at the given channel of the convertor.
func (hd *XPT2046) readValue(chl ChannelSelect) (int, error) {
	controlByte := hd.makeControlByte(chl)

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

const (
	XY_READ_TIMS = 10 //读取次数
)

func (hd *XPT2046) readFilterValue(chl ChannelSelect) (int, error) {

	var readValue [XY_READ_TIMS]int
	var readNum int = 0
	for i := 0; i < XY_READ_TIMS; i++ { //读取XY_READ_TIMS次结果
		if val, err := hd.readValue(chl); err != nil {
			readValue[readNum] = val
			readNum = readNum + 1
		}
	}

	//---软件滤波---//
	//---先大到小排序，除去最高值，除去最低值，求其平均值---//
	for i := 0; i < readNum-1; i++ { //从大到小排序
		for j := i + 1; j < readNum; j++ {
			if readValue[i] < readValue[j] {
				readValue[i], readValue[j] = readValue[j], readValue[i]
			}
		}
	}
	var endValue int
	for i := 2; i < readNum-2; i++ {
		endValue += readValue[i]
	}
	endValue = endValue / (readNum - 4) //求平均值

	return endValue, nil
}

//	class ChannelSelect(object):
type ChannelSelect byte

const (
	ChannelXPosition      ChannelSelect = 0x50 // 0b0101 0000
	ChannelYPosition      ChannelSelect = 0x10 // 0b0001 0000
	ChannelZ1Position     ChannelSelect = 0x30 // 0b0011 0000
	ChannelZ2Position     ChannelSelect = 0x40 // 0b0100 0000
	ChannelTemp0          ChannelSelect = 0x00 // 0b0000 0000
	ChannelTemp1          ChannelSelect = 0x70 // 0b0111 0000
	ChannelBatteryVoltage ChannelSelect = 0x20 // 0b0010 0000
	ChannelAuxiliary      ChannelSelect = 0x60 // 0b0110 0000
)

const (
	startBit byte = 0x80 // 0b1000 0000
)

type ConversionSelect byte

const (
	conversion8Bit  ConversionSelect = 0x08 //0b0000 1000
	conversion12Bit ConversionSelect = 0x00 //0b0000 0000
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
