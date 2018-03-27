package stonelizard

import (
   "io"
   "fmt"
   "sync"
   "runtime"
   "strings"
   "reflect"
   "net/http"
   "math/rand"
   "encoding/xml"
   "encoding/json"
   "compress/gzip"
   "golang.org/x/net/websocket"
)

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   var parms                []interface{}
   var authparms              map[string]interface{}
   var j                      int
   var endpoint              *UrlNode
   var resp                   Response
   var ok                     bool
   var err                    error
   var mrsh                   Marshaler
   var umrsh                  Unmarshaler
   var outWriter              io.Writer
   var encRequest           []string
   var enc, e                 string
   var gzw                   *gzip.Writer
   var header                 string
   var qry                    string
   var buf                  []byte
   var httpstat               int
   var authinfo               interface{}
   var useWebSocket           bool
   var cryp                   string
   var httpStatus             int
   var proto                  string

   Goose.Serve.Logf(1,"Access %s+%s %s from %s", r.Proto, r.Method, r.URL.Path, r.RemoteAddr)
   Goose.Serve.Logf(1,"Goose/stonelizard/Serve %d", uint8(Goose.Serve))

   if r.URL.Path=="/crtlogin" {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte(" "))
      return
   }

   if r.TLS != nil {
      cryp = "S"
   }

   hd := w.Header()
   hd.Add("Access-Control-Allow-Origin","*")
   hd.Add("Vary","Origin")


   for _hd, val := range r.Header {
      Goose.Serve.Logf(1,"Header %s:%#v",_hd, val)
   }

   proto = strings.Split(r.Proto,"/")[0] + cryp
   Goose.Serve.Logf(1,"Going to check if it is a websocket connection")
   for _, upg := range r.Header["Upgrade"] {
      Goose.Serve.Logf(1,"Header Upgrade:%s",upg)
      if upg == "websocket" {
         proto = "WS" + cryp
         useWebSocket = true
      }
   }

   Goose.Serve.Logf(6,"Will check if swagger.json is requested: %#v",svc.Swagger)
   if r.URL.Path=="/swagger.json" {
      defer func() {
         if r := recover(); r != nil {
            Goose.Serve.Logf(1,"Internal server error writing response body for swagger.json: %#v",r)
         }
      }()
      hd.Add("Content-Type","application/json; charset=utf-8")
//      w.WriteHeader(http.StatusOK)
      Goose.Serve.Logf(6,"Received request of swagger.json: %#v",svc.Swagger)
//      mrsh = json.NewEncoder(w)
//      err = mrsh.Encode(svc.Swagger)
      buf, err = json.Marshal(svc.Swagger)
      if err!=nil {
         Goose.Serve.Logf(1,"Internal server error marshaling swagger.json: %s",err)
      }
      hd.Add("Content-Length", fmt.Sprintf("%d",len(buf)))
      _, err = io.WriteString(w,string(buf))
      if err!=nil {
         Goose.Serve.Logf(1,"Internal server error writing response body for swagger.json: %s",err)
      }
      return
   }

   endpoint, parms, authparms, httpStatus = svc.FetchEndpointHandler(proto, r.Method, r.URL.Path)
   if httpStatus > 0 {
      w.WriteHeader(httpStatus)
      w.Write([]byte(fmt.Sprintf("Invalid service handler: %s", r.URL.Path)))
      return
   }

   Goose.Serve.Logf(5,"Original parms: %#v",parms)

   if r.Method == "OPTIONS" {
      Goose.Serve.Logf(4,"CORS Options called on " + r.URL.Path)
      hd.Add("Access-Control-Allow-Methods","POST, GET, OPTIONS, PUT, DELETE")
//Access-Control-Allow-Origin: http://foo.example
//Access-Control-Allow-Methods: POST, GET, OPTIONS
//Access-Control-Allow-Headers: X-PINGOTHER
//Access-Control-Allow-Origin: *
      hd.Add("Access-Control-Allow-Headers", strings.Join(endpoint.Headers,", "))
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("OK"))
      return
   }

   if len(endpoint.Query) > 0 {
      r.ParseForm()
      for _, qry = range endpoint.Query {
         if _, ok := r.Form[qry]; !ok {
            errmsg := fmt.Sprintf("%s: %s",ErrorMissingRequiredQueryField,qry)
            Goose.Serve.Logf(1,errmsg)
            w.WriteHeader(http.StatusBadRequest)
            w.Write([]byte(errmsg))
            return
         }
         parms = append(parms,r.Form[qry][0]) // TODO array support
         authparms[endpoint.ParmNames[j]] = r.Form[qry][0]
         j++
      }
   }

   Goose.Serve.Logf(5,"Parms with query: %#v",parms)

   for _, header = range endpoint.Headers {
      if (r.Header[header]==nil) || (len(r.Header[header])==0) {
         errmsg := fmt.Sprintf("%s: %s",ErrorMissingRequiredHTTPHeader,header)
         Goose.Serve.Logf(1,errmsg)
         Goose.Serve.Logf(6,"HTTP Headers found: %#v",r.Header)
         w.WriteHeader(http.StatusBadRequest)
         w.Write([]byte(errmsg))
         return
      }
      parms = append(parms,r.Header[header][0]) // TODO array support
      authparms[endpoint.ParmNames[j]] = r.Header[header][0]
      j++
   }

   Goose.Serve.Logf(5,"Parms with headers: %#v",parms)

   Goose.Serve.Logf(5,"checking marshalers: %s, %s",endpoint.consumes,endpoint.produces)

   if endpoint.consumes == "application/json" {
      umrsh = json.NewDecoder(r.Body)
   } else if endpoint.consumes == "application/xml" {
      umrsh = xml.NewDecoder(r.Body)
   } else if endpoint.consumes == "multipart/form-data" {
//      bdy, err = ioutil.ReadAll(r.Body)
//      ioutil.WriteFile("upload.bin", bdy, 0600)
//      Goose.Serve.Logf(6,"body=%s",bdy)
      umrsh, err = NewMultipartUnmarshaler(r,endpoint.Body)
      if err != nil {
         errmsg := fmt.Sprintf("Error initializing multipart/formdata unmarshaller for %s: %s", r.URL.Path, err)
         Goose.Serve.Logf(1,errmsg)
         w.WriteHeader(http.StatusBadRequest)
         w.Write([]byte(errmsg))
         return
      }
   } else {
      errmsg := fmt.Sprintf("Unsupported input mimetype")
      Goose.Serve.Logf(1,errmsg)
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte(errmsg))
      return
   }

   Goose.Serve.Logf(6,"umrsh=%#v",umrsh)

   outWriter = w

   if encRequest, ok = r.Header["Accept-Encoding"] ; ok {
      Goose.Serve.Logf(6,"Accept-Encoding: %#v",encRequest)
      if svc.AllowGzip == true {
         Goose.Serve.Logf(5,"svc.AllowGzip == true")
gzipcheck:
         for _, enc = range encRequest {
            for _, e = range strings.Split(enc,", ") {
               Goose.Serve.Logf(5,"Encoding: %s",e)
               if e == "gzip" {
                  Goose.Serve.Logf(5,"Using gzip")
                  gzw = gzip.NewWriter(w)
                  outWriter = gzHttpResponseWriter{
                     Writer: gzw,
                     ResponseWriter: w,
                  }
                  defer gzw.Close()
                  hd.Add("Vary", "Accept-Encoding")
                  hd.Add("Content-Encoding","gzip")
                  break gzipcheck
               }
            }
         }
      }
   }


   if endpoint.produces == "application/json" {
      mrsh = json.NewEncoder(outWriter)
      hd.Add("Content-Type","application/json")
   } else if endpoint.produces == "application/xml" {
      mrsh = xml.NewEncoder(outWriter)
      hd.Add("Content-Type","application/xml")
   } else if (endpoint.produces == "text/css") || (endpoint.produces == "text/html") || (endpoint.produces == "text/javascript") || (endpoint.produces == "application/javascript") {
      mrsh = NewStaticEncoder(outWriter)
      hd.Add("Content-Type",endpoint.produces + "; charset=utf-8")
   } else {
      errmsg := fmt.Sprintf("Internal server error determining response mimetype")
      Goose.Serve.Logf(1,errmsg)
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(errmsg))
      return
   }

   Goose.Serve.Logf(5,"svc.Access: %d",svc.Access)
   err = nil
   if svc.Access != AccessNone {

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

            Goose.Serve.Logf(0,"panic (%s): calling %s -> %s with %s @ %s", panicerr, endpoint.Path, authparms, src)
         }
      }()

      httpstat, authinfo, err = svc.Authorizer.Authorize(endpoint.Path, authparms, r.RemoteAddr, r.TLS, svc.SavePending)
   }

   if err == nil {
      Goose.Serve.Logf(5,"Authorization returned HTTP status %d",httpstat)
      if svc.Access != AccessAuthInfo && svc.Access != AccessVerifyAuthInfo {
         authinfo = nil
      }

      resp = endpoint.Handle(parms,umrsh,authinfo,r.Host,r.RemoteAddr)
      if useWebSocket && resp.Status < http.StatusMultipleChoices && resp.Status >= http.StatusOK {
         websocket.Handler(func (ws *websocket.Conn) {
            //endpoint.Handle(parms,umrsh,authinfo)
            var op *WSocketOperation
            var err error
            var codec websocket.Codec
            var request []interface{}
            var trackid int64
            var opName string
            var opi interface{}
            var obj reflect.Value
            var ins []reflect.Value
            var WSResponse Response
            var wg sync.WaitGroup
            var sessionId = rand.Intn(100000)
            var evHandlers map[string]*WSEventTrigger
            var name string
            var ev *WSEventTrigger
            var i int
            var evName interface{}
            var sIns string
            var rval reflect.Value

            if endpoint.consumes == "application/json" {
               codec = websocket.JSON
//            } else if endpoint.consumes == "application/xml" {
//               codec = websocket.Message
            }

            obj = reflect.ValueOf(resp.Body)

            evHandlers = map[string]*WSEventTrigger{}

            wsEventHandle(ws, codec, resp.Body, wg, evHandlers)

            wsMainLoop:
            for {
               err = codec.Receive(ws, &request)
               if err != nil {
                  if err == io.EOF {
                     Goose.Serve.Logf(1,"Ending websocket session %d for %s",sessionId, endpoint.Path)
                  } else {
                     Goose.Serve.Logf(1,"Websocket protocol error: %s",err)
                  }
                  if err = ws.Close(); err != nil {
                     Goose.Serve.Logf(1,"Error closing websocket connection: %s",err)
                  }
                  break
               }

               // callID , procedureName, requestData
               if len(request) < 3 {
                  Goose.Serve.Logf(1,"Websocket protocol error: %s",WrongParameterLength)
                  break
               }

               switch request[0].(type) {
                  case float64:
                  trackid = int64(request[0].(float64))
                  default:
                     Goose.Serve.Logf(1,"Websocket protocol error %s @0: %T",WrongParameterType,request[0])
                     break wsMainLoop
               }

               switch request[1].(type) {
                  case string:
                     opName = request[1].(string)
                  default:
                     Goose.Serve.Logf(1,"[%d] Websocket protocol error: %d @1", trackid, WrongParameterType)
                     break wsMainLoop
               }


               if opName == "bind" { // reserved word
                  parmOk := true
                  bindFor:
                  for i, evName = range request[2:] {
                     switch evName.(type) {
                        case string:
                           if eh, ok := evHandlers[evName.(string)]; ok {
                              eh.Status = true
                           } else {
                              parmOk = false
                              Goose.Serve.Logf(1,"[%d] Websocket bind event %s not found, id: %d", trackid, evName.(string))
                              codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                              break bindFor
                           }
                        default:
                           parmOk = false
                           Goose.Serve.Logf(1,"[%d] Websocket bind event protocol error %s: @%d", trackid, WrongParameterType, i)
                           codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                           break bindFor
                     }
                  }
                  if parmOk {
                     Goose.Serve.Logf(3,"[%d] Websocket bound event %s", trackid, evName.(string))
                     codec.Send(ws, []interface{}{trackid, http.StatusOK})
                  }
               } else if opName == "unbind" { // reserved word
                  parmOk := true
                  unbindFor:
                  for i, evName = range request[2:] {
                     switch evName.(type) {
                        case string:
                           if eh, ok := evHandlers[evName.(string)]; ok {
                              eh.Status = false
                           } else {
                              parmOk = false
                              Goose.Serve.Logf(1,"[%d] Websocket unbind event %s not found, id: %d", trackid, evName.(string))
                              codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                              break unbindFor
                           }
                        default:
                           parmOk = false
                           Goose.Serve.Logf(1,"[%d] Websocket unbind event protocol error %s: @%d", trackid, WrongParameterType, i)
                           codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                           break unbindFor
                     }
                  }
                  if parmOk {
                     Goose.Serve.Logf(3,"[%d] Websocket unbound event %s", evName.(string), trackid)
                     codec.Send(ws, []interface{}{trackid, http.StatusOK})
                  }
               } else {
                  opi, err = endpoint.WSocketOperations.Get(opName)
                  if err != nil {
                     Goose.Serve.Logf(1,"Operation lookup failure %s",err)
                     break
                  }
                  op = opi.(*WSocketOperation)

                  if op.CallByRef {
                     ins, err = pushParms(request[2:], obj, op.Method)
                  } else {
                     ins, err = pushParms(request[2:], reflect.Indirect(obj), op.Method)
                  }

                  if err != nil {
                     Goose.Serve.Logf(1,"[%d] Websocket protocol error: %s @1", trackid, err)
                     break
                  }

                  for _, rval = range ins {
                     sIns += fmt.Sprintf("%#v, ",rval.Interface())
                  }
                  if len(sIns) > 0 {
                     sIns = sIns[:len(sIns)-2]
                  }
                  Goose.OpHandle.Logf(5,"Calling websocket operation with: %s",sIns)
                  retData := op.Method.Func.Call(ins)
                  sIns = ""
                  for _, rval = range retData {
                     sIns += fmt.Sprintf("%#v, ",rval.Interface())
                  }
                  if len(sIns) > 0 {
                     sIns = sIns[:len(sIns)-2]
                  }
                  Goose.OpHandle.Logf(5,"Websocket operation retData: %s",retData)

                  WSResponse = retData[0].Interface().(Response)

                  // callID , status, response
                  Goose.Serve.Logf(1,"[%d] Websocket send %#v", trackid, WSResponse.Body)
                  codec.Send(ws, []interface{}{trackid, WSResponse.Status, WSResponse.Body})
               }

               Goose.Serve.Logf(1,"[%d] Websocket message sent", trackid)
            }

            // stop all event triggers
            for name, ev = range evHandlers {
               Goose.Serve.Logf(1,"Event channel close for %s",name)
               close(ev.EventData)
            }

            wg.Wait()
         }).ServeHTTP(w,r)
         return
      }

      if resp.Status != 0 {
         for k, v := range resp.Header {
            hd.Add(k, v)
         }
         w.WriteHeader(resp.Status)
      }


      if resp.Status != http.StatusNoContent {
         err = mrsh.Encode(resp.Body)
         if err!=nil {
            Goose.Serve.Logf(1,"Internal server error writing response body (no status sent to client): %s",err)
            return
         }
      }
   } else {
      Goose.Serve.Logf(1,"Authorization failure with HTTP Status %d and error %s", httpstat, err)
      w.WriteHeader(httpstat)
      if httpstat != http.StatusNoContent {
         err = mrsh.Encode(fmt.Sprintf("%s",err))
         if err!=nil {
            Goose.Serve.Logf(1,"Internal server error writing response body (no body sent to client): %s",err)
            return
         }
      }
   }
}


