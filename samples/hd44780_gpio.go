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

/* circuit connects  with "RPI 3+" GPIO .
   +---------+                    +----------+
   |         |VSS                 |          |
   |         |-------GND          |          |
   |         |VDD                 |          |
   |         |-------+5V          |          |
   |         |V0                  |          |
   |         |-------GND          |          |
   |         |RS           GPIO_22|          |
   |         |--------------------|          |
   |         |RW                  | MCU      |
   |         |-------GND          |          |
   |         |E            GPIO_23|          |
   |LCD16X2  |--------------------|          |
   |         |                    |          |
   |hd44780  |                    |          |
   |         |D4           GPIO_24|          |
   |         |--------------------|          |
   |         |D5           GPIO_25|          |
   |         |--------------------|          |
   |         |D6           GPIO_26|          |
   |         |--------------------|          |
   |         |D7           GPIO_27|          |
   |         |--------------------|          |
   |         |A                   +----------+
   |         |------+5V
   |         |K
   |         |------GND
   +---------+
*/

func main() {
	flag.Parse()
	embd.SetHost(embd.HostRPi, 50420202)
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	hd, err := hd44780.NewGPIO("GPIO_22" /*rs*/, "GPIO_23", /*en*/
		"GPIO_24" /*d4*/, "GPIO_25" /*d5*/, "GPIO_26" /*d6*/, "GPIO_27" /*d7*/, nil, /*backlight*/
		false, /*BacklightPolarity*/
		hd44780.RowAddress16Col, /*RowAddress*/
		hd44780.TwoLine,
		hd44780.BlinkOn,
	)
	if err != nil {
		panic(err)
	}
	defer hd.Close()

	hd.Clear()

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
