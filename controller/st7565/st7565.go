package main

//http://www.nxp.com/documents/data_sheet/PCF8574.pdf
import (
	"time"

	"flag"

	"github.com/golang/glog"
	"github.com/kidoman/embd"

	_ "github.com/kidoman/embd/host/all"

	"errors"
	"fmt"
)

func main() {
	flag.Parse()
	embd.SetHost(embd.HostRPi, 1000000)
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	hd, err := NewGPIO("P1_7" /*cs*/, "P1_11" /*wr*/, "P1_26" /*reset*/, "P1_12", /*rs*/
		"P1_15" /*db0*/, "P1_16" /*db1*/, "P1_18", /*db2*/
		"P1_19" /*db3*/, "P1_22" /*db4*/, "P1_21" /*db5*/, "P1_23" /*db6*/, "P1_24" /*db7*/)

	if err != nil {
		panic(err)
	}
	defer hd.Close()

	//dis := time.Now().Sub(start).Seconds()

	s := time.Now()
	hd.Clear()
	dis := time.Now().Sub(s).Minutes()
	fmt.Printf("hd.Clear using minutes %v\n", dis)

	s = time.Now()
	for i := 500; i > 0; i-- {
	}
	d := time.Now().Sub(s).Nanoseconds()
	fmt.Printf("500 loop consume %v\n", d)

	hd.Clear()
	hd.Ascii168(0, 0)

	/*
		running := true
		reader := bufio.NewReader(os.Stdin)

		for i:=0;running;i++ {
			fmt.Printf("show %v\n",i)
			//GUI_Write32CnChar(100, 20, "¼¼¿ÆÖÐÆÕ", RED, BLACK);
			hd.GUI_ShowPicture(74, 69, 151, 150);
			data, _, _ := reader.ReadLine()
			command := string(data)
			if command == "stop" {
				running = false
			}
		}
	*/
}

/*--  文字:  普  --*/
/*--  宋体12;  此字体下对应的点阵为：宽x高=16x16   --*/
var ASCII168 = [...]byte{0x00, 0x40, 0x44, 0x54, 0x64, 0x45, 0x7E, 0x44, 0x44, 0x44, 0x7E, 0x45, 0x64, 0x54, 0x44, 0x40,
	0x00, 0x00, 0x00, 0x00, 0xFF, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0xFF, 0x00, 0x00, 0x00}

func (hd *ST7565) Ascii168(xPos, yPos byte) {
	hd.setPage(yPos, xPos)
	for i := 0; i < 8; i++ {
		hd.WriteData(ASCII168[i])
	}
	hd.setPage(yPos+1, xPos)
	for i := 8; i < 16; i++ {
		hd.WriteData(ASCII168[i])
	}
}

const (
	//---定义屏幕大小---//
	//TFT_XMAX = 329
	//TFT_YMAX = 479

	writeDelay = 2 * time.Microsecond
	clearDelay = 2 * time.Microsecond
)

const (
//WHITE   = 0xFFFF
//BLACK   = 0x0000
//BLUE    = 0x001F
//RED     = 0xF800
//MAGENTA = 0xF81F
//GREEN   = 0x07E0
//CYAN    = 0x7FFF
//YELLOW  = 0xFFE0
)

// Close closes the underlying Connection.
func (hd *ST7565) Close() error {
	return hd.Connection.Close()
}

func (hd *ST7565) setPage(page, column byte) error {
	if page > 8 { //
		return errors.New("according to datasheet(P42), not beyond 8 pages")
	}
	page = page | cmdSetPageAddr
	if err := hd.WriteCmd(page); err != nil {
		return err
	}

	if column > 0x83 { //
		return errors.New("too large,according to datasheet(P43), the column address increment is topped at 83H")
	}
	lsb := column & 0x0f
	msb := (column & 0xf0) >> 4
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
func (hd *ST7565) Clear() error {

	//--表格第3个命令，设置Y的坐标--//
	//--Y轴有64个，一个坐标8位，也就是有8个坐标--//
	//所以一般我们使用的也就是从0xB0到0x07,就够了--//
	for i := 0; i < 8; i++ {
		//--表格第4个命令，设置X坐标--//
		//--当你的段初始化为0xA1时，X坐标从0x10,0x04到0x18,0x04,一共128位--//
		//--当你的段初始化为0xA0时，X坐标从0x10,0x00到0x18,0x00,一共128位--//
		//--在写入数据之后X坐标的坐标是会自动加1的，我们初始化使用0xA0所以--//
		//--我们的X坐标从0x10,0x00开始---//
		if err := hd.setPage(byte(i) /*page*/, 0x04 /*column*/); err != nil {
			return err
		}

		//--X轴有128位，就一共刷128次，X坐标会自动加1，所以我们不用再设置坐标--//
		for j := 0; j < 128; j++ {
			hd.WriteData(0x00) //如果设置背景为白色时，清屏选择0XFF
		}
	}
	hd.WriteCSX(embd.High)
	//GPIO.output(LCD_CS, True)
	//fmt.Println("leave Clear")

	return nil
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Connection abstracts the different methods of communicating with an R61526.
type Connection interface {
	// Close closes all open resources.
	Close() error
	// WriteReset  write value to RESET gpio port
	WriteReset(val int) error
	// WriteCSX  write value to CS gpio port
	WriteCSX(val int) error
	// WriteDCX  write value to DCX gpio port
	WriteDCX(val int) error
	// WriteWR  write value to WR gpio port
	WriteWR(val int) error
	// fillDB8  fills the Db7~Db0 port
	fillDB8(value byte) error
}

// GPIOConnection   implements Connection using XXXX bus.
type GPIOConnection struct {
	CS, WR, RESET, DCX                     embd.DigitalPin
	DB0, DB1, DB2, DB3, DB4, DB5, DB6, DB7 embd.DigitalPin
}

// ST7565 represents an ST7565-compatible character LCD controller.
type ST7565 struct {
	Connection
	//eMode entryMode
	//dMode displayMode
	//fMode functionMode
}

// WriteCmd for st7565
func (hd *ST7565) WriteCmd(cmd byte) error {
	var err error

	if err = hd.WriteCSX(embd.Low); err != nil { //chip select,打开片选
		return err
	}
	//hd.WriteRD(embd.High)  //disable read，读失能
	if err = hd.WriteDCX(embd.Low); err != nil { //select command，选择命令
		return err
	}
	if err = hd.WriteWR(embd.Low); err != nil { //select write，选择写模式
		return err
	}

	//_nop_();
	//_nop_();
	value := byte(cmd)
	if err = hd.fillDB8(value); err != nil {
		return err
	}

	//DATA_PORT = cmd; //put command，放置命令
	//_nop_();
	//_nop_();

	if err = hd.WriteWR(embd.High); err != nil { //command writing ，写入命令
		return err
	}
	return nil
}

// WriteData for st7565
func (hd *ST7565) WriteData(dat byte) {

	hd.WriteCSX(embd.Low) //chip select,打开片选
	//hd.WriteRD(embd.High)  //disable read，读失能
	hd.WriteDCX(embd.High) //select data，选择数据
	hd.WriteWR(embd.Low)   //select write，选择写模式

	//_nop_();
	//_nop_();

	hd.fillDB8(dat) //put data，放置数据
	//_nop_();
	//_nop_();

	hd.WriteWR(embd.High) //data writing，写数据
}

/****************************************************************************
*      * wide  图片宽度
*      * high  图片高度
****************************************************************************/

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

	//common output mode select, datasheet p46
	cmdSetComNormal  = 0xC0
	cmdSetComReverse = 0xC8

	//power controller set, datasheet p47
	cmdSetPowerControl  = 0x28
	cmdSetResistorRATIO = 0x20
	cmdSetVolumeFIRST   = 0x81
	cmdSetVolumeSECOND  = 0x00
	cmdSetStaticOFF     = 0xAC
	cmdSetStaticON      = 0xAD
	cmdSetStaticREG     = 0x00
)

// Init   initialize the st7565, the command is from the datasheet
func (hd *ST7565) Init() error {
	//	uchar i;
	//	LCD12864_RSET = 0;
	//	for (i=0; i<100; i++);
	//	LCD12864_CS = 0;
	//	LCD12864_RSET = 1;
	functions := []func() error{
		func() error { return hd.WriteReset(embd.Low) },
		func() error { return hd.WriteCSX(embd.Low) },
		func() error { return hd.WriteReset(embd.High) },
	}
	for _, f := range functions {
		if err := f(); err != nil {
			return err
		}
		//time.Sleep(delay)
	}

	if err := hd.WriteCmd(cmdInternalReset); err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)

	functions = []func() error{
		func() error { return hd.WriteCmd(cmdSetADCReverse) },   //--表格第8个命令，0xA0段（左右）方向选择正常方向（0xA1为反方向）--//
		func() error { return hd.WriteCmd(cmdSetComReverse) },   //--表格第15个命令，0xC8普通(上下)方向选择选择反向，0xC0为正常方向--//
		func() error { return hd.WriteCmd(cmdSetDispNormal) },   //--表格第9个命令，0xA6为设置字体为黑色，背景为白色.--0xA7为设置字体为白色，背景为黑色-//
		func() error { return hd.WriteCmd(cmdSetAllptsNormal) }, //--表格第10个命令，0xA4像素正常显示，0xA5像素全开--//
		func() error { return hd.WriteCmd(cmdSetLCDBias9) },     ////--表格第11个命令，0xA3偏压为1/7,0xA2偏压为1/9--//
		func() error { return hd.WriteCmd(0xF8) },               //--表格第19个命令，这个是个双字节的命令，0xF800选择增压为4X;--//
		func() error { return hd.WriteCmd(0x01) },               //--0xF801,选择增压为5X，其实效果差不多--//
		func() error { return hd.WriteCmd(0x81) },               //--表格第18个命令，这个是个双字节命令，高字节为0X81，低字节可以选择从0x00到0X3F。用来设置背景光对比度。
		func() error { return hd.WriteCmd(0x23) },
		func() error { return hd.WriteCmd(0x25) },                //--表格第17个命令，选择调节电阻率--//
		func() error { return hd.WriteCmd(0x2F) },                //--表格第16个命令，电源设置。--//
		func() error { return hd.WriteCmd(cmdSetDispStartLine) }, //--表格第2个命令  0x40，设置显示开始位置--//
		func() error { return hd.WriteCmd(cmdDisplyON) },         //--表格第1个命令，开启显示--//
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
		//time.Sleep(delay)
	}

	//time.Sleep(delay)

	return nil
}

// NewGPIO creates a new R61526 connected by XXX bus.
func NewGPIO(
	cs, wr, reset, dcx, db0, db1, db2, db3, db4, db5, db6, db7 interface{}) (*ST7565, error) {
	pinKeys := []interface{}{cs, wr, reset, dcx, db0, db1, db2, db3, db4, db5, db6, db7}
	pins := [13]embd.DigitalPin{}
	for idx, key := range pinKeys {
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
				glog.V(1).Infof("hX8357: error creating digital pin %+v: %s", key, err)
				return nil, err
			}
		}
		pins[idx] = digitalPin
	}
	for _, pin := range pins {
		if pin == nil {
			continue
		}
		err := pin.SetDirection(embd.Out)
		if err != nil {
			glog.Errorf("hX8357: error setting pin %+v to out direction: %s", pin, err)
			return nil, err
		}
	}
	return newSt7565(
		newGPIOConnection(
			pins[0],
			pins[1],
			pins[2],
			pins[3],
			pins[4],
			pins[5],
			pins[6],
			pins[7],
			pins[8],
			pins[9],
			pins[10],
			pins[11],
		))
}

// newGPIOConnection returns a new Connection based on a 4-bit GPIO bus.
func newGPIOConnection(cs, wr, reset, dcx, db0, db1, db2, db3,
	db4, db5, db6, db7 embd.DigitalPin) *GPIOConnection {
	return &GPIOConnection{
		CS:    cs,
		WR:    wr,
		RESET: reset,
		DCX:   dcx,
		DB0:   db0,
		DB1:   db1,
		DB2:   db2,
		DB3:   db3,
		DB4:   db4,
		DB5:   db5,
		DB6:   db6,
		DB7:   db7}
}

// newHX8357 creates a new R61526 connected by a Connection bus.
func newSt7565(bus Connection) (*ST7565, error) {
	controller := &ST7565{
		Connection: bus,
	}

	glog.V(2).Info("ST7565: initializing display")
	err := controller.Init()
	if err != nil {
		return nil, err
	}
	return controller, nil
}

// Close closes all open DigitalPins.
func (conn *GPIOConnection) Close() error {
	glog.V(2).Info("ST7565: closing all GPIO pins")
	pins := []embd.DigitalPin{
		conn.CS,
		conn.WR,
		conn.RESET,
		conn.DCX,
		conn.DB0,
		conn.DB1,
		conn.DB2,
		conn.DB3,
		conn.DB4,
		conn.DB5,
		conn.DB6,
		conn.DB7,
	}

	for _, pin := range pins {
		err := pin.Close()
		if err != nil {
			glog.Errorf("ST7565: error closing pin %+v: %s", pin, err)
			return err
		}
	}
	return nil
}

//  fillDB write value to DB0~7 GPIO
func (conn *GPIOConnection) fillDB8(value byte) error {
	functions := []func() error{
		func() error { return conn.DB4.Write(int((value >> 4) & 0x01)) },
		func() error { return conn.DB5.Write(int((value >> 5) & 0x01)) },
		func() error { return conn.DB6.Write(int((value >> 6) & 0x01)) },
		func() error { return conn.DB7.Write(int((value >> 7) & 0x01)) },
		func() error { return conn.DB0.Write(int(value & 0x01)) },
		func() error { return conn.DB1.Write(int((value >> 1) & 0x01)) },
		func() error { return conn.DB2.Write(int((value >> 2) & 0x01)) },
		func() error { return conn.DB3.Write(int((value >> 3) & 0x01)) },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

// to-do
func (conn *GPIOConnection) write8(iscmd bool, value byte) error {

	v := embd.Low
	if iscmd == false {
		v = embd.High
	}

	functions := []func() error{
		func() error { return conn.WR.Write(embd.High) }, //初始化WR
		func() error { return conn.CS.Write(embd.Low) },  //打开片选
		// func() error { return conn.RD.Write(embd.High) }, // invalid RDX. we set it allways high
		func() error { return conn.DCX.Write(v) }, //indicate the value  is command or data
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}

	if err := conn.fillDB8(value); err != nil {
		return err
	}

	// write command through rising edage of WRX
	functions = []func() error{
		func() error { return conn.WR.Write(embd.Low) }, //写入时序
		func() error { return conn.WR.Write(embd.High) },
		func() error { return conn.CS.Write(embd.High) }, //关闭片选
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteReset  write value to RESET gpio port
func (conn *GPIOConnection) WriteReset(val int) error {
	return conn.RESET.Write(val)
}

// WriteCSX  write value to CS gpio port
func (conn *GPIOConnection) WriteCSX(val int) error {
	return conn.CS.Write(val)
}

// WriteDCX  write value to DCX gpio port
func (conn *GPIOConnection) WriteDCX(val int) error {
	return conn.DCX.Write(val)
}

// WriteWR  write value to WR gpio port
func (conn *GPIOConnection) WriteWR(val int) error {
	return conn.WR.Write(val)
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
//https://github.com/Bodmer/TFT_HX8357

//https://github.com/baran0119/ALCATEL_LCM/tree/master/OT_903D/mediatek/custom/common/kernel/lcm/r61526

/*
func (hd *ST7565) TFT_WriteColorData(color uint16) error {
	red := (color & 0x1F)        //取5位蓝色
	green := (color >> 5) & 0x3F //取6位绿色
	blue := (color >> 11) & 0x1F //取5位红色

	rgb := uint16((red << 11) | (green << 6) | blue)

	return hd.WriteData16(rgb)
}


func (hd *ST7565) TFT_SetWindow(xStart, yStart, xEnd, yEnd uint16) {
}

func (hd *ST7565) GUI_ShowPicture(x, y uint16, wide, high uint16) {

	hd.TFT_SetWindow(x, y, uint16(x+wide)-1, y+high-1)
	num := wide * high * 2
	var tmp uint16 = 0
	var temp uint16
	for tmp < num {
		temp = uint16(pic[tmp+1])
		temp = temp << 8
		temp = temp | uint16(pic[tmp])
		//TFT_WriteData(~temp);//逐点显示
		hd.TFT_WriteColorData(temp)
		tmp += 2
	}
}
*/
