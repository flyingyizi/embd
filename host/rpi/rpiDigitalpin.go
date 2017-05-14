// Digital IO support.
// This driver requires kernel version 3.8+ and should work uniformly
// 实现rpi 直接memory map操纵GPIO.  实现package embd中“type DigitalPin interface”的所有接口

package rpi

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

// Peripheral Offsets for the Raspberry Pi
const (
	GpioBase        = 0x00200000 // GPIO registers
	SizeOfuint32    = 4          // bytes
	uint32BlockSize = SizeOfuint32 * 1024
)

// Pull Up / Down / Off
const (
	PullOff int = iota
	PullDown
	PullUp
)

var (
	gpioArry        []uint32
	gpio            []byte
	memlock         sync.Mutex
	gpioInitialized bool
)

// Close unmaps GPIO memory
func RPIMemClose() (err error) {
	memlock.Lock()
	defer memlock.Unlock()

	err = syscall.Munmap(gpio)
	return
}

//
func init() {
	gpioInitialized = false
	_, piGpioBase, err := GetBoardInfo()
	if err != nil {
		return
	}

	// Set the offsets into the memory interface.
	GPIO_BASE := piGpioBase + GpioBase

	//	Try /dev/mem. If that fails, then
	//	try /dev/gpiomem. If that fails then game over.
	file, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0660)
	if err != nil {
		file, err = os.OpenFile("/dev/gpiomem", os.O_RDWR|os.O_SYNC, 0660) //|os.O_CLOEXEC
		if err != nil {
			return //errors.New("can not open /dev/mem or /dev/gpiomem, maybe try sudo")
		}
	}
	//fd can be closed after memory mapping
	defer file.Close()

	//	GPIO:
	gpio, err = syscall.Mmap(int(file.Fd()), GPIO_BASE, uint32BlockSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return //errors.New("mmap (GPIO) failed")
	}

	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&gpio))
	// The length and capacity of the slice are different.
	header.Len /= SizeOfuint32
	header.Cap /= SizeOfuint32
	// Convert slice header to an []uint32
	gpioArry = *(*[]uint32)(unsafe.Pointer(&header))
	gpioInitialized = true
	return
}

func mapMem() (gpioArry []uint32, gpio []byte, err error) {
	_, piGpioBase, err := GetBoardInfo()
	if err != nil {
		return
	}

	// Set the offsets into the memory interface.
	GPIO_BASE := piGpioBase + GpioBase

	//	Try /dev/mem. If that fails, then
	//	try /dev/gpiomem. If that fails then game over.
	file, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0660)
	if err != nil {
		file, err = os.OpenFile("/dev/gpiomem", os.O_RDWR|os.O_SYNC, 0660) //|os.O_CLOEXEC
		if err != nil {
			return //errors.New("can not open /dev/mem or /dev/gpiomem, maybe try sudo")
		}
	}
	//fd can be closed after memory mapping
	defer file.Close()

	//	GPIO:
	gpio, err = syscall.Mmap(int(file.Fd()), GPIO_BASE, uint32BlockSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return //errors.New("mmap (GPIO) failed")
	}

	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&gpio))
	// The length and capacity of the slice are different.
	header.Len /= SizeOfuint32
	header.Cap /= SizeOfuint32
	// Convert slice header to an []uint32
	gpioArry = *(*[]uint32)(unsafe.Pointer(&header))
	return
}

type rpiDigitPin struct {
	id string
	n  int

	drv embd.GPIODriver

	//dir *os.File
	//val       *os.File
	activeLow *os.File

	readBuf []byte

	initialized bool
}

func NewRPIDigitalPin(pd *embd.PinDesc, drv embd.GPIODriver) embd.DigitalPin {
	return &rpiDigitPin{id: pd.ID, n: pd.DigitalLogical, drv: drv, readBuf: make([]byte, 1)}
}

func (p *rpiDigitPin) N() int {
	return p.n
}

func (p *rpiDigitPin) init() error {
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
	//if p.val, err = p.valueFile(); err != nil {
	//	return err
	//}
	if p.activeLow, err = p.activeLowFile(); err != nil {
		return err
	}

	p.initialized = true

	return nil
}

func (p *rpiDigitPin) export() error {
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

func (p *rpiDigitPin) unexport() error {
	unexporter, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer unexporter.Close()
	_, err = unexporter.WriteString(strconv.Itoa(p.n))
	return err
}

func (p *rpiDigitPin) basePath() string {
	return fmt.Sprintf("/sys/class/gpio/gpio%v", p.n)
}

func (p *rpiDigitPin) openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, os.ModeExclusive)
}

func (p *rpiDigitPin) directionFile() (*os.File, error) {
	return p.openFile(path.Join(p.basePath(), "direction"))
}

func (p *rpiDigitPin) activeLowFile() (*os.File, error) {
	return p.openFile(path.Join(p.basePath(), "active_low"))
}

// SetDirection sets the direction of the pin (in/out).     in “type DigitalPin interface”
func (p *rpiDigitPin) SetDirection(dir embd.Direction) error {
	if err := p.init(); err != nil {
		return err
	}

	bcmNumber := uint(p.n)
	//In the datasheet at page 91 we find that the GPFSEL registers are organised per 10 pins.
	//So one 32-bit register contains the setup bits for 10 pins. *gpio.addr + ((g))/10 is
	// the register address that contains the GPFSEL bits of the pin "g"
	// Pin fsel register, 0 or 1 depending on bank
	fsel := (bcmNumber) / 10
	//There are three GPFSEL bits per pin (000: input, 001: output). The location
	//of these three bits inside the GPFSEL register is given by ((g)%10)*3
	shift := ((bcmNumber) % 10) * 3
	memlock.Lock()
	defer memlock.Unlock()

	if dir == embd.In {
		gpioArry[fsel] = gpioArry[fsel] &^ (7 << shift) //7:0b111 - pinmode is 3 bits
	} else {
		//This is also the reason that the comment says to "always use INP_GPIO(x) before using
		//OUT_GPIO(x)". This way you are sure that the other 2 bits are 0, and justifies the
		//use of a OR operation here. If you don't do that, you are not sure those bits will
		//be zero and you might have given the pin "g" a different setup.
		gpioArry[fsel] = gpioArry[fsel] &^ (7 << shift)
		gpioArry[fsel] = (gpioArry[fsel] &^ (7 << shift)) | (1 << shift)
	}

	//#define INP_GPIO(g)   *(gpio.addr + ((g)/10)) &= ~(7<<(((g)%10)*3))
	//#define OUT_GPIO(g)   *(gpio.addr + ((g)/10)) |=  (1<<(((g)%10)*3))

	return nil
}

// Read reads the value from the pin.    in “type DigitalPin interface”
func (pp *rpiDigitPin) Read() (int, error) {
	if err := pp.init(); err != nil {
		return 0, err
	}

	p := pp.n

	// Input level register offset (13 / 14 depending on bank)
	//In the datasheet on page 96, we seet that the GPLEVn register is
	//located 13 or 14 32-bit registers further than the gpio base register. GPLEV0 STORE 0~31,GPLEV1 STORE 32~53,
	levelReg := (p)/32 + 13

	if (gpioArry[levelReg] & (1 << uint8(p))) != 0 {
		return 1, nil
	}

	return 0, nil
}

// Write writes the provided value to the pin.   in “type DigitalPin interface”
func (pp *rpiDigitPin) Write(val int) error {
	if err := pp.init(); err != nil {
		return err
	}

	p := uint(pp.n)

	// Clear register, 10 / 11 depending on bank
	// Set register, 7 / 8 depending on bank
	//In the datasheet on page 90, we seet that the GPSET register is
	//located 10 32-bit registers further than the gpio base register. GPCLR0 STORE 0~31,GPCLR1 STORE 32~53,
	clearReg := p/32 + 10
	//In the datasheet on page 90, we seet that the GPSET register is
	//located 7 32-bit registers further than the gpio base register. GPSET0 STORE 0~31,GPSET1 STORE 32~53,
	setReg := p/32 + 7

	memlock.Lock()
	defer memlock.Unlock()

	if val == embd.Low {
		gpioArry[clearReg] = 1 << (p & 31)
	} else {
		gpioArry[setReg] = 1 << (p & 31)
	}
	return nil
}

// TimePulse measures the duration of a pulse on the pin.     in “type DigitalPin interface”
func (p *rpiDigitPin) TimePulse(state int) (time.Duration, error) {
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
func (p *rpiDigitPin) ActiveLow(b bool) error {
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
func (p *rpiDigitPin) PullUp() error {
	gpioPullMode(uint8(p.n), PullUp)
	return nil

}

// PullDown pulls the pin down.     in “type DigitalPin interface”
func (p *rpiDigitPin) PullDown() error {
	gpioPullMode(uint8(p.n), PullDown)
	return nil
}

// Close releases the resources associated with the pin.     in “type DigitalPin interface”
func (p *rpiDigitPin) Close() error {
	if err := p.StopWatching(); err != nil {
		return err
	}

	if err := p.drv.Unregister(p.id); err != nil {
		return err
	}

	if !p.initialized {
		return nil
	}

	//if err := p.dir.Close(); err != nil {
	//	return err
	//}
	//if err := p.val.Close(); err != nil {
	//	return err
	//}
	if err := p.activeLow.Close(); err != nil {
		return err
	}
	if err := p.unexport(); err != nil {
		return err
	}

	p.initialized = false

	return nil
}

func (p *rpiDigitPin) setEdge(edge embd.Edge) error {
	file, err := p.openFile(path.Join(p.basePath(), "edge"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(edge))
	return err
}

func (p *rpiDigitPin) Watch(edge embd.Edge, handler func(embd.DigitalPin)) error {
	return errors.New("gpio: not implemented")
	//if err := p.setEdge(edge); err != nil {
	//	return err
	//}
	//return registerInterrupt(p, handler)
}

func (p *rpiDigitPin) StopWatching() error {
	return errors.New("gpio: not implemented")
	//return unregisterInterrupt(p)
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

func gpioPullMode(bcmNumber uint8, pull int) {
	// Pull up/down/off register has offset 38 / 39, pull is 37
	pullClkReg := (bcmNumber)/32 + 38
	pullReg := 37
	shift := ((bcmNumber) % 32) // get 0 or 1 bank

	memlock.Lock()
	defer memlock.Unlock()

	switch pull {
	case PullDown, PullUp:
		gpioArry[pullReg] = gpioArry[pullReg]&^3 | uint32(pull)
	case PullOff:
		gpioArry[pullReg] = gpioArry[pullReg] &^ 3
	}

	// Wait for value to clock in, this is ugly, sorry :(
	time.Sleep(time.Microsecond)

	gpioArry[pullClkReg] = 1 << shift

	// Wait for value to clock in
	time.Sleep(time.Microsecond)

	gpioArry[pullReg] = gpioArry[pullReg] &^ 3
	gpioArry[pullClkReg] = 0

}
