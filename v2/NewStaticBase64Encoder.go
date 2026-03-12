package stonelizard

import (
   "io"
   "encoding/base64"
)

func NewStaticBase64Encoder(w io.Writer) *StaticBase64 {
   return &StaticBase64{
      w: base64.NewEncoder(base64.StdEncoding, w),
   }
}
