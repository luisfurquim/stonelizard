package stonelizard

import (
   "io"
   "fmt"
   "runtime"
   "strings"
   "reflect"
   "net/http"
   "encoding/base64"
)


// Constructs the operation handler:
// a) prepares the data received,
// b) calls the application method
// c) prepares the data returned
// d) send it back to the caller
func buildHandle(this reflect.Value, isPtr bool, met reflect.Method, posttype []reflect.Type, accesstype uint8, useWebSocket bool) (func ([]interface{}, Unmarshaler, interface{}, string, string) Response) {
   return func (parms []interface{}, Unmarshal Unmarshaler, authinfo interface{}, host string, remoteAddr string) Response {
      var httpResp Response
      var j int
//       var outs []reflect.Value
      var ins []reflect.Value
      var err error
      var postvalue reflect.Value
      var errmsg string
      var num int
      var rval reflect.Value
      var sIns string
      var enc []string
      var decoded []byte

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
            var parms string
            var p reflect.Value
            var tmp string
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

            for _, p = range ins[1:] {
               tmp = fmt.Sprintf(" %#v; ", p.Interface())
               if len(tmp) > 40 {
                  tmp = tmp[:40] + "; "
               }
               parms += tmp
            }

            Goose.Serve.Logf(1,"panic (%s): calling %s -> %s with %s @ %s", panicerr, met.PkgPath, met.Name, parms, src)
         }
      }()

      // Prepare the calling parameters
      if isPtr {
         ins, err = pushParms(parms, this, met)
      } else {
         ins, err = pushParms(parms, this.Elem(), met)
      }

      // On error return HTTP error message
      if err != nil {
         httpResp.Status = http.StatusInternalServerError
         httpResp.Body   = "Internal server error"
         httpResp.Header = map[string]string{}
         return httpResp
      }

      // Just debugging ;)
      if this.Type().Kind() == reflect.Struct {
         if _, ok := this.Type().FieldByName("Session"); ok {
            if this.Kind()==reflect.Ptr {
               Goose.OpHandle.Logf(1,"THIS SESSION: %#v",this.Elem().FieldByName("Session"))
            } else {
               Goose.OpHandle.Logf(1,"THIS SESSION: %#v",this.FieldByName("Session"))
            }
         }
      }


//                  Goose.OpHandle.Logf(1,"posttype: %#v, postdata: %s",posttype, postdata)
      Goose.OpHandle.Logf(5,"posttype: %#v",posttype)
      if posttype != nil { // Adds data sent through HTTP POST
         err = nil
         j   = 0
         for (err==nil) && (j<len(posttype)) {
            Goose.OpHandle.Logf(6,"posttype[%d]: %#v",j,posttype[j])
            // Allocate room for the next parameter
            postvalue = reflect.New(posttype[j])
            // Decode it from the HTTP body
            err = Unmarshal.Decode(postvalue.Interface())
            if err != nil && err != io.EOF {
               // Return HTTP error
               httpResp.Status = http.StatusInternalServerError
               httpResp.Body   = "Internal server error"
               httpResp.Header = map[string]string{}

               Goose.OpHandle.Logf(0,"posttype[%d]: %#v",j,posttype[j])
               if postvalue.Kind() == reflect.Ptr && !postvalue.IsNil() {
                  Goose.OpHandle.Logf(0,"posttype: %#v",postvalue.Elem().Kind())
                  Goose.OpHandle.Logf(1,"Internal server error parsing post body: %s - postvalue: %#v",err,postvalue.Elem().Interface())
               } else {
                  Goose.OpHandle.Logf(1,"Internal server error parsing post body: %s - postvalue: %#v",err,postvalue.Interface())
               }
               return httpResp
            }

            if err != io.EOF {
               Goose.OpHandle.Logf(6,"postvalue.Kind() = %d",reflect.Indirect(postvalue).Kind())
               Goose.OpHandle.Logf(8,"postvalue = %s",reflect.Indirect(postvalue).Interface())
               if reflect.Indirect(postvalue).Kind() == reflect.String {
                  enc = isBase64DataURL.FindStringSubmatch(reflect.Indirect(postvalue).Interface().(string))
                  if len(enc) == 2 {
                     Goose.OpHandle.Logf(6,"Data URL detected! %s",enc[1])
                     decoded, err = base64.StdEncoding.DecodeString(enc[1])
                     if err == nil {
                        postvalue = reflect.ValueOf(string(decoded))
                        Goose.OpHandle.Logf(8,"Data URL decoded! %s",postvalue.Interface())
                     }
                  }
               }

               // Adds the post variable to the method parameter array
               Goose.OpHandle.Logf(8,"postvalue: %#v",postvalue)
               ins = append(ins,reflect.Indirect(postvalue))
               Goose.OpHandle.Logf(8,"ins: %d:%s",len(ins),ins)
//               Goose.OpHandle.Logf(5,"ins2: %c-%c-%c-%c",(*postvalue.Interface().(*string))[0],(*postvalue.Interface().(*string))[1],(*postvalue.Interface().(*string))[2],(*postvalue.Interface().(*string))[3])
               j++
            }
         }
      }

      num = met.Type.NumIn()

      // If the application required, we must provide the authenticated user information to the method.
      // This is done by adding it as the last parameter
      Goose.OpHandle.Logf(8,"ins3: %d:%s",len(ins),ins)
      if accesstype == AccessAuthInfo || accesstype == AccessVerifyAuthInfo{
         Goose.OpHandle.Logf(6,"Checking the need for appending authinfo")
         if (len(ins)+1) == num || (len(ins)+3) == num {
            Goose.OpHandle.Logf(6,"Appending authinfo: %s",reflect.ValueOf(authinfo).Elem())
            ins = append(ins,reflect.ValueOf(authinfo))
         }
      }

      // Checks if the calling parameter count matches the method parameter count
      if len(ins) != num {
         var sdbg string

         if (len(ins) == (num-2)) && (num>=2) && (met.Type.In(num-2).Kind()==reflect.String) && (met.Type.In(num-1).Kind()==reflect.String) {
            Goose.OpHandle.Logf(5, "Appending request info: host:%s, RemoteAddr: %s", host, remoteAddr)
            ins = append(ins,reflect.ValueOf(host),reflect.ValueOf(remoteAddr))

            for _, rval = range ins[1:] {
               sIns += fmt.Sprintf("%#v, ",rval.Interface())
            }
            if len(sIns) > 0 {
               sIns = sIns[:len(sIns)-2]
            }
            Goose.OpHandle.Logf(8,"Operation has these parameters: this, %s",sIns)
         } else {
            for _, in := range ins[1:] {
               sdbg += fmt.Sprintf("[%s], ",in.Interface())
            }
            errmsg = fmt.Sprintf("Operation call with wrong input argument count: expected:%d, received:%d -> %s", num, len(ins), sdbg)
            Goose.OpHandle.Logf(1,errmsg)
            return Response {
               Status:            http.StatusBadRequest,
               Body:              errmsg,
            }
         }
      }

      // Finally calls the method
      retData := met.Func.Call(ins)

      Goose.OpHandle.Logf(8,"retData: %#v",retData)
      return retData[0].Interface().(Response)
   }
}

