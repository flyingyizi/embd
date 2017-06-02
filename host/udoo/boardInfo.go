package udoo

import (
	"io/ioutil"
	"strings"
)

/*
udooer@udooneo:~$ cat  /proc/cpuinfo
processor	: 0
model name	: ARMv7 Processor rev 10 (v7l)
BogoMIPS	: 7.54
Features	: swp half thumb fastmult vfp edsp thumbee neon vfpv3 tls vfpd32
CPU implementer	: 0x41
CPU architecture: 7
CPU variant	: 0x2
CPU part	: 0xc09
CPU revision	: 10

Hardware	: Freescale i.MX6 SoloX (Device Tree)
Revision	: 0000
Serial		: 0000000000000000


udooer@udooneo:/opt/udoo-web-conf/shscripts$ cat serial.sh
#!/bin/bash

H=`cat /sys/fsl_otp/HW_OCOTP_CFG0 |sed -e 's/0x//'`
L=`cat /sys/fsl_otp/HW_OCOTP_CFG1 |sed -e 's/0x//'`

SERIAL=$H$L
SERIAL=${SERIAL^^}

echo $SERIAL

udooer@udooneo:/opt/udoo-web-conf/shscripts$ sudo ./serial.sh
[sudo] password for udooer:
DF669959170DC1D4
udooer@udooneo:/opt/udoo-web-conf/shscripts$
*/

const (
	PeripheralBaseUnknown int64 = 0
	PeripheralBase2835    int64 = 0x20000000
	PeripheralBase2836    int64 = 0x3f000000
	PeripheralBase2837    int64 = 0x3f000000
)

// ModelT :  Raspberry Pi Revision :: Model
type ModelT int

const (
	ModelNeoBasic  ModelT = iota // udoo neo basic,	//  0
	ModelNeoExtend               // udoo neo extend,	//  1
	ModelNeoFull                 // udoo neo full,	//  2
	Modelquad
	ModelDual

	Modelx86
	ModelSecosbcA62
)

// MemoryT :  Raspberry Pi Revision :: memeory type
type MemoryT int

// the value is from PRI post PI2 revision
const (
	UnknownMB MemoryT = -1
	512MB     MemoryT = 1
	1024MB    MemoryT = 2
)

type ProcessorT int

const (
	Unknown ProcessorT = -1
	IMX6    ProcessorT = 0
)

type InfoT struct {
	model            ModelT
	hasM4, hasLvds15 bool
	/*	mem       MemoryT
		processor ProcessorT
		revision uint64
	*/
}

func (info *InfoT) ModelName() (modelname string) {
	return ModelName[info.model]
}

var ModelName = map[ModelT]string{
	ModelNeoBasic:  "Udoo Neo basic",
	ModelNeoExtend: "Udoo Neo extend",
	ModelNeoFull:   "Udoo Neo full",

	Modelx86:        "model",
	Modelquad:       "",
	ModelDual:       "",
	ModelSecosbcA62: "",
}

func getRevision() (revision string, err error) {
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return
	}

	lines := strings.Split(string(cpuinfo), "\n")
	for _, l := range lines {
		fields := strings.Split(l, ":")
		if len(fields) == 2 {
			k := strings.TrimSpace(fields[0])
			v := strings.TrimSpace(fields[1])
			if k == "Revision" {
				revision = v
				break
			}
		}
	}
	return
}

func GetBoardInfo() (info InfoT /*, periphereBase int64*/, err error) {
	model, err := ioutil.ReadFile("/proc/device-tree/model")
	if err != nil {
		return
	}
	str := strings.TrimSpace(string(model))

	switch str {
	case "UDOO Quad Board":
		info.model = Modelquad
		info.hasM4 = false
		info.hasLvds15 = true
	case "UDOO Dual-lite Board":
		info.model = ModelDual
		info.hasM4 = false
		info.hasLvds15 = true
	case "UDOO Neo Extended":
		info.model = ModelNeoExtend
		info.hasM4 = true
		info.hasLvds15 = false
	case "UDOO Neo Full":
		info.model = ModelNeoFull
		info.hasM4 = true
		info.hasLvds15 = false
	case "UDOO Neo Basic Kickstarter":
		fallthrough
	case "UDOO Neo Basic":
		info.model = ModelNeoBasic
		info.hasM4 = true
		info.hasLvds15 = false
	}

	return
}
