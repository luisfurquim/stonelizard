package stonelizard

import (
   "io"
   "os"
   "fmt"
   "net"
   "sync"
   "time"
   "bytes"
   "errors"
   "regexp"
   "strconv"
   "strings"
   "reflect"
//   "net/url"
   "net/http"
   "io/ioutil"
   "crypto/tls"
   "crypto/x509"
   "encoding/xml"
   "encoding/json"
   "compress/gzip"
   "path/filepath"
)

// http://www.hydrogen18.com/blog/stop-listening-http-server-go.html

var ErrorNoRoot error = errors.New("No service root specified")
var ErrorServiceSyntax error = errors.New("Syntax error on service definition")

//Init the listener for the service.
func NewListener(l net.Listener) (*StoppableListener, error) {
   tcpL, ok := l.(*net.TCPListener)

   if !ok {
      return nil, errors.New("Cannot wrap listener")
   }

   retval := &StoppableListener{}
   retval.TCPListener = tcpL
   retval.stop = make(chan int)

   return retval, nil
}

//Implements a wrapper on the system accept
func (sl *StoppableListener) Accept() (net.Conn, error) {

   for {
      Goose.Logf(6,"Start accept loop")

      //Wait up to one second for a new connection
      sl.SetDeadline(time.Now().Add(5*time.Second))

      newConn, err := sl.TCPListener.Accept()

      //Check for the channel being closed
      select {
         case <-sl.stop:
            Goose.Logf(2,"Received stop request")
            return nil, ErrorStopped
         default:
            Goose.Logf(6,"channel still open")
            // If the channel is still open, continue as normal
      }

      if err != nil {
         Goose.Logf(6,"Err not nil: %s",err)
         netErr, ok := err.(net.Error)

         // If this is a timeout, then continue to wait for
         // new connections
         if ok && netErr.Timeout() && netErr.Temporary() {
            Goose.Logf(6,"continue")
            continue
         }
      }

      if err==nil {
         Goose.Logf(2,"done listening")
         return newConn, nil
      }

      Goose.Logf(2,"done listening, err=%s",err)
      return nil, err
   }
}

//Stop the service and releases the chanel
func (sl *StoppableListener) Stop() {
   close(sl.stop)
}


//Close the webservices and the listeners
func (svc Service) Close() {
   svc.Listener.Stop()
   svc.Listener.Close()
   svc.CRLListener.Stop()
   svc.CRLListener.Close()
}

func GetSwaggerType(parm reflect.Type) (*SwaggerParameterT, error) {
   var item, subItem *SwaggerParameterT
   var field          reflect.StructField
   var doc            string
   var description   *string
   var err            error
   var i              int
   var fieldType      string

   Goose.Logf(6,"Tipo do parametro: %d: %s",parm.Kind(),parm)

   if parm == voidType {
      return nil, nil
   }

   if parm.Kind() == reflect.Interface {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"*"}}, nil
   }

   if parm.Kind() == reflect.Bool {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"boolean"}}, nil
   }

   if (parm.Kind()>=reflect.Int) && (parm.Kind()<=reflect.Int32) {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer", Format: "int32"}}, nil
   }

   if parm.Kind()==reflect.Int64 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer", Format: "int64"}}, nil
   }

   if (parm.Kind()>=reflect.Uint) && (parm.Kind()<=reflect.Uint64) {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer"}}, nil
   }

   if parm.Kind()==reflect.Float32 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"number", Format: "float"}}, nil
   }

   if parm.Kind()==reflect.Float64 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"number", Format: "double"}}, nil
   }

   if parm.Kind()==reflect.String {
      Goose.Logf(6,"Got string")
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"string"}}, nil
   }

   if parm.Kind()==reflect.Ptr {
      return GetSwaggerType(parm.Elem())
   }

   if (parm.Kind()==reflect.Array) || (parm.Kind()==reflect.Slice) {
      item, err = GetSwaggerType(parm.Elem())
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         return nil, err
      }
      return &SwaggerParameterT{
         Type:"array",
         Items: &SwaggerItemT{
            Type:             item.Schema.Type,
            Format:           item.Schema.Format,
            Items:            item.Schema.Items,
         },
         Schema: &SwaggerSchemaT{
            Type: item.Schema.Type,
         },
         CollectionFormat: "csv",
      }, nil // TODO: allow more collection formats
   }

   if parm.Kind()==reflect.Map {
      item, err = GetSwaggerType(parm.Elem())
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         return nil, err
      }

      kname   := parm.Key().Name()
      ktype   := ""
      kformat := ""
      if kname == "string" {
         ktype = "string"
      } else {
         ktype   = "integer"
         if kname[len(kname)-2:] == "64" {
            kformat = "int64"
         } else {
            kformat = "int32"
         }
      }

      return &SwaggerParameterT{
         Type:"array",
         Items: &SwaggerItemT{
            Type:              item.Schema.Type,
            Format:            item.Schema.Format,
            Items:             item.Schema.Items,
         },
         XKeyType:             ktype,
         XKeyFormat:           kformat,
         CollectionFormat:     "csv",
         XCollectionFormat:    "cskv",
      }, nil // TODO: allow more collection formats
   }

   if parm.Kind()==reflect.Struct {
      item = &SwaggerParameterT{
         Type:"object",
         Schema: &SwaggerSchemaT{
            Required: []string{},
            Properties: map[string]SwaggerSchemaT{},
//            Description: description,
         },
      }
      Goose.Logf(6,"Got struct: %#v",item)
      for i=0; i<parm.NumField(); i++ {
         field = parm.Field(i)
         Goose.Logf(6,"Struct field: %s",field.Name)
         doc   = field.Tag.Get("doc")
         if doc != "" {
            description    = new(string)
            (*description) = doc
         } else {
            description = nil
         }

         subItem, err = GetSwaggerType(field.Type)
         if (subItem==nil) || (err != nil) || (subItem.Schema==nil) {
            return nil, err
         }
         item.Schema.Required = append(item.Schema.Required,field.Name)
         if subItem.Type != "" {
            fieldType = subItem.Type
         } else {
            fieldType = subItem.Schema.Type
         }

         item.Schema.Properties[field.Name] = SwaggerSchemaT{
            Type:             fieldType,
            Format:           subItem.Format,
            Items:            subItem.Items,
            Description:      description,
            Required:         subItem.Schema.Required,
            Properties:       subItem.Schema.Properties,
         }
      }
      Goose.Logf(6,"Got final struct: %#v",item)
      return item, nil
   }

   name := parm.Name()
   if name == "" {
      name = fmt.Sprintf("%s /// %#v",parm,parm)
   }

   return nil, errors.New(fmt.Sprintf("invalid parameter %s type",name))
}


//From the endpoint, defined in the Service struct, init the variables, the certificates infrastructure and the server listeners.
//Besides, load de configuration file to start basic data required for the proposed solution.
func initSvc(svcElem EndPointHandler) (*Service, error) {
   var err             error
   var resp           *Service
//   var svcRecv         reflect.Value
   var cfg             io.Reader
   var ls              net.Listener
   var ClientCertPem []byte
   var ClientCert     *x509.Certificate
   var CertIdStr     []string
   var CertId          int

   resp = &Service{
      AuthRequired: false,
   }

   cfg, err = svcElem.GetConfig()
   if err != nil {
      Goose.Logf(1,"Failed opening config: %s", err)
      return nil, err
   }

   err = json.NewDecoder(cfg).Decode(&resp)

   if (err!=nil) && (err!=io.EOF) {
      Goose.Logf(1,"Failed parsing config file: %s", err)
      return nil, err
   }

   ls, err = net.Listen("tcp", resp.ListenAddress)
   if err != nil {
      Goose.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.Listener, err = NewListener(ls)
   if err != nil {
      Goose.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }


   ls, err = net.Listen("tcp", resp.CRLListenAddress)
   if err != nil {
      Goose.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.CRLListener, err = NewListener(ls)
   if err != nil {
      Goose.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }

   resp.PageNotFound, err = ioutil.ReadFile(resp.PageNotFoundPath)
   if err != nil {
      Goose.Logf(1,"Failed reading %s file: %s", resp.PageNotFoundPath, err)
      return nil, err
   }

   err = resp.ReadCert(&resp.Auth.CACertPem,     &resp.Auth.CACert,     resp.PemPath + "/rootCA.crt")
   if err != nil {
      Goose.Logf(1,"Failed reading rootCA.crt file: %s", err)
      return nil, err
   }
   err = resp.ReadCert(&resp.Auth.ServerCertPem, &resp.Auth.ServerCert, resp.PemPath + "/server.crt")
   if err != nil {
      Goose.Logf(1,"Failed reading server.crt file: %s", err)
      return nil, err
   }

   err = resp.ReadCRL(&resp.Auth.CACRL, "rootCA.crl")
   if err != nil {
      Goose.Logf(1,"Failed reading rootCA.crl file: %s", err)
      return nil, err
   }


//   err = resp.readEcdsaKey(&resp.Auth.CAKeyPem,       &resp.Auth.CAKey,      "rootCA.key")
//   if err != nil {
//      Goose.Logf(1,"Failed reading rootCA.key file: %s", err)
//      return nil, err
//   }
   err = resp.ReadRsaKey(&resp.Auth.CAKeyPem,   &resp.Auth.CAKey,  resp.PemPath + "/rootCA.key")
   if err != nil {
      Goose.Logf(1,"Failed reading rootCA.key file: %s", err)
      return nil, err
   }

   err = resp.ReadRsaKey(&resp.Auth.ServerKeyPem,   &resp.Auth.ServerKey,  resp.PemPath + "/server.key")
   if err != nil {
      Goose.Logf(1,"Failed reading server.key file: %s", err)
      return nil, err
   }

   resp.CertPool = x509.NewCertPool()
   resp.CertPool.AddCert(resp.Auth.CACert)

   err = filepath.Walk(resp.PemPath + "/client", func (path string, f os.FileInfo, err error) error {
      if (len(path)<4) || (path[len(path)-4:]!=".crt") {
         return nil
      }

      err = resp.ReadCert(&ClientCertPem, &ClientCert, path)
      if err != nil {
         Goose.Logf(1,"Failed reading %s file: %s", path, err)
         return err
      }

      CertIdStr = strings.Split(ClientCert.Subject.CommonName,":")
      if len(CertIdStr) > 2 {
         Goose.Logf(1,"Failed extracting %s subject name",ClientCert.Subject.CommonName)
         return err
      }

      fmt.Sscanf(CertIdStr[len(CertIdStr)-1],"%d",&CertId)
      if resp.UserCerts == nil {
         resp.UserCerts = map[int]*x509.Certificate{CertId:ClientCert}
      } else {
         resp.UserCerts[CertId] = ClientCert
      }

     return nil
   })

   return resp, err
}

func buildHandle(this reflect.Value, met reflect.Method, posttype reflect.Type) (func ([]string, Unmarshaler) Response) {
   return func (parms []string, Unmarshal Unmarshaler) Response {
      var httpResp Response
      var i int
      var p, ptmp string
//                  var outs []reflect.Value
      var ins []reflect.Value
      var parm reflect.Value
      var parmType reflect.Type
      var parmTypeName string
      var err error
      var postvalue reflect.Value
      var elemDelim, keyDelim string
      var keyval string
      var arrKeyVal []string

      ins = []reflect.Value{this}
      for i, p = range parms {
         Goose.Logf(5,"parm: %d:%s",i+1,p)
         parmType = met.Type.In(i+1)
         parmTypeName = parmType.Name()
         if parmTypeName == "string" {
            p = "\"" + p + "\""

         } else if (parmType.Kind() == reflect.Array) || (parmType.Kind() == reflect.Slice) {
            parmTypeName = "[]" + parmType.Elem().Name()
            if parmType.Elem().Name() == "string" {
               p = "[\"" + strings.Replace(p,",","\",\"",-1) + "\"]"
            } else {
               p = "[" + p + "]"
            }
         } else if parmType.Kind() == reflect.Map {
            parmTypeName = "map[" + parmType.Key().Name() + "]" + parmType.Elem().Name()
            if parmType.Elem().Name() == "string" {
               elemDelim = "\""
            } else {
               elemDelim = ""
            }

            if parmType.Key().Name() == "string" {
               keyDelim = "\""
            } else {
               keyDelim = ""
            }
            ptmp = ""
            for _, keyval = range strings.Split(p,",") {
               arrKeyVal = strings.Split(keyval,":")
               if len(arrKeyVal) != 2 {
                  Goose.Logf(1,"map parameter encoding error: %s",err)
                  httpResp.Status = http.StatusInternalServerError
                  httpResp.Body   = "Internal server error"
                  httpResp.Header = map[string]string{}
                  Goose.Logf(1,"Internal server error on map parameter encoding %s: %s",p,err)
                  return httpResp
               }
               if len(ptmp)>0 {
                  ptmp += ","
               }
               ptmp += keyDelim + arrKeyVal[0] + keyDelim + ":" + elemDelim + arrKeyVal[1] + elemDelim
            }
            p = "{" + ptmp + "}"
         }
         Goose.Logf(5,"parmtype: %s",parmTypeName)
         Goose.Logf(4,"parmcoding: %s",p)
         parm = reflect.New(parmType)
         err = json.Unmarshal([]byte(p),parm.Interface())
         if err != nil {
            Goose.Logf(1,"unmarshal error: %s",err)
            httpResp.Status = http.StatusInternalServerError
            httpResp.Body   = "Internal server error"
            httpResp.Header = map[string]string{}
            Goose.Logf(1,"Internal server error parsing %s: %s",p,err)
            return httpResp
         }
         ins = append(ins,reflect.Indirect(parm))
         Goose.Logf(5,"ins: %d:%s",len(ins),ins)
      }

//                  Goose.Logf(1,"posttype: %#v, postdata: %s",posttype, postdata)
      Goose.Logf(6,"posttype: %#v",posttype)
      if posttype != nil {
         postvalue = reflect.New(posttype)
         err = Unmarshal.Decode(postvalue.Interface())
         if err != nil {
            httpResp.Status = http.StatusInternalServerError
            httpResp.Body   = "Internal server error"
            httpResp.Header = map[string]string{}
            Goose.Logf(1,"Internal server error parsing post body: %s - postvalue: %s",err,postvalue.Interface())
            return httpResp
         }
         Goose.Logf(6,"postvalue: %#v",postvalue)
         ins = append(ins,reflect.Indirect(postvalue))
         Goose.Logf(5,"ins: %d:%s",len(ins),ins)
      }

      return met.Func.Call(ins)[0].Interface().(Response)
   }
}

func ParseFieldList(listEncoding string, parmcountIn int, fld reflect.StructField, method reflect.Method, methodName string, swaggerParametersIn []SwaggerParameterT) (list []string, parmcount int, pt reflect.Type, swaggerParameters  []SwaggerParameterT, err error) {
   var lstFlds            string
   var lstFld             string
   var doc                string
   var SwaggerParameter  *SwaggerParameterT

   parmcount         = parmcountIn
   swaggerParameters = swaggerParametersIn

   list = []string{}
   lstFlds = fld.Tag.Get(listEncoding)
   if lstFlds != "" {
      for _, lstFld = range strings.Split(lstFlds,",") {
         parmcount++
         if (parmcount+1) > method.Type.NumIn() {
            Goose.Logf(1,"%s (with query) at method %s", ErrorWrongParameterCount, methodName)
            err = ErrorWrongParameterCount
            return
         }
         pt = method.Type.In(parmcount)
         SwaggerParameter, err = GetSwaggerType(pt)
         if err != nil {
            return
         }

         if SwaggerParameter == nil {
            err = ErrorInvalidNilParam
            return
         }

/*
         if SwaggerParameter.Schema.Required != nil {
            Goose.Logf(1,"%s: %s",ErrorInvalidParameterType,lstFld)
            Goose.Logf(1,"SwaggerParameter: %#v",SwaggerParameter)
            err = ErrorInvalidParameterType
            return
         }
*/

         doc = fld.Tag.Get(lstFld)
         if doc != "" {
            SwaggerParameter.Description = doc
         }

         SwaggerParameter.Name     = lstFld
         SwaggerParameter.In       = listEncoding
         SwaggerParameter.Required = true
         SwaggerParameter.Schema   = nil

         if pt.Kind() == reflect.Map {
            if SwaggerParameter.Items == nil {
               SwaggerParameter.Items = &SwaggerItemT{}
            }
            kname := pt.Key().Name()
            if kname == "string" {
               SwaggerParameter.Items.XKeyType = "string"
            } else {
               SwaggerParameter.Items.XKeyType   = "integer"
               if kname[len(kname)-2:] == "64" {
                  SwaggerParameter.Items.XKeyFormat = "int64"
               } else {
                  SwaggerParameter.Items.XKeyFormat = "int32"
               }
            }
         }

         swaggerParameters = append(swaggerParameters,*SwaggerParameter)
         list              = append(list,lstFld)
      }
   }

   Goose.Logf(6,"parm: %s, count: %d, met.in:%d",methodName, parmcount,method.Type.NumIn()) // 3, 4
   return
}



func New(svcs ...EndPointHandler) (*Service, error) {
   var resp                *Service
   var svc                  EndPointHandler
   var svcElem              EndPointHandler
   var svcRecv              reflect.Value
   var consumes             string
   var svcConsumes          string
   var produces             string
   var svcProduces          string
   var allowGzip            string
   var enableCORS           string
   var svcRoot              string
   var i, j                 int
   var typ                  reflect.Type
   var typPtr               reflect.Type
   var pt                   reflect.Type
   var fld                  reflect.StructField
   var method               reflect.Method
   var parmcount            int
   var httpmethod, path     string
   var methodName           string
   var tk                   string
   var ok                   bool
   var re                   string
   var c                    rune
   var err                  error
   var stmp                 string
   var SwaggerParameter    *SwaggerParameterT
   var swaggerParameters  []SwaggerParameterT
   var swaggerInfo          SwaggerInfoT
   var swaggerLicense       SwaggerLicenseT
   var swaggerContact       SwaggerContactT
   var globalDataCount      int
   var responseOk           string
   var responses map[string]SwaggerResponseT
   var fldType              reflect.Type
   var doc                  string
   var description         *string
   var headers            []string
   var query              []string

   for _, svc = range svcs {
      Goose.Logf(6,"Elem: %#v (Kind: %#v)", reflect.ValueOf(svc), reflect.ValueOf(svc).Kind())
      if reflect.ValueOf(svc).Kind() == reflect.Ptr {
         Goose.Logf(6,"Elem: %#v", reflect.ValueOf(svc).Elem())
         svcElem = reflect.ValueOf(svc).Elem().Interface().(EndPointHandler)
         Goose.Logf(6,"Elem type: %s, ptr type: %s", reflect.TypeOf(svcElem), reflect.TypeOf(svc))
      } else {
         svcElem = svc
         Goose.Logf(6,"Elem type: %s", reflect.TypeOf(svcElem))
      }

      if resp == nil {
         resp, err = initSvc(svcElem)
         if err != nil {
            return nil, err
         }
      }

      typ = reflect.ValueOf(svcElem).Type()
      if typ.Kind()==reflect.Ptr {
         typPtr     = typ
         typ        = typ.Elem()
      } else {
         typPtr = reflect.PtrTo(typ)
      }

      if resp.Swagger == nil {
         for i=0; (i<typ.NumField()) && (globalDataCount<4); i++ {
            if svcRoot == "" {
               svcRoot = typ.Field(i).Tag.Get("root")
               if svcRoot != "" {
                  svcConsumes = typ.Field(i).Tag.Get("consumes")
                  svcProduces = typ.Field(i).Tag.Get("produces")
                  allowGzip   = typ.Field(i).Tag.Get("allowGzip")
                  enableCORS  = typ.Field(i).Tag.Get("EnableCORS")
                  globalDataCount++
               }
            }
            if swaggerInfo.Title == "" {
               stmp = typ.Field(i).Tag.Get("title")
               if stmp != "" {
                  swaggerInfo.Title          = stmp
                  swaggerInfo.Description    = typ.Field(i).Tag.Get("description")
                  swaggerInfo.TermsOfService = typ.Field(i).Tag.Get("tos")
                  swaggerInfo.Version        = typ.Field(i).Tag.Get("version")
                  globalDataCount++
               }
            }
            if swaggerContact.Name == "" {
               stmp = typ.Field(i).Tag.Get("contact")
               if stmp != "" {
                  swaggerContact.Name  = stmp
                  swaggerContact.Url   = typ.Field(i).Tag.Get("url")
                  swaggerContact.Email = typ.Field(i).Tag.Get("email")
                  globalDataCount++
               }
            }
            if swaggerLicense.Name == "" {
               stmp = typ.Field(i).Tag.Get("license")
               if stmp != "" {
                  swaggerLicense.Name  = stmp
                  swaggerLicense.Url   = typ.Field(i).Tag.Get("url")
                  globalDataCount++
               }
            }
         }

         swaggerInfo.Contact = swaggerContact
         swaggerInfo.License = swaggerLicense

         svcRoot     = strings.Trim(strings.Trim(svcRoot," "),"/") + "/"
         svcConsumes = strings.Trim(svcConsumes," ")
         svcProduces = strings.Trim(svcProduces," ")

         if (svcRoot=="") || (svcConsumes=="") || (svcProduces=="") {
            Goose.Logf(1,"Err: %s",ErrorNoRoot)
            return nil, ErrorNoRoot
         }


         hostport := strings.Split(resp.ListenAddress,":")
         if hostport[0] == "" {
            hostport[0] = resp.Auth.ServerCert.DNSNames[0]
         }

         resp.Swagger = &SwaggerT{
            Version:     "2.0",
            Info:        swaggerInfo,
            Host:        strings.Join(hostport,":"),
            BasePath:    "/" + svcRoot[:len(svcRoot)-1],
            Schemes:     []string{"https"},
            Consumes:    []string{svcConsumes},
            Produces:    []string{svcProduces},
            Paths:       map[string]SwaggerPathT{},
            Definitions: map[string]SwaggerSchemaT{},
         }

         Goose.Logf(6,"enableCORS: [%s]",enableCORS)
         if enableCORS != "" {
            resp.EnableCORS, err = strconv.ParseBool(enableCORS)
            Goose.Logf(6,"resp.EnableCORS: %#v",resp.EnableCORS)
            if err != nil {
               Goose.Logf(1,"Err: %s",ErrorServiceSyntax)
               return nil, ErrorServiceSyntax
            }
         }

         Goose.Logf(6,"allowGzip: [%s]",allowGzip)
         if allowGzip != "" {
            resp.AllowGzip, err = strconv.ParseBool(allowGzip)
            Goose.Logf(6,"resp.AllowGzip: %#v",resp.AllowGzip)
            if err != nil {
               Goose.Logf(1,"Err: %s",ErrorServiceSyntax)
               return nil, ErrorServiceSyntax
            }
         }
      }

      for i=0; i<typ.NumField(); i++ {
         fld = typ.Field(i)
         httpmethod = fld.Tag.Get("method")
         if httpmethod != "" {
            methodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]
            svcRecv = reflect.ValueOf(svcElem)
            if method, ok = typ.MethodByName(methodName); !ok {
               if method, ok = typPtr.MethodByName(methodName); !ok {
                  Goose.Logf(5,"|methods|=%d",typ.NumMethod())
                  Goose.Logf(5,"type=%s.%s",typ.PkgPath(),typ.Name())
                  for j=0; j<typ.NumMethod(); j++ {
                     mt := typ.Method(j)
                     Goose.Logf(5,"%d: %s",j,mt.Name)
                  }

                  Goose.Logf(1,"Method not found: %s, Data: %#v",methodName,typ)
                  return nil, errors.New(fmt.Sprintf("Method not found: %s",methodName))
               } else {
                  Goose.Logf(1,"Pointer method found, type of svcElem: %s",reflect.TypeOf(svcElem))
                  svcRecv = reflect.ValueOf(svc)
                  Goose.Logf(5,"Pointer method found: %s",methodName)
               }
            }
            path   = fld.Tag.Get("path")

            if _, ok := resp.Swagger.Paths[path]; !ok {
               resp.Swagger.Paths[path] = SwaggerPathT{}
   //         } else if _, ok := resp.Swagger.Paths[path][httpmethod]; !ok {
   //            resp.Swagger.Paths[path][httpmethod] = SwaggerOperationT{}
            }

            swaggerParameters = []SwaggerParameterT{}

            re = "^" + strings.ToUpper(httpmethod) + ":/" + svcRoot

            parmcount = 0
            for _, tk = range strings.Split(strings.Trim(path,"/"),"/") {
               if tk!="" {
                  if (tk[0]=='{') && (tk[len(tk)-1]=='}') {
                     re += "([^/]+)/"
                     parmcount++
                     SwaggerParameter, err = GetSwaggerType(method.Type.In(parmcount))
                     if err != nil {
                        return nil, err
                     }

                     if SwaggerParameter == nil {
                        return nil, ErrorInvalidNilParam
                     }

                     if (SwaggerParameter.Items != nil) || (SwaggerParameter.CollectionFormat!="") || (SwaggerParameter.Schema.Required != nil) {
                        Goose.Logf(1,"%s: %s",tk[1:len(tk)-1])
                        return nil, ErrorInvalidParameterType
                     }

                     doc = fld.Tag.Get(tk[1:len(tk)-1])
                     if doc != "" {
                        description    = new(string)
                        (*description) = doc
                     } else {
                        description = SwaggerParameter.Schema.Description
                     }

                     xkeytype := ""
                     if SwaggerParameter.Schema != nil {
                        xkeytype = SwaggerParameter.Schema.XKeyType
                     }

                     swaggerParameters = append(
                        swaggerParameters,
                        SwaggerParameterT{
                           Name: tk[1:len(tk)-1],
                           In:   "path",
                           Required: true,
                           Type: SwaggerParameter.Schema.Type,
                           XKeyType: xkeytype,
                           Description: *description,
                           Format: SwaggerParameter.Format,
                        })
                  } else if (tk[0]!='{') && (tk[len(tk)-1]!='}') {
                     for _, c = range tk {
                        re += fmt.Sprintf("\\x{%x}",c)
                     }
                     re += "/"
                  } else {
                     return nil, errors.New("syntax error at " + tk)
                  }
               }
            }

            if resp.Svc == nil {
               resp.Svc = []UrlNode{}
            }

            re += "{0,1}$"

            Goose.Logf(4,"Service " + strings.ToUpper(httpmethod) + ":/" + svcRoot + path + ", RE=" + re )

            query, parmcount, pt, swaggerParameters, err = ParseFieldList("query", parmcount, fld, method, methodName, swaggerParameters)
            if err != nil {
               return nil, err
            }

            headers, parmcount, pt, swaggerParameters, err = ParseFieldList("header", parmcount, fld, method, methodName, swaggerParameters)
            if err != nil {
               return nil, err
            }

            if fld.Tag.Get("postdata") != "" {
               parmcount++
               if (parmcount+1) != method.Type.NumIn() {
                  return nil, errors.New("Wrong parameter count (with post) at method " + methodName)
               }
               pt = method.Type.In(parmcount)
               SwaggerParameter, err = GetSwaggerType(pt)
               if err != nil {
                  return nil, err
               }

               if SwaggerParameter == nil {
                  return nil, ErrorInvalidNilParam
               }

               doc = fld.Tag.Get(fld.Tag.Get("postdata"))
               if doc != "" {
                  SwaggerParameter.Schema.Description    = new(string)
                  (*SwaggerParameter.Schema.Description) = doc
               }

               SwaggerParameter.Name     = fld.Tag.Get("postdata")
               SwaggerParameter.In       = "body"
               SwaggerParameter.Required = true

               swaggerParameters = append(swaggerParameters,*SwaggerParameter)
            } else {
               if (parmcount+1) != method.Type.NumIn() {
                  return nil, errors.New("Wrong parameter count at method " + methodName)
               }
               pt = nil
            }

            Goose.Logf(5,"Registering: %s",re)
            consumes = fld.Tag.Get("consumes")
            if consumes == "" {
               consumes = svcConsumes
            }

            produces = fld.Tag.Get("produces")
            if produces == "" {
               produces = svcProduces
            }

            responses = map[string]SwaggerResponseT{}

            responseOk = fld.Tag.Get("ok")
            if responseOk != "" {
               fldType = fld.Type
               if fldType.Kind() == reflect.Ptr {
                  fldType = fldType.Elem()
               }

               SwaggerParameter, err = GetSwaggerType(fldType)
               if err != nil {
                  return nil, err
               }

               if SwaggerParameter == nil {
                  responses[fmt.Sprintf("%d",http.StatusNoContent)] = SwaggerResponseT{
                     Description: responseOk,
                  }
               } else {

                  doc = fld.Tag.Get(fld.Name)

                  if doc != "" {
                     SwaggerParameter.Schema.Description    = new(string)
                     (*SwaggerParameter.Schema.Description) = doc
                  }

                  if SwaggerParameter.Schema == nil {
                     SwaggerParameter.Schema = &SwaggerSchemaT{}
                  }

                  if (SwaggerParameter.Schema.Type=="") && (SwaggerParameter.Type!="") {
                     SwaggerParameter.Schema.Type = SwaggerParameter.Type
                  }

                  responses[fmt.Sprintf("%d",http.StatusOK)] = SwaggerResponseT{
                     Description: responseOk,
                     Schema:      SwaggerParameter.Schema,
                  }
                  //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
                  //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
                  Goose.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusOK)].Schema))
               }
            } else {
               if responseFunc, ok := typ.MethodByName(fld.Name + "Responses"); ok {
                  responseList := responseFunc.Func.Call([]reflect.Value{})[0].Interface().(map[string]ResponseT)
                  for responseStatus, responseSchema := range responseList {
                     SwaggerParameter, err = GetSwaggerType(reflect.TypeOf(responseSchema.TypeReturned))
                     if err != nil {
                        return nil, err
                     }
                     if SwaggerParameter == nil {
                        responses[responseStatus] = SwaggerResponseT{
                           Description: responseSchema.Description,
                        }
                     } else {
                        responses[responseStatus] = SwaggerResponseT{
                           Description: responseSchema.Description,
                           Schema:      SwaggerParameter.Schema,
                        }
                     }
                  }
               }
            }


            resp.Swagger.Paths[path][strings.ToLower(httpmethod)] = SwaggerOperationT{
               Schemes: []string{"https"},
               OperationId: methodName,
               Parameters: swaggerParameters,
               Responses: responses,
               Consumes: []string{consumes},
               Produces: []string{produces},
            }

            Goose.Logf(5,"Registering marshalers: %s, %s",consumes,produces)

            resp.Svc = append(resp.Svc,UrlNode{
               Path: path,
               Matcher: regexp.MustCompile(re),
               consumes: consumes,
               produces: produces,
               Headers: headers,
               Query: query,
               Handle: buildHandle(svcRecv,method,pt),
            })
         }
      }
   }

   return resp, nil
}


//Init a webserver and wait for http requests.
func (svc *Service) ListenAndServeTLS() error {
   var err    error
   var aType  tls.ClientAuthType
   var wg     sync.WaitGroup
   var crypls net.Listener
   var hn                   string

   if svc.ListenAddress[0] == ':' {
      hn, err = os.Hostname()
      if err!=nil {
         Goose.Logf(1,"Error checking hostname: %s", err)
         return err
      }
   }

   if svc.AuthRequired {
      aType = tls.RequireAndVerifyClientCert
   } else {
//      aType = tls.RequireAnyClientCert
      aType = tls.RequestClientCert
   }

   Goose.Logf(6,"auth: %#v",aType)

   srv := &http.Server{
      Addr: hn + svc.ListenAddress,
      Handler: svc,

      TLSConfig: &tls.Config{
         ClientAuth: aType,
//         ClientCAs: svc.CertPool,
//         InsecureSkipVerify: true,
         Certificates: make([]tls.Certificate, 1),
      },
   }

   srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(svc.PemPath + "/server.crt", svc.PemPath + "/server.key")
   if err != nil {
      Goose.Logf(1,"Failed reading server certificates: %s",err)
      return err
   }

   srv.TLSConfig.BuildNameToCertificate()

   crypls = tls.NewListener(svc.Listener,srv.TLSConfig)

   srvcrl := &http.Server{
      Addr: hn + svc.CRLListenAddress,
      Handler: svc.Auth,
   }

   wg.Add(2)

   go func() {
      Goose.Logf(5,"CRL Listen Address: %s",svc.CRLListenAddress)
      defer wg.Done()
      err = srvcrl.Serve(svc.CRLListener)

//      Goose.Logf(5,"CRL Listen is serving")
//      err = http.ListenAndServe(svc.CRLListenAddress,svc.Auth)
//      if err != nil {
//         Goose.Fatalf(1,"Error serving CRL: %s",err)
//      }
      Goose.Logf(5,"CRL Listen ended listening")
   }()


   go func() {
      defer wg.Done()
      err = srv.Serve(crypls)
      Goose.Logf(5,"Main Listen ended listening")
   }()

   wg.Wait()

   Goose.Logf(1,"Ending listening")

   return err
}


func (casvc ServerCert) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   if r.URL.Path == "/rootCA.crl" {
      w.WriteHeader(http.StatusOK)
      w.Header().Set("Content-Type", "application/pkix-crl")
      w.Write(casvc.CACRL)
      return
   }

   w.WriteHeader(http.StatusNotFound)
}

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   var match              [][]string
   var endpoint               UrlNode
   var resp                   Response
   var cert, UserCert        *x509.Certificate
   var CertIdStr            []string
   var CertId                 int
   var found                  bool
   var  ok              bool
//   var ok              bool
//   var body                 []byte
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


   Goose.Logf(5,"Peer certificates")
   found = false
   for _, cert = range r.TLS.PeerCertificates {
      Goose.Logf(5,"Peer certificate: #%s, ID: %s, Issuer: %s, Subject: %s, \n\n\n",cert.SerialNumber,cert.SubjectKeyId,cert.Issuer.CommonName,cert.Subject.CommonName)
      CertIdStr = strings.Split(cert.Subject.CommonName,":")
      if len(CertIdStr) != 2 {
         continue
      }

      fmt.Sscanf(CertIdStr[1],"%d",&CertId)
      if UserCert, ok = svc.UserCerts[CertId]; ok {
         if bytes.Equal(UserCert.Raw,cert.Raw) {
            found = true
            break
         }
      }
   }



   if r.URL.Path=="/crtlogin" {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte(" "))
      return
   }


   if !found {
      Goose.Logf(1,"Unauthorized access attempt, method: %s",r.Method)
      w.WriteHeader(http.StatusNotFound)
      w.Write(svc.PageNotFound)
      return
   }

   Goose.Logf(7,"TLS:%#v",r.TLS)


   hd := w.Header()
   hd.Add("Access-Control-Allow-Origin","*")
   hd.Add("Vary","Origin")

   if r.Method == "OPTIONS" {
      hd.Add("Access-Control-Allow-Methods","POST, GET, OPTIONS, PUT, DELETE")
      hd.Add("Access-Control-Allow-Headers", "*")
//Access-Control-Allow-Origin: http://foo.example
//Access-Control-Allow-Methods: POST, GET, OPTIONS
//Access-Control-Allow-Headers: X-PINGOTHER
//Access-Control-Allow-Origin: *
   }

   Goose.Logf(6,"Will check if swagger.json is requested: %#v",svc.Swagger)
   if r.URL.Path=="/swagger.json" {
      defer func() {
         Goose.Logf(5,"Testing the need to recover...")
         if r := recover(); r != nil {
            Goose.Logf(1,"Internal server error writing response body for swagger.json: %#v",r)
         }
         Goose.Logf(5,"Tested the need to recover!!!")
      }()
      hd.Add("Content-Type","application/json")
//      w.WriteHeader(http.StatusOK)
      Goose.Logf(6,"Received request of swagger.json: %#v",svc.Swagger)
//      mrsh = json.NewEncoder(w)
//      err = mrsh.Encode(svc.Swagger)
      buf, err = json.Marshal(svc.Swagger)
      if err!=nil {
         Goose.Logf(1,"Internal server error marshaling swagger.json: %s",err)
      }
      hd.Add("Content-Length", fmt.Sprintf("%d",len(buf)))
      _, err = io.WriteString(w,string(buf))
      if err!=nil {
         Goose.Logf(1,"Internal server error writing response body for swagger.json: %s",err)
      }
      return
   }

   for _, endpoint = range svc.Svc {
      Goose.Logf(5,"trying %s with endpoint: %s",r.Method+":"+r.URL.Path,endpoint.Path)
      match = endpoint.Matcher.FindAllStringSubmatch(r.Method+":"+r.URL.Path,-1)
      if len(match) > 0 {
         Goose.Logf(5,"Found endpoint %s for: %s",endpoint.Path,r.URL.Path)
         break
      }
   }

   if len(match) == 0 {
      Goose.Logf(1,"Invalid service handler " + r.URL.Path)
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Invalid service handler " + r.URL.Path))
      return
   }

   if len(endpoint.Query) > 0 {
      r.ParseForm()
      for _, qry = range endpoint.Query {
         if _, ok := r.Form[qry]; !ok {
            Goose.Logf(1,"%s: %s",ErrorMissingRequiredQueryField,qry)
            return
         }
         match[0] = append(match[0],r.Form[qry][0]) // TODO array support
      }
   }

   for _, header = range endpoint.Headers {
      if (r.Header[header]==nil) || (len(r.Header[header])==0) {
         Goose.Logf(1,"%s: %s",ErrorMissingRequiredHTTPHeader,header)
         Goose.Logf(6,"HTTP Headers found: %#v",r.Header)
         return
      }
      match[0] = append(match[0],r.Header[header][0]) // TODO array support
   }

   Goose.Logf(5,"checking marshalers: %s, %s",endpoint.consumes,endpoint.produces)

   if endpoint.consumes == "application/json" {
      umrsh = json.NewDecoder(r.Body)
   } else if endpoint.consumes == "application/xml" {
      umrsh = xml.NewDecoder(r.Body)
   }

   Goose.Logf(6,"umrsh=%#v",umrsh)

   resp = endpoint.Handle(match[0][1:],umrsh)

   outWriter = w

   if encRequest, ok = r.Header["Accept-Encoding"] ; ok {
      Goose.Logf(6,"Accept-Encoding: %#v",encRequest)
      if svc.AllowGzip == true {
         Goose.Logf(5,"svc.AllowGzip == true")
gzipcheck:
         for _, enc = range encRequest {
            for _, e = range strings.Split(enc,", ") {
               Goose.Logf(5,"Encoding: %s",e)
               if e == "gzip" {
                  Goose.Logf(5,"Using gzip")
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

   if resp.Status != 0 {
      for k, v := range resp.Header {
         hd.Add(k, v)
      }
      w.WriteHeader(resp.Status)
   }


   if resp.Status != http.StatusNoContent {
      err = mrsh.Encode(resp.Body)
      if err!=nil {
         Goose.Logf(1,"Internal server error writing response body (no status sent to client): %s",err)
         return
      }
   }

}

