package stonelizard

import (
   "io"
   "fmt"
)


func (d *Static) Encode(v interface{}) error {
   var err error
   var src io.Reader
   var ok bool

   if src, ok = v.(io.Reader); ok {
      _, err = io.Copy(d.w, src)
   } else {
      _, err = fmt.Fprintf(d.w,"%s",v)
   }

   return err
}


