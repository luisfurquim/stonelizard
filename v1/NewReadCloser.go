package stonelizard

import (
   "bytes"
)

func NewReadCloser(src []byte) ReadCloser {
   return ReadCloser{ Rd: bytes.NewReader(src)}
}

func (rc ReadCloser) Read(p []byte) (n int, err error) {
   return rc.Rd.Read(p)
}

func (rc ReadCloser) Close() error {
   return nil
}

