package stonelizard

import (
   "io"
   "os"
   "fmt"
   "net"
   "sync"
   "time"
   "errors"
   "regexp"
   "strconv"
   "strings"
   "reflect"
   "runtime"
//   "net/url"
   "net/http"
   "math/rand"
//   "io/ioutil"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
   "encoding/xml"
   "encoding/json"
   "compress/gzip"
//   "path/filepath"
   "golang.org/x/net/websocket"
   "github.com/luisfurquim/strtree"
)

func init() {
   gorootRE = regexp.MustCompile("^\\s+" + os.Getenv("GOROOT") + "[^\\s]+\\.go:[0-9]+")
   gosrcRE = regexp.MustCompile("^\\s+[^\\s]+\\.go:[0-9]+")
}

//Convert the UrlNode field value, from the Service struct, into a string
func (svc Service) String() string {
   var s string

   if svc.Config.PageNotFound() != nil {
      s = "404: " + string(svc.Config.PageNotFound()) + "\n"
   }

   s += fmt.Sprintf("%s",svc.Svc)
   return s
}

//Convert the Handle field value, from UrlNode struct, into a Go-syntax string (method signature)
func (u UrlNode) String() string {
   var s string

   if u.Handle != nil {
      s = fmt.Sprintf("Method: %#v\n",u.Handle)
   }

   return s
}



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

// http://www.hydrogen18.com/blog/stop-listening-http-server-go.html
//Implements a wrapper on the system accept
func (sl *StoppableListener) Accept() (net.Conn, error) {

   for {
      Goose.Listener.Logf(6,"Start accept loop")

      //Wait up to one second for a new connection
      sl.SetDeadline(time.Now().Add(5*time.Second))

      newConn, err := sl.TCPListener.Accept()

      //Check for the channel being closed
      select {
         case <-sl.stop:
            Goose.Listener.Logf(2,"Received stop request")
            return nil, ErrorStopped
         default:
            Goose.Listener.Logf(6,"channel still open")
            // If the channel is still open, continue as normal
      }

      if err != nil {
         Goose.Listener.Logf(6,"Err not nil: %s",err)
         netErr, ok := err.(net.Error)

         // If this is a timeout, then continue to wait for
         // new connections
         if ok && netErr.Timeout() && netErr.Temporary() {
            Goose.Listener.Logf(6,"continue")
            continue
         }
      }

      if err==nil {
         Goose.Listener.Logf(3,"done listening")
         return newConn, nil
      }

      Goose.Listener.Logf(2,"done listening, err=%s",err)
      return nil, err
   }
}

//Stop the service and releases the chanel
func (sl *StoppableListener) Stop() {
   close(sl.stop)
}


//Close the webservices and the listeners
func (svc Service) Close() {
   Goose.Listener.Logf(2,"Closing listeners")
   svc.Listener.Stop()
   svc.Listener.Close()
   svc.CRLListener.Stop()
   svc.CRLListener.Close()
   Goose.Listener.Logf(3,"All listeners closed")
}

func GetWebSocketSpec(field reflect.StructField, WSMethodName string, WSMethod reflect.Method) (*SwaggerOperationT, error) {
   var operation            *SwaggerOperationT
   var err                   error
   var inTag                 string
   var tags                []string
   var parmName              string
   var SwaggerParameter     *SwaggerParameterT
   var parmcount             int
   var responses  map[string]SwaggerResponseT

   inTag = field.Tag.Get("in")
   tags  = strings.Split(field.Tag.Get("tags"),",")
   if len(tags) == 0 {
      tags = nil
   }

   if (len(tags) == 1) || (tags[0]=="") {
      tags = nil
   }

   responses, _, err = getResponses(field, field.Type)
   if err != nil {
      return nil, err
   }

   operation = &SwaggerOperationT{
      Tags:        tags,
      Description: field.Tag.Get("doc"),
      OperationId: WSMethodName,
      //Consumes []string `json:"consumes,omitempty"`
      //Produces []string `json:"produces,omitempty"`
      Parameters: []SwaggerParameterT{},
      //Responses:   map[string]SwaggerResponseT{},
      //Schemes:  []string{"ocpp"},
      Responses: responses, // callID , status, response
   }

   for parmcount, parmName = range strings.Split(strings.Trim(inTag,""),",") {
      if parmName=="" {
         Goose.Swagger.Logf(1,"%s",ErrorParmListSyntax)
         return nil, ErrorInvalidType
      }

      SwaggerParameter, err = GetSwaggerType(WSMethod.Type.In(parmcount+1))
      if err != nil {
         return nil, err
      }

      if SwaggerParameter == nil {
         return nil, ErrorInvalidNilParam
      }

      if (SwaggerParameter.Items != nil) || (SwaggerParameter.CollectionFormat != "") || (SwaggerParameter.Schema.Required != nil && len(SwaggerParameter.Schema.Required)>0) {
         Goose.New.Logf(1,"%s: sch_req:%#v %#v",parmName,SwaggerParameter.Schema.Required,SwaggerParameter)
         return nil, ErrorInvalidParameterType
      }

      xkeytype := ""
      if SwaggerParameter.Schema != nil {
         xkeytype = SwaggerParameter.Schema.XKeyType
      }

      operation.Parameters = append(
         operation.Parameters,
         SwaggerParameterT{
            Name: parmName,
            In:   "path",
            Required: true,
            Type: SwaggerParameter.Schema.Type,
            XKeyType: xkeytype,
            Format: SwaggerParameter.Format,
         })
   }

   return operation, nil
}


func GetSwaggerType(parm reflect.Type) (*SwaggerParameterT, error) {
   var item, subItem *SwaggerParameterT
   var field          reflect.StructField
   var doc            string
   var description   *string
   var err            error
   var i              int
   var fieldType      string

   Goose.Swagger.Logf(6,"Tipo do parametro: %d: %s",parm.Kind(),parm)

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
      Goose.Swagger.Logf(6,"Got string")
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"string"}}, nil
   }

   if parm.Kind()==reflect.Ptr {
      return GetSwaggerType(parm.Elem())
   }

   if (parm.Kind()==reflect.Array) || (parm.Kind()==reflect.Slice) {
      item, err = GetSwaggerType(parm.Elem())
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         Goose.Swagger.Logf(6,"Error get array elem type item=%#v, err:%s", item, err)
         if (item!=nil) && (err==nil) && (item.Schema==nil) {
            Goose.Swagger.Logf(6,"And also error get array elem type item.Schema=%#v",item.Schema)
         }
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
      Goose.Swagger.Logf(6,"map elem type item=%#v, err:%s", item, err)
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         Goose.Swagger.Logf(6,"Error get map elem type item=%#v, err:%s", item, err)
         if (item!=nil) && (err==nil) && (item.Schema==nil) {
            Goose.Swagger.Logf(6,"And also error get map elem type item.Schema=%#v",item.Schema)
         }
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
         Schema: &SwaggerSchemaT{
            Type: item.Schema.Type,
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
      Goose.Swagger.Logf(6,"Got struct: %#v",item)
      for i=0; i<parm.NumField(); i++ {
         field = parm.Field(i)

         if field.Name[:1] == strings.ToLower(field.Name[:1]) {
            continue // Unexported field
         }


         Goose.Swagger.Logf(6,"Struct field: %s",field.Name)
         doc   = field.Tag.Get("doc")
         if doc != "" {
            description    = new(string)
            (*description) = doc
         } else {
            description = nil
         }

         item.Schema.Required = append(item.Schema.Required,field.Name)

         subItem, err = GetSwaggerType(field.Type)
         Goose.Swagger.Logf(6,"struct subitem=%#v, err:%s", subItem, err)
         if (subItem==nil) || (err != nil) || (subItem.Schema==nil) {
            if err == ErrorInvalidParameterType {
               Goose.Swagger.Logf(5,"%s on subitem %s, just ignoring", err, field.Name)
               continue
            }
            Goose.Swagger.Logf(1,"Error getting type of subitem %s: %s",field.Name,err)
            return nil, err
         }

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

      Goose.Swagger.Logf(6,"Got final struct: %#v",item)
      return item, nil
   }

   return nil, ErrorInvalidParameterType
}


// From the endpoint, defined in the Service struct, init the variables and the server listeners.
// Besides, load de configuration file to start basic data required for the proposed solution.
func initSvc(svcElem EndPointHandler) (*Service, error) {
   var err     error
   var i       int
   var resp   *Service
   var cfg     Shaper
   var ls      net.Listener
   var met     reflect.Method
   var ok      bool
   var typ     reflect.Type
   var fld     reflect.StructField
   var auth    AuthT

   resp = &Service{
//      AuthRequired: false,
      MatchedOps: map[int]int{},
   }

   cfg, err = svcElem.GetConfig()
   if err != nil {
      Goose.Initialize.Logf(1,"Failed opening config: %s", err)
      return nil, err
   }

   if cfg == nil {
      return nil, nil
   }

/*
   //TODO: shaper -> remover
   err = json.NewDecoder(cfg).Decode(&resp)

   if (err!=nil) && (err!=io.EOF) {
      Goose.Initialize.Logf(1,"Failed parsing config file: %s", err)
      return nil, err
   }
*/

   resp.Config = cfg

   ls, err = net.Listen("tcp", cfg.ListenAddress())
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.Listener, err = NewListener(ls)
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }


   ls, err = net.Listen("tcp", cfg.CRLListenAddress())
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.CRLListener, err = NewListener(ls)
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }

   typ = reflect.ValueOf(svcElem).Type()
   if typ.Kind()==reflect.Ptr {
      typ = typ.Elem()
   }


   met, ok = typ.MethodByName("SavePending")
   if ok {
      resp.SavePending = func(info interface{}) error {
         var err error

         errIFace := met.Func.Call([]reflect.Value{reflect.ValueOf(resp.Config),reflect.ValueOf(info)})[0].Interface()
         switch errIFace.(type) {
            case error:
               err = errIFace.(error)
         }

         if err != nil {
            Goose.Auth.Logf(1,"Internal server error saving unauthorized access attempt info: %s",err)
            Goose.Auth.Logf(5,"Dump of info on internal server error saving unauthorized access attempt info: %#v",info)
         }
         return err
      }
   } else {
      resp.SavePending = func(interface{}) error {
         return nil
      }
   }

   Goose.Auth.Logf(5,"cfg: %#v",cfg)
   Goose.Auth.Logf(5,"cfg.CertKit(): %#v",cfg.CertKit())
   Goose.Auth.Logf(5,"auth: %#v",auth)
   Goose.Auth.Logf(5,"auth: %#v",reflect.TypeOf((*AuthT)(nil)).Elem())
   if cfg.CertKit()!=nil && reflect.TypeOf(cfg.CertKit()).Implements(reflect.TypeOf((*AuthT)(nil)).Elem()) {
      resp.Authorizer = cfg.CertKit()
   } else {
      resp.Authorizer = PublicAccessT{}
   }

   for i=0; i<typ.NumField(); i++ {
      fld = typ.Field(i)
      if fld.Anonymous && fld.Type.Implements(reflect.TypeOf((*AuthT)(nil)).Elem()) {
         resp.Authorizer = reflect.ValueOf(svcElem).Field(i).Interface().(AuthT)
         break
      }
   }

   return resp, err
}

func isZero(v reflect.Value) bool {
   return reflect.Zero(v.Type()) == v
}

func f64Conv(i interface{}, t reflect.Type) (v reflect.Value, err error) {
   var u reflect.Value

   v = reflect.ValueOf(i)
   if !v.Type().ConvertibleTo(t) {
      return reflect.Zero(t), ErrorInvalidType
   }

   u = v.Convert(t)
   if !isZero(v) && isZero(u) {
      return reflect.Zero(t), ErrorConversionOverflow
   }

   return u, nil
}

func pushParms(parms []interface{}, obj reflect.Value, met reflect.Method) ([]reflect.Value, error) {
   var i int
   var iface interface{}
   var p string
   var buf []byte
   var parm reflect.Value
   var v reflect.Value
   var parmType reflect.Type
   var parmTypeName string
   var elemDelim string
   var keyDelim string
   var ptmp string
   var keyval string
   var arrKeyVal []string
   var err error
   var ins []reflect.Value

   ins = []reflect.Value{obj}

   for i, iface = range parms {
      parmType = met.Type.In(i+1)
      Goose.OpHandle.Logf(1,"parsing parm %#v",iface)

      switch iface.(type) {
         case string:
            p = iface.(string)
            Goose.OpHandle.Logf(5,"parm: %d:%s",i+1,p)
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
                     Goose.OpHandle.Logf(1,"Internal server error on map parameter encoding %s: %s",p,MapParameterEncodingError)
                     return nil, MapParameterEncodingError
                  }
                  if len(ptmp)>0 {
                     ptmp += ","
                  }
                  ptmp += keyDelim + arrKeyVal[0] + keyDelim + ":" + elemDelim + arrKeyVal[1] + elemDelim
               }
               p = "{" + ptmp + "}"
            }
            Goose.OpHandle.Logf(4,"parmcoding: %s",p)
            buf = []byte(p)
         case bool, []interface{}, map[string]interface{}:
            buf, err = json.Marshal(iface)
            if err != nil {
               Goose.OpHandle.Logf(1,"unmarshal error: %s",err)
               Goose.OpHandle.Logf(1,"Internal server error parsing %s: %s",buf,err)
               return nil, err
            }

         case float64:
            Goose.OpHandle.Logf(5,"parmtype: %s",parmType.Name())
            switch parmType.Kind() {
               case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32:
                  v, err = f64Conv(iface, parmType)
                  if err != nil {
                     Goose.OpHandle.Logf(1,"Internal server error parsing number parameter %q -> %q: %s", iface, v, err)
                     return nil, err
                  }
                  ins = append(ins,v)
                  continue
               case reflect.Float64:
               default:
                  Goose.OpHandle.Logf(1,"Internal server error parsing number parameter %q: %s", iface, WrongParameterType)
                  return nil, WrongParameterType
            }
         default:
            if parmType == reflect.TypeOf(iface) {
               ins = append(ins,reflect.ValueOf(iface))
               continue
            }
            Goose.OpHandle.Logf(1,"Internal server error parsing parameter %q: %s", iface, WrongParameterType)
            return nil, err
      }
      Goose.OpHandle.Logf(5,"parmtype: %s",parmTypeName)
      parm = reflect.New(parmType)
      Goose.OpHandle.Logf(1,"adding parm %s",buf)
      err = json.Unmarshal(buf,parm.Interface())
      if err != nil {
         Goose.OpHandle.Logf(1,"unmarshal error: %s",err)
         Goose.OpHandle.Logf(1,"Internal server error parsing [%s]: %s",buf,err)
         Goose.OpHandle.Logf(1,"parms %#v",parms)
         return nil, err
      }

      Goose.OpHandle.Logf(1,"added parm %#v",reflect.Indirect(parm).Interface())

      ins = append(ins,reflect.Indirect(parm))
      Goose.OpHandle.Logf(5,"ins: %d:%s",len(ins),ins)
   }

   return ins, nil
}

func buildHandle(this reflect.Value, isPtr bool, met reflect.Method, posttype []reflect.Type, accesstype uint8, useWebSocket bool) (func ([]interface{}, Unmarshaler, interface{}) Response) {
   return func (parms []interface{}, Unmarshal Unmarshaler, authinfo interface{}) Response {
      var httpResp Response
      var j int
//       var outs []reflect.Value
      var ins []reflect.Value
      var err error
      var postvalue reflect.Value
      var errmsg string

      defer func() {
         if panicerr := recover(); panicerr != nil {
            const size = 64 << 10
            var buf []byte
            var srcs []string
            var src string

            buf  = make([]byte, size)
            buf  = buf[:runtime.Stack(buf, false)]
            srcs = strings.Split(string(buf),"\n")
            for _, src = range srcs {
               if gosrcRE.MatchString(src) && (!gorootRE.MatchString(src)) {
                  break
               }
            }
            Goose.Serve.Logf(0,"panic (%s): calling %#v @ %s", panicerr, ins, src)
         }
      }()

      if isPtr {
         ins, err = pushParms(parms, this, met)
      } else {
         ins, err = pushParms(parms, this.Elem(), met)
      }

      if err != nil {
         httpResp.Status = http.StatusInternalServerError
         httpResp.Body   = "Internal server error"
         httpResp.Header = map[string]string{}
         return httpResp
      }

      if this.Type().Kind() == reflect.Struct {
         if _, ok := this.Type().FieldByName("Session"); ok {
            if this.Kind()==reflect.Ptr {
               Goose.OpHandle.Logf(0,"THIS SESSION: %#v",this.Elem().FieldByName("Session"))
            } else {
               Goose.OpHandle.Logf(0,"THIS SESSION: %#v",this.FieldByName("Session"))
            }
         }
      }

//                  Goose.OpHandle.Logf(1,"posttype: %#v, postdata: %s",posttype, postdata)
      Goose.OpHandle.Logf(6,"posttype: %#v",posttype)
      if posttype != nil {
         err = nil
         j   = 0
         for (err==nil) && (j<len(posttype)) {
            Goose.OpHandle.Logf(6,"posttype[%d]: %#v",j,posttype[j])
            postvalue = reflect.New(posttype[j])
            err = Unmarshal.Decode(postvalue.Interface())
            if err != nil && err != io.EOF {
               httpResp.Status = http.StatusInternalServerError
               httpResp.Body   = "Internal server error"
               httpResp.Header = map[string]string{}
               Goose.OpHandle.Logf(1,"Internal server error parsing post body: %s - postvalue: %s",err,postvalue.Elem().Interface())
               return httpResp
            }
            if err != io.EOF {
               Goose.OpHandle.Logf(6,"postvalue: %#v",postvalue)
               ins = append(ins,reflect.Indirect(postvalue))
               Goose.OpHandle.Logf(5,"ins: %d:%s",len(ins),ins)
//               Goose.OpHandle.Logf(5,"ins2: %c-%c-%c-%c",(*postvalue.Interface().(*string))[0],(*postvalue.Interface().(*string))[1],(*postvalue.Interface().(*string))[2],(*postvalue.Interface().(*string))[3])
               j++
            }
         }
      }

      Goose.OpHandle.Logf(5,"ins3: %d:%s",len(ins),ins)
      if accesstype == AccessAuthInfo || accesstype == AccessVerifyAuthInfo{
         Goose.OpHandle.Logf(5,"Checking the need for appending authinfo")
         if (len(ins)+1) == met.Type.NumIn() {
            Goose.OpHandle.Logf(5,"Appending authinfo: %s",reflect.ValueOf(authinfo).Elem())
            ins = append(ins,reflect.ValueOf(authinfo))
         }
      }

      if len(ins) != met.Type.NumIn() {
         errmsg = fmt.Sprintf("Operation call with wrong input argument count: received:%d, expected:%d",len(ins), met.Type.NumIn())
         Goose.OpHandle.Logf(1,errmsg)
         return Response {
            Status:            http.StatusBadRequest,
            Body:              errmsg,
         }
      }

      retData := met.Func.Call(ins)

      Goose.OpHandle.Logf(5,"retData: %#v",retData)
      return retData[0].Interface().(Response)
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
            Goose.ParseFields.Logf(1,"%s (with query) at method %s", ErrorWrongParameterCount, methodName)
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
            Goose.ParseFields.Logf(1,"%s: %s",ErrorInvalidParameterType,lstFld)
            Goose.ParseFields.Logf(1,"SwaggerParameter: %#v",SwaggerParameter)
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

   Goose.ParseFields.Logf(6,"parm: %s, count: %d, met.in:%d",methodName, parmcount,method.Type.NumIn()) // 3, 4
   return
}

func validateProtoList(proto []string) (bool, bool, error) {
   var plain, crypt, http, ws bool
   var p string

   for _, p = range proto {
      switch p {
         case "http" :
            plain = true
            http  = true
         case "https" :
            crypt = true
            http  = true
         case "ws" :
            plain = true
            ws    = true
         case "wss" :
            crypt = true
            ws    = true
         default:
            return false, false, ErrorInvalidProtocol
      }
   }

   return http && ws, plain && crypt, nil
}

func buildSubHandles(OpType reflect.Type, produces []string, consumes []string) (*strtree.Node, map[string]*SwaggerOperationT, error) {
   var Op reflect.Type
   var fld reflect.StructField
   var method reflect.Method
   var methodName string
   var callByRef, ok bool
   var i, j int
   var t strtree.Node
   var spec map[string]*SwaggerOperationT
   var operation *SwaggerOperationT
   var err error
   var inTag string

Goose.New.Logf(0,"OpType: %#v",OpType)

   if OpType.Kind() != reflect.Struct {
      Goose.New.Logf(1,"Error parsing object %s.%s to handle websocket: %s",OpType.PkgPath(),OpType.Name(), ErrorWrongHandlerKind)
      return nil, nil, ErrorWrongHandlerKind
   }

   spec = map[string]*SwaggerOperationT{}

   for i=0; i<OpType.NumField(); i++ {
      fld = OpType.Field(i)
      if strings.ToUpper(fld.Name) == "ON" { // ON is a reserved name: it is used to register event callbacks
         Goose.New.Logf(1,"Reserved word conflict: field %s ignored",fld.Name)
         continue
      }

      inTag = fld.Tag.Get("in")
      if inTag == "" {
         continue // not a websocket operation because input parameters not defined ...
      }

      if fld.Name[:1] == strings.ToLower(fld.Name[:1]) {
         callByRef = false
         methodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]
         if method, ok = OpType.MethodByName(methodName); !ok {
            Op = reflect.PtrTo(OpType)
            if method, ok = Op.MethodByName(methodName); !ok {
               Goose.New.Logf(5,"|wsmethods|=%d",Op.NumMethod())
               Goose.New.Logf(5,"wstype=%s.%s",Op.PkgPath(),Op.Name())
               for j=0; j<Op.NumMethod(); j++ {
                  mt := Op.Method(j)
                  Goose.New.Logf(5,"%d: %s",j,mt.Name)
               }

               Goose.New.Logf(1,"Method not found: %s, Data: %#v",methodName,Op)
               continue // Unexported field but not a websocket operation ...
            } else {
               Goose.New.Logf(3,"Pointer method %s found, type of operation: %s",methodName,OpType)
               callByRef = true
            }
         }
      } else {
         continue // not a websocket operation (first identifier letter must be lowercase in struct and uppercase in method name) ...
      }

      operation, err = GetWebSocketSpec(fld, methodName, method)
      if err != nil {
         return nil, nil, err
      }

      t.Set(
         methodName,
         &WSocketOperation{
            ParmNames: strings.Split(inTag,","),
            CallByRef: callByRef,
            Method:    method,
         })

      spec[methodName] = operation
   }

   return &t, spec, nil
}




func getResponses(fld reflect.StructField, typ reflect.Type) (map[string]SwaggerResponseT, reflect.Type, error) {
   var responseOk              string
   var responseCreated         string
   var responseAccepted        string
   var responses    map[string]SwaggerResponseT
   var fldType                 reflect.Type
   var SwaggerParameter       *SwaggerParameterT
   var err                     error
   var doc                     string
   var responseFunc            reflect.Method
   var ok                      bool
   var responseList map[string]ResponseT

   responses = map[string]SwaggerResponseT{}

   responseOk = fld.Tag.Get("ok")
   responseCreated = fld.Tag.Get("created")
   responseAccepted = fld.Tag.Get("accepted")
   if responseOk != "" || responseCreated != "" || responseAccepted != "" {
      if responseOk != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
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
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusOK)].Schema))
         }
      }
      if responseCreated != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
         }

         if SwaggerParameter == nil {
            responses[fmt.Sprintf("%d",http.StatusCreated)] = SwaggerResponseT{
               Description: responseCreated,
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

            responses[fmt.Sprintf("%d",http.StatusCreated)] = SwaggerResponseT{
               Description: responseCreated,
               Schema:      SwaggerParameter.Schema,
            }
            //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
            //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusCreated)].Schema))
         }
      }
      if responseAccepted != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
         }

         if SwaggerParameter == nil {
            responses[fmt.Sprintf("%d",http.StatusAccepted)] = SwaggerResponseT{
               Description: responseAccepted,
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

            responses[fmt.Sprintf("%d",http.StatusAccepted)] = SwaggerResponseT{
               Description: responseAccepted,
               Schema:      SwaggerParameter.Schema,
            }
            //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
            //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusAccepted)].Schema))
         }
      }
   } else if responseFunc, ok = typ.MethodByName(fld.Name + "Responses"); ok {
      responseList = responseFunc.Func.Call([]reflect.Value{})[0].Interface().(map[string]ResponseT)
      for responseStatus, responseSchema := range responseList {
         SwaggerParameter, err = GetSwaggerType(reflect.TypeOf(responseSchema.TypeReturned))
         if err != nil {
            return nil, nil, err
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

   return responses, fldType, nil
}


func New(svcs ...EndPointHandler) (*Service, error) {
   var resp                *Service
   var svc                  EndPointHandler
   var svcElem              EndPointHandler
//   var svcRecv              reflect.Value
   var consumes             string
   var svcConsumes          string
   var produces             string
   var svcProduces          string
   var allowGzip            string
   var enableCORS           string
   var proto              []string
   var svcProto           []string
   var svcRoot              string
   var i, j, k              int
   var typ                  reflect.Type
   var typPtr               reflect.Type
   var pt                 []reflect.Type
   var fld                  reflect.StructField
   var method               reflect.Method
   var parmcount            int
   var httpmethod, path     string
   var methodName           string
   var tk                   string
   var ok                   bool
   var re                   string
   var reAllOps             string
   var reComp              *regexp.Regexp
   var c                    rune
   var err                  error
   var stmp                 string
   var SwaggerParameter    *SwaggerParameterT
   var swaggerParameters  []SwaggerParameterT
   var swaggerInfo          SwaggerInfoT
   var swaggerLicense       SwaggerLicenseT
   var swaggerContact       SwaggerContactT
   var globalDataCount      int
   var responses map[string]SwaggerResponseT
   var fldType              reflect.Type
   var doc                  string
   var description         *string
   var headers            []string
   var query              []string
   var optIndex  map[string]int
   var HdrNew, HdrOld       string
   var MatchedOpsIndex      int
   var postFields         []string
   var postField            string
   var postdata             string
//   var accesstype           uint8
   var parmnames          []string
   var callByRef            bool
   var mixedProto           bool
   var mixedEnc             bool
   var pos                  int

   for _, svc = range svcs {

      Goose.New.Logf(6,"Elem: %#v (Kind: %#v)", reflect.ValueOf(svc), reflect.ValueOf(svc).Kind())
      if reflect.ValueOf(svc).Kind() == reflect.Ptr {
         Goose.New.Logf(6,"Elem: %#v", reflect.ValueOf(svc).Elem())
         svcElem = reflect.ValueOf(svc).Elem().Interface().(EndPointHandler)
         Goose.New.Logf(6,"Elem type: %s, ptr type: %s", reflect.TypeOf(svcElem), reflect.TypeOf(svc))
      } else {
         svcElem = svc
         Goose.New.Logf(6,"Elem type: %s", reflect.TypeOf(svcElem))
      }

      // The first endpoint handler MUST have a config defined, otherwise we'll ignore endpoint handlers until we find one which provides a configuration
      if resp == nil {
         resp, err = initSvc(svcElem)
         if err != nil {
            return nil, err
         }
         if resp == nil {
            continue // If we still don't have a config defined and the endpoint handler has no config defined it WILL BE IGNORED!!!
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
                  enableCORS  = typ.Field(i).Tag.Get("enableCORS")
                  if typ.Field(i).Tag.Get("proto") != "" {
                     svcProto    = strings.Split(strings.ToLower(strings.Trim(typ.Field(i).Tag.Get("proto")," ")),",")
                     _, _, err = validateProtoList(svcProto)
                     if err != nil {
                        Goose.New.Logf(1,"Error validating global protocol list: %s", err)
                        return nil, err
                     }
                  } else {
                     svcProto    = []string{"https"}
                  }

                  Goose.New.Logf(3,"Access tag: %s", typ.Field(i).Tag.Get("access"))

                  if typ.Field(i).Tag.Get("access") != "" {
                     switch strings.ToLower(strings.Trim(typ.Field(i).Tag.Get("access")," ")) {
                        case "none":
                           resp.Access = AccessNone
                        case "auth":
                           resp.Access = AccessAuth
                        case "authinfo":
                           resp.Access = AccessAuthInfo
                        case "verifyauth":
                           resp.Access = AccessVerifyAuth
                        case "verifyauthinfo":
                           resp.Access = AccessVerifyAuthInfo
                     }
                     Goose.New.Logf(3,"Custom access type: %d", resp.Access)
                  } else {
                     resp.Access = AccessAuthInfo
                     Goose.New.Logf(3,"Default access type: %d", resp.Access)
                  }
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
            Goose.New.Logf(1,"Err: %s",ErrorNoRoot)
            return nil, ErrorNoRoot
         }

         hostport := strings.Split(resp.Config.ListenAddress(),":")
         if hostport[0] == "" {
            hostport[0] = resp.Authorizer.GetDNSNames()[0]
         }

         resp.Swagger = &SwaggerT{
            Version:     "2.0",
            Info:        swaggerInfo,
            Host:        strings.Join(hostport,":"),
            BasePath:    "/" + svcRoot[:len(svcRoot)-1],
            Schemes:     svcProto,
            Consumes:    []string{svcConsumes},
            Produces:    []string{svcProduces},
            Paths:       map[string]SwaggerPathT{},
            Definitions: map[string]SwaggerSchemaT{},
         }

         resp.Proto = svcProto

         Goose.New.Logf(6,"enableCORS: [%s]",enableCORS)
         if enableCORS != "" {
            resp.EnableCORS, err = strconv.ParseBool(enableCORS)
            Goose.New.Logf(6,"resp.EnableCORS: %#v",resp.EnableCORS)
            if err != nil {
               Goose.New.Logf(1,"Err: %s",ErrorServiceSyntax)
               return nil, ErrorServiceSyntax
            }
         }

         Goose.New.Logf(6,"allowGzip: [%s]",allowGzip)
         if allowGzip != "" {
            resp.AllowGzip, err = strconv.ParseBool(allowGzip)
            Goose.New.Logf(6,"resp.AllowGzip: %#v",resp.AllowGzip)
            if err != nil {
               Goose.New.Logf(1,"Err: %s",ErrorServiceSyntax)
               return nil, ErrorServiceSyntax
            }
         }
      }

      for i=0; i<typ.NumField(); i++ {
         fld = typ.Field(i)
         httpmethod = fld.Tag.Get("method")
         if httpmethod != "" {
            methodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]

            callByRef = false
            if method, ok = typ.MethodByName(methodName); !ok {
               if method, ok = typPtr.MethodByName(methodName); !ok {
                  Goose.New.Logf(5,"|methods|=%d",typ.NumMethod())
                  Goose.New.Logf(5,"type=%s.%s",typ.PkgPath(),typ.Name())
                  for j=0; j<typ.NumMethod(); j++ {
                     mt := typ.Method(j)
                     Goose.New.Logf(5,"%d: %s",j,mt.Name)
                  }

                  Goose.New.Logf(1,"Method not found: %s, Data: %#v",methodName,typ)
                  return nil, errors.New(fmt.Sprintf("Method not found: %s",methodName))
               } else {
                  Goose.New.Logf(1,"Pointer method found, type of svcElem: %s",reflect.TypeOf(svcElem))
                  callByRef = true
                  Goose.New.Logf(5,"Pointer method found: %s",methodName)
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
            parmnames = []string{}

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
                        Goose.New.Logf(1,"%s: %s",tk[1:len(tk)-1])
                        return nil, ErrorInvalidParameterType
                     }

                     doc = fld.Tag.Get(tk[1:len(tk)-1])
                     if doc != "" {
                        description    = new(string)
                        (*description) = doc
                     } else {
                        description = SwaggerParameter.Schema.Description
                        if description == nil {
                           description = new(string)
                        }
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
                     parmnames = append(parmnames,tk[1:len(tk)-1])
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

            Goose.New.Logf(4,"Service " + strings.ToUpper(httpmethod) + ":/" + svcRoot + path + ", RE=" + re )

            query, parmcount, _, swaggerParameters, err = ParseFieldList("query", parmcount, fld, method, methodName, swaggerParameters)
            if err != nil {
               return nil, err
            }

            headers, parmcount, _, swaggerParameters, err = ParseFieldList("header", parmcount, fld, method, methodName, swaggerParameters)
            if err != nil {
               return nil, err
            }

            parmnames = append(parmnames, query...)
            parmnames = append(parmnames, headers...)

            postdata = fld.Tag.Get("postdata")
            if postdata != "" {
               // Body fields definitions
               postFields = strings.Split(postdata,",")
               pt = make([]reflect.Type,len(postFields))
               for k, postField = range postFields {
                  parmcount++
                  pt[k] = method.Type.In(parmcount)
                  SwaggerParameter, err = GetSwaggerType(pt[k])
                  if err != nil {
                     return nil, err
                  }

                  if SwaggerParameter == nil {
                     return nil, ErrorInvalidNilParam
                  }

                  doc = fld.Tag.Get(postField)
                  if doc != "" {
                     SwaggerParameter.Schema.Description    = new(string)
                     (*SwaggerParameter.Schema.Description) = doc
                  }

                  parmnames                 = append(parmnames, postField)
                  SwaggerParameter.Name     = postField
                  SwaggerParameter.In       = "body"
                  SwaggerParameter.Required = true

                  swaggerParameters = append(swaggerParameters,*SwaggerParameter)
               }

/*
               if resp.Access == AccessAuthInfo || resp.Access == AccessVerifyAuthInfo {
                  parmcount++
               }

               if (parmcount+len(postFields)+1) != method.Type.NumIn() {
                  return nil, errors.New("Wrong parameter count (with post) at method " + methodName)
               }
*/
            } else {
               pt = nil
            }

            if resp.Access == AccessAuthInfo || resp.Access == AccessVerifyAuthInfo {
               if (parmcount+1) != method.Type.NumIn() {
                  parmcount++
                  if (parmcount+1) != method.Type.NumIn() {
                     return nil, errors.New("Wrong parameter (with info) count at method " + methodName)
                  }
               }
            } else {
               if (parmcount+1) != method.Type.NumIn() {
                  return nil, errors.New("Wrong parameter count at method " + methodName)
               }
            }

            Goose.New.Logf(5,"Registering: %s",re)
            consumes = fld.Tag.Get("consumes")
            Goose.New.Logf(1,"op:%s consumes: %s tag:%#v",methodName,consumes,fld.Tag)
            if consumes == "" {
               consumes = svcConsumes
            }

            produces = fld.Tag.Get("produces")
            if produces == "" {
               produces = svcProduces
            }

            if fld.Tag.Get("proto") != "" {
               proto = strings.Split(strings.ToLower(strings.Trim(typ.Field(i).Tag.Get("proto")," ")),",")
               mixedProto, mixedEnc, err = validateProtoList(proto)
               if err != nil {
                  Goose.New.Logf(1,"Error validating global protocol list: %s", err)
                  return nil, err
               }

               if mixedProto {
                  Goose.New.Logf(1,"Error validating global protocol list: %s", ErrorMixedProtocol)
                  return nil, ErrorMixedProtocol
               }

               if mixedEnc {
                  Goose.New.Logf(2,"Warning validating global protocol list on operation %s: both plain and encrypted protocols selected", methodName)
               }
            } else {
               mixedProto, _, _ = validateProtoList(svcProto)
               if mixedProto {
                  Goose.New.Logf(2,"Warning using just the first global protocol on operation %s to prevent mixed use of http/https with ws/wss", methodName)
                  proto = []string{svcProto[0]}
               } else {
                  proto = svcProto
               }
            }

            responses, fldType, err = getResponses(fld, typ)
            if err != nil {
               return nil, err
            }

            resp.Swagger.Paths[path][strings.ToLower(httpmethod)] = &SwaggerOperationT{
               Schemes:     proto,
               OperationId: methodName,
               Parameters:  swaggerParameters,
               Responses:   responses,
               Consumes:  []string{consumes},
               Produces:  []string{produces},
            }

            Goose.New.Logf(0,"Registering marshalers: %s, %s",consumes,produces)

            resp.MatchedOps[MatchedOpsIndex] = len(resp.Svc)
            reComp                           = regexp.MustCompile(re)
            MatchedOpsIndex                 += reComp.NumSubexp() + 1

/*
            switch strings.ToLower(fld.Tag.Get("access")) {
               case "none":     accesstype = AccessNone
               case "auth":     accesstype = AccessAuth
               case "authinfo": accesstype = AccessAuthInfo
               default:         accesstype = AccessAuth
            }
*/

            pos = len(resp.Svc)

            resp.Svc = append(resp.Svc,UrlNode{
               Proto:     proto,
               Path:      path,
               consumes:  consumes,
               produces:  produces,
               Headers:   headers,
               Query:     query,
               Body:      postFields,
               ParmNames: parmnames,
               Handle:    buildHandle(reflect.ValueOf(svc),callByRef,method,pt,resp.Access,proto[0][0] == 'w'),
//               Access:    resp.Access,
            })

            if proto[0][0] == 'w' { // if this is a web service handler
               resp.Svc[pos].WSocketOperations, resp.Swagger.Paths[path][strings.ToLower(httpmethod)].XWebSocket, err = buildSubHandles(fldType,strings.Split(produces,","),strings.Split(consumes,","))
               if err != nil {
                  return nil, err
               }
            }

            reAllOps += "|(" + re + ")"
            Goose.New.Logf(6,"Partial Matcher for %s is %s",path,reAllOps)

            if resp.EnableCORS {
               index := len(resp.Svc)
               if optIndex == nil {
                  optIndex = map[string]int{path:index}
               } else if index, ok = optIndex[path]; ok {
                  for _, HdrNew = range headers {
                     for _, HdrOld = range resp.Svc[index].Headers {
                        if HdrOld == HdrNew {
                           break
                        }
                     }
                     if HdrOld != HdrNew {
                        resp.Svc[index].Headers = append(resp.Svc[index].Headers, HdrNew)
                     }
                  }
                  continue
               } else {
                  optIndex[path] = len(resp.Svc)
               }

               re = "^OPTIONS" + re[len(httpmethod)+1:]
               resp.MatchedOps[MatchedOpsIndex] = len(resp.Svc)
               reComp                           = regexp.MustCompile(re)
               MatchedOpsIndex                 += reComp.NumSubexp() + 1

               resp.Svc = append(resp.Svc,UrlNode{
                  Path: path,
                  Headers: headers,
               })
               reAllOps += "|(" + re + ")"
               Goose.New.Logf(6,"Partial Matcher with options for %s is %s",path,reAllOps)
            }
         }
      }
   }

   Goose.New.Logf(6,"Operations matcher: %s\n",reAllOps[1:])
   Goose.New.Logf(6,"Operations %#v\n",resp.Svc)
   resp.Matcher = regexp.MustCompile(reAllOps[1:]) // Cutting the leading '|'
   return resp, nil
}


//Init a webserver and wait for http requests.
func (svc *Service) ListenAndServeTLS() error {
   var err      error
   var wg       sync.WaitGroup
   var crypls   net.Listener
   var hn       string
   var useHttps bool
   var useHttp  bool
   var unode    UrlNode
   var p        string
   var auth     AuthT
   var tc      *tls.Config

   if svc.Config.ListenAddress()[0] == ':' {
      hn, err = os.Hostname()
      if err!=nil {
         Goose.InitServe.Logf(1,"Error checking hostname: %s", err)
         return err
      }
   }

FindEncLoop:
   for _, unode = range svc.Svc {
      for _, p = range unode.Proto {
         if p == "https" || p == "wss" {
            useHttps = true
         }

         if p == "http" || p == "ws" {
            useHttp = true
         }

         if useHttp && useHttps {
            break FindEncLoop
         }
      }
   }


   if useHttps {
      auth = svc.Authorizer

      wg.Add(2)

      go func() {
         var errcrl error
         defer wg.Done()
         if svc.Config.CRLListenAddress() != "" {
            errcrl = auth.StartCRLServer(hn + svc.Config.CRLListenAddress(), svc.CRLListener)
            if errcrl != nil {
               Goose.InitServe.Logf(5,"Error starting CRL listener: %s", errcrl)
               err = errcrl
            } else {
               Goose.InitServe.Logf(5,"CRL Listen started listening")
            }
         }
      }()

      go func() {
         var errsrv error
         defer wg.Done()

//         tc, errsrv = auth.GetTLSConfig(svc.AuthRequired)
         tc, errsrv = auth.GetTLSConfig(svc.Access)
         if errsrv != nil {
            Goose.InitServe.Logf(1,"Failed setting up listener encryption parameters: %s",err)
            err = errsrv
            return
         }

         srv := &http.Server{
            Addr: hn + svc.Config.ListenAddress(),
            Handler: svc,
            TLSConfig: tc,
         }

         crypls = tls.NewListener(svc.Listener,srv.TLSConfig)

         errsrv = srv.Serve(crypls)
         if errsrv != nil {
            Goose.InitServe.Logf(5,"Error starting service listener: %s", errsrv)
            err = errsrv
         } else {
            Goose.InitServe.Logf(5,"Service started listening")
         }
      }()

      wg.Wait()

   } else if useHttp { // TODO: provide 2 listeners and change this XOR to an OR
      srv := &http.Server{
         Addr: hn + svc.Config.ListenAddress(),
         Handler: svc,
      }
      err = srv.Serve(svc.Listener)
      if err != nil {
         Goose.InitServe.Logf(5,"Error starting service listener: %s", err)
      } else {
         Goose.InitServe.Logf(5,"Service started listening")
      }
   }

   Goose.InitServe.Logf(1,"Ending listening")

   return err
}


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
      if useWebSocket && resp.Status < http.StatusMultipleChoices {
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

            wg.Add(1)
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
                              eh.stat = true
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
                              eh.stat = false
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
               close(ev.ch)
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


func wsEventHandle(ws *websocket.Conn, codec websocket.Codec, obj interface{}, wg sync.WaitGroup, evHandlers map[string]*WSEventTrigger) {
   var i int
   var ev string
   var t reflect.Type
   var v reflect.Value

   v = reflect.ValueOf(obj)
   if v.Kind() == reflect.Ptr {
      v = v.Elem()
   }

   t = v.Type()
   Goose.Serve.Logf(1,"Looking for event channel for type %#v",t)
   for i=0; i<t.NumField(); i++ {
      Goose.Serve.Logf(1,"Looking for event channel on field %s",t.Field(i).Name)
      if t.Field(i).Type.AssignableTo(typeWSEventTrigger) {
         ev = t.Field(i).Tag.Get("event")
         if ev != "" {
            Goose.Serve.Logf(1,"Event setup for %s",ev)
            evHandlers[ev] = v.Field(i).Interface().(*WSEventTrigger)
            go func(c reflect.Value, name string) {
               var ok bool
               var v reflect.Value

               for {
                  Goose.Serve.Logf(1,"Event comm loop will wait on channel")
                  v, ok = c.Recv()
                  Goose.Serve.Logf(1,"Event comm loop test %#v received %#v",ok, v.Interface())
                  if !ok {
                     // End event triggering if wsStopEvent has closed the channel
                     wg.Done()
                     return
                  }

                  Goose.Serve.Logf(1,"Event comm Recv: %#v",v.Interface())
                  // callID , status, response
                  codec.Send(ws, []interface{}{name, StatusTrigEvent, v.Interface()})
               }

            }(reflect.ValueOf(evHandlers[ev].ch), ev)
         }
      }
   }
   Goose.Serve.Logf(1,"Event channels for type %#v all configured",t.Name)
}


func (pa PublicAccessT) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   return http.StatusOK, nil, nil
}

//func (pa PublicAccessT)  GetTLSConfig(AuthRequired bool) (*tls.Config, error) {
func (pa PublicAccessT)  GetTLSConfig(Access uint8) (*tls.Config, error) {
   return nil, nil
/*
   var err        error
   var tlsConfig *tls.Config

   Goose.Auth.Logf(6,"authtype: %#v",tls.NoClientCert)

   tlsConfig = &tls.Config{
      ClientAuth: tls.NoClientCert,
//      InsecureSkipVerify: true,
      Certificates: []tls.Certificate{ck.ServerX509KeyPair},
   }

   Goose.Auth.Logf(5,"X509KeyPair used: %#v",tlsConfig.Certificates[0])
   tlsConfig.BuildNameToCertificate()

   return tlsConfig, nil
*/
}

func (pa PublicAccessT) StartCRLServer(listenAddress string, listener *StoppableListener) error {
   return nil

/*
   srvcrl := &http.Server{
      Addr: listenAddress,
      Handler: ck.CertKit(),
   }

   Goose.Auth.Logf(5,"CRL Listen Address: %s",listenAddress)
   return srvcrl.Serve(listener)
*/

//   Goose.InitServe.Logf(5,"CRL Listen is serving")
//   err = http.ListenAndServe(svc.CRLListenAddress,svc.Auth)
//   if err != nil {
//      Goose.InitServe.Fatalf(1,"Error serving CRL: %s",err)
//   }

}


func (pa PublicAccessT) GetDNSNames() []string {
   var hn  string
   var err error
   hn, err = os.Hostname()
   if err != nil {
      return []string{"localhost"}
   }
   return []string{hn}
}

func (pa PublicAccessT) GetServerCert() *x509.Certificate {
   return nil
}

func (pa PublicAccessT) GetServerKey() *rsa.PrivateKey {
   return nil
}

func (pa PublicAccessT) GetCACert() *x509.Certificate {
   return nil
}

func (pa PublicAccessT) GetCAKey() *rsa.PrivateKey {
   return nil
}

func (pa PublicAccessT) GetServerX509KeyPair() tls.Certificate {
   return tls.Certificate{}
}

func (pa PublicAccessT) GetCertPool() *x509.CertPool {
   return nil
}

func (pa PublicAccessT) ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadCertificate(fname string) (*x509.Certificate, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadDecryptRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadDecryptRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) Setup(udata map[string]interface{}) error {
   return nil
}

func (pa PublicAccessT) LoadUserData() error {
   return nil
}

func (pa PublicAccessT) Trust(id string) error {
   return nil
}

func (pa PublicAccessT) GetPending() (map[string]interface{}, error) {
   return map[string]interface{}{}, nil
}

func (pa PublicAccessT) GetTrusted() (map[string]interface{}, error) {
   return map[string]interface{}{}, nil
}

func (pa PublicAccessT) Reject(id string) error {
   return nil
}

func (pa PublicAccessT) Drop(id string) error {
   return nil
}

func (pa PublicAccessT) Delete(tree, id string) error {
   return nil
}

