// Digital IO support.
// This driver requires kernel version 3.8+ and should work uniformly
// 实现rpi 直接memory map操纵GPIO.  实现package embd中“type DigitalPin interface”的所有接口

//go:generate go run mkiomuxc_ignore.go

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
			//pin.crl.confReg
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
