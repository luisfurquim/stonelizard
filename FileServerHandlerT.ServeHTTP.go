package stonelizard

import (
	"os"
   "fmt"
   "strings"
   "runtime"
   "net/http"
   "path/filepath"
)

func (fs FileServerHandlerT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   var err error
   var httpstat int
   var extAuth ExtAuthT
   var authparms map[string]interface{}
   var qry, header string
   var ok bool
   var etag string
   var fi os.FileInfo
   var exported string

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
               Goose.Serve.Logf(1,"panic loop %d/%d", i, len(srcs))
               src = srcs[i]
               if len(src) > 0 {
                  break
               }
            }
         }

         Goose.Serve.Logf(1,"panic (%s): calling %s -> %s @ %s", panicerr, r.URL.Path, src)
      }
   }()


   Goose.Serve.Logf(2,"svc.Access: %s from %s with level %d", r.RequestURI,  r.RemoteAddr, fs.svc.Access)
   if fs.svc.Access != AccessNone && fs.svc.Access != AccessInfo {

      r.ParseForm()
      authparms =  map[string]interface{}{}
      for qry, _ = range r.Form {
         authparms[qry] = r.Form[qry][0]
      }

      for header, _ = range r.Header {
         authparms[header] = r.Header[header][0]
      }

//      Goose.Serve.Fatalf(0,"authparms: %#v", authparms)

      if extAuth, ok = fs.svc.Authorizer.(ExtAuthT); ok {
         httpstat, _, err = extAuth.ExtAuthorize(fs.svc.ch, fs.path, authparms, w, r, fs.svc.SavePending)
      } else {
         httpstat, _, err = fs.svc.Authorizer.Authorize(fs.path, nil, r.RemoteAddr, r.TLS, fs.svc.SavePending)
      }

      if err != nil {
         Goose.Serve.Logf(1,"svc.Access: %s from %s with level %d error: %s", r.RequestURI,  r.RemoteAddr, fs.svc.Access, err)
         w.WriteHeader(httpstat)
         w.Write([]byte(fmt.Sprintf("%s",err)))
         return
      }
   }

	if fs.exported == "" {
		fs.exported = "."
	}

	if (len(fs.path) > len(r.URL.Path)) || (fs.path != r.URL.Path[:len(fs.path)]) {
		Goose.Serve.Logf(1,"Invalid path %s", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(fmt.Sprintf("%s",err)))
//		w.Write([]byte(fs.exported + "/" + r.URL.Path))
		return
	}

	exported, err = filepath.Abs(fs.exported)
	Goose.Serve.Logf(1,"exported: [%s], path: [%s], err:[%s]", exported, r.URL.Path[len(fs.path):], err)

	if err != nil {
		Goose.Serve.Logf(1,"Error normalizing exported path %s: %s", exported, err)
		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte(fmt.Sprintf("%s",err)))
//		w.Write([]byte(fs.exported + "/" + r.URL.Path))
		return
	}

	fi, err = os.Stat(exported + "/" + r.URL.Path[len(fs.path):])
	if err != nil {
		Goose.Serve.Logf(1,"Error stating path %s: %s", exported + "/" + r.URL.Path[len(fs.path):], err)
		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte(fmt.Sprintf("%s",err)))
//		w.Write([]byte(exported + "/" + r.URL.Path))
		return
	}

	etag = `"` + fi.ModTime().Format("20060102150405") + `"`
	w.Header().Set("Etag", etag)


   if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

   fs.hnd.ServeHTTP(w,r)
}
