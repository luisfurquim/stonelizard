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


	if len(b) < 120 {
		Goose.Serve.Logf(4,"ByteArray json: %s", b[6:])
	} else {
		Goose.Serve.Logf(4,"ByteArray json~: %s", b[6:120])
	}
	
	return b[6:], nil
}

