package stonelizard

import (
   "fmt"
   "errors"
)

// Call this method from your websocket to trigger an event
// all parameters data will be sent through the websocket to the
// client.
// The data parameters MUST be compliant (length and types) to what you defined
// in the StructTag for your particular event.
func (wset *WSEventTrigger) Trigger(data ...interface{}) (err error) {
	var lbuf string

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

   // If the application decided to end the operation
   // we must refuse to send the data
   if !wset.Status {
      err = ErrorStopEventTriggering
      return err
   }

   // Send data to wsEventHandle.1 which has access to the HTTP/Websocket connection
	lbuf = fmt.Sprintf("%#v", data)
	if len(lbuf) > 120 {
		lbuf = lbuf[:120]
	}

   Goose.Serve.Logf(4,"Send will send %s",data)
   wset.EventData <- data
   Goose.Serve.Logf(5,"Send has sent data")

   return err
}
