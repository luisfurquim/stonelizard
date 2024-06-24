package stonelizard

import (
   "io"
   "fmt"
   "sync"
   "bytes"
   "runtime"
   "strings"
   "reflect"
   "net/http"
   "math/rand"
   "crypto/x509"
   "encoding/xml"
   "encoding/json"
   "compress/gzip"
   "encoding/base64"
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
   var extAuth                ExtAuthT
   var origHost               string
   var parts                []string
   var optional					bool

   Goose.Serve.Logf(1,"Access %s+%s %s from %s", r.Proto, r.Method, r.URL.Path, r.RemoteAddr)

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
      Goose.Serve.Logf(6,"Header %s:%#v",_hd, val)
   }

   proto = strings.Split(r.Proto,"/")[0] + cryp
   Goose.Serve.Logf(6,"Going to check if it is a websocket connection")
   for _, upg := range r.Header["Upgrade"] {
      Goose.Serve.Logf(1,"Header Upgrade:%s",upg)
      if upg == "websocket" {
         proto = "WS" + cryp
         useWebSocket = true
      }
   }

   Goose.Serve.Logf(5,"Will check if swagger.json is requested: %#v", svc.Swagger)
   Goose.Serve.Logf(4,"svc.SwaggerPath: %s", svc.SwaggerPath)
   if r.URL.Path==(svc.SwaggerPath+"/swagger.json") {
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
//      Goose.Serve.Logf(0,"Received request of swagger.json: %#v",svc.Swagger.Paths["/pleito"]["get"].Responses["200"].Schema.Items)

      if r.Header.Get("X-Forwarded-Host") != "" {
         origHost = svc.Swagger.Host
         svc.Swagger.Host = r.Header.Get("X-Forwarded-Host")
      }

      buf, err = json.Marshal(svc.Swagger)

      if origHost != "" {
         svc.Swagger.Host = origHost
      }

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

   Goose.Serve.Logf(6,"Original parms: %#v",parms)

   if r.Method == "OPTIONS" {
      Goose.Serve.Logf(4,"CORS Options called on " + r.URL.Path)
      hd.Add("Access-Control-Allow-Methods","POST, GET, OPTIONS, PUT, DELETE")
//Access-Control-Allow-Origin: http://foo.example
//Access-Control-Allow-Methods: POST, GET, OPTIONS
//Access-Control-Allow-Headers: X-PINGOTHER
//Access-Control-Allow-Origin: *
      hd.Add("Access-Control-Allow-Headers", strings.Join(append(endpoint.Headers,"Content-Type","X-Request-Signer","X-Request-Signature"),", "))
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("OK"))
      return
   }

   r.ParseForm()
   for _, qry = range endpoint.Query {
		if qry[0] == '?' {
			optional = true
			qry = qry[1:]
		} else {
			optional = false
		}
      if _, ok := r.Form[qry]; !ok {
			if optional {
				parms = append(parms,"") // TODO array support
				j++
			} else {
				errmsg := fmt.Sprintf("%s: %s",ErrorMissingRequiredQueryField,qry)
				Goose.Serve.Logf(1,errmsg)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(errmsg))
				return
			}
      } else {
			parms = append(parms,r.Form[qry][0]) // TODO array support
			j++
		}
   }
   for qry, _ = range r.Form {
      authparms[qry] = r.Form[qry][0]
   }

   Goose.Serve.Logf(6,"Parms with query: %#v",parms)
   Goose.Serve.Logf(6,"authparms with query: %#v",authparms)
   Goose.Serve.Logf(6,"r.Form with query: %#v",r.Form)

   for _, header = range endpoint.Headers {
		if header[0] == '?' {
			optional = true
			header = header[1:]
		} else {
			optional = false
		}

		header = strings.ToUpper(header)

		for hdrKey, _ := range r.Header {
			if strings.ToUpper(hdrKey) == header {
				header = hdrKey
				break
			}
		}
		
      if _, ok = r.Header[header]; !ok {
//      if (r.Header[header]==nil) || (len(r.Header[header])==0) {
			if optional {
				parms = append(parms,"") // TODO array support
				j++
			} else {
				errmsg := fmt.Sprintf("%s: %s",ErrorMissingRequiredHTTPHeader,header)
				Goose.Serve.Logf(1,errmsg)
				Goose.Serve.Logf(6,"HTTP Headers found: %#v",r.Header)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(errmsg))
				return
			}
      } else {
			parms = append(parms,r.Header[header][0]) // TODO array support
			j++
		}
   }
   for header, _ = range r.Header {
      authparms[header] = r.Header[header][0]
   }

   Goose.Serve.Logf(6,"Parms with headers: %#v",parms)
   Goose.Serve.Logf(6,"authparms with headers: %#v",authparms)

   Goose.Serve.Logf(5,"checking marshalers: cons:%s, prod:%s",endpoint.consumes,endpoint.produces)

   dbg := &bytes.Buffer{}
   io.Copy(dbg, r.Body)
//   Goose.Serve.Logf(0,"Body: %s", dbg.Bytes())

   r.Body = io.NopCloser(dbg)

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
   } else if len(endpoint.consumes)>=24  && endpoint.consumes[:24] == "application/octet-stream" {
      if len(endpoint.consumes)>=31  && endpoint.consumes[24:31] == ";base64" {
         umrsh = NewBase64Unmarshaler(r)
      } else {
         umrsh = NewDummyUnmarshaler(r)
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
         Goose.Serve.Logf(6,"svc.AllowGzip == true")
gzipcheck:
         for _, enc = range encRequest {
            for _, e = range strings.Split(enc,", ") {
               Goose.Serve.Logf(6,"Encoding: %s",e)
               if e == "gzip" {
                  Goose.Serve.Logf(6,"Using gzip")
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

   Goose.Serve.Logf(5,"svc.Access level: %d",svc.Access)
   err = nil
   if svc.Access != AccessNone && svc.Access != AccessInfo {

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
            const size = 64 << 14
            var buf []byte
            var srcs, srcs2 []string
            var src string
            var i int

            buf  = make([]byte, size)
            i    = runtime.Stack(buf, false)
            buf  = buf[:i]
            srcs = strings.Split(string(buf),"\n")
            for _, src = range srcs {
               if gosrcRE.MatchString(src) && (!gorootRE.MatchString(src)) {
						if len(gosrcFNameRE.FindStringSubmatch(src)) > 1 {
							srcs2 = append(srcs2,gosrcFNameRE.FindStringSubmatch(src)[1])
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

            Goose.Serve.Logf(1,"panic (%s): calling %s with %s @ %s", panicerr, endpoint.Path, authparms, src)
         }
      }()

      if extAuth, ok = svc.Authorizer.(ExtAuthT); ok {
         httpstat, authinfo, err = extAuth.ExtAuthorize(svc.ch, endpoint.Path, authparms, w, r, svc.SavePending)
      } else {
         httpstat, authinfo, err = svc.Authorizer.Authorize(endpoint.Path, authparms, r.RemoteAddr, r.TLS, svc.SavePending)
      }
      //Goose.Serve.Fatalf(0,"%d %#v %s", httpstat, authinfo, err)
   }

   if err != nil {
      Goose.Serve.Logf(1,"Authorization failure with HTTP Status %d and error %s", httpstat, err)
      w.WriteHeader(httpstat)
      return
   }

//   Goose.Serve.Logf(0,"check authinfo")
   Goose.Serve.Logf(5,"Authorization returned HTTP status %d",httpstat)
   if svc.Access != AccessAuthInfo && svc.Access != AccessVerifyAuthInfo && svc.Access != AccessInfo{
      authinfo = nil
   } else if svc.Access == AccessInfo {
      certBuf := r.Header.Get("X-Request-Signer-Certificate")
      if len(certBuf) > 0 {
         cert, err := base64.StdEncoding.DecodeString(certBuf)
         if err == nil {
            authinfo, err = x509.ParseCertificate(cert)
            if err != nil {
               authinfo = nil
            }
         }
      }
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
         var startwg sync.WaitGroup
         var sessionId = rand.Intn(100000)
         var evHandlers map[string]*WSEventTrigger
         var name string
         var ev *WSEventTrigger
         var i int
         var evName interface{}
         var sIns string
         var rval reflect.Value
         var met reflect.Method
         var ok bool

         Goose.OpHandle.Logf(2,"Websocket start") //5

         if endpoint.consumes == "application/json" {
            codec = websocket.JSON
//            } else if endpoint.consumes == "application/xml" {
//               codec = websocket.Message
         }

         obj = reflect.ValueOf(resp.Body)

         evHandlers = map[string]*WSEventTrigger{}

         Goose.OpHandle.Logf(2,"Websocket configuring events") //5
         startwg.Add(1)
         wsEventHandle(ws, codec, resp.Body, &wg, evHandlers, &startwg)
         startwg.Wait()
         Goose.OpHandle.Logf(3,"Websocket events configured") //5

         wsMainLoop:
         for {
            Goose.OpHandle.Logf(5,"Websocket loop start") //5
            err = codec.Receive(ws, &request)
            Goose.OpHandle.Logf(6,"Websocket Received: %s",request) //5
            if err != nil {
               if err == io.EOF {
                  Goose.Serve.Logf(1,"Ending websocket session %d for %s",sessionId, endpoint.Path)
               } else {
                  Goose.Serve.Logf(1,"Websocket protocol error: %s",err)
               }
               if err = ws.Close(); err != nil {
                  Goose.Serve.Logf(1,"Error closing websocket connection: %s",err)
               }

               met, ok = obj.Type().MethodByName("OnClose")
               if ok {
                  met.Func.Call([]reflect.Value{obj})
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
               Goose.Serve.Logf(6,"[%d] Websocket bind request %#v", trackid, request)//3
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
                  Goose.Serve.Logf(3,"[%d] Websocket bound event %s", trackid, evName.(string))//3
                  codec.Send(ws, []interface{}{trackid, http.StatusOK})
                  met, ok = obj.Type().MethodByName("OnBind")
                  if ok {
                     met.Func.Call([]reflect.Value{obj, reflect.ValueOf(evName.(string))})
                  }
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
                  met, ok = obj.Type().MethodByName("OnUnbind")
                  if ok {
                     met.Func.Call([]reflect.Value{obj, reflect.ValueOf(evName.(string))})
                  }
               }
            } else {
               opi, err = endpoint.WSocketOperations.Get(opName)
               if err != nil {
                  Goose.Serve.Logf(1,"Operation lookup failure %s",err)
                  break
               }
               op = opi.(*WSocketOperation)

               request, ok = request[2].([]interface{})

               if !ok {
                  Goose.Serve.Logf(1,"[%d] Websocket protocol error: 3rd parameter must be array", trackid)
                  break
               }

               if op.CallByRef {
                  ins, err = pushParms(request, obj, op.Method)
               } else {
                  ins, err = pushParms(request, reflect.Indirect(obj), op.Method)
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
               Goose.OpHandle.Logf(5,"Calling websocket operation with: %s",sIns) //5
               retData := op.Method.Func.Call(ins)
               sIns = ""
               for _, rval = range retData {
                  sIns += fmt.Sprintf("%#v, ",rval.Interface())
               }
               if len(sIns) > 0 {
                  sIns = sIns[:len(sIns)-2]
               }
               Goose.OpHandle.Logf(5,"Websocket operation retData: %s",retData) //5

               WSResponse = retData[0].Interface().(Response)

               if WSResponse.Body == nil {
                  // callID , status (no content)
                  Goose.Serve.Logf(4,"[%d] Websocket send no content", trackid)
                  codec.Send(ws, []interface{}{trackid, WSResponse.Status})
               } else {
                  // callID , status, response
                  Goose.Serve.Logf(4,"[%d] Websocket send %#v", trackid, WSResponse.Body)
                  codec.Send(ws, []interface{}{trackid, WSResponse.Status, WSResponse.Body})
               }
            }

            Goose.Serve.Logf(4,"[%d] Websocket message sent", trackid)
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

   if endpoint.produces == "application/json" {
      Goose.Serve.Logf(4,"Using json encoder")
      mrsh = json.NewEncoder(outWriter)
      hd.Add("Content-Type","application/json")
   } else if endpoint.produces == "application/xml" {
      mrsh = xml.NewEncoder(outWriter)
      hd.Add("Content-Type","application/xml")
   } else if (endpoint.produces == "application/javascript") || (len(endpoint.produces)>=5 && endpoint.produces[:5] == "text/") {
      mrsh = NewStaticEncoder(outWriter)
      hd.Add("Content-Type",endpoint.produces + "; charset=utf-8")
	} else if len(endpoint.produces)>=12 && endpoint.produces[:12] == "application/" {
      mrsh = NewStaticEncoder(outWriter)
      hd.Add("Content-Type", endpoint.produces)
   } else {
      parts = strings.Split(endpoint.produces,";")
      if parts[0] == "*" {
         Goose.Serve.Logf(1,"Using static encoder")
         if len(parts)>1 && parts[1]=="base64" {
            mrsh = NewStaticBase64Encoder(outWriter)
         } else {
            mrsh = NewStaticEncoder(outWriter)
         }
      } else {
         errmsg := fmt.Sprintf("Internal server error determining response mimetype")
         Goose.Serve.Logf(1,errmsg)
         w.WriteHeader(http.StatusInternalServerError)
         w.Write([]byte(errmsg))
         return
      }
   }

   if resp.Status != 0 {
      for k, v := range resp.Header {
         hd.Add(k, v)
      }
      w.WriteHeader(resp.Status)
      if resp.Status != http.StatusNoContent {
         err = mrsh.Encode(resp.Body)
         if err!=nil {
            Goose.Serve.Logf(1,"Internal server error writing response body (no status sent to client): %s",err)
            return
         }
      }
   }

}


