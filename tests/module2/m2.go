package module2

import (
	"BCDns_0.1/tests/modules"
	"fmt"
)

var (
	J int
)

func init() {
	fmt.Println("m2", modules.I, J)
}
