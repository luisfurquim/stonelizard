package stonelizard

import (
	"strconv"
)

func (ba ByteArray) MarshalJSON() ([]byte, error) {
	var s string = "["
	var c byte

	if len(ba)>0 {
		s += strconv.Itoa(int(ba[0]))
	}

	for _, c = range ba[1:] {
		s += "," + strconv.Itoa(int(c))
	}

	if len(s) < 120 {
		Goose.Serve.Logf(4,"ByteArray json: %s", s[21:])
	} else {
		Goose.Serve.Logf(4,"ByteArray json~: %s", s[21:120])
	}
	
	return []byte(s+"]"), nil
}

