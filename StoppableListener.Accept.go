package stonelizard

import (
   "net"
   "time"
)

// http://www.hydrogen18.com/blog/stop-listening-http-server-go.html
//Implements a wrapper on the system accept
func (sl *StoppableListener) Accept() (net.Conn, error) {

   for {
      Goose.Listener.Logf(6,"Start accept loop")

      //Wait up to one second for a new connection
      sl.SetDeadline(time.Now().Add(5*time.Second))

      newConn, err := sl.TCPListener.Accept()

      //Check for the channel being closed
      select {
         case <-sl.stop:
            Goose.Listener.Logf(2,"Received stop request")
            return nil, ErrorStopped
         default:
            Goose.Listener.Logf(6,"channel still open")
            // If the channel is still open, continue as normal
      }

      if err != nil {
         Goose.Listener.Logf(6,"Err not nil: %s",err)
         netErr, ok := err.(net.Error)

         // If this is a timeout, then continue to wait for
         // new connections
         if ok && netErr.Timeout() && netErr.Temporary() {
            Goose.Listener.Logf(6,"continue")
            continue
         }
      }

      if err==nil {
         Goose.Listener.Logf(3,"done listening")
         return newConn, nil
      }

      Goose.Listener.Logf(2,"done listening, err=%s",err)
      return nil, err
   }
}

