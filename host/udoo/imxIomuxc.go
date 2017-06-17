//+build ignore

package udoo

//e.g Pad Mux Register (IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO00)    1717
//e.g Pad Control Register (IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO00).2004

//pad control register ：pad控制
//pad mux register：pad 复用
//select input register：输入模式选择

//iomuxc memory map
const (
	iomuxcBase = 0x020E0000 //   p187 define iomuxc size is 16k
)
