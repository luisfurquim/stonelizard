package stonelizard

import (
   "io"
   "fmt"
   "context"
   "runtime"
   "strings"
   "reflect"
   "net/http"
   "crypto/x509"
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
         if panicerr := recover(); panicerr != nil {
            const size = 64 << 10
            var buf []byte
            var srcs, srcs2 []string
            var src string
            var parms string
            var p reflect.Value
            var tmp string
            var i int
            var fNamePart []string

            buf  = make([]byte, size)
            buf  = buf[:runtime.Stack(buf, false)]
            srcs = strings.Split(string(buf),"\n")
            for _, src = range srcs {
               if gosrcRE.MatchString(src) && (!gorootRE.MatchString(src)) {
						fNamePart = gosrcFNameRE.FindStringSubmatch(src)
						if len(fNamePart) > 1 {
							srcs2 = append(srcs2,fNamePart[1])
						} else {
							srcs2 = append(srcs2,"___---NoName---___")
						}
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

            for _, p = range ins[2:] {
               tmp = fmt.Sprintf(" %#v; ", p.Interface())
               if len(tmp) > 40 {
                  tmp = tmp[:40] + "; "
               }
               parms += tmp
            }

            Goose.Serve.Logf(1,"panic (%s): calling %s -> %s with %s @ %s", panicerr, met.PkgPath, met.Name, parms, src)
         }
      }()

      // Build context carrying request environment (certificate, host, remoteAddr).
      // This context is always the first parameter of every service method in v1.
      ctx := context.Background()
      ctx = context.WithValue(ctx, contextKeyHost, host)
      ctx = context.WithValue(ctx, contextKeyRemoteAddr, remoteAddr)
      if authinfo != nil {
         if reflect.ValueOf(authinfo).IsValid() && !reflect.ValueOf(authinfo).IsNil() {
            if cert, ok := authinfo.(*x509.Certificate); ok {
               ctx = context.WithValue(ctx, contextKeyCertificate, cert)
            }
         }
      }

      // Prepare the calling parameters; context is injected as first param (after receiver)
      if isPtr {
         ins, err = pushParms(parms, this, met, ctx)
      } else {
         ins, err = pushParms(parms, this.Elem(), met, ctx)
      }

      // On error return HTTP error message
      if err != nil {
         httpResp.Status = http.StatusInternalServerError
         httpResp.Body   = "Internal server error"
         httpResp.Header = map[string]string{}
         return httpResp
      }

      if this.Type().Kind() == reflect.Struct {
         if _, ok := this.Type().FieldByName("Session"); ok {
            if this.Kind()==reflect.Ptr {
               Goose.OpHandle.Logf(1,"THIS SESSION: %#v",this.Elem().FieldByName("Session"))
            } else {
               Goose.OpHandle.Logf(1,"THIS SESSION: %#v",this.FieldByName("Session"))
            }
         }
      }

      Goose.OpHandle.Logf(5,"posttype: %#v",posttype)
      if posttype != nil { // Adds data sent through HTTP POST
         err = nil
         j   = 0
         for (err==nil) && (j<len(posttype)) {
            Goose.OpHandle.Logf(6,"posttype[%d]: %#v",j,posttype[j])
            postvalue = reflect.New(posttype[j])
            Goose.OpHandle.Logf(6,"postvalue.Interface(): %#v",postvalue.Interface())
			   err = Unmarshal.Decode(postvalue.Interface())
            if err != nil && err != io.EOF {
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

               Goose.OpHandle.Logf(6,"postvalue.Kind() = %d",reflect.Indirect(postvalue).Kind())
               Goose.OpHandle.Logf(6,"postvalue = %s",reflect.Indirect(postvalue).Interface())
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

               Goose.OpHandle.Logf(8,"postvalue: %#v",postvalue)
               ins = append(ins,reflect.Indirect(postvalue))
               Goose.OpHandle.Logf(8,"ins: %d:%s",len(ins),ins)
               j++
         }
      }

      num = met.Type.NumIn()

      // Verify parameter count matches exactly - no optional params in v1
      Goose.OpHandle.Logf(8,"ins3: %d:%s, num:%d", len(ins), ins, num)
      if len(ins) != num {
         var sdbg string
         for ii, in := range ins[2:] {
            sdbg += fmt.Sprintf("[%s], ",in.Interface())
            Goose.OpHandle.Logf(1,"%d: [%#v], ", ii, in.Interface())
         }
         errmsg = fmt.Sprintf("Operation call with wrong input argument count: expected:%d, received:%d -> %s", num, len(ins), sdbg)
         Goose.OpHandle.Logf(1,errmsg)
         return Response {
            Status:            http.StatusBadRequest,
            Body:              errmsg,
         }
      }

      for _, rval = range ins[2:] {
         sIns += fmt.Sprintf("%#v, ",rval.Interface())
      }
      if len(sIns) > 0 {
         sIns = sIns[:len(sIns)-2]
      }
      Goose.OpHandle.Logf(8,"Operation has these parameters: this, ctx, %s",sIns)

      // Finally calls the method
      retData := met.Func.Call(ins)

      Goose.OpHandle.Logf(8,"retData: %#v",retData)
      return retData[0].Interface().(Response)
   }
}
