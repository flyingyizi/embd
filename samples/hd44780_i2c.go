// +build ignore

package main

import (
	"bufio"
	"flag"
	"os"
	"time"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/hd44780"

	_ "github.com/kidoman/embd/host/all"
)

/* circuit connects  with "RPI 3+" i2c .
+---------+
|         |VSS
|         |-------GND
|         |VDD                                +----------+
|         |-------+5V                         |          |
|         |V0            +--------+           |          |
|         |-------GND    |        |           |          |
|         |RS          P0|        |sda   sda.1|          |
|         |--------------|        |-----------|          |
|         |RW          P1|        |           | MCU      |
|         |--------------|        |           |          |
|         |E           P2|PCF8574 |scl   scl.1|          |
|LCD16X2  |--------------|        |-----------|          |
|hd44780  |              |        |           |          |
|         |D4          P4|        |vcc        |          |
|         |--------------|        |----+5v    |          |
|         |D5          P5|        |           |          |
|         |--------------|        |GND        |          |
|         |D6          P6|        |----GND    |          |
|         |--------------|        |           |          |
|         |D7          P7|        |           |          |
|         |--------------|        |           |          |
|         |A             +--------+           +----------+
|         |------+5V
|         |K
|         |------GND
+---------+


*/

func main() {
	flag.Parse()
	embd.SetHost(embd.HostRPi, 50420202)

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	hd, err := hd44780.NewI2C(
		bus,
		0x20,
		hd44780.PCF8574PinMap,
		hd44780.RowAddress20Col,
		hd44780.TwoLine,
		hd44780.BlinkOn,
	)
	if err != nil {
		panic(err)
	}
	defer hd.Close()

	hd.Clear()
	/*
		message := "Hello, world!"
		bytes := []byte(message)
		for _, b := range bytes {
			hd.WriteChar(b)
		}
		hd.SetCursor(0, 1)

		message = "@embd | hd44780"
		bytes = []byte(message)
		for _, b := range bytes {
			hd.WriteChar(b)
		}
		time.Sleep(10 * time.Second)
	*/
	outPut(hd)
	hd.BacklightOff()
}

func outPut(hd *hd44780.HD44780) {
	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "stop" {
			running = false
		}
		hd.Clear()

		bytes := []byte(command)
		l := len(bytes)
		for _, b := range bytes {
			hd.WriteChar(b)
		}

		for n := l - 16; n > 0; n-- {
			hd.ShiftLeft()
			time.Sleep(500 * time.Millisecond)
		}
		for n := l - 16; n > 0; n-- {
			hd.ShiftRight()
			time.Sleep(500 * time.Millisecond)
		}

	}

}
