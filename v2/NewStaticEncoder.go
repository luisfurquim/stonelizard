package stonelizard

import (
   "io"
)

func NewStaticEncoder(w io.Writer) *Static {
   return &Static{
      w: w,
   }
}
