package stonelizard

import (
	"net/http"
)

type SLFileSystem struct {
	Dir http.Dir
}

func (fs SLFileSystem) Open(name string) (http.File, error) {
	var fh http.File
	var err error

	fh, err = fs.Dir.Open(name)

	Goose.Serve.Logf(0,"Open: %s -> %#v [%s]", name, fh, err)

	return fh, err
}

