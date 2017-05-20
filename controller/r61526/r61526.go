package r61526

import (
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"

	"fmt"
)

/*

 */
type LCD struct {
	R61526
}

const (
	// LCD Parameters
	LcdWidth  = 319
	LcdHeight = 479

	WHITE   = 0xFFFF
	BLACK   = 0x0000
	BLUE    = 0x001F
	RED     = 0xF800
	MAGENTA = 0xF81F
	GREEN   = 0x07E0
	CYAN    = 0x7FFF
	YELLOW  = 0xFFE0
)

//SetCursor
func (hd *LCD) SetCursor(xStart, yStart, xEnd, yEnd uint16) error {
	//if x >= LcdWidth || x < 0 { //
	//	return errors.New("according to datasheet(P43), beyond")
	//}
	//if y > LcdHeight || y < 0 { //
	//	return errors.New("according to datasheet(P42), not beyond 8 pages")
	//}

	if err := hd.R61526.WriteCmd(cmdSetColumnAddr); err != nil {
		return err
	}
	if err := hd.R61526.WriteData16(xStart); err != nil {
		return err
	}
	if err := hd.R61526.WriteData16(xEnd); err != nil {
		return err
	}

	if err := hd.R61526.WriteCmd(cmdSetPageAddr); err != nil {
		return err
	}
	if err := hd.R61526.WriteData16(yStart); err != nil {
		return err
	}
	if err := hd.R61526.WriteData16(yEnd); err != nil {
		return err
	}

	if err := hd.R61526.WriteCmd(0x2c); err != nil {
		return err
	}
	return nil
}

func (hd *LCD) TFT_Changegrb(color uint16) (p uint16) {

	red := (color & 0x1F)        //5 bits
	green := (color >> 5) & 0x3F //6 bits
	blue := (color >> 11) & 0x1F //5 bits

	p = ((red << 11) | (green << 6) | blue)
	return
}

// Clear clears the display and mode settings sets the cursor to the home position.
func (hd *LCD) Clear(color uint16) error {
	grb := hd.TFT_Changegrb(color)

	if err := hd.SetCursor(0, 0, LcdWidth, LcdHeight); err != nil {
		return err
	}

	var i uint16
	for i = 0; i < LcdHeight; i++ {
		for j := 0; j < LcdWidth; j++ {
			hd.R61526.WriteData16(grb)

		}
	}

	return nil
}

func (hd *LCD) ExecUserCmd(cmd byte, paramenters ...byte) error {
	if err := hd.R61526.WriteCmd(cmd); err != nil {
		return err
	}

	for _, param := range paramenters {
		if err := hd.R61526.WriteData(param); err != nil {
			return err
		}
	}
	return nil
}

// Init   initialize the st7565,
//special, if the cmd is 0xff, the func will deay 5us
func (hd *LCD) Init(cmds []byte) error {

	if err := hd.R61526.WriteCS(embd.Low); err != nil { //chip select
		return err
	}

	seq := []int{embd.Low, embd.High}
	for _, sig := range seq {
		if err := hd.R61526.WriteRST(sig); err != nil {
			return err
		}
	}
	usDealy(10)
	hd.ExecUserCmd(0xB0, 0x3F, 0x3F)
	usDealy(5)

	hd.ExecUserCmd(0xB3, 0x02, 0x00, 0x00, 0x00, 0x00)

	hd.ExecUserCmd(0xB4, 0x00)

	hd.ExecUserCmd(0xC0, 0x33, 0x4F, 0x00, 0x10, 0xA2, 0x00, 0x01, 0x00)

	hd.ExecUserCmd(0xC1, 0x01, 0x02, 0x20, 0x08, 0x08)
	usDealy(50)

	hd.ExecUserCmd(0xC3, 0x01, 0x00, 0x28, 0x08, 0x08)
	usDealy(5)

	hd.ExecUserCmd(0xC4, 0x11, 0x01, 0x23, 0x04, 0x00)

	hd.ExecUserCmd(0xC8,
		0x05, 0x0C, 0x0b, 0x15, 0x11, 0x09, 0x05, 0x07, 0x13, 0x10, 0x20,
		0x13, 0x07, 0x05, 0x09, 0x11, 0x15, 0x0b, 0x0c, 0x05, 0x05, 0x02)

	hd.ExecUserCmd(0xC9,
		0x05, 0x0C, 0x05, 0x15, 0x11, 0x09, 0x05, 0x07, 0x13, 0x10, 0x20,
		0x13, 0x07, 0x05, 0x09, 0x11, 0x15, 0x0b, 0x0c, 0x05, 0x05, 0x02)

	hd.ExecUserCmd(0xCA,
		0x05, 0x0C, 0x0b, 0x15, 0x11, 0x09, 0x05, 0x07, 0x13, 0x10, 0x20,
		0x13, 0x07, 0x05, 0x09, 0x11, 0x15, 0x0b, 0x0c, 0x05, 0x05, 0x02)

	hd.ExecUserCmd(0xD0,
		0x33, 0x53, 0x87, 0x3b, 0x30, 0x00)

	hd.ExecUserCmd(0xD1, 0x2c, 0x61, 0x10)

	hd.ExecUserCmd(0xD2, 0x03, 0x24)

	hd.ExecUserCmd(0xD4, 0x03, 0x24)

	hd.ExecUserCmd(0xE2, 0x3f)
	usDealy(5)

	hd.ExecUserCmd(0x35, 0x00)

	hd.ExecUserCmd(0x36, 0x40)

	hd.ExecUserCmd(0x3A, 0x55) //55 16bit color

	hd.ExecUserCmd(cmdSetColumnAddr, 0x00, 0x00, 0x00, 0xEF)

	hd.ExecUserCmd(cmdSetPageAddr, 0x00, 0x00, 0x01, 0x3F)

	hd.ExecUserCmd(0x11)
	usDealy(5)
	hd.ExecUserCmd(cmdSetDispOn) //TFT_WriteCmd(0x29);
	usDealy(5)
	hd.ExecUserCmd(0x2C) //TFT_WriteCmd(0x2C) ;
	usDealy(5)

	hd.R61526.WriteCS(embd.High) //diable chip select

	return nil
}

func (hd *LCD) Close() error {
	return hd.R61526.Close()
}

//Writeascii168Str   write string, the char size 8X8 in the string
func (hd *LCD) Writeascii168Str(str string, x /*column*/, y /*page*/ uint16) error {

	for index, by := range str {
		c := byte(by - 32)
		ind := uint16(index)
		hd.lcd_ascii168(c, x+(ind*8), y)
	}
	return nil

}

func (hd *LCD) lcd_ascii168(char byte, x /*column*/, y /*page*/ uint16) error {
	//hd.SetCursor(x , y )

	val := Ascii168[char]
	for i := 0; i < 8; i++ {
		hd.WriteData(val[i])
	}
	//hd.SetCursor(x /*column*/, y+1 /*page*/)
	for i := 8; i < 16; i++ {
		hd.WriteData(val[i])
	}
	return nil
}

// ST7565 represents an ST7565-compatible character LCD controller.
type R61526 interface {
	Close() error

	WriteCmd(cmd byte) error
	WriteCS(val int) error
	WriteRST(val int) error
	WriteData(cmd byte) error
	WriteData16(cmd uint16) error
}

//NewGpio
//m: input Parallel8080 or Parallel6800
//cmds: initial command ,if it is nil, it will initilized with defaultInitCmd
func NewGpio(m interface{}, cmds ...byte) (*LCD, error) {

	var lcd LCD
	switch inst := m.(type) {
	case Parallel8080:
		if con, err := newGPIOPins(inst.CS, inst.WR, inst.RST, inst.RS, inst.RD,
			inst.DB0, inst.DB1, inst.DB2, inst.DB3,
			inst.DB4, inst.DB5, inst.DB6, inst.DB7); err == nil {
			inst.Connection = con
			lcd.R61526 = &inst
			if err := lcd.Init(cmds); err != nil {
				return nil, fmt.Errorf("init fail")
			}

			return &lcd, nil
		}
	}
	return nil, fmt.Errorf("unknow")
}

// Close closes the underlying Connection.
func (hd *Parallel8080) Close() error {
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

	PinInst(pin interface{}) *embd.DigitalPin
}

type Parallel8080 struct {
	Connection
	CS, WR, RST, RS, RD                    interface{}
	DB0, DB1, DB2, DB3, DB4, DB5, DB6, DB7 interface{}
}

/*
//---¶¨ÒåÊ¹ÓÃµÄIO¿Ú---//
sbit	    TFT_RS  = P3^2;	  //Êý¾ÝÃüÁîÑ¡Ôñ¶Ë
sbit	    TFT_RST = P3^3;   //¸´Î»
sbit	    TFT_WR  = P2^5;	  //¶ÁÐ´¿ØÖÆ
sbit        TFT_RD  = P2^6;   //¶ÁÐ´¿ØÖÆ
sbit	    TFT_CS  = P2^7;	  //Æ¬Ñ¡
*/

var (
	DefaultMap8080 Parallel8080 = Parallel8080{
		RS:  "P1_11",
		RST: "P1_12",
		WR:  "P1_15",
		RD:  "P1_40",
		CS:  "P1_7",

		DB0: "P1_19", DB1: "P1_21", DB2: "P1_23", DB3: "P1_16",
		DB4: "P1_18", DB5: "P1_22", DB6: "P1_24", DB7: "P1_26"}
	//8X8 ascii
	Ascii168 = [...][16]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, //" ",0
		{0x00, 0x00, 0x00, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x33, 0x30, 0x00, 0x00, 0x00}, //"!",1
		{0x00, 0x10, 0x0C, 0x06, 0x10, 0x0C, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, //""",2
		{0x40, 0xC0, 0x78, 0x40, 0xC0, 0x78, 0x40, 0x00, 0x04, 0x3F, 0x04, 0x04, 0x3F, 0x04, 0x04, 0x00}, //"#",3
		{0x00, 0x70, 0x88, 0xFC, 0x08, 0x30, 0x00, 0x00, 0x00, 0x18, 0x20, 0xFF, 0x21, 0x1E, 0x00, 0x00}, //"$",4
		{0xF0, 0x08, 0xF0, 0x00, 0xE0, 0x18, 0x00, 0x00, 0x00, 0x21, 0x1C, 0x03, 0x1E, 0x21, 0x1E, 0x00}, //"%",5
		{0x00, 0xF0, 0x08, 0x88, 0x70, 0x00, 0x00, 0x00, 0x1E, 0x21, 0x23, 0x24, 0x19, 0x27, 0x21, 0x10}, //"&",6
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xB0, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00}, //",",7
		{0x00, 0x00, 0x00, 0xE0, 0x18, 0x04, 0x02, 0x00, 0x00, 0x00, 0x00, 0x07, 0x18, 0x20, 0x40, 0x00}, //"(",8
		{0x00, 0x02, 0x04, 0x18, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x40, 0x20, 0x18, 0x07, 0x00, 0x00, 0x00}, //")",9
		{0x40, 0x40, 0x80, 0xF0, 0x80, 0x40, 0x40, 0x00, 0x02, 0x02, 0x01, 0x0F, 0x01, 0x02, 0x02, 0x00}, //"*",10
		{0x00, 0x00, 0x00, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x1F, 0x01, 0x01, 0x01, 0x00}, //"+",11
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xB0, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00}, //",",12
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}, //"-",13
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00}, //".",14
		{0x00, 0x00, 0x00, 0x00, 0x80, 0x60, 0x18, 0x04, 0x00, 0x60, 0x18, 0x06, 0x01, 0x00, 0x00, 0x00}, //"/",15
		{0x00, 0xE0, 0x10, 0x08, 0x08, 0x10, 0xE0, 0x00, 0x00, 0x0F, 0x10, 0x20, 0x20, 0x10, 0x0F, 0x00}, //"0",16
		{0x00, 0x10, 0x10, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x20, 0x3F, 0x20, 0x20, 0x00, 0x00}, //"1",17
		{0x00, 0x70, 0x08, 0x08, 0x08, 0x88, 0x70, 0x00, 0x00, 0x30, 0x28, 0x24, 0x22, 0x21, 0x30, 0x00}, //"2",18
		{0x00, 0x30, 0x08, 0x88, 0x88, 0x48, 0x30, 0x00, 0x00, 0x18, 0x20, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"3",19
		{0x00, 0x00, 0xC0, 0x20, 0x10, 0xF8, 0x00, 0x00, 0x00, 0x07, 0x04, 0x24, 0x24, 0x3F, 0x24, 0x00}, //"4",20
		{0x00, 0xF8, 0x08, 0x88, 0x88, 0x08, 0x08, 0x00, 0x00, 0x19, 0x21, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"5",21
		{0x00, 0xE0, 0x10, 0x88, 0x88, 0x18, 0x00, 0x00, 0x00, 0x0F, 0x11, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"6",22
		{0x00, 0x38, 0x08, 0x08, 0xC8, 0x38, 0x08, 0x00, 0x00, 0x00, 0x00, 0x3F, 0x00, 0x00, 0x00, 0x00}, //"7",23
		{0x00, 0x70, 0x88, 0x08, 0x08, 0x88, 0x70, 0x00, 0x00, 0x1C, 0x22, 0x21, 0x21, 0x22, 0x1C, 0x00}, //"8",24
		{0x00, 0xE0, 0x10, 0x08, 0x08, 0x10, 0xE0, 0x00, 0x00, 0x00, 0x31, 0x22, 0x22, 0x11, 0x0F, 0x00}, //"9",25
		{0x00, 0x00, 0x00, 0xC0, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x30, 0x00, 0x00, 0x00}, //":",26
		{0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x60, 0x00, 0x00, 0x00, 0x00}, //";",27
		{0x00, 0x00, 0x80, 0x40, 0x20, 0x10, 0x08, 0x00, 0x00, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x00}, //"<",28
		{0x40, 0x40, 0x40, 0x40, 0x40, 0x40, 0x40, 0x00, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x00}, //"=",29
		{0x00, 0x08, 0x10, 0x20, 0x40, 0x80, 0x00, 0x00, 0x00, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01, 0x00}, //">",30
		{0x00, 0x70, 0x48, 0x08, 0x08, 0x08, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x30, 0x36, 0x01, 0x00, 0x00}, //"?",31
		{0xC0, 0x30, 0xC8, 0x28, 0xE8, 0x10, 0xE0, 0x00, 0x07, 0x18, 0x27, 0x24, 0x23, 0x14, 0x0B, 0x00}, //"@",32
		{0x00, 0x00, 0xC0, 0x38, 0xE0, 0x00, 0x00, 0x00, 0x20, 0x3C, 0x23, 0x02, 0x02, 0x27, 0x38, 0x20}, //"A",33
		{0x08, 0xF8, 0x88, 0x88, 0x88, 0x70, 0x00, 0x00, 0x20, 0x3F, 0x20, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"B",34
		{0xC0, 0x30, 0x08, 0x08, 0x08, 0x08, 0x38, 0x00, 0x07, 0x18, 0x20, 0x20, 0x20, 0x10, 0x08, 0x00}, //"C",35
		{0x08, 0xF8, 0x08, 0x08, 0x08, 0x10, 0xE0, 0x00, 0x20, 0x3F, 0x20, 0x20, 0x20, 0x10, 0x0F, 0x00}, //"D",36
		{0x08, 0xF8, 0x88, 0x88, 0xE8, 0x08, 0x10, 0x00, 0x20, 0x3F, 0x20, 0x20, 0x23, 0x20, 0x18, 0x00}, //"E",37
		{0x08, 0xF8, 0x88, 0x88, 0xE8, 0x08, 0x10, 0x00, 0x20, 0x3F, 0x20, 0x00, 0x03, 0x00, 0x00, 0x00}, //"F",38
		{0xC0, 0x30, 0x08, 0x08, 0x08, 0x38, 0x00, 0x00, 0x07, 0x18, 0x20, 0x20, 0x22, 0x1E, 0x02, 0x00}, //"G",39
		{0x08, 0xF8, 0x08, 0x00, 0x00, 0x08, 0xF8, 0x08, 0x20, 0x3F, 0x21, 0x01, 0x01, 0x21, 0x3F, 0x20}, //"H",40
		{0x00, 0x08, 0x08, 0xF8, 0x08, 0x08, 0x00, 0x00, 0x00, 0x20, 0x20, 0x3F, 0x20, 0x20, 0x00, 0x00}, //"I",41
		{0x00, 0x00, 0x08, 0x08, 0xF8, 0x08, 0x08, 0x00, 0xC0, 0x80, 0x80, 0x80, 0x7F, 0x00, 0x00, 0x00}, //"J",42
		{0x08, 0xF8, 0x88, 0xC0, 0x28, 0x18, 0x08, 0x00, 0x20, 0x3F, 0x20, 0x01, 0x26, 0x38, 0x20, 0x00}, //"K",43
		{0x08, 0xF8, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x3F, 0x20, 0x20, 0x20, 0x20, 0x30, 0x00}, //"L",44
		{0x08, 0xF8, 0xF8, 0x00, 0xF8, 0xF8, 0x08, 0x00, 0x20, 0x3F, 0x00, 0x3F, 0x00, 0x3F, 0x20, 0x00}, //"M",45
		{0x08, 0xF8, 0x30, 0xC0, 0x00, 0x08, 0xF8, 0x08, 0x20, 0x3F, 0x20, 0x00, 0x07, 0x18, 0x3F, 0x00}, //"N",46
		{0xE0, 0x10, 0x08, 0x08, 0x08, 0x10, 0xE0, 0x00, 0x0F, 0x10, 0x20, 0x20, 0x20, 0x10, 0x0F, 0x00}, //"O",47
		{0x08, 0xF8, 0x08, 0x08, 0x08, 0x08, 0xF0, 0x00, 0x20, 0x3F, 0x21, 0x01, 0x01, 0x01, 0x00, 0x00}, //"P",48
		{0xE0, 0x10, 0x08, 0x08, 0x08, 0x10, 0xE0, 0x00, 0x0F, 0x18, 0x24, 0x24, 0x38, 0x50, 0x4F, 0x00}, //"Q",49
		{0x08, 0xF8, 0x88, 0x88, 0x88, 0x88, 0x70, 0x00, 0x20, 0x3F, 0x20, 0x00, 0x03, 0x0C, 0x30, 0x20}, //"R",50
		{0x00, 0x70, 0x88, 0x08, 0x08, 0x08, 0x38, 0x00, 0x00, 0x38, 0x20, 0x21, 0x21, 0x22, 0x1C, 0x00}, //"S",51
		{0x18, 0x08, 0x08, 0xF8, 0x08, 0x08, 0x18, 0x00, 0x00, 0x00, 0x20, 0x3F, 0x20, 0x00, 0x00, 0x00}, //"T",52
		{0x08, 0xF8, 0x08, 0x00, 0x00, 0x08, 0xF8, 0x08, 0x00, 0x1F, 0x20, 0x20, 0x20, 0x20, 0x1F, 0x00}, //"U",53
		{0x08, 0x78, 0x88, 0x00, 0x00, 0xC8, 0x38, 0x08, 0x00, 0x00, 0x07, 0x38, 0x0E, 0x01, 0x00, 0x00}, //"V",54
		{0xF8, 0x08, 0x00, 0xF8, 0x00, 0x08, 0xF8, 0x00, 0x03, 0x3C, 0x07, 0x00, 0x07, 0x3C, 0x03, 0x00}, //"W",55
		{0x08, 0x18, 0x68, 0x80, 0x80, 0x68, 0x18, 0x08, 0x20, 0x30, 0x2C, 0x03, 0x03, 0x2C, 0x30, 0x20}, //"X",56
		{0x08, 0x38, 0xC8, 0x00, 0xC8, 0x38, 0x08, 0x00, 0x00, 0x00, 0x20, 0x3F, 0x20, 0x00, 0x00, 0x00}, //"Y",57
		{0x10, 0x08, 0x08, 0x08, 0xC8, 0x38, 0x08, 0x00, 0x20, 0x38, 0x26, 0x21, 0x20, 0x20, 0x18, 0x00}, //"Z",58
		{0x00, 0x00, 0x00, 0xFE, 0x02, 0x02, 0x02, 0x00, 0x00, 0x00, 0x00, 0x7F, 0x40, 0x40, 0x40, 0x00}, //"[",59
		{0x00, 0x00, 0x00, 0x00, 0x80, 0x60, 0x18, 0x04, 0x00, 0x60, 0x18, 0x06, 0x01, 0x00, 0x00, 0x00}, //"/",60
		{0x00, 0x02, 0x02, 0x02, 0xFE, 0x00, 0x00, 0x00, 0x00, 0x40, 0x40, 0x40, 0x7F, 0x00, 0x00, 0x00}, //"]",61
		{0x00, 0x00, 0x04, 0x02, 0x02, 0x02, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, //"^",62
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}, //"-",63
		{0x00, 0x0C, 0x30, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x06, 0x38, 0xC0, 0x00}, //"\",64
		{0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x00, 0x19, 0x24, 0x22, 0x22, 0x22, 0x3F, 0x20}, //"a",65
		{0x08, 0xF8, 0x00, 0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x3F, 0x11, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"b",66
		{0x00, 0x00, 0x00, 0x80, 0x80, 0x80, 0x00, 0x00, 0x00, 0x0E, 0x11, 0x20, 0x20, 0x20, 0x11, 0x00}, //"c",67
		{0x00, 0x00, 0x00, 0x80, 0x80, 0x88, 0xF8, 0x00, 0x00, 0x0E, 0x11, 0x20, 0x20, 0x10, 0x3F, 0x20}, //"d",68
		{0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x00, 0x1F, 0x22, 0x22, 0x22, 0x22, 0x13, 0x00}, //"e",69
		{0x00, 0x80, 0x80, 0xF0, 0x88, 0x88, 0x88, 0x18, 0x00, 0x20, 0x20, 0x3F, 0x20, 0x20, 0x00, 0x00}, //"f",70
		{0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x6B, 0x94, 0x94, 0x94, 0x93, 0x60, 0x00}, //"g",71
		{0x08, 0xF8, 0x00, 0x80, 0x80, 0x80, 0x00, 0x00, 0x20, 0x3F, 0x21, 0x00, 0x00, 0x20, 0x3F, 0x20}, //"h",72
		{0x00, 0x80, 0x98, 0x98, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x20, 0x3F, 0x20, 0x20, 0x00, 0x00}, //"i",73
		{0x00, 0x00, 0x00, 0x80, 0x98, 0x98, 0x00, 0x00, 0x00, 0xC0, 0x80, 0x80, 0x80, 0x7F, 0x00, 0x00}, //"j",74
		{0x08, 0xF8, 0x00, 0x00, 0x80, 0x80, 0x80, 0x00, 0x20, 0x3F, 0x24, 0x02, 0x2D, 0x30, 0x20, 0x00}, //"k",75
		{0x00, 0x08, 0x08, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x20, 0x3F, 0x20, 0x20, 0x00, 0x00}, //"l",76
		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x20, 0x3F, 0x20, 0x00, 0x3F, 0x20, 0x00, 0x3F}, //"m",77
		{0x80, 0x80, 0x00, 0x80, 0x80, 0x80, 0x00, 0x00, 0x20, 0x3F, 0x21, 0x00, 0x00, 0x20, 0x3F, 0x20}, //"n",79
		{0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x00, 0x1F, 0x20, 0x20, 0x20, 0x20, 0x1F, 0x00}, //"o",80
		{0x80, 0x80, 0x00, 0x80, 0x80, 0x00, 0x00, 0x00, 0x80, 0xFF, 0xA1, 0x20, 0x20, 0x11, 0x0E, 0x00}, //"p",81
		{0x00, 0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x0E, 0x11, 0x20, 0x20, 0xA0, 0xFF, 0x80}, //"q",82
		{0x80, 0x80, 0x80, 0x00, 0x80, 0x80, 0x80, 0x00, 0x20, 0x20, 0x3F, 0x21, 0x20, 0x00, 0x01, 0x00}, //"r",83
		{0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x33, 0x24, 0x24, 0x24, 0x24, 0x19, 0x00}, //"s",84
		{0x00, 0x80, 0x80, 0xE0, 0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1F, 0x20, 0x20, 0x00, 0x00}, //"t",85
		{0x80, 0x80, 0x00, 0x00, 0x00, 0x80, 0x80, 0x00, 0x00, 0x1F, 0x20, 0x20, 0x20, 0x10, 0x3F, 0x20}, //"u",86
		{0x80, 0x80, 0x80, 0x00, 0x00, 0x80, 0x80, 0x80, 0x00, 0x01, 0x0E, 0x30, 0x08, 0x06, 0x01, 0x00}, //"v",87
		{0x80, 0x80, 0x00, 0x80, 0x00, 0x80, 0x80, 0x80, 0x0F, 0x30, 0x0C, 0x03, 0x0C, 0x30, 0x0F, 0x00}, //"w",88
		{0x00, 0x80, 0x80, 0x00, 0x80, 0x80, 0x80, 0x00, 0x00, 0x20, 0x31, 0x2E, 0x0E, 0x31, 0x20, 0x00}, //"x",89
		{0x80, 0x80, 0x80, 0x00, 0x00, 0x80, 0x80, 0x80, 0x80, 0x81, 0x8E, 0x70, 0x18, 0x06, 0x01, 0x00}, //"y",90
		{0x00, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x21, 0x30, 0x2C, 0x22, 0x21, 0x30, 0x00}, //"z",91
		{0x00, 0x00, 0x00, 0x00, 0x80, 0x7C, 0x02, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3F, 0x40, 0x40}, //"{",92
		{0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00}, //"|",93
		{0x00, 0x02, 0x02, 0x7C, 0x80, 0x00, 0x00, 0x00, 0x00, 0x40, 0x40, 0x3F, 0x00, 0x00, 0x00, 0x00}, //"}",94
		{0x00, 0x06, 0x01, 0x01, 0x02, 0x02, 0x04, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, //"~",95
	}
)

// WriteCmd for 8080 MPU
func (hd *Parallel8080) WriteCmd(cmd byte) error {
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
	usDealy(5)
	hd.WriteCS(embd.High) //diable chip select

	return nil
}

// WriteData for 8080 MPU
func (hd *Parallel8080) WriteData(dat byte) error {
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

	usDealy(5)
	hd.WriteCS(embd.High) //diable chip select

	return nil
}

// WriteData for 8080 MPU
func (hd *Parallel8080) WriteData16(val uint16) error {
	if err := hd.WriteCS(embd.Low); err != nil { //chip select,打开片选
		return err
	}

	if err := hd.writeRD(embd.High); err != nil { //disable read，读失能
		return err
	}

	if err := hd.writeRS(embd.High); err != nil { //select data，选择数据
		return err
	}

	//write msb
	dataH := byte(val >> 8)
	dataL := byte(val & 0xff)

	//write dataH
	if err := hd.fillDB8(dataH); err != nil { //put data，放置数据
		return err
	}
	if err := hd.writeWR(embd.Low); err != nil { //select write，选择写模式
		return err
	}
	//trigger WR rising edge to latch into LCD
	hd.writeWR(embd.High)

	//write dataL
	if err := hd.fillDB8(dataL); err != nil { //put data，放置数据
		return err
	}
	if err := hd.writeWR(embd.Low); err != nil { //select write，选择写模式
		return err
	}
	//trigger WR rising edge to latch into LCD
	hd.writeWR(embd.High)
	hd.WriteCS(embd.High) //diable chip select

	return nil
}

func (hd *Parallel8080) WriteRST(val int) error {
	return hd.Connection.Write(hd.RST, val)
}
func (hd *Parallel8080) writeWR(val int) error {
	return hd.Connection.Write(hd.WR, val)
}
func (hd *Parallel8080) writeRS(val int) error {
	return hd.Connection.Write(hd.RS, val)
}
func (hd *Parallel8080) writeRD(val int) error {
	return hd.Connection.Write(hd.RD, val)
}
func (hd *Parallel8080) WriteCS(val int) error {
	return hd.Connection.Write(hd.CS, val)
}

//  fillDB write value to DB0~7 GPIO
func (conn *Parallel8080) fillDB8(value byte) error {
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

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

type GPIOConnection struct {
	// con is the list of registered GPIO.
	con map[interface{}]embd.DigitalPin
}

func newGPIOPins(pins ...interface{}) (*GPIOConnection, error) {

	var ll = GPIOConnection{}
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

// PinInst  return gpio pin instance according to input
func (ll *GPIOConnection) PinInst(pin interface{}) *embd.DigitalPin {
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

	return &digitalPin
}

func usDealy(val time.Duration) {
	time.Sleep(val * time.Microsecond)
	//Millisecond
}

//commands from st7565 datasheet
const (
	//display on/of, datasheet P42

	//This command defines the column an area on the frame memory that can be accessed by the MPU
	cmdSetColumnAddr = 0x2A //  4 paramenters, datasheet p94
	cmdSetPageAddr   = 0x2B //  4 paramenters

	cmdSetDispOn = 0x29 // datasheet p93
)

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
