// Digital IO support.
// This driver requires kernel version 3.8+ and should work uniformly
// 实现rpi 直接memory map操纵GPIO.  实现package embd中“type DigitalPin interface”的所有接口

//go:generate go run mkiomuxc.go

package udoo

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/kidoman/embd"
)

// Pull Up / Down / Off
const (
	PullOff   int = -1
	Pus100KPD     = 0
	Pus470KPU     = 1
	Pus100KPU     = 2
	Pus22KPU      = 3

//00 PUS_100KOHM_PD — 100K Ohm Pull Down
//01 PUS_47KOHM_PU — 47K Ohm Pull Up
//10 PUS_100KOHM_PU — 100K Ohm Pull Up
//11 PUS_22KOHM_PU — 22K Ohm Pull Up

)

type gpioReg struct {
	offset   uint32 // in each bank, the offset(u32 unit) from the gpio bank base. ref P1527  GPIO memorymap
	readOnly bool
}

var (
	// P1527  GPIO memorymap,
	dr      = gpioReg{offset: 0, readOnly: false}
	gdir    = gpioReg{offset: 1, readOnly: false}
	psr     = gpioReg{offset: 2, readOnly: true}
	icr1    = gpioReg{offset: 3, readOnly: false}
	icr2    = gpioReg{offset: 4, readOnly: false}
	imr     = gpioReg{offset: 5, readOnly: false} //GPIO interrupt mask register
	isr     = gpioReg{offset: 6, readOnly: false} //GPIO interrupt status register
	edgeSel = gpioReg{offset: 7, readOnly: false} //GPIO edge select register
)

type GpioPin struct {
	n     int
	aliay string
	crl   PinCrtl
	//padctrlOft uint32
	//muxOft     uint32
}

var pins = [...]GpioPin{
	GpioPin{n: 0, aliay: "GPIO1_IO00", crl: MX6SX_PAD_GPIO1_IO00__GPIO1_IO_0}, //padctrl code ref P283, pad mux reg mem offset ref p1691
	GpioPin{n: 1, aliay: "GPIO1_IO01", crl: MX6SX_PAD_GPIO1_IO01__GPIO1_IO_1},
	GpioPin{n: 2, aliay: "GPIO1_IO02", crl: MX6SX_PAD_GPIO1_IO02__GPIO1_IO_2},
	GpioPin{n: 4, aliay: "GPIO1_IO04", crl: MX6SX_PAD_GPIO1_IO04__GPIO1_IO_4},
	GpioPin{n: 5, aliay: "GPIO1_IO05", crl: MX6SX_PAD_GPIO1_IO05__GPIO1_IO_5},
	GpioPin{n: 6, aliay: "GPIO1_IO06", crl: MX6SX_PAD_GPIO1_IO06__GPIO1_IO_6},
	GpioPin{n: 7, aliay: "GPIO1_IO07", crl: MX6SX_PAD_GPIO1_IO07__GPIO1_IO_7},
	GpioPin{n: 8, aliay: "GPIO1_IO08", crl: MX6SX_PAD_GPIO1_IO08__GPIO1_IO_8},
	GpioPin{n: 9, aliay: "GPIO1_IO09", crl: MX6SX_PAD_GPIO1_IO09__GPIO1_IO_9},
	GpioPin{n: 10, aliay: "GPIO1_IO10", crl: MX6SX_PAD_GPIO1_IO10__GPIO1_IO_10},
	GpioPin{n: 11, aliay: "GPIO1_IO11", crl: MX6SX_PAD_GPIO1_IO11__GPIO1_IO_11},
	GpioPin{n: 12, aliay: "GPIO1_IO12", crl: MX6SX_PAD_GPIO1_IO12__GPIO1_IO_12},
	GpioPin{n: 13, aliay: "GPIO1_IO13", crl: MX6SX_PAD_GPIO1_IO13__GPIO1_IO_13},
	GpioPin{n: 14, aliay: "GPIO1_IO14", crl: MX6SX_PAD_CSI_DATA00__GPIO1_IO_14},
	GpioPin{n: 15, aliay: "GPIO1_IO15", crl: MX6SX_PAD_CSI_DATA01__GPIO1_IO_15},
	GpioPin{n: 16, aliay: "GPIO1_IO16", crl: MX6SX_PAD_CSI_DATA02__GPIO1_IO_16},
	GpioPin{n: 17, aliay: "GPIO1_IO17", crl: MX6SX_PAD_CSI_DATA03__GPIO1_IO_17},
	GpioPin{n: 18, aliay: "GPIO1_IO18", crl: MX6SX_PAD_CSI_DATA04__GPIO1_IO_18},
	GpioPin{n: 19, aliay: "GPIO1_IO19", crl: MX6SX_PAD_CSI_DATA05__GPIO1_IO_19},
	GpioPin{n: 20, aliay: "GPIO1_IO20", crl: MX6SX_PAD_CSI_DATA06__GPIO1_IO_20},
	GpioPin{n: 21, aliay: "GPIO1_IO21", crl: MX6SX_PAD_CSI_DATA07__GPIO1_IO_21},
	GpioPin{n: 22, aliay: "GPIO1_IO22", crl: MX6SX_PAD_CSI_HSYNC__GPIO1_IO_22},
	GpioPin{n: 23, aliay: "GPIO1_IO23", crl: MX6SX_PAD_CSI_MCLK__GPIO1_IO_23},
	GpioPin{n: 24, aliay: "GPIO1_IO24", crl: MX6SX_PAD_CSI_PIXCLK__GPIO1_IO_24},
	GpioPin{n: 25, aliay: "GPIO1_IO25", crl: MX6SX_PAD_CSI_VSYNC__GPIO1_IO_25},

	GpioPin{n: 96, aliay: "GPIO4_IO00", crl: MX6SX_PAD_NAND_ALE__GPIO4_IO_0}, //padctrl code ref P285, pad mux reg mem offset ref p1695
	GpioPin{n: 97, aliay: "GPIO4_IO01", crl: MX6SX_PAD_NAND_CE0_B__GPIO4_IO_1},
	GpioPin{n: 98, aliay: "GPIO4_IO02", crl: MX6SX_PAD_NAND_CE1_B__GPIO4_IO_2},
	GpioPin{n: 99, aliay: "GPIO4_IO03", crl: MX6SX_PAD_NAND_CLE__GPIO4_IO_3},

	GpioPin{n: 100, aliay: "GPIO4_IO04", crl: MX6SX_PAD_NAND_DATA00__GPIO4_IO_4},
	GpioPin{n: 101, aliay: "GPIO4_IO05", crl: MX6SX_PAD_NAND_DATA01__GPIO4_IO_5},
	GpioPin{n: 102, aliay: "GPIO4_IO06", crl: MX6SX_PAD_NAND_DATA02__GPIO4_IO_6},
	GpioPin{n: 103, aliay: "GPIO4_IO07", crl: MX6SX_PAD_NAND_DATA03__GPIO4_IO_7},
	GpioPin{n: 104, aliay: "GPIO4_IO08", crl: MX6SX_PAD_NAND_DATA04__GPIO4_IO_8},
	GpioPin{n: 105, aliay: "GPIO4_IO09", crl: MX6SX_PAD_NAND_DATA05__GPIO4_IO_9},
	GpioPin{n: 106, aliay: "GPIO4_IO10", crl: MX6SX_PAD_NAND_DATA06__GPIO4_IO_10},
	GpioPin{n: 107, aliay: "GPIO4_IO11", crl: MX6SX_PAD_NAND_DATA07__GPIO4_IO_11},
	/*ToDo
	GpioPin{n: 108, aliay: "GPIO4_IO12", crl:},
	GpioPin{n: 109, aliay: "GPIO4_IO13", crl:},
	GpioPin{n: 110, aliay: "GPIO4_IO14", crl:},
	GpioPin{n: 111, aliay: "GPIO4_IO15", crl:},
	GpioPin{n: 112, aliay: "GPIO4_IO16", crl:},
	GpioPin{n: 113, aliay: "GPIO4_IO17", crl:},
	GpioPin{n: 114, aliay: "GPIO4_IO18", crl:},
	GpioPin{n: 115, aliay: "GPIO4_IO19", crl:},
	GpioPin{n: 116, aliay: "GPIO4_IO20", crl:},
	GpioPin{n: 117, aliay: "GPIO4_IO21", crl:},
	GpioPin{n: 118, aliay: "GPIO4_IO22", crl:},
	GpioPin{n: 119, aliay: "GPIO4_IO23", crl:},
	GpioPin{n: 120, aliay: "GPIO4_IO24", crl:},
	GpioPin{n: 121, aliay: "GPIO4_IO25", crl:},
	GpioPin{n: 122, aliay: "GPIO4_IO26", crl:},
	GpioPin{n: 123, aliay: "GPIO4_IO27", crl:},
	GpioPin{n: 124, aliay: "GPIO4_IO28", crl:},
	GpioPin{n: 125, aliay: "GPIO4_IO29", crl:},
	GpioPin{n: 126, aliay: "GPIO4_IO30", crl:},
	GpioPin{n: 127, aliay: "GPIO4_IO31", crl:},

	GpioPin{n: 160, aliay: "GPIO6_IO00",crl:MX6SX_PAD_SD1_CLK__GPIO6_IO_0}, //padctrl code ref P287, pad mux reg mem offset ref p1697
	GpioPin{n: 161, aliay: "GPIO6_IO01",
	GpioPin{n: 162, aliay: "GPIO6_IO02", },
	GpioPin{n: 163, aliay: "GPIO6_IO03", },
	GpioPin{n: 164, aliay: "GPIO6_IO04", },
	GpioPin{n: 165, aliay: "GPIO6_IO05", },
	GpioPin{n: 166, aliay: "GPIO6_IO06",
	GpioPin{n: 167, aliay: "GPIO6_IO07",
	GpioPin{n: 168, aliay: "GPIO6_IO08", },
	GpioPin{n: 169, aliay: "GPIO6_IO09", },
	GpioPin{n: 170, aliay: "GPIO6_IO10", crl:MX6SX_PAD_SD2_DATA2__GPIO6_IO_10},
	GpioPin{n: 171, aliay: "GPIO6_IO11", crl:MX6SX_PAD_SD2_DATA3__GPIO6_IO_11},
	GpioPin{n: 172, aliay: "GPIO6_IO12",
	GpioPin{n: 173, aliay: "GPIO6_IO13",
	GpioPin{n: 174, aliay: "GPIO6_IO14", },
	GpioPin{n: 175, aliay: "GPIO6_IO15", },
	GpioPin{n: 176, aliay: "GPIO6_IO16", },
	GpioPin{n: 177, aliay: "GPIO6_IO17", },
	GpioPin{n: 178, aliay: "GPIO6_IO18", },
	GpioPin{n: 179, aliay: "GPIO6_IO19", },
	GpioPin{n: 180, aliay: "GPIO6_IO20", },
	GpioPin{n: 181, aliay: "GPIO6_IO21", },
	GpioPin{n: 182, aliay: "GPIO6_IO22", },  */
}

var (
	//p188  , one bank is 16K
	gpioBankBase                                    = [...]int32{0, 0x0209C000 /*bank1*/, 0x020A0000, 0x020A4000, 0x020A8000, 0x020AC000, 0x020B0000, 0x020B4000}
	gpio1, gpio2, gpio3, gpio4, gpio5, gpio6, gpio7 []uint32
	memlock                                         sync.Mutex
	gpioInitialized                                 bool
)

//
func init() {
	var err error
	gpio1, _, err = mapMem(gpioBankBase[1], 16) //map 16K
	if err != nil {
		return
	}
	if gpio2, _, err = mapMem(gpioBankBase[2], 16); err != nil {
		return
	}
	if gpio3, _, err = mapMem(gpioBankBase[3], 16); err != nil {
		return
	}
	if gpio4, _, err = mapMem(gpioBankBase[4], 16); err != nil {
		return
	}
	if gpio5, _, err = mapMem(gpioBankBase[5], 16); err != nil {
		return
	}
	if gpio6, _, err = mapMem(gpioBankBase[6], 16); err != nil {
		return
	}
	if gpio7, _, err = mapMem(gpioBankBase[7], 16); err != nil {
		return
	}
	//
	iomuxc, _, err = mapMem(iomuxcBase, 16) //map 16K
	if err != nil {
		return
	}

	gpioInitialized = true
	return
}

// paddr it should be 4k aligned
// count how many size  (uint is 1K)
func mapMem(paddr int32, count int) (u32Array []uint32, byteArray []byte, err error) {

	//	Try /dev/mem. If that fails, then
	//	try /dev/gpiomem. If that fails then game over.
	file, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0660)
	if err != nil {
		err = errors.New("can not open /dev/mem , maybe try sudo")
		return //
	}
	//fd can be closed after memory mapping
	defer file.Close()

	//	GPIO:
	byteArray, err = syscall.Mmap(int(file.Fd()), int64(paddr), 1024*count*4, /*4 because we map cout K uint32*/
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return //errors.New("mmap  failed")
	}

	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&byteArray))
	// The length and capacity of the slice are different.
	header.Len /= 4
	header.Cap /= 4
	// Convert slice header to an []uint32
	u32Array = *(*[]uint32)(unsafe.Pointer(&header))
	return
}

func readGpioPadCtl(n int) (uint32, error) {
	for _, pin := range pins {
		if pin.n == n {
			oft := (pin.padctrlOft) / 4 // because iomuxc is uint32 array
			return iomuxc[oft], nil
		}
	}

	return 0, fmt.Errorf("the gpio pad ctrl reg dont register")
}

func writeGpioPadCtl(n int, val uint32) error {
	for _, pin := range pins {
		if pin.n == n {
			oft := (pin.padctrlOft) / 4 // because iomuxc is uint32 array
			iomuxc[oft] = val
			return nil
		}
	}

	return fmt.Errorf("the gpio pad ctrl reg dont register")
}

func bankBase(bank int) ([]uint32, error) {
	if bank <= 0 || bank > 7 {
		return nil, fmt.Errorf("bank: %+v is invalid", bank)
	}
	var vAddr []uint32

	switch bank {
	case 1:
		vAddr = gpio1
	case 2:
		vAddr = gpio2
	case 3:
		vAddr = gpio3
	case 4:
		vAddr = gpio4
	case 5:
		vAddr = gpio5
	case 6:
		vAddr = gpio6
	case 7:
		vAddr = gpio7
	default:
		return nil, fmt.Errorf("reg is invliad")
	}
	return vAddr, nil
}

func readGpioReg(bank int, io uint, reg gpioReg) (uint32, error) {
	if io <= 0 || io > 31 {
		return 0, fmt.Errorf("io: %+v is invalid", io)
	}
	if gpioInitialized == false {
		return 0, fmt.Errorf("memory map not init")
	}
	vAddr, err := bankBase(bank)
	if err != nil {
		return 0, err
	}

	v := (vAddr[reg.offset] & (1 << io)) >> io
	return v, nil
}
func writeGpioReg(bank int, io uint, reg gpioReg, high bool) error {
	if bank <= 0 || bank > 7 {
		return fmt.Errorf("bank: %+v is invalid", bank)
	}
	if io <= 0 || io > 31 {
		return fmt.Errorf("io: %+v is invalid", io)
	}
	if gpioInitialized == false {
		return fmt.Errorf("memory map not init")
	}

	if reg.readOnly == true {
		return fmt.Errorf("reg: %+v can not be writed according to data sheet", reg)
	}

	vAddr, err := bankBase(bank)
	if err != nil {
		return err
	}

	if high == true {
		vAddr[reg.offset] |= uint32(1 << io)
	} else {
		v := uint32(1 << io)
		vAddr[reg.offset] = vAddr[reg.offset] ^ v
	}
	return nil
}

type udooDigitPin struct {
	id string
	n  int

	drv embd.GPIODriver

	//dir *os.File
	val       *os.File
	activeLow *os.File

	readBuf []byte

	initialized bool
}

//NewUdooDigitalPin   export to digitPin driver
func NewUdooDigitalPin(pd *embd.PinDesc, drv embd.GPIODriver) embd.DigitalPin {
	return &udooDigitPin{id: pd.ID, n: pd.DigitalLogical, drv: drv, readBuf: make([]byte, 1)}
}

func (p *udooDigitPin) N() int {
	return p.n
}

func (p *udooDigitPin) init() error {
	if p.initialized {
		return nil
	}

	if gpioInitialized == false {
		return fmt.Errorf("memory map not init")
	}

	var err error
	if err = p.export(); err != nil {
		return err
	}
	//if p.dir, err = p.directionFile(); err != nil {
	//	return err
	//}
	if p.val, err = p.valueFile(); err != nil {
		return err
	}
	if p.activeLow, err = p.activeLowFile(); err != nil {
		return err
	}

	p.initialized = true

	return nil
}

func (p *udooDigitPin) export() error {
	exporter, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer exporter.Close()
	_, err = exporter.WriteString(strconv.Itoa(p.n))
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.EBUSY {
		return nil // EBUSY -> the pin has already been exported
	}
	return err
}

func (p *udooDigitPin) unexport() error {
	unexporter, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer unexporter.Close()
	_, err = unexporter.WriteString(strconv.Itoa(p.n))
	return err
}

func (p *udooDigitPin) basePath() string {
	return fmt.Sprintf("/sys/class/gpio/gpio%v", p.n)
}

func (p *udooDigitPin) openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, os.ModeExclusive)
}

func (p *udooDigitPin) directionFile() (*os.File, error) {
	return p.openFile(path.Join(p.basePath(), "direction"))
}

func (p *udooDigitPin) valueFile() (*os.File, error) {
	return p.openFile(path.Join(p.basePath(), "value"))
}

func (p *udooDigitPin) activeLowFile() (*os.File, error) {
	return p.openFile(path.Join(p.basePath(), "active_low"))
}

// SetDirection sets the direction of the pin (in/out).     in “type DigitalPin interface”
func (p *udooDigitPin) SetDirection(dir embd.Direction) error {
	if err := p.init(); err != nil {
		return err
	}

	//(bank-1)*32 +io  -->normally GPIO num
	bank := p.n/32 + 1
	io := uint(p.n % 32)
	// 30.5.2 GPIO direction register (GPIOx_GDIR)
	//0 INPUT — GPIO is configured as input.
	//1 OUTPUT — GPIO is configured as output.
	if dir == embd.In {
		if err := writeGpioReg(bank, io, gdir, false); err != nil {
			return err
		}
	} else {
		if err := writeGpioReg(bank, io, gdir, true); err != nil {
			return err
		}
	}

	return nil
}

// Read reads the value from the pin.    in “type DigitalPin interface”
func (p *udooDigitPin) Read() (int, error) {
	if err := p.init(); err != nil {
		return 0, err
	}

	//(bank-1)*32 +io  -->normally GPIO num
	bank := p.n/32 + 1
	io := uint(p.n % 32)

	val, err := readGpioReg(bank, io, psr)

	return int(val), err
}

// Write writes the provided value to the pin.   in “type DigitalPin interface”
func (p *udooDigitPin) Write(val int) error {
	if err := p.init(); err != nil {
		return err
	}

	//(bank-1)*32 +io  -->normally GPIO num
	bank := p.n/32 + 1
	io := uint(p.n % 32)

	if val == embd.Low {
		return writeGpioReg(bank, io, dr, false)
	}

	return writeGpioReg(bank, io, dr, true)

}

// TimePulse measures the duration of a pulse on the pin.     in “type DigitalPin interface”
func (p *udooDigitPin) TimePulse(state int) (time.Duration, error) {
	if err := p.init(); err != nil {
		return 0, err
	}

	aroundState := embd.Low
	if state == embd.Low {
		aroundState = embd.High
	}

	// Wait for any previous pulse to end
	for {
		v, err := p.Read()
		if err != nil {
			return 0, err
		}

		if v == aroundState {
			break
		}
	}

	// Wait until ECHO goes high
	for {
		v, err := p.Read()
		if err != nil {
			return 0, err
		}

		if v == state {
			break
		}
	}

	startTime := time.Now() // Record time when ECHO goes high

	// Wait until ECHO goes low
	for {
		v, err := p.Read()
		if err != nil {
			return 0, err
		}

		if v == aroundState {
			break
		}
	}

	return time.Since(startTime), nil // Calculate time lapsed for ECHO to transition from high to low
}

// ActiveLow makes the pin active low. A low logical state is represented by
// a high state on the physical pin, and vice-versa.     in “type DigitalPin interface”
func (p *udooDigitPin) ActiveLow(b bool) error {
	if err := p.init(); err != nil {
		return err
	}

	str := "0"
	if b {
		str = "1"
	}
	_, err := p.activeLow.WriteString(str)
	return err
}

// PullUp pulls the pin up.     in “type DigitalPin interface”
func (p *udooDigitPin) PullUp() error {
	gpioPullMode((p.n), Pus100KPU)
	return nil

}

// PullDown pulls the pin down.     in “type DigitalPin interface”
func (p *udooDigitPin) PullDown() error {
	gpioPullMode((p.n), Pus100KPD)
	return nil
}

// Close releases the resources associated with the pin.     in “type DigitalPin interface”
func (p *udooDigitPin) Close() error {
	//rpiDigitPin dont implement watch
	//if err := p.StopWatching(); err != nil {
	//	return err
	//}

	if err := p.drv.Unregister(p.id); err != nil {
		return err
	}

	if !p.initialized {
		return nil
	}

	//if err := p.dir.Close(); err != nil {
	//	return err
	//}
	if err := p.val.Close(); err != nil {
		return err
	}
	if err := p.activeLow.Close(); err != nil {
		return err
	}
	if err := p.unexport(); err != nil {
		return err
	}

	p.initialized = false

	return nil
}

func (p *udooDigitPin) setEdge(edge embd.Edge) error {
	file, err := p.openFile(path.Join(p.basePath(), "edge"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(edge))
	return err
}

func (p *udooDigitPin) Watch(edge embd.Edge, handler func(embd.DigitalPin)) error {
	//return errors.New("gpio: not implemented")
	if err := p.setEdge(edge); err != nil {
		return err
	}
	return registerInterrupt(p, handler)
}

func (p *udooDigitPin) StopWatching() error {
	//return errors.New("gpio: not implemented")
	return unregisterInterrupt(p)

}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

func gpioPullMode(n int, pull int) error {

	//15-14 PUS Pull Up / Down Config. Field
	//Select one of next values for pad:
	//00 PUS_100KOHM_PD — 100K Ohm Pull Down
	//01 PUS_47KOHM_PU — 47K Ohm Pull Up
	//10 PUS_100KOHM_PU — 100K Ohm Pull Up
	//11 PUS_22KOHM_PU — 22K Ohm Pull Up

	//13 PUE Pull / Keep Select Field
	//Select one of next values for pad: CSI_MCLK.
	//0 KEEP — Keeper Enabled
	//1 PULL — Pull Enabled
	///
	//12 PKE  Pull / Keep Enable Field
	//Pull / Keep Enable Field
	//Select one of next values for pad: CSI_MCLK.
	//0 DISABLED — Pull/Keeper Disabled
	//1 ENABLED — Pull/Keeper Enabled

	//(bank-1)*32 +io  -->normally GPIO num
	//bank := n/32 + 1
	//io := uint(n % 32)
	//gpio := fmt.Sprintf("GPIO%d_IO%2d", bank, io)

	val, err := readGpioPadCtl(n)
	if err != nil {
		return err
	}

	memlock.Lock()
	defer memlock.Unlock()

	if pull != PullOff {
		val &= ^(uint32(1) << 12)                   //PKE set to 0
		val |= (1 << 13)                            //PUE set to 1
		val = (uint32(pull) << 14) | (val & 0x3FFF) //set PUS and keep other
	} else {
		val &= ^(uint32(1) << 12)
	}

	// Wait for value to clock in, this is ugly, sorry :(
	time.Sleep(time.Microsecond)
	return nil
}
