package stonelizard

import (
   "fmt"
)

//Convert the UrlNode field value, from the Service struct, into a string
func (svc Service) String() string {
   var s string

   if svc.Config.PageNotFound() != nil {
      s = "404: " + string(svc.Config.PageNotFound()) + "\n"
   }

   s += fmt.Sprintf("%s",svc.Svc)
   return s
}

