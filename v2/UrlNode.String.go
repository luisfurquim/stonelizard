package stonelizard

import (
   "fmt"
)

//Convert the Handle field value, from UrlNode struct, into a Go-syntax string (method signature)
func (u UrlNode) String() string {
   var s string

   if u.Handle != nil {
      s = fmt.Sprintf("Method: %#v\n",u.Handle)
   }

   return s
}



