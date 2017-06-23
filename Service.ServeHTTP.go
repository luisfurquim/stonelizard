package stonelizard

import (
   "io"
   "fmt"
   "sync"
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
   var match                []string
   var parms                []interface{}
   var authparms              map[string]interface{}
   var i, j                   int
   var endpoint               UrlNode
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

   Goose.Serve.Logf(1,"Access %s from %s", r.URL.Path, r.RemoteAddr)

   if r.URL.Path=="/crtlogin" {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte(" "))
      return
   }

   hd := w.Header()
   hd.Add("Access-Control-Allow-Origin","*")
   hd.Add("Vary","Origin")

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

   match = svc.Matcher.FindStringSubmatch(r.Method+":"+r.URL.Path)
   Goose.Serve.Logf(6,"Matcher found this %#v\n", match)
   if len(match) == 0 {
      Goose.Serve.Logf(1,"Invalid service handler " + r.URL.Path)
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Invalid service handler " + r.URL.Path))
      return
   }

   parms = []interface{}{}
   authparms = map[string]interface{}{}

//   for _, endpoint = range svc.Svc {
   for i=1; i<len(match); i++ {
      Goose.Serve.Logf(5,"trying %s:%s with endpoint: %s",r.Method,r.URL.Path,svc.Svc[svc.MatchedOps[i-1]].Path)
      if len(match[i]) > 0 {
         Goose.Serve.Logf(5,"Found endpoint %s for: %s",svc.Svc[svc.MatchedOps[i-1]].Path,r.URL.Path)
         endpoint = svc.Svc[svc.MatchedOps[i-1]]
         for j=i+1; (j<len(match)) && (len(match[j])>0); j++ {
            authparms[endpoint.ParmNames[j-i-1]] = match[j]
         }
         for k := i+1; k<j; k++ { // parms = []interface{}(match[i+1:j])
            parms = append(parms,match[k])
         }
         j -= i + 1
         break
      }
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
            Goose.Serve.Logf(1,"%s: %s",ErrorMissingRequiredQueryField,qry)
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
         Goose.Serve.Logf(1,"%s: %s",ErrorMissingRequiredHTTPHeader,header)
         Goose.Serve.Logf(6,"HTTP Headers found: %#v",r.Header)
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
         Goose.Serve.Logf(1,"Error initializing multipart/formdata unmarshaller for %s: %s", r.URL.Path, err)
         return
      }
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
   }

   Goose.Serve.Logf(5,"svc.Access: %d",svc.Access)
   err = nil
   if svc.Access != AccessNone {
      httpstat, authinfo, err = svc.Authorizer.Authorize(endpoint.Path, authparms, r.RemoteAddr, r.TLS, svc.SavePending)
   }

   if err == nil {
      Goose.Serve.Logf(5,"Authorization returned HTTP status %d",httpstat)
      if svc.Access != AccessAuthInfo && svc.Access != AccessVerifyAuthInfo {
         authinfo = nil
      }


      for _, p := range endpoint.Proto {
         if p=="ws" || p=="wss" {
            useWebSocket = true
            break
         }
      }

      resp = endpoint.Handle(parms,umrsh,authinfo)
      if useWebSocket && resp.Status < http.StatusMultipleChoices && resp.Status >= http.StatusOK {
         websocket.Handler(func (ws *websocket.Conn) {
            //endpoint.Handle(parms,umrsh,authinfo)
            var op *WSocketOperation
            var err error
            var codec websocket.Codec
            var request []interface{}
            var trackid string
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

            if endpoint.consumes == "application/json" {
               codec = websocket.JSON
//            } else if endpoint.consumes == "application/xml" {
//               codec = websocket.Message
            }

            obj = reflect.ValueOf(resp.Body)

            evHandlers = map[string]*WSEventTrigger{}

            wsEventHandle(ws, codec, resp.Body, wg, evHandlers)

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
                  case string:
                  trackid = request[0].(string)
                  default:
                     Goose.Serve.Logf(1,"Websocket protocol error: %s @0",WrongParameterType)
                     break
               }

               switch request[1].(type) {
                  case string:
                     opName = request[1].(string)
                  default:
                     Goose.Serve.Logf(1,"[%s] Websocket protocol error: %s @1", trackid, WrongParameterType)
                     break
               }


               if opName == "bind" { // reserved word
                  for i, evName = range request[2:] {
                     switch evName.(type) {
                        case string:
                           if eh, ok := evHandlers[evName.(string)]; ok {
                              eh.Status = true
                           } else {
                              Goose.Serve.Logf(1,"Websocket bind event not found: %s", trackid, evName.(string))
                              codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                           }
                        default:
                           Goose.Serve.Logf(1,"[%s] Websocket bind event protocol error: @%d", trackid, WrongParameterType, i)
                           codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                     }
                  }
               } else if opName == "unbind" { // reserved word
                  for i, evName = range request[2:] {
                     switch evName.(type) {
                        case string:
                           if eh, ok := evHandlers[evName.(string)]; ok {
                              eh.Status = false
                           } else {
                              Goose.Serve.Logf(1,"Websocket unbind event not found: %s", trackid, evName.(string))
                              codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                           }
                        default:
                           Goose.Serve.Logf(1,"[%s] Websocket unbind event protocol error: @%d", trackid, WrongParameterType, i)
                           codec.Send(ws, []interface{}{trackid, WrongParameterType, i})
                     }
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
                     Goose.Serve.Logf(1,"[%s] Websocket protocol error: %s @1", trackid, err)
                     break
                  }

                  retData := op.Method.Func.Call(ins)
                  Goose.OpHandle.Logf(5,"retData: %#v",retData)

                  WSResponse = retData[0].Interface().(Response)

                  // callID , status, response
                  Goose.Serve.Logf(1,"Websocket [%s] send %#v", trackid, WSResponse.Body)
                  codec.Send(ws, []interface{}{trackid, WSResponse.Status, WSResponse.Body})
               }

               Goose.Serve.Logf(1,"Websocket [%s] message sent", trackid)
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
            Goose.Serve.Logf(1,"Internal server error writing response body (no status sent to client): %s",err)
            return
         }
      }
   }
}


