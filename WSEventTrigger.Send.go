package stonelizard

import (
   "fmt"
   "errors"
)

func (wset *WSEventTrigger) Send(data interface{}) (err error) {

   err = nil

   defer func() {
      if r := recover(); r != nil {
         if fmt.Sprintf("%s",r) == "send on closed channel" {
            err = ErrorEndEventTriggering
         } else {
            err = errors.New(fmt.Sprintf("%s",r))
         }
      }
   }()

   if !wset.stat {
      err = ErrorStopEventTriggering
      return err
   }

   Goose.Serve.Logf(1,"Send will send %#v",data)
   wset.ch <- data
   Goose.Serve.Logf(1,"Send has sent %#v",data)

   return err
}
