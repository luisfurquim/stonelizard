package stonelizard

import (
	"fmt"
)

func (ba ByteArray) MarshalJSON() ([]byte, error) {
	var s string
	var b []byte
	s = fmt.Sprintf("%#v",ba)
	b = []byte(s)
	b[len(b)-1] = ']'
	b[6] = '['
	return b[6:], nil
}

