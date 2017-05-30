package main

import (
	"bufio"
	"os"
	"strconv"

	"flag"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/r61526"

	_ "github.com/kidoman/embd/host/all"

	"fmt"
)

func main() {
	flag.Parse()
	embd.SetHost(embd.HostRPi, 50420202)
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	//for lcd display
	lcd, err := r61526.NewGpio(r61526.DefaultMap8080)

	if err != nil {
		panic(err)
	}
	defer lcd.Close()

	//dis := time.Now().Sub(start).Seconds()

	//s := time.Now()
	//dis := time.Now().Sub(s).Minutes()
	//fmt.Printf("touch %v;  spi %v;  irq %v;  \n", touch, spiBus, penIRQ)

	lcd.Clear(r61526.WHITE)
	//lcd.Clear(0x0000)
	lcd.DrawLine(100, 54, 220, 54, 0xF800)
	lcd.WriteAscii16x24Str(5, 0, "RST", 0xF800, 0x0000)

	modifyInput(lcd)
	//time.Sleep(1 * time.Minute)
}

func modifyInput(lcd *r61526.LCD) error {
	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "stop" {
			running = false
		}
		abb, _ := strconv.Atoi(command)
		fmt.Printf("received value is %v\n", abb)
		if abb == 1 {
			lcd.DrawLine(0, 0, 100, 300, r61526.BLACK)
			lcd.WriteAscii16x24Str(100, 20, "it is black line", r61526.BLACK, r61526.WHITE)
		} else if abb == 0 {
			lcd.Clear(0x0000)
			lcd.DrawLine(100, 54, 220, 54, 0xF800)
			lcd.WriteAscii16x24Str(5, 0, "RST", 0xF800, 0x0000)

			//lcd.DrawLine(0, 0, 100, 300, r61526.RED)
			//lcd.WriteAscii16x24Str(100, 20, "it is red line", r61526.RED, r61526.WHITE)
		} else if abb == 2 {
			lcd.DrawLine(0, 0, 100, 300, r61526.WHITE)
		} else if abb == 3 {
			lcd.DrawLine(100, 0, 100, 300, r61526.BLUE)
		} else if abb == 4 {
			lcd.DrawLine(0, 300, 100, 300, r61526.MAGENTA)
		} else if abb == 5 {
			lcd.DrawLine(0, 0, 100, 300, r61526.GREEN)
		} else if abb == 6 {
			lcd.DrawLine(0, 0, 100, 300, r61526.CYAN)
		} else if abb == 7 {
			lcd.DrawLine(0, 0, 100, 300, r61526.YELLOW)
		}
	}
	return nil
}
