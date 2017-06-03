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


*/

// ModelT :  Raspberry Pi Revision :: Model
type ModelT int

const (
	//ModelNeoBasic   Udoo neo basic
	ModelNeoBasic ModelT = iota
	//ModelNeoExtend   Udoo neo extend
	ModelNeoExtend
	//ModelNeoFull   udoo neo full,	//  2
	ModelNeoFull
	//Modelquad  Udoo quad
	Modelquad
	//ModelDual   Udoo dual
	ModelDual
	//Modelx86    Udoo x86
	Modelx86
	//ModelSecosbcA62  Udoo secosbc a62
	ModelSecosbcA62
)

// MemoryT :  Raspberry Pi Revision :: memeory type
type MemoryT int

// the value is from PRI post PI2 revision
const (
	UnknownMB MemoryT = -1
	M512MB    MemoryT = 1
	M1024MB   MemoryT = 2
)

//InfoT  store board information
type InfoT struct {
	Model            ModelT
	HasM4, HasLvds15 bool
	Uid              string
	/*	mem       MemoryT
		processor ProcessorT
		revision uint64
	*/
}

//ModelName  map to board name
func (info *InfoT) ModelName() (modelname string) {
	return modelName[info.Model]
}

var modelName = map[ModelT]string{
	ModelNeoBasic:  "Udoo Neo basic",
	ModelNeoExtend: "Udoo Neo extend",
	ModelNeoFull:   "Udoo Neo full",

	Modelx86:        "Udoo x86",
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

//GetBoardInfo  returns board information
func GetBoardInfo() (info InfoT /*, periphereBase int64*/, err error) {
	model, err := ioutil.ReadFile("/proc/device-tree/model")
	if err != nil {
		return
	}
	str := strings.TrimSpace(string(model))

	switch str {
	case "UDOO Quad Board":
		info.Model = Modelquad
		info.HasM4 = false
		info.HasLvds15 = true
	case "UDOO Dual-lite Board":
		info.Model = ModelDual
		info.HasM4 = false
		info.HasLvds15 = true
	case "UDOO Neo Extended":
		info.Model = ModelNeoExtend
		info.HasM4 = true
		info.HasLvds15 = false
	case "UDOO Neo Full":
		info.Model = ModelNeoFull
		info.HasM4 = true
		info.HasLvds15 = false
	case "UDOO Neo Basic Kickstarter":
		fallthrough
	case "UDOO Neo Basic":
		info.Model = ModelNeoBasic
		info.HasM4 = true
		info.HasLvds15 = false
	}

	if h, err := ioutil.ReadFile("/sys/fsl_otp/HW_OCOTP_CFG0"); err == nil {
		if l, err := ioutil.ReadFile("/sys/fsl_otp/HW_OCOTP_CFG1"); err == nil {
			hs := strings.TrimSpace(string(h))
			ls := strings.TrimSpace(string(l))
			info.Uid = hs + ls
		}
	}

	return
}
