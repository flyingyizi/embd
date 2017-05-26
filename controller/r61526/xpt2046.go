package r61526

import (
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"

	"fmt"
	"math"
)

//https://github.com/MatthewLowden/RPi-XPT2046-Touchscreen-Python/blob/master/SPIManager.py
/*

 */
type Touch struct {
	Xpt2046
	Conversion ConversionSelect
}

// Init   initialize the st7565,
//special, if the cmd is 0xff, the func will deay 5us
func (hd *Touch) Init() error {
	hd.Conversion = conversion12Bit
	return nil
}

func (hd *Touch) SetMode(c ConversionSelect) {
	hd.Conversion = c
}
func (hd *Touch) makeControlByte(chl ChannelSelect) byte {
	return (startBit | byte(chl) | byte(hd.Conversion))
}

func (hd *Touch) TOUCH_XPT_ReadXY() (x uint16, y uint16, err error) {
	xPositionCmd := hd.makeControlByte(channelXPosition)
	yPositionCmd := hd.makeControlByte(channelYPosition)

	err = hd.MockSPIStart()
	if err != nil {
		return 0, 0, err
	}
	//---分别读两次X值和Y值, 交叉着读可以提高一些读取精度---//
	var x1, x2, y1, y2 uint16
	if x1, err = hd.XptReadData(xPositionCmd); err != nil {
		return 0, 0, err
	}
	if y1, err = hd.XptReadData(yPositionCmd); err != nil {
		return 0, 0, err
	}
	if x2, err = hd.XptReadData(xPositionCmd); err != nil {
		return 0, 0, err
	}
	if y2, err = hd.XptReadData(yPositionCmd); err != nil {
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

	x &= 0xFFF0 //去掉低四位
	y &= 0xFFF0

	//---确定XY值的范围，用在触摸屏大于TFT时---//
	if (x < 100) || (y > 4000) {
		return 0, 0, fmt.Errorf("da")
	}
	return
}

func (hd *Touch) Close() error {
	return hd.Xpt2046.Close()
}

// ST7565 represents an ST7565-compatible character LCD controller.
type Xpt2046 interface {
	Close() error

	MockSPIStart() error
	XptReadData(cmd byte) (uint16, error)
}

//NewGpio
//m: input Parallel8080 or Parallel6800
//cmds: initial command ,if it is nil, it will initilized with defaultInitCmd
func NewTouchGpio(m interface{}) (*Touch, error) {

	var th Touch
	switch inst := m.(type) {
	case Xpt2046GPIO:
		if con, err := newTouchGPIOPins(inst.DO, inst.CLK, inst.DIN, inst.CS,
			inst.BUSY); err == nil {
			inst.TouchConnection = con
			inst.InitDirection()
			th.Xpt2046 = &inst
			if err := th.Init(); err != nil {
				return nil, fmt.Errorf("init fail")
			}
			return &th, nil
		}
	}
	return nil, fmt.Errorf("unknow")
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

/*

//---定义使用的IO口---//
sbit TOUCH_DOUT = P2^0;
sbit TOUCH_CLK  = P2^1;
sbit TOUCH_DIN  = P2^2;
sbit TOUCH_CS   = P2^3;
sbit TOUCH_PEN  = P2^4;
*/
type Xpt2046GPIO struct {
	TouchConnection
	DO, CLK, DIN, CS, BUSY interface{}
}

var (
	DefaultTouchMap Xpt2046GPIO = Xpt2046GPIO{
		//PEN: "P1_3", //p27  //检测触摸屏响应信号
		/*BUSY PIN*/
		CS:  "P1_37", //p26  //片选
		DIN: "P1_32", //p25  //输入
		CLK: "P1_36", //p33  //时钟
		DO:  "P1_38", //p32  //输出
	}
)

func (hd *Xpt2046GPIO) MockSPIStart() error {

	/*
		TOUCH_CLK = 0;
		TOUCH_CS  = 1;
		TOUCH_DIN = 1;
		TOUCH_CLK = 1;
		TOUCH_CS  = 0;
	*/
	hd.writeCLK(embd.Low)
	hd.writeCS(embd.High)
	hd.writeDIN(embd.High)
	hd.writeCLK(embd.High)
	hd.writeCS(embd.Low)

	return nil
}

func (hd *Xpt2046GPIO) mockSPIWrite(dat byte) {

	hd.writeCLK(embd.Low)
	for i := 0; i < 8; i++ {
		hd.writeDIN(int(dat >> 7)) //
		dat <<= 1
		hd.writeCLK(embd.Low) //rising edage to latch data

		hd.writeCLK(embd.High)

	}
}

func (hd *Xpt2046GPIO) mockSPIRead() int {
	hd.writeCLK(embd.Low)
	var dat int
	for i := 0; i < 12; i++ { //receuve 12 but
		dat <<= 1

		hd.writeCLK(embd.High)
		hd.writeCLK(embd.Low)
		val, _ := hd.readDO()
		dat |= val

	}
	return dat
}

const (
	XY_READ_TIMS = 10 //读取次数
)

func (hd *Xpt2046GPIO) XptReadData(cmd byte) (uint16, error) {

	var readValue [XY_READ_TIMS]int

	hd.writeCLK(embd.Low) //先拉低时间
	hd.writeCS(embd.Low)  //选中芯片

	for i := 0; i < XY_READ_TIMS; i++ { //读取XY_READ_TIMS次结果
		hd.mockSPIWrite(cmd) //发送转换命令
		//Delay_6us();
		for j := 6; j > 0; j-- {
		} //延时等待转换结果
		hd.writeCLK(embd.High) //发送一个时钟周期，清除BUSY
		//_nop_()
		//_nop_()
		hd.writeCLK(embd.Low)
		//_nop_()
		//_nop_()

		readValue[i] = hd.mockSPIRead()
	}
	hd.writeCS(embd.High) //释放片选

	//---软件滤波---//
	//---先大到小排序，除去最高值，除去最低值，求其平均值---//
	for i := 0; i < XY_READ_TIMS-1; i++ { //从大到小排序
		for j := i + 1; j < XY_READ_TIMS; j++ {
			if readValue[i] < readValue[j] {
				readValue[i], readValue[j] = readValue[j], readValue[i]
			}
		}
	}
	var endValue int
	for i := 2; i < XY_READ_TIMS-2; i++ {
		endValue += readValue[i]
	}
	endValue = endValue / (XY_READ_TIMS - 4) //求平均值

	return uint16(endValue), nil
}

//func (hd *Xpt2046GPIO) ReadPEN() (int, error) {
//	return hd.TouchConnection.Read(hd.PEN)
//}
func (hd *Xpt2046GPIO) readDO() (int, error) {
	return hd.TouchConnection.Read(hd.DO)
}
func (hd *Xpt2046GPIO) writeCLK(val int) error {
	return hd.TouchConnection.Write(hd.CLK, val)
}
func (hd *Xpt2046GPIO) writeDIN(val int) error {
	return hd.TouchConnection.Write(hd.DIN, val)
}
func (hd *Xpt2046GPIO) writeCS(val int) error {
	return hd.TouchConnection.Write(hd.CS, val)
}
func (hd *Xpt2046GPIO) InitDirection() error {

	functions := []func() error{
		func() error { return hd.PinInst(hd.CS).SetDirection(embd.Out) },
		func() error { return hd.PinInst(hd.DIN).SetDirection(embd.Out) },
		func() error { return hd.PinInst(hd.CLK).SetDirection(embd.Out) },
		func() error { return hd.PinInst(hd.DO).SetDirection(embd.In) },
		//func() error { return hd.PinInst(hd.PEN).SetDirection(embd.In) },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			glog.Errorf("xpt2046: error setting   direction: %s", err)
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
// Connection abstracts the different methods of communicating.
type TouchConnection interface {
	// Close closes all open resources.
	Close() error
	Write(pin interface{}, val int) error
	Read(pin interface{}) (int, error)

	PinInst(pin interface{}) embd.DigitalPin
}

type TouchGPIOConnection struct {
	// con is the list of registered GPIO.
	con map[interface{}]embd.DigitalPin
}

func newTouchGPIOPins(pins ...interface{}) (*TouchGPIOConnection, error) {

	var ll = TouchGPIOConnection{}
	ll.con = make(map[interface{}]embd.DigitalPin)

	for _, key := range pins {
		if key == nil {
			continue
		}
		var digitalPin embd.DigitalPin
		if pin, ok := key.(embd.DigitalPin); ok {
			digitalPin = pin
		} else {
			var err error
			digitalPin, err = embd.NewDigitalPin(key)
			if err != nil {
				glog.V(1).Infof("GPIO: error creating digital pin %+v: %s", key, err)

				return nil, err
			}
		}
		ll.con[key] = digitalPin
	}

	for _, pin := range ll.con {
		if pin == nil {
			continue
		}

	}

	return &ll, nil
}

// Close closes all open DigitalPins.
func (ll *TouchGPIOConnection) Close() error {
	glog.V(2).Info("GPIO: closing all GPIO pins")

	for _, pin := range ll.con {
		err := pin.Close()
		if err != nil {
			glog.Errorf("GPIO: error closing pin %+v: %s", pin, err)
			return err
		}
	}
	return nil
}

// Write  write value to pin gpio port
func (ll *TouchGPIOConnection) Write(pin interface{}, val int) error {
	if pin == nil {
		return nil
	}
	var digitalPin embd.DigitalPin
	if p, ok := pin.(embd.DigitalPin); ok {
		digitalPin = p
	} else {
		if p, ok := ll.con[pin]; ok {
			digitalPin = p
		}
	}

	return digitalPin.Write(val)
}
func (ll *TouchGPIOConnection) Read(pin interface{}) (int, error) {
	if pin == nil {
		return 0, nil
	}
	var digitalPin embd.DigitalPin
	if p, ok := pin.(embd.DigitalPin); ok {
		digitalPin = p
	} else {
		if p, ok := ll.con[pin]; ok {
			digitalPin = p
		}
	}

	return digitalPin.Read()
}

// PinInst  return gpio pin instance according to input
func (ll *TouchGPIOConnection) PinInst(pin interface{}) embd.DigitalPin {
	if pin == nil {
		return nil
	}
	var digitalPin embd.DigitalPin
	if p, ok := pin.(embd.DigitalPin); ok {
		digitalPin = p
	} else {
		if p, ok := ll.con[pin]; ok {
			digitalPin = p
		}
	}

	return digitalPin
}

func usDealy_(val time.Duration) {
	time.Sleep(val * time.Microsecond)
	//Millisecond
}

//commands from xpt2046 datasheet datasheet p18 p14
const (
	//---定义芯片命令字节---//  xpt2046 datasheet p14
	//XPT_CMD_X = 0xD0 //读取X轴的命令
	XPT_CMD_Y = 0x90 //读取Y轴的命令
)

//	class ChannelSelect(object):
type ChannelSelect byte

const (
	channelXPosition      ChannelSelect = 0x50 // 0b0101 0000
	channelYPosition      ChannelSelect = 0x10 // 0b0001 0000
	channelZ1Position     ChannelSelect = 0x30 // 0b0011 0000
	channelZ2Position     ChannelSelect = 0x40 // 0b0100 0000
	channelTemp0          ChannelSelect = 0x00 // 0b0000 0000
	channelTemp1          ChannelSelect = 0x70 // 0b0111 0000
	channelBatteryVoltage ChannelSelect = 0x20 // 0b0010 0000
	channelAuxiliary      ChannelSelect = 0x60 // 0b0110 0000
)

const (
	startBit byte = 0x80 // 0b1000 0000
)

type ConversionSelect byte

const (
	conversion8Bit  ConversionSelect = 0x08 //0b0000 1000
	conversion12Bit ConversionSelect = 0x00 //0b0000 0000
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
