package stonelizard

import (
   "io"
   "fmt"
)


func (d *Static) Encode(v interface{}) error {
   var err error
   var src io.Reader
   var ch <-chan []byte
   var buf []byte
   var ok bool
   var n int64
   var m int
   var sum int
   var closer io.WriteCloser

   if src, ok = v.(io.Reader); ok {
      Goose.Serve.Logf(4,"Using reader")
      n, err = io.Copy(d.w, src)
      Goose.Serve.Logf(4,"Written %d bytes", n)
   } else if ch, ok = v.(<-chan []byte); ok {
      Goose.Serve.Logf(3,"Using channel")
      for buf = range ch {
         m, err = fmt.Fprintf(d.w,"%s",buf)
         sum += m
      }
      Goose.Serve.Logf(3,"Written %d bytes", sum)
   } else {
      Goose.Serve.Logf(4,"Using printf")
      m, err = fmt.Fprintf(d.w,"%s",v)
      Goose.Serve.Logf(4,"Written %d bytes", m)
   }

   if closer, ok = v.(io.WriteCloser); ok {
      closer.Close()
   }

   return err
}


