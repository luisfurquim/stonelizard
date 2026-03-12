package stonelizard

import (
   "net"
)

//Init the listener for the service.
func NewListener(l net.Listener) (*StoppableListener, error) {
   tcpL, ok := l.(*net.TCPListener)

   if !ok {
      return nil, ErrorCannotWrapListener
   }

   retval := &StoppableListener{}
   retval.TCPListener = tcpL
   retval.stop = make(chan int)

   return retval, nil
}

