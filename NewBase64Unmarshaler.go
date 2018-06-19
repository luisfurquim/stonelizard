package stonelizard

import (
   "net/http"
   "encoding/base64"
)

// Creates a Base64 Unmarshaler
func NewBase64Unmarshaler(r *http.Request) *Base64Unmarshaler {
   return &Base64Unmarshaler{r:base64.NewDecoder(base64.StdEncoding,r.Body)}
}

