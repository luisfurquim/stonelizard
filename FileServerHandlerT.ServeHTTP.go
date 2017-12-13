package stonelizard

import (
   "fmt"
   "strings"
   "runtime"
   "net/http"
)

func (fs FileServerHandlerT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   var err error
   var httpstat int

   defer func() {
      // Tries to survive to any panic from the application.
      // Additionally, tries to extract useful data from the panic message and log it.
      // This is needed because the panic message is too long and have newlines in it,
      // when such messages reach the system log, only the first line goes into the log.
      // In my experience, the most needed data is where in my code the panic ocurred,
      // so here we search for the first line referencing a GOLANG sourcecode which is not
      // a system package and select it to put in the log
      // TODO: add the error message to the log too
      if panicerr := recover(); panicerr != nil {
         const size = 64 << 10
         var buf []byte
         var srcs, srcs2 []string
         var src string
         var i int

         buf  = make([]byte, size)
         buf  = buf[:runtime.Stack(buf, false)]
         srcs = strings.Split(string(buf),"\n")
         for _, src = range srcs {
            if gosrcRE.MatchString(src) && (!gorootRE.MatchString(src)) {
               srcs2 = append(srcs2,gosrcFNameRE.FindStringSubmatch(src)[1])
            }
         }

         src = strings.Join(srcs2,", ")

         if len(src) == 0 {
            for i=len(srcs)-1; i>0; i-- {
               Goose.Serve.Logf(0,"panic loop %d/%d", i, len(srcs))
               src = srcs[i]
               if len(src) > 0 {
                  break
               }
            }
         }

         Goose.Serve.Logf(0,"panic (%s): calling %s -> %s @ %s", panicerr, r.URL.Path, src)
      }
   }()


   Goose.Serve.Logf(5,"svc.Access: %d",fs.svc.Access)
   if fs.svc.Access != AccessNone {
      httpstat, _, err = fs.svc.Authorizer.Authorize(fs.path, nil, r.RemoteAddr, r.TLS, fs.svc.SavePending)
      if err != nil {
         w.WriteHeader(httpstat)
         w.Write([]byte(fmt.Sprintf("%s",err)))
         return
      }
   }

   fs.hnd.ServeHTTP(w,r)
}
