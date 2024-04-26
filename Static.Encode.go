package stonelizard

import (
   "io"
   "fmt"
   "bytes"
   "strings"
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
   var s string
   var sp *string
   var bufp *[]byte
   var closer io.WriteCloser

   if src, ok = v.(io.Reader); ok {
      Goose.Serve.Logf(4,"Using reader")
      n, err = io.Copy(d.w, src)
   } else if ch, ok = v.(<-chan []byte); ok {
      Goose.Serve.Logf(3,"Using channel")
      for buf = range ch {
         m, err = fmt.Fprintf(d.w,"%s",buf)
         sum += m
      }
      m = sum
   } else {
      Goose.Serve.Logf(4,"Using buffer")
      if s, ok = v.(string) ; ok {
			src = strings.NewReader(s)
		} else if buf, ok = v.([]byte) ; ok {
			src = bytes.NewReader(buf)
		} else if sp, ok = v.(*string) ; ok {
			src = strings.NewReader(*sp)
		} else if bufp, ok = v.(*[]byte) ; ok {
			src = bytes.NewReader(*bufp)
		} else {
			src = strings.NewReader(fmt.Sprintf("%s",v))
		}
      n, err = io.Copy(d.w, src)
   }

	Goose.Serve.Logf(4,"Written %d bytes", n)
	if err != nil {
		Goose.Serve.Logf(1,"Error writing body: %s", err)
	}

   if closer, ok = d.w.(io.WriteCloser); ok {
      closer.Close()
   }

   return err
}


