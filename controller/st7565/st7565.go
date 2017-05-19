package st7565

//http://www.nxp.com/documents/data_sheet/PCF8574.pdf

//http://www.rototron.info/raspberry-pi-graphics-lcd-display-tutorial/
import (
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"

	"errors"
	"fmt"
)

/*

 */
type LCD struct {
	ST7565
}

const (
	// LCD Parameters
	LcdWidth  = 128
	LcdHeight = 64
	//lcdPageCount   ：according to datasheet(P42), not beyond 8 pages
	lcdPageCount = 8
	LCD_CONTRAST = 0x19
)

//# LCD Page Order
//var pagemap = [...]byte{3, 2, 1, 0, 7, 6, 5, 4}
var pagemap = [...]byte{0, 1, 2, 3, 4, 5, 6, 7}

//SetCursor
func (hd *LCD) SetCursor(x /*column*/, y /*page*/ byte) error {
	if x >= LcdWidth || x < 0 { //
		return errors.New("according to datasheet(P43), beyond")
	}
	if y > lcdPageCount-1 || y < 0 { //
		return errors.New("according to datasheet(P42), not beyond 8 pages")
	}

	//set page
	pagexxx := pagemap[y] | cmdSetPageAddr
	if err := hd.WriteCmd(pagexxx); err != nil {
		return err
	}

	//set upper/lower bits of column
	lsb := x & 0x0f
	msb := (x & 0xf0) >> 4
	msb = msb | cmdSetColumnUpper
	if err := hd.WriteCmd(msb); err != nil {
		return err
	}
	lsb = lsb | cmdSetColumnLower
	if err := hd.WriteCmd(lsb); err != nil {
		return err
	}
	return nil
}

// Clear clears the display and mode settings sets the cursor to the home position.
func (hd *LCD) Clear() error {

	for i := 0; i < 8; i++ {
		if err := hd.SetCursor(0, byte(i)); err != nil {
			return err
		}

		for j := 0; j < 128; j++ {
			hd.WriteData(0x00)

		}
	}
	hd.WriteCS(embd.High)

	return nil
}

//SetVolumeMode    The Electronic Volume Mode Set  , datasheet p47
//for display contrast ratio
func (hd *LCD) SetVolumeMode(levl /*voltage levels*/ byte) error {
	if levl > 0x3F || levl < 0 { //
		return errors.New("according to datasheet(P47), beyond")
	}

	if err := hd.WriteCmd(cmdSetVolumeFIRST); err != nil {
		return err
	}
	if err := hd.WriteCmd(levl); err != nil {
		return err
	}
	return nil
}

//SetResistorRATIOMode    V0 Voltage Regulator Internal Resistor Ratio Set, , datasheet p47
//for display contrast ratio
func (hd *LCD) SetResistorRATIOMode(ratio /*resistor ratio*/ byte) error {
	if ratio > 7 || ratio < 0 { //
		return errors.New("according to datasheet(P47), beyond")
	}
	ratio = cmdSetResistorRATIO | ratio
	if err := hd.WriteCmd(ratio); err != nil {
		return err
	}
	return nil
}

//SetPowerControlMode     refer datasheet p47
func (hd *LCD) SetPowerControlMode(boosterON, VRON, VFON bool) error {
	//refer datasheet p31
	var d2, d1, d0 byte
	if boosterON {
		d2 = 0x01 << 2
	}
	if VRON {
		d1 = 0x01 << 1
	}
	if VFON {
		d0 = 0x01

	}
	v := cmdSetPowerControl | d2 | d1 | d0
	if err := hd.WriteCmd(v); err != nil {
		return err
	}
	return nil
}

//SetBoosterRatioMode      ratio: 0-->2x/3x/4x;  1->5x; 3->6x
func (hd *LCD) SetBoosterRatioMode(ratio byte) error {
	if ratio > 3 || ratio < 0 { //
		return errors.New("according to datasheet(P49), beyond")
	}

	if err := hd.WriteCmd(cmdSetBoosterRatio); err != nil {
		return err
	}
	if err := hd.WriteCmd(ratio); err != nil {
		return err
	}
	return nil
}

// ST7565 represents an ST7565-compatible character LCD controller.
type ST7565 interface {
	Close() error

	WriteCmd(cmd byte) error
	WriteCS(val int) error
	WriteData(cmd byte) error
	//eMode entryMode
	//dMode displayMode
	//fMode functionMode
}

//NewGpio
//m: input LCD8080 or LCD6800
//cmds: initial command ,if it is nil, it will initilized with defaultInitCmd
func NewGpio(m interface{}, cmds ...byte) (*LCD, error) {
	if cmds == nil {
		cmds = defaultInitCmd
	}

	var lcd LCD
	switch inst := m.(type) {
	case LCD8080:
		if con, err := newGPIOPins(inst.CS, inst.WR, inst.RST, inst.RS, inst.RD,
			inst.DB0, inst.DB1, inst.DB2, inst.DB3,
			inst.DB4, inst.DB5, inst.DB6, inst.DB7); err == nil {
			inst.Connection = con
			if err := inst.Init(cmds); err != nil {
				return nil, fmt.Errorf("init fail")
			}
			lcd.ST7565 = &inst
			return &lcd, nil
		}
	case LCD6800:
		//todo
		//if con, err := newGPIOPins(inst.CS, inst.WR, inst.RST, inst.RS, inst.RD,
		//	inst.DB0, inst.DB1, inst.DB2, inst.DB3,
		//	inst.DB4, inst.DB5, inst.DB6, inst.DB7); err == nil {
		//	inst.Connection = con
		//	if err := inst.Init(); err != nil {
		//		return nil, fmt.Errorf("init fail")
		//	}
		//	lcd.ST7565 = &inst
		//	return &lcd, nil
		//}
		return nil, fmt.Errorf("unknow")
	}
	return nil, fmt.Errorf("unknow")
}

// Init   initialize the st7565, the command is from the datasheet
func (hd *LCD8080) Init(cmds []byte) error {
	if err := hd.WriteCS(embd.Low); err != nil {
		return err
	}
	if err := hd.reset(); err != nil {
		return err
	}

	for _, cmd := range cmds {
		if err := hd.WriteCmd(cmd); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying Connection.
func (hd *LCD8080) Close() error {
	return hd.Connection.Close()
}

// # Initialize back_buffer
var back_buffer = [LcdHeight][LcdWidth]byte{}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Connection abstracts the different methods of communicating with an R61526.
type Connection interface {
	// Close closes all open resources.
	Close() error
	// Write  write value to pin gpio port
	Write(pin interface{}, val int) error
}

type LCD8080 struct {
	Connection
	CS, WR, RST, RS, RD                    interface{}
	DB0, DB1, DB2, DB3, DB4, DB5, DB6, DB7 interface{}
}

var (
	DefaultMap8080 LCD8080 = LCD8080{
		CS:  "P1_7",
		WR:  "P1_11",
		RST: "P1_12",
		RS:  "P1_15",
		RD:  "P1_40",
		DB0: "P1_19", DB1: "P1_21", DB2: "P1_23", DB3: "P1_16",
		DB4: "P1_18", DB5: "P1_22", DB6: "P1_24", DB7: "P1_26"}
	DefaultMap6800 LCD6800 = LCD6800{
		CS:  "P1_7",
		RW:  "P1_11",
		RST: "P1_12",
		A0:  "P1_15",
		E:   "P1_40",
		DB0: "P1_19", DB1: "P1_21", DB2: "P1_23", DB3: "P1_16",
		DB4: "P1_18", DB5: "P1_22", DB6: "P1_24", DB7: "P1_26"}
)

// WriteCmd for 8080 MPU
func (hd *LCD8080) WriteCmd(cmd byte) error {
	var err error

	if err = hd.WriteCS(embd.Low); err != nil { //chip select
		return err
	}

	if err = hd.writeRD(embd.High); err != nil { //disable read
		return err
	}

	if err = hd.writeRS(embd.Low); err != nil { //select command
		return err
	}

	if err = hd.writeWR(embd.Low); err != nil { //select write
		return err
	}

	//_nop_();
	//_nop_();
	if err = hd.fillDB8(cmd); err != nil {
		return err
	}
	usDealy(5)
	//trigger WR rising edge to latch into LCD
	if err = hd.writeWR(embd.High); err != nil { //command writing ，写入命令
		return err
	}
	return nil
}

// WriteData for 8080 MPU
func (hd *LCD8080) WriteData(dat byte) error {
	if err := hd.WriteCS(embd.Low); err != nil { //chip select,打开片选
		return err
	}

	if err := hd.writeRD(embd.High); err != nil { //disable read，读失能
		return err
	}

	if err := hd.writeRS(embd.High); err != nil { //select data，选择数据
		return err
	}

	if err := hd.writeWR(embd.Low); err != nil { //select write，选择写模式
		return err
	}

	//_nop_();
	//_nop_();

	if err := hd.fillDB8(dat); err != nil { //put data，放置数据
		return err
	}

	//_nop_();
	//_nop_();

	usDealy(5)
	//trigger WR rising edge to latch into LCD
	hd.writeWR(embd.High)

	return nil
}

func (hd *LCD8080) reset() error {
	seq := []int{embd.Low, embd.High}
	for _, sig := range seq {
		if err := hd.writeReset(sig); err != nil {
			return err
		}
	}
	usDealy(10)
	if err := hd.WriteCmd(cmdInternalReset); err != nil {
		return err
	}

	return nil
}

func (hd *LCD8080) writeReset(val int) error {
	return hd.Write(hd.RST, val)
}
func (hd *LCD8080) writeWR(val int) error {
	return hd.Write(hd.WR, val)
}
func (hd *LCD8080) writeRS(val int) error {
	return hd.Write(hd.RS, val)
}
func (hd *LCD8080) writeRD(val int) error {
	return hd.Write(hd.RD, val)
}
func (hd *LCD8080) WriteCS(val int) error {
	return hd.Write(hd.CS, val)
}

//  fillDB write value to DB0~7 GPIO
func (conn *LCD8080) fillDB8(value byte) error {
	functions := []func() error{
		func() error { return conn.Write(conn.DB0, int(value&0x01)) },
		func() error { return conn.Write(conn.DB1, int((value>>1)&0x01)) },
		func() error { return conn.Write(conn.DB2, int((value>>2)&0x01)) },
		func() error { return conn.Write(conn.DB3, int((value>>3)&0x01)) },
		func() error { return conn.Write(conn.DB4, int((value>>4)&0x01)) },
		func() error { return conn.Write(conn.DB5, int((value>>5)&0x01)) },
		func() error { return conn.Write(conn.DB6, int((value>>6)&0x01)) },
		func() error { return conn.Write(conn.DB7, int((value>>7)&0x01)) },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////

type LCD6800 struct {
	Connection
	CS, RW, RST, A0, E                     interface{}
	DB0, DB1, DB2, DB3, DB4, DB5, DB6, DB7 interface{}
}

// Init   initialize the st7565, the command is from the datasheet
func (hd *LCD6800) Init(cmds []byte) error {
	if err := hd.WriteCS(embd.Low); err != nil {
		return err
	}
	if err := hd.reset(); err != nil {
		return err
	}

	for _, cmd := range cmds {
		if err := hd.WriteCmd(cmd); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying Connection.
func (hd *LCD6800) Close() error {
	return hd.Connection.Close()
}

// WriteCmd for 8080 MPU
func (hd *LCD6800) WriteCmd(cmd byte) error {
	var err error

	if err = hd.WriteCS(embd.Low); err != nil { //chip select
		return err
	}

	if err = hd.writeRD(embd.High); err != nil { //disable read
		return err
	}

	if err = hd.writeRS(embd.Low); err != nil { //select command
		return err
	}

	if err = hd.writeWR(embd.Low); err != nil { //select write
		return err
	}

	//_nop_();
	//_nop_();
	if err = hd.fillDB8(cmd); err != nil {
		return err
	}
	usDealy(5)
	//trigger WR rising edge to latch into LCD
	if err = hd.writeWR(embd.High); err != nil { //command writing ，写入命令
		return err
	}
	return nil
}

// WriteData for 8080 MPU
func (hd *LCD6800) WriteData(dat byte) error {
	if err := hd.WriteCS(embd.Low); err != nil { //chip select,打开片选
		return err
	}

	if err := hd.writeRD(embd.High); err != nil { //disable read，读失能
		return err
	}

	if err := hd.writeRS(embd.High); err != nil { //select data，选择数据
		return err
	}

	if err := hd.writeWR(embd.Low); err != nil { //select write，选择写模式
		return err
	}

	//_nop_();
	//_nop_();

	if err := hd.fillDB8(dat); err != nil { //put data，放置数据
		return err
	}

	//_nop_();
	//_nop_();

	usDealy(5)
	//trigger WR rising edge to latch into LCD
	hd.writeWR(embd.High)

	return nil
}

func (hd *LCD6800) writeReset(val int) error {
	return hd.Write(hd.RST, val)
}
func (hd *LCD6800) writeRW(val int) error {
	return hd.Write(hd.RW, val)
}
func (hd *LCD6800) writeA0(val int) error {
	return hd.Write(hd.A0, val)
}
func (hd *LCD6800) writeE(val int) error {
	return hd.Write(hd.E, val)
}
func (hd *LCD6800) WriteCS(val int) error {
	return hd.Write(hd.CS, val)
}

//  fillDB write value to DB0~7 GPIO
func (conn *LCD6800) fillDB8(value byte) error {
	functions := []func() error{
		func() error { return conn.Write(conn.DB0, int(value&0x01)) },
		func() error { return conn.Write(conn.DB1, int((value>>1)&0x01)) },
		func() error { return conn.Write(conn.DB2, int((value>>2)&0x01)) },
		func() error { return conn.Write(conn.DB3, int((value>>3)&0x01)) },
		func() error { return conn.Write(conn.DB4, int((value>>4)&0x01)) },
		func() error { return conn.Write(conn.DB5, int((value>>5)&0x01)) },
		func() error { return conn.Write(conn.DB6, int((value>>6)&0x01)) },
		func() error { return conn.Write(conn.DB7, int((value>>7)&0x01)) },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

//commands from st7565 datasheet
const (
	//display on/of, datasheet P42
	cmdDisplyOFF = 0xAE
	cmdDisplyON  = 0xAF

	//display start line set, datasheet p42
	cmdSetDispStartLine = 0x40 //0x40~0x7F, map 0 to 63

	// page address set,datasheet p42
	cmdSetPageAddr = 0xB0 //0xB0~0xB8, map 0 to 8

	//column address set, datasheet p43
	cmdSetColumnUpper = 0x10
	cmdSetColumnLower = 0x00

	//ADC select(segment driver direction select), datasheet p44
	//ref p26 to know the detail
	cmdSetADCNormal  = 0xA0
	cmdSetADCReverse = 0xA1

	//display Normal/reverse, lit and unlit display without overwriting the content of the display data RAM
	//datasheet p44
	cmdSetDispNormal  = 0xA6
	cmdSetDispReverse = 0xA7

	//display all points on/off, datasheet p44
	cmdSetAllptsNormal = 0xA4
	cmdSetAllptsON     = 0xA5

	//LCD Bias set, , datasheet p45
	cmdSetLCDBias9 = 0xA2
	cmdSetLCDBias7 = 0xA3

	//Reset, datasheet p46
	cmdInternalReset = 0xE2
	/*
		sT7565_LCD_CMD_NOP               = 0xE3
		sT7565_LCD_CMD_RMW               = 0xE0
		sT7565_LCD_CMD_RMW_CLEAR         = 0xEE
		sT7565_LCD_CMD_SET_COM_NORMAL    = 0xC0
		sT7565_LCD_CMD_SET_STATIC_OFF    = 0xAC
		sT7565_LCD_CMD_SET_STATIC_ON     = 0xAD
		sT7565_LCD_CMD_SET_BOOSTER_FIRST = 0xF8
		sT7565_LCD_CMD_SET_BOOSTER_234   = 0
		sT7565_LCD_CMD_SET_BOOSTER_5     = 1
		sT7565_LCD_CMD_SET_BOOSTER_6     = 3
		sT7565_LCD_CMD_TEST              = 0xF0
	*/
	//common output mode select, datasheet p46
	cmdSetComNormal  = 0xC0
	cmdSetComReverse = 0xC8

	//power controller set, datasheet p47
	cmdSetPowerControl = 0x28
	cmdSetVolumeSECOND = 0x00
	cmdSetStaticOFF    = 0xAC
	cmdSetStaticON     = 0xAD
	cmdSetStaticREG    = 0x00

	//The Electronic Volume Mode Set  , datasheet p47
	// Once the electronic volume mode has been set, no other command except for the
	//electronic volume register  command (0~0x3f) can be used.
	cmdSetVolumeFIRST = 0x81

	//V0 Voltage Regulator Internal Resistor Ratio Set, datasheet p47
	//This command sets the V0 voltage regulator internal resistor ratio (0~7)
	cmdSetResistorRATIO = 0x20

	//Booster Ratio , datasheet p49
	cmdSetBoosterRatio = 0xF8
)

var defaultInitCmd []byte = []byte{
	(cmdSetADCNormal),  //设置SEG输出方向
	(cmdSetComReverse), //设置公共端输出扫描方向

	cmdSetVolumeFIRST,
	0x14,

	cmdSetResistorRATIO | 5,

	cmdSetPowerControl | 4 | 2 | 1,

	cmdSetBoosterRatio,
	0x1,

	cmdSetDispStartLine | 0, //   = 0x40 //0x40~0x7F, map 0 to 63

	cmdDisplyOFF,

	(cmdSetLCDBias7), //设置偏压比
	(cmdDisplyON),
}

type GPIOConnection struct {
	// Describers is a global list of registered GPIO.
	con map[interface{}]embd.DigitalPin
}

func newGPIOPins(pins ...interface{}) (*GPIOConnection, error) {

	var ll = GPIOConnection{}

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
		err := pin.SetDirection(embd.Out)
		if err != nil {
			glog.Errorf("hX8357: error setting pin %+v to out direction: %s", pin, err)
			return nil, err
		}
	}

	return &ll, nil
}

// Close closes all open DigitalPins.
func (ll *GPIOConnection) Close() error {
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
func (ll *GPIOConnection) Write(pin interface{}, val int) error {
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

func usDealy(val time.Duration) {
	time.Sleep(val * time.Microsecond)
	//Millisecond
}

/************************************************************************/
/* Font Definitions                                                     */
/************************************************************************/
/**
 * 5x7 LCD font 'flipped' for the ST7565 - public domain
 * @note This is a 256 character font. Delete glyphs in order to save Flash
 */
/* Controls the definition of the Font array and character spcing */
var Full_font = []byte{
	0x0, 0x0, 0x0, 0x0, 0x0, /* ASC(00) */
	0x7C, 0xDA, 0xF2, 0xDA, 0x7C, /* ASC(01) */
	0x7C, 0xD6, 0xF2, 0xD6, 0x7C, /* ASC(02) */
	0x38, 0x7C, 0x3E, 0x7C, 0x38, /* ASC(03) */
	0x18, 0x3C, 0x7E, 0x3C, 0x18, /* ASC(04) */
	0x38, 0xEA, 0xBE, 0xEA, 0x38, /* ASC(05) */
	0x38, 0x7A, 0xFE, 0x7A, 0x38, /* ASC(06) */
	0x0, 0x18, 0x3C, 0x18, 0x0, /* ASC(07) */
	0xFF, 0xE7, 0xC3, 0xE7, 0xFF, /* ASC(08) */
	0x0, 0x18, 0x24, 0x18, 0x0, /* ASC(09) */
	0xFF, 0xE7, 0xDB, 0xE7, 0xFF, /* ASC(10) */
	0xC, 0x12, 0x5C, 0x60, 0x70, /* ASC(11) */
	0x64, 0x94, 0x9E, 0x94, 0x64, /* ASC(12) */
	0x2, 0xFE, 0xA0, 0xA0, 0xE0, /* ASC(13) */
	0x2, 0xFE, 0xA0, 0xA4, 0xFC, /* ASC(14) */
	0x5A, 0x3C, 0xE7, 0x3C, 0x5A, /* ASC(15) */
	0xFE, 0x7C, 0x38, 0x38, 0x10, /* ASC(16) */
	0x10, 0x38, 0x38, 0x7C, 0xFE, /* ASC(17) */
	0x28, 0x44, 0xFE, 0x44, 0x28, /* ASC(18) */
	0xFA, 0xFA, 0x0, 0xFA, 0xFA, /* ASC(19) */
	0x60, 0x90, 0xFE, 0x80, 0xFE, /* ASC(20) */
	0x0, 0x66, 0x91, 0xA9, 0x56, /* ASC(21) */
	0x6, 0x6, 0x6, 0x6, 0x6, /* ASC(22) */
	0x29, 0x45, 0xFF, 0x45, 0x29, /* ASC(23) */
	0x10, 0x20, 0x7E, 0x20, 0x10, /* ASC(24) */
	0x8, 0x4, 0x7E, 0x4, 0x8, /* ASC(25) */
	0x10, 0x10, 0x54, 0x38, 0x10, /* ASC(26) */
	0x10, 0x38, 0x54, 0x10, 0x10, /* ASC(27) */
	0x78, 0x8, 0x8, 0x8, 0x8, /* ASC(28) */
	0x30, 0x78, 0x30, 0x78, 0x30, /* ASC(29) */
	0xC, 0x1C, 0x7C, 0x1C, 0xC, /* ASC(30) */
	0x60, 0x70, 0x7C, 0x70, 0x60, /* ASC(31) */
	0x3C, 0x64, 0xC4, 0x64, 0x3C, /* ASC(127) */
	0x78, 0x85, 0x85, 0x86, 0x48, /* ASC(128) */
	0x5C, 0x2, 0x2, 0x4, 0x5E, /* ASC(129) */
	0x1C, 0x2A, 0x2A, 0xAA, 0x9A, /* ASC(130) */
	0x84, 0xAA, 0xAA, 0x9E, 0x82, /* ASC(131) */
	0x84, 0x2A, 0x2A, 0x1E, 0x82, /* ASC(132) */
	0x84, 0xAA, 0x2A, 0x1E, 0x2, /* ASC(133) */
	0x4, 0x2A, 0xAA, 0x9E, 0x2, /* ASC(134) */
	0x30, 0x78, 0x4A, 0x4E, 0x48, /* ASC(135) */
	0x9C, 0xAA, 0xAA, 0xAA, 0x9A, /* ASC(136) */
	0x9C, 0x2A, 0x2A, 0x2A, 0x9A, /* ASC(137) */
	0x9C, 0xAA, 0x2A, 0x2A, 0x1A, /* ASC(138) */
	0x0, 0x0, 0xA2, 0x3E, 0x82, /* ASC(139) */
	0x0, 0x40, 0xA2, 0xBE, 0x42, /* ASC(140) */
	0x0, 0x80, 0xA2, 0x3E, 0x2, /* ASC(141) */
	0xF, 0x94, 0x24, 0x94, 0xF, /* ASC(142) */
	0xF, 0x14, 0xA4, 0x14, 0xF, /* ASC(143) */
	0x3E, 0x2A, 0xAA, 0xA2, 0x0, /* ASC(144) */
	0x4, 0x2A, 0x2A, 0x3E, 0x2A, /* ASC(145) */
	0x3E, 0x50, 0x90, 0xFE, 0x92, /* ASC(146) */
	0x4C, 0x92, 0x92, 0x92, 0x4C, /* ASC(147) */
	0x4C, 0x12, 0x12, 0x12, 0x4C, /* ASC(148) */
	0x4C, 0x52, 0x12, 0x12, 0xC, /* ASC(149) */
	0x5C, 0x82, 0x82, 0x84, 0x5E, /* ASC(150) */
	0x5C, 0x42, 0x2, 0x4, 0x1E, /* ASC(151) */
	0x0, 0xB9, 0x5, 0x5, 0xBE, /* ASC(152) */
	0x9C, 0x22, 0x22, 0x22, 0x9C, /* ASC(153) */
	0xBC, 0x2, 0x2, 0x2, 0xBC, /* ASC(154) */
	0x3C, 0x24, 0xFF, 0x24, 0x24, /* ASC(155) */
	0x12, 0x7E, 0x92, 0xC2, 0x66, /* ASC(156) */
	0xD4, 0xF4, 0x3F, 0xF4, 0xD4, /* ASC(157) */
	0xFF, 0x90, 0x94, 0x6F, 0x4, /* ASC(158) */
	0x3, 0x11, 0x7E, 0x90, 0xC0, /* ASC(159) */
	0x4, 0x2A, 0x2A, 0x9E, 0x82, /* ASC(160) */
	0x0, 0x0, 0x22, 0xBE, 0x82, /* ASC(161) */
	0xC, 0x12, 0x12, 0x52, 0x4C, /* ASC(162) */
	0x1C, 0x2, 0x2, 0x44, 0x5E, /* ASC(163) */
	0x0, 0x5E, 0x50, 0x50, 0x4E, /* ASC(164) */
	0xBE, 0xB0, 0x98, 0x8C, 0xBE, /* ASC(165) */
	0x64, 0x94, 0x94, 0xF4, 0x14, /* ASC(166) */
	0x64, 0x94, 0x94, 0x94, 0x64, /* ASC(167) */
	0xC, 0x12, 0xB2, 0x2, 0x4, /* ASC(168) */
	0x1C, 0x10, 0x10, 0x10, 0x10, /* ASC(169) */
	0x10, 0x10, 0x10, 0x10, 0x1C, /* ASC(170) */
	0xF4, 0x8, 0x13, 0x35, 0x5D, /* ASC(171) */
	0xF4, 0x8, 0x14, 0x2C, 0x5F, /* ASC(172) */
	0x0, 0x0, 0xDE, 0x0, 0x0, /* ASC(173) */
	0x10, 0x28, 0x54, 0x28, 0x44, /* ASC(174) */
	0x44, 0x28, 0x54, 0x28, 0x10, /* ASC(175) */
	0x55, 0x0, 0xAA, 0x0, 0x55, /* ASC(176) */
	0x55, 0xAA, 0x55, 0xAA, 0x55, /* ASC(177) */
	0xAA, 0x55, 0xAA, 0x55, 0xAA, /* ASC(178) */
	0x0, 0x0, 0x0, 0xFF, 0x0, /* ASC(179) */
	0x8, 0x8, 0x8, 0xFF, 0x0, /* ASC(180) */
	0x28, 0x28, 0x28, 0xFF, 0x0, /* ASC(181) */
	0x8, 0x8, 0xFF, 0x0, 0xFF, /* ASC(182) */
	0x8, 0x8, 0xF, 0x8, 0xF, /* ASC(183) */
	0x28, 0x28, 0x28, 0x3F, 0x0, /* ASC(184) */
	0x28, 0x28, 0xEF, 0x0, 0xFF, /* ASC(185) */
	0x0, 0x0, 0xFF, 0x0, 0xFF, /* ASC(186) */
	0x28, 0x28, 0x2F, 0x20, 0x3F, /* ASC(187) */
	0x28, 0x28, 0xE8, 0x8, 0xF8, /* ASC(188) */
	0x8, 0x8, 0xF8, 0x8, 0xF8, /* ASC(189) */
	0x28, 0x28, 0x28, 0xF8, 0x0, /* ASC(190) */
	0x8, 0x8, 0x8, 0xF, 0x0, /* ASC(191) */
	0x0, 0x0, 0x0, 0xF8, 0x8, /* ASC(192) */
	0x8, 0x8, 0x8, 0xF8, 0x8, /* ASC(193) */
	0x8, 0x8, 0x8, 0xF, 0x8, /* ASC(194) */
	0x0, 0x0, 0x0, 0xFF, 0x8, /* ASC(195) */
	0x8, 0x8, 0x8, 0x8, 0x8, /* ASC(196) */
	0x8, 0x8, 0x8, 0xFF, 0x8, /* ASC(197) */
	0x0, 0x0, 0x0, 0xFF, 0x28, /* ASC(198) */
	0x0, 0x0, 0xFF, 0x0, 0xFF, /* ASC(199) */
	0x0, 0x0, 0xF8, 0x8, 0xE8, /* ASC(200) */
	0x0, 0x0, 0x3F, 0x20, 0x2F, /* ASC(201) */
	0x28, 0x28, 0xE8, 0x8, 0xE8, /* ASC(202) */
	0x28, 0x28, 0x2F, 0x20, 0x2F, /* ASC(203) */
	0x0, 0x0, 0xFF, 0x0, 0xEF, /* ASC(204) */
	0x28, 0x28, 0x28, 0x28, 0x28, /* ASC(205) */
	0x28, 0x28, 0xEF, 0x0, 0xEF, /* ASC(206) */
	0x28, 0x28, 0x28, 0xE8, 0x28, /* ASC(207) */
	0x8, 0x8, 0xF8, 0x8, 0xF8, /* ASC(208) */
	0x28, 0x28, 0x28, 0x2F, 0x28, /* ASC(209) */
	0x8, 0x8, 0xF, 0x8, 0xF, /* ASC(210) */
	0x0, 0x0, 0xF8, 0x8, 0xF8, /* ASC(211) */
	0x0, 0x0, 0x0, 0xF8, 0x28, /* ASC(212) */
	0x0, 0x0, 0x0, 0x3F, 0x28, /* ASC(213) */
	0x0, 0x0, 0xF, 0x8, 0xF, /* ASC(214) */
	0x8, 0x8, 0xFF, 0x8, 0xFF, /* ASC(215) */
	0x28, 0x28, 0x28, 0xFF, 0x28, /* ASC(216) */
	0x8, 0x8, 0x8, 0xF8, 0x0, /* ASC(217) */
	0x0, 0x0, 0x0, 0xF, 0x8, /* ASC(218) */
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, /* ASC(219) */
	0xF, 0xF, 0xF, 0xF, 0xF, /* ASC(220) */
	0xFF, 0xFF, 0xFF, 0x0, 0x0, /* ASC(221) */
	0x0, 0x0, 0x0, 0xFF, 0xFF, /* ASC(222) */
	0xF0, 0xF0, 0xF0, 0xF0, 0xF0, /* ASC(223) */
	0x1C, 0x22, 0x22, 0x1C, 0x22, /* ASC(224) */
	0x3E, 0x54, 0x54, 0x7C, 0x28, /* ASC(225) */
	0x7E, 0x40, 0x40, 0x60, 0x60, /* ASC(226) */
	0x40, 0x7E, 0x40, 0x7E, 0x40, /* ASC(227) */
	0xC6, 0xAA, 0x92, 0x82, 0xC6, /* ASC(228) */
	0x1C, 0x22, 0x22, 0x3C, 0x20, /* ASC(229) */
	0x2, 0x7E, 0x4, 0x78, 0x4, /* ASC(230) */
	0x60, 0x40, 0x7E, 0x40, 0x40, /* ASC(231) */
	0x99, 0xA5, 0xE7, 0xA5, 0x99, /* ASC(232) */
	0x38, 0x54, 0x92, 0x54, 0x38, /* ASC(233) */
	0x32, 0x4E, 0x80, 0x4E, 0x32, /* ASC(234) */
	0xC, 0x52, 0xB2, 0xB2, 0xC, /* ASC(235) */
	0xC, 0x12, 0x1E, 0x12, 0xC, /* ASC(236) */
	0x3D, 0x46, 0x5A, 0x62, 0xBC, /* ASC(237) */
	0x7C, 0x92, 0x92, 0x92, 0x0, /* ASC(238) */
	0x7E, 0x80, 0x80, 0x80, 0x7E, /* ASC(239) */
	0x54, 0x54, 0x54, 0x54, 0x54, /* ASC(240) */
	0x22, 0x22, 0xFA, 0x22, 0x22, /* ASC(241) */
	0x2, 0x8A, 0x52, 0x22, 0x2, /* ASC(242) */
	0x2, 0x22, 0x52, 0x8A, 0x2, /* ASC(243) */
	0x0, 0x0, 0xFF, 0x80, 0xC0, /* ASC(244) */
	0x7, 0x1, 0xFF, 0x0, 0x0, /* ASC(245) */
	0x10, 0x10, 0xD6, 0xD6, 0x10, /* ASC(246) */
	0x6C, 0x48, 0x6C, 0x24, 0x6C, /* ASC(247) */
	0x60, 0xF0, 0x90, 0xF0, 0x60, /* ASC(248) */
	0x0, 0x0, 0x18, 0x18, 0x0, /* ASC(249) */
	0x0, 0x0, 0x8, 0x8, 0x0, /* ASC(250) */
	0xC, 0x2, 0xFF, 0x80, 0x80, /* ASC(251) */
	0x0, 0xF8, 0x80, 0x80, 0x78, /* ASC(252) */
	0x0, 0x98, 0xB8, 0xE8, 0x48, /* ASC(253) */
	0x0, 0x3C, 0x3C, 0x3C, 0x3C /* ASC(254) */}

var Gca_font = []byte{
	0x0, 0x0, 0x0, 0x0, 0x0, /* ASC(32) */
	0x0, 0x0, 0xFA, 0x0, 0x0, /* ASC(33) */
	0x0, 0xE0, 0x0, 0xE0, 0x0, /* ASC(34) */
	0x28, 0xFE, 0x28, 0xFE, 0x28, /* ASC(35) */
	0x24, 0x54, 0xFE, 0x54, 0x48, /* ASC(36) */
	0xC4, 0xC8, 0x10, 0x26, 0x46, /* ASC(37) */
	0x6C, 0x92, 0x6A, 0x4, 0xA, /* ASC(38) */
	0x0, 0x10, 0xE0, 0xC0, 0x0, /* ASC(39) */
	0x0, 0x38, 0x44, 0x82, 0x0, /* ASC(40) */
	0x0, 0x82, 0x44, 0x38, 0x0, /* ASC(41) */
	0x54, 0x38, 0xFE, 0x38, 0x54, /* ASC(42) */
	0x10, 0x10, 0x7C, 0x10, 0x10, /* ASC(43) */
	0x0, 0x1, 0xE, 0xC, 0x0, /* ASC(44) */
	0x10, 0x10, 0x10, 0x10, 0x10, /* ASC(45) */
	0x0, 0x0, 0x6, 0x6, 0x0, /* ASC(46) */
	0x4, 0x8, 0x10, 0x20, 0x40, /* ASC(47) */
	0x7C, 0x8A, 0x92, 0xA2, 0x7C, /* ASC(48) */
	0x0, 0x42, 0xFE, 0x2, 0x0, /* ASC(49) */
	0x4E, 0x92, 0x92, 0x92, 0x62, /* ASC(50) */
	0x84, 0x82, 0x92, 0xB2, 0xCC, /* ASC(51) */
	0x18, 0x28, 0x48, 0xFE, 0x8, /* ASC(52) */
	0xE4, 0xA2, 0xA2, 0xA2, 0x9C, /* ASC(53) */
	0x3C, 0x52, 0x92, 0x92, 0x8C, /* ASC(54) */
	0x82, 0x84, 0x88, 0x90, 0xE0, /* ASC(55) */
	0x6C, 0x92, 0x92, 0x92, 0x6C, /* ASC(56) */
	0x62, 0x92, 0x92, 0x94, 0x78, /* ASC(57) */
	0x0, 0x0, 0x28, 0x0, 0x0, /* ASC(58) */
	0x0, 0x2, 0x2C, 0x0, 0x0, /* ASC(59) */
	0x0, 0x10, 0x28, 0x44, 0x82, /* ASC(60) */
	0x28, 0x28, 0x28, 0x28, 0x28, /* ASC(61) */
	0x0, 0x82, 0x44, 0x28, 0x10, /* ASC(62) */
	0x40, 0x80, 0x9A, 0x90, 0x60, /* ASC(63) */
	0x7C, 0x82, 0xBA, 0x9A, 0x72, /* ASC(64) */
	0x3E, 0x48, 0x88, 0x48, 0x3E, /* ASC(65) */
	0xFE, 0x92, 0x92, 0x92, 0x6C, /* ASC(66) */
	0x7C, 0x82, 0x82, 0x82, 0x44, /* ASC(67) */
	0xFE, 0x82, 0x82, 0x82, 0x7C, /* ASC(68) */
	0xFE, 0x92, 0x92, 0x92, 0x82, /* ASC(69) */
	0xFE, 0x90, 0x90, 0x90, 0x80, /* ASC(70) */
	0x7C, 0x82, 0x82, 0x8A, 0xCE, /* ASC(71) */
	0xFE, 0x10, 0x10, 0x10, 0xFE, /* ASC(72) */
	0x0, 0x82, 0xFE, 0x82, 0x0, /* ASC(73) */
	0x4, 0x2, 0x82, 0xFC, 0x80, /* ASC(74) */
	0xFE, 0x10, 0x28, 0x44, 0x82, /* ASC(75) */
	0xFE, 0x2, 0x2, 0x2, 0x2, /* ASC(76) */
	0xFE, 0x40, 0x38, 0x40, 0xFE, /* ASC(77) */
	0xFE, 0x20, 0x10, 0x8, 0xFE, /* ASC(78) */
	0x7C, 0x82, 0x82, 0x82, 0x7C, /* ASC(79) */
	0xFE, 0x90, 0x90, 0x90, 0x60, /* ASC(80) */
	0x7C, 0x82, 0x8A, 0x84, 0x7A, /* ASC(81) */
	0xFE, 0x90, 0x98, 0x94, 0x62, /* ASC(82) */
	0x64, 0x92, 0x92, 0x92, 0x4C, /* ASC(83) */
	0xC0, 0x80, 0xFE, 0x80, 0xC0, /* ASC(84) */
	0xFC, 0x2, 0x2, 0x2, 0xFC, /* ASC(85) */
	0xF8, 0x4, 0x2, 0x4, 0xF8, /* ASC(86) */
	0xFC, 0x2, 0x1C, 0x2, 0xFC, /* ASC(87) */
	0xC6, 0x28, 0x10, 0x28, 0xC6, /* ASC(88) */
	0xC0, 0x20, 0x1E, 0x20, 0xC0, /* ASC(89) */
	0x86, 0x9A, 0x92, 0xB2, 0xC2, /* ASC(90) */
	0x0, 0xFE, 0x82, 0x82, 0x82, /* ASC(91) */
	0x40, 0x20, 0x10, 0x8, 0x4, /* ASC(92) */
	0x0, 0x82, 0x82, 0x82, 0xFE, /* ASC(93) */
	0x20, 0x40, 0x80, 0x40, 0x20, /* ASC(94) */
	0x2, 0x2, 0x2, 0x2, 0x2, /* ASC(95) */
	0x0, 0xC0, 0xE0, 0x10, 0x0, /* ASC(96) */
	0x4, 0x2A, 0x2A, 0x1E, 0x2, /* ASC(97) */
	0xFE, 0x14, 0x22, 0x22, 0x1C, /* ASC(98) */
	0x1C, 0x22, 0x22, 0x22, 0x14, /* ASC(99) */
	0x1C, 0x22, 0x22, 0x14, 0xFE, /* ASC(100) */
	0x1C, 0x2A, 0x2A, 0x2A, 0x18, /* ASC(101) */
	0x0, 0x10, 0x7E, 0x90, 0x40, /* ASC(102) */
	0x18, 0x25, 0x25, 0x39, 0x1E, /* ASC(103) */
	0xFE, 0x10, 0x20, 0x20, 0x1E, /* ASC(104) */
	0x0, 0x22, 0xBE, 0x2, 0x0, /* ASC(105) */
	0x4, 0x2, 0x2, 0xBC, 0x0, /* ASC(106) */
	0xFE, 0x8, 0x14, 0x22, 0x0, /* ASC(107) */
	0x0, 0x82, 0xFE, 0x2, 0x0, /* ASC(108) */
	0x3E, 0x20, 0x1E, 0x20, 0x1E, /* ASC(109) */
	0x3E, 0x10, 0x20, 0x20, 0x1E, /* ASC(110) */
	0x1C, 0x22, 0x22, 0x22, 0x1C, /* ASC(111) */
	0x3F, 0x18, 0x24, 0x24, 0x18, /* ASC(112) */
	0x18, 0x24, 0x24, 0x18, 0x3F, /* ASC(113) */
	0x3E, 0x10, 0x20, 0x20, 0x10, /* ASC(114) */
	0x12, 0x2A, 0x2A, 0x2A, 0x24, /* ASC(115) */
	0x20, 0x20, 0xFC, 0x22, 0x24, /* ASC(116) */
	0x3C, 0x2, 0x2, 0x4, 0x3E, /* ASC(117) */
	0x38, 0x4, 0x2, 0x4, 0x38, /* ASC(118) */
	0x3C, 0x2, 0xC, 0x2, 0x3C, /* ASC(119) */
	0x22, 0x14, 0x8, 0x14, 0x22, /* ASC(120) */
	0x32, 0x9, 0x9, 0x9, 0x3E, /* ASC(121) */
	0x22, 0x26, 0x2A, 0x32, 0x22, /* ASC(122) */
	0x0, 0x10, 0x6C, 0x82, 0x0, /* ASC(123) */
	0x0, 0x0, 0xEE, 0x0, 0x0, /* ASC(124) */
	0x0, 0x82, 0x6C, 0x10, 0x0, /* ASC(125) */
	0x40, 0x80, 0x40, 0x20, 0x40 /* ASC(126) */}
