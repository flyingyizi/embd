package main

import (
	"bufio"
	"os"
	"strconv"

	"flag"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/r61526"
	"github.com/kidoman/embd/convertors/xpt2046"

	_ "github.com/kidoman/embd/host/all"

	"fmt"
)

const (
	channel = 0
	speed   = 500000
	bpw     = 8
	delay   = 0
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

	// for lcd touch
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus := embd.NewSPIBus(embd.SPIMode0, channel, speed, bpw, delay)
	defer spiBus.Close()

	touch := xpt2046.New(spiBus)

	if err = touch.Watch("P1_37"); err != nil {
		panic(err)
	}
	defer touch.StopWatching()
	//dis := time.Now().Sub(start).Seconds()

	//s := time.Now()
	//dis := time.Now().Sub(s).Minutes()
	//fmt.Printf("touch %v;  spi %v;  irq %v;  \n", touch, spiBus, penIRQ)

	lcd.Clear(r61526.WHITE)
	//lcd.Clear(0x0000)
	lcd.DrawLine(100, 54, 220, 54, 0xF800)
	lcd.WriteAscii16x24Str(5, 0, "RST", 0xF800, 0x0000)
	go func(tft *r61526.LCD, xpt *xpt2046.XPT2046) {
		for {
			select {
			//case v, ok := <-xpt.XY:

			case v := <-xpt.XY:
				//	if ok {
				//fmt.Printf("x: %+v ;y: %+v； \n", v.X, v.Y)
				tft.DrawDot(uint16(v.X), uint16(v.Y), 0x001F)

				tft.WriteAscii16x24Str(0, 170, "X:", 0xF800, r61526.WHITE)
				tft.WriteAscii16x24Str(32, 170, "           ", 0xF800, r61526.WHITE)
				tft.WriteAscii16x24Str(32, 170, strconv.Itoa(int(v.X)), 0xF800, r61526.WHITE)
				tft.WriteAscii16x24Str(0, 195, "Y:", 0xF800, r61526.WHITE)
				tft.WriteAscii16x24Str(32, 195, "           ", 0xF800, r61526.WHITE)
				tft.WriteAscii16x24Str(32, 195, strconv.Itoa(int(v.Y)), 0xF800, r61526.WHITE)
				//	}
			}
		}
	}(lcd, touch)

	modifyInput(lcd)

	//touchMody(lcd, touch)
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

func touchMody(lcd *r61526.LCD, x, y int) {
	//if rst == 1
	{
		lcd.Clear(r61526.WHITE)
		//lcd.Clear(0x0000)
		lcd.DrawLine(100, 54, 220, 54, 0xF800)
		lcd.WriteAscii16x24Str(5, 0, "RST", 0xF800, 0x0000)
		//rst = 0
	}

	//quit <- p
	//x = (x - 256) * 320 / 3638
	//y = (y - 160) * 480 / 3716

	//x, _ := t.ReadX()
	//y, _ := t.ReadY()
	fmt.Printf("x: %+v ;y: %+v； \n", x, y)
	//x, y, err := t.TOUCH_XPT_ReadXY()
	//z1, _ := t.ReadZ1()
	//z2, _ := t.ReadZ2()
	//pressure, _ := t.ReadTouchPressure()
	//fmt.Printf("x: %+v ;y: %+v;z1: %+v;z2: %+v;press: %+v:\n", x, y, z1, z2, pressure)

	//lcd.DrawDot(uint16(319-x), uint16(y), 0x001F)
	lcd.DrawDot(uint16(x), uint16(y), 0x001F)
	//--显示AD值--//
	lcd.WriteAscii16x24Str(0, 170, "X:", 0xF800, 0x0000)
	lcd.WriteAscii16x24Str(32, 170, "           ", 0xF800, 0x0000)
	lcd.WriteAscii16x24Str(32, 170, strconv.Itoa(int(x)), 0xF800, 0x0000)
	lcd.WriteAscii16x24Str(0, 195, "Y:", 0xF800, 0x0000)
	lcd.WriteAscii16x24Str(32, 195, "           ", 0xF800, 0x0000)
	lcd.WriteAscii16x24Str(32, 195, strconv.Itoa(int(y)), 0xF800, 0x0000)

	//	fmt.Printf(" %v was pressed.\n", <-quit)

	/*
		var rst = 1

		for true {
			if rst == 1 {
				lcd.Clear(r61526.WHITE)

				//lcd.Clear(0x0000)
				lcd.DrawLine(100, 54, 220, 54, 0xF800)
				lcd.WriteAscii16x24Str(5, 0, "RST", 0xF800, 0x0000)
				rst = 0
			}
			x, y, err := t.TOUCH_XPT_ReadXY()
			//x = (x - 256) * 320 / 3638
			//y = (y - 160) * 480 / 3716

			//x, _ := t.ReadX()
			//y, _ := t.ReadY()
			if err == nil {
				fmt.Printf("x: %+v ;y: %+v； \n", x, y)
				//x, y, err := t.TOUCH_XPT_ReadXY()
				z1, _ := t.ReadZ1()
				z2, _ := t.ReadZ2()
				pressure, _ := t.ReadTouchPressure()
				fmt.Printf("x: %+v ;y: %+v;z1: %+v;z2: %+v;press: %+v:\n", x, y, z1, z2, pressure)

				//lcd.DrawDot(uint16(319-x), uint16(y), 0x001F)
				lcd.DrawDot(uint16(x), uint16(y), 0x001F)
				//--显示AD值--//
				lcd.WriteAscii16x24Str(0, 170, "X:", 0xF800, 0x0000)
				lcd.WriteAscii16x24Str(32, 170, "           ", 0xF800, 0x0000)
				lcd.WriteAscii16x24Str(32, 170, strconv.Itoa(int(x)), 0xF800, 0x0000)
				lcd.WriteAscii16x24Str(0, 195, "Y:", 0xF800, 0x0000)
				lcd.WriteAscii16x24Str(32, 195, "           ", 0xF800, 0x0000)
				lcd.WriteAscii16x24Str(32, 195, strconv.Itoa(int(y)), 0xF800, 0x0000)
			}
				if err == nil {
					//--如果触摸跟显示发生偏移，可以根据显示AD值
					//--调整下面公式里面的数值--//
					x = (x - 256) * 320 / 3638
					y = (y - 160) * 480 / 3716

					if x > 319 {
						x = 318
					}
					if y > 479 {
						y = 478
					}
					if (x > 280) && (y < 30) {
						rst = 1
					} else {
						lcd.DrawDot((319 - x), y, 0x001F)
						//--计算读取到的AD值--//
						//--由于添加了显示AD值，计算需要时间，所以触摸有一点延迟--//xpt_xy.
						//xValue[1] = byte((x % 10000 / 1000)) + '0'
						//xValue[2] = byte((x % 1000 / 100)) + '0'
						//xValue[3] = byte((x % 100 / 10)) + '0'
						//xValue[4] = byte((x % 10)) + '0'

						//yValue[1] = byte((y % 10000 / 1000)) + '0'
						//yValue[2] = byte((y % 1000 / 100)) + '0'
						//yValue[3] = byte((y % 100 / 10)) + '0'
						//yValue[4] = byte((y % 10)) + '0'
						fmt.Printf("x: %+v ;y: %+v:\n", x, y)
						//--显示AD值
						lcd.WriteAscii16x24Str(0, 170, "X:", 0xF800, 0x0000)
						lcd.WriteAscii16x24Str(32, 170, "           ", 0xF800, 0x0000)
						lcd.WriteAscii16x24Str(32, 170, strconv.Itoa(int(x)), 0xF800, 0x0000)
						lcd.WriteAscii16x24Str(0, 195, "Y:", 0xF800, 0x0000)
						lcd.WriteAscii16x24Str(32, 195, "           ", 0xF800, 0x0000)
						lcd.WriteAscii16x24Str(32, 195, strconv.Itoa(int(y)), 0xF800, 0x0000)
					}
				}
		}
	*/
}
