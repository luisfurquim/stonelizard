package stonelizard

import (
   "fmt"
)


func (d *Static) Encode(v interface{}) error {
   var err error
   _, err = fmt.Fprintf(d.w,"%s",v)

   return err
}
