package stonelizard

import (
   "net/http"
)

// Creates a Dummy Unmarshaler
func NewDummyUnmarshaler(r *http.Request) *DummyUnmarshaler {
   return &DummyUnmarshaler{r:r.Body}
}

