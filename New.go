package stonelizard

import (
   "os"
   "fmt"
   "regexp"
   "errors"
   "strings"
   "reflect"
   "strconv"
)

func New(svcs ...EndPointHandler) (*Service, error) {
   var resp                    *Service
   var svc                      EndPointHandler
   var svcElem                  EndPointHandler
//   var svcRecv                  reflect.Value
   var consumes                 string
   var svcConsumes              string
   var produces                 string
   var svcProduces              string
   var allowGzip                string
   var enableCORS               string
   var proto                  []string
   var svcProto               []string
   var svcRoot                  string
   var i, j, k                  int
   var typ                      reflect.Type
   var typPtr                   reflect.Type
   var pt                     []reflect.Type
   var fld                      reflect.StructField
   var method                   reflect.Method
   var parmcount                int
   var httpmethod, path         string
   var methodName               string
   var tk                       string
   var ok                       bool
   var re                       string
   var reAllOps                 string
   var reComp                  *regexp.Regexp
   var c                        rune
   var err                      error
   var stmp                     string
   var SwaggerParameter        *SwaggerParameterT
   var swaggerParameters      []SwaggerParameterT
   var swaggerInfo              SwaggerInfoT
   var swaggerLicense           SwaggerLicenseT
   var swaggerContact           SwaggerContactT
   var globalDataCount          int
   var responses     map[string]SwaggerResponseT
   var fldType                  reflect.Type
   var doc                      string
   var description             *string
   var headers                []string
   var query                  []string
   var optIndex      map[string]int
   var HdrNew, HdrOld           string
   var MatchedOpsIndex          int
   var postFields             []string
   var postField                string
   var postdata                 string
//   var accesstype               uint8
   var parmnames              []string
   var callByRef                bool
   var mixedProto               bool
   var mixedEnc                 bool
   var pos                      int
//   var webSocketOps map[string]*SwaggerWSOperationT
//   var webSocketSpec           *SwaggerWSOperationT
   var static                   string
   var exported                 string
   var exportedValue          []reflect.Value
   var this                     reflect.Value
   var num                      int
   var mod, out, outvar         string
   var extAuth                  ExtAuthT
   var swpath                   string
   var tags                 [][]string
   var tag                    []string
   var propParm                *SwaggerParameterT
   var propField                reflect.StructField

   Goose.New.Logf(6,"Initializing services: %#v", svcs)

   for _, svc = range svcs {
      Goose.New.Logf(0,"-----------")
//      Goose.New.Logf(6,"Elem: %#v (Kind: %#v)", reflect.ValueOf(svc), reflect.ValueOf(svc).Kind())
//      Goose.New.Logf(0,"-----------")
      if reflect.ValueOf(svc).Kind() == reflect.Ptr {
//         Goose.New.Logf(6,"Elem: %#v", reflect.ValueOf(svc).Elem())
         svcElem = reflect.ValueOf(svc).Elem().Interface().(EndPointHandler)
//         Goose.New.Logf(6,"Elem type: %s, ptr type: %s", reflect.TypeOf(svcElem), reflect.TypeOf(svc))
      } else {
         svcElem = svc
//         Goose.New.Logf(6,"Elem type: %s", reflect.TypeOf(svcElem))
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

         resp.SecureStatic = map[string]string{}
         resp.PlainStatic = map[string]string{}
      }

      typ = reflect.ValueOf(svcElem).Type()
      if typ.Kind()==reflect.Ptr {
         typPtr     = typ
         typ        = typ.Elem()
      } else {
         typPtr = reflect.PtrTo(typ)
      }

      if resp.Swagger == nil {
//         for i=0; (i<typ.NumField()) && (globalDataCount<4); i++ {
         for i=0; i < typ.NumField(); i++ {
            fld = typ.Field(i)
            if svcRoot == "" {
               svcRoot = fld.Tag.Get("root")
               if svcRoot != "" {
                  svcConsumes = fld.Tag.Get("consumes")
                  svcProduces = fld.Tag.Get("produces")
                  allowGzip   = fld.Tag.Get("allowGzip")
                  enableCORS  = fld.Tag.Get("enableCORS")
                  swpath      = fld.Tag.Get("swagger")

                  if swpath == "root" {
                     resp.SwaggerPath = svcRoot
                  }

                  if fld.Tag.Get("proto") != "" {
                     svcProto    = strings.Split(strings.ToLower(strings.Trim(fld.Tag.Get("proto")," ")),",")
                     _, _, err = validateProtoList(svcProto)
                     if err != nil {
                        Goose.New.Logf(1,"Error validating global protocol list: %s", err)
                        return nil, err
                     }
                  } else {
                     svcProto    = []string{"https"}
                  }

                  Goose.New.Logf(3,"Access tag: %s", fld.Tag.Get("access"))

                  if fld.Tag.Get("access") != "" {
                     switch strings.ToLower(strings.Trim(fld.Tag.Get("access")," ")) {
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
                        case "info":
                           resp.Access = AccessInfo
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
               stmp = fld.Tag.Get("title")
               if stmp != "" {
                  swaggerInfo.Title          = stmp
                  swaggerInfo.Description    = fld.Tag.Get("description")
                  swaggerInfo.TermsOfService = fld.Tag.Get("tos")
                  swaggerInfo.Version        = fld.Tag.Get("version")
                  globalDataCount++
               }
            }
            if swaggerContact.Name == "" {
               stmp = fld.Tag.Get("contact")
               if stmp != "" {
                  swaggerContact.Name  = stmp
                  swaggerContact.Url   = fld.Tag.Get("url")
                  swaggerContact.Email = fld.Tag.Get("email")
                  globalDataCount++
               }
            }
            if swaggerLicense.Name == "" {
               stmp = fld.Tag.Get("license")
               if stmp != "" {
                  swaggerLicense.Name  = stmp
                  swaggerLicense.Url   = fld.Tag.Get("url")
                  globalDataCount++
               }
            }

            Goose.New.Logf(1,"Is %s a module?", fld.Name)
            if fld.Type == ModType {
               Goose.New.Logf(1,"Module %s found", fld.Name)
               if swaggerInfo.XModules == nil {
                  Goose.New.Logf(1,"Module %s Init", fld.Name)
                  swaggerInfo.XModules = map[string]map[string]*SwaggerSchemaT{}
               }

               swaggerInfo.XModules[fld.Name] = map[string]*SwaggerSchemaT{}

               tags = tagRE.FindAllStringSubmatch(string(fld.Tag),-1)
               for _, tag = range tags {
                  Goose.New.Logf(1,"Prop %s: %s", tag[1], tag[2])
                  propField, ok = typ.FieldByName(tag[2])
                  if !ok {
                     Goose.New.Logf(1,"Error: %s", ErrorUndefPropType)
                     return nil, ErrorUndefPropType
                  }

                  propParm, err = GetSwaggerType(propField.Type)
                  if err != nil {
                     Goose.New.Logf(1,"Error checking type of module: %s",err)
                     return nil, err
                  }
                  swaggerInfo.XModules[fld.Name][tag[1]] = propParm.Schema
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
            if resp.Authorizer == nil {
               hostport[0] = "0.0.0.0"
            } else {
               authdns := resp.Authorizer.GetDNSNames()
               if len(authdns) > 0 {
                  if len(strings.Split(authdns[0],".")) == 1 {
                     hostport[0] = resp.Authorizer.GetServerCert().IPAddresses[0].String()
                  } else {
                     hostport[0] = authdns[0]
                  }
               } else {
                  hostport[0] = "0.0.0.0"
               }
            }
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

                  Goose.New.Logf(5,"*|methods|=%d",typPtr.NumMethod())
                  Goose.New.Logf(5,"*type=%s.%s",typPtr.PkgPath(),typPtr.Name())
                  for j=0; j<typPtr.NumMethod(); j++ {
                     mt := typPtr.Method(j)
                     Goose.New.Logf(5,"%d: *%s",j,mt.Name)
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

//            re = "^" + strings.ToUpper(httpmethod) + ":/" + svcRoot
            re = svcRoot

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

//                     if (SwaggerParameter.Items != nil) || (SwaggerParameter.CollectionFormat!="") || (SwaggerParameter.Schema.Required != nil) {
                     if (SwaggerParameter.CollectionFormat!="") || (SwaggerParameter.Schema.Required != nil) {
                        Goose.New.Logf(0,"%s on %s: %s",ErrorInvalidParameterType,methodName,tk[1:len(tk)-1])
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

            num = method.Type.NumIn()
            if resp.Access == AccessAuthInfo || resp.Access == AccessVerifyAuthInfo || resp.Access == AccessInfo {
               if (parmcount+1) != num {
                  parmcount++
                  if (parmcount != (num-3)) || (num<2) || (method.Type.In(num-2).Kind()!=reflect.String) || (method.Type.In(num-1).Kind()!=reflect.String) {
                     return nil, errors.New(fmt.Sprintf("Wrong parameter (with info) count at method %s, got %d want %d",methodName,parmcount,num))
                  }
               }
            } else {
               if (parmcount+1) != num {
                  return nil, errors.New(fmt.Sprintf("Wrong parameter count at method %s, got %d want %d",methodName,parmcount,num))
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

            mod = fld.Tag.Get("mod")
            out = fld.Tag.Get("out")
            outvar = fld.Tag.Get("outvar")

            resp.Swagger.Paths[path][strings.ToLower(httpmethod)] = &SwaggerOperationT{
               Schemes:     proto,
               OperationId: methodName,
               Parameters:  swaggerParameters,
               Responses:   responses,
               Consumes:  []string{consumes},
               Produces:  []string{produces},
               XModule:     mod,
               XOutput:     out,
               XOutputVar:  outvar,
            }


/*
            for _, prt := range proto {
               if prt[0:2] == "ws" {
                  if fld.Tag.Get("in") == "" {
                     Goose.New.Logf(1,"Error %s getting operations of %s",ErrorMissingWebsocketInTagSyntax,methodName)
                     return nil, ErrorMissingWebsocketInTagSyntax
                  }

                  webSocketOps, err = getWebSocketOps(fld)
                  if err != nil {
                     return nil, err
                  }
//                  webSocketSpec, err = GetWebSocketSpec(fld, methodName, method)
//                  if err != nil {
//                     return nil, err
//                  }
               }
            }
*/

            for _, prt := range proto {
               rePrtMethod := strings.ToUpper("^" + prt + "\\+" + httpmethod + ":/")

               Goose.New.Logf(0,"Registering marshalers: %s, %s",consumes,produces)

               resp.MatchedOps[MatchedOpsIndex] = len(resp.Svc)
               reComp                           = regexp.MustCompile(rePrtMethod + re)
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
                  Proto:     []string{prt},
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

               if prt[0:2] == "ws" { // if this is a web service handler
                  resp.Svc[pos].WSocketOperations, resp.Swagger.Paths[path][strings.ToLower(httpmethod)].XWSOperations, err = buildSubHandles(fldType,strings.Split(produces,","),strings.Split(consumes,","))
                  if err != nil {
                     return nil, err
                  }

                  resp.Swagger.Paths[path][strings.ToLower(httpmethod)].XWSEvents, err = GetWebSocketEventSpec(fld, methodName, method)
                  if err != nil {
                     return nil, err
                  }
               }

               reAllOps += "|(" + rePrtMethod + re + ")"
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

                  reCORS := "^" + strings.ToUpper(prt) + "\\+OPTIONS:/" + re
                  resp.MatchedOps[MatchedOpsIndex] = len(resp.Svc)
                  reComp                           = regexp.MustCompile(reCORS)
                  MatchedOpsIndex                 += reComp.NumSubexp() + 1

                  resp.Svc = append(resp.Svc,UrlNode{
                     Path: path,
                     Headers: headers,
                  })
                  reAllOps += "|(" + reCORS + ")"
               }
               Goose.New.Logf(6,"Partial Matcher with options for %s is %s",path,reAllOps)
            }

         } else {
            Goose.New.Logf(2,"Checking %s for static handlers",fld.Name)
            static = fld.Tag.Get("path")
            if static == "" {
               continue
            }

            Goose.New.Logf(2,"Checking %s for static handlers has path %s", fld.Name, static)
            exported = fld.Tag.Get("export")
            if exported =="" {
               continue
            }

            Goose.New.Logf(2,"Checking %s for static handlers exports %s", fld.Name, exported)
            methodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]
            if method, ok = typ.MethodByName(methodName); !ok {
               Goose.New.Logf(1,"Checking %s for static handlers for method %s not found", fld.Name, methodName)
               continue
            }

            Goose.New.Logf(2,"Checking %s for static handlers has defined method", fld.Name)
            this = reflect.ValueOf(svc)
            if this.Kind() == reflect.Ptr {
               this = this.Elem()
            }

            exportedValue = method.Func.Call([]reflect.Value{this})
            if len(exportedValue) != 1 {
               Goose.New.Logf(2,"Checking %s for static handlers has exportedValue %#v", fld.Name, exportedValue)
               continue
            }

            switch e := exportedValue[0].Interface().(type) {
            case string:
               exported = strings.Replace(fmt.Sprintf("%s%s",e,exported),"/",string([]byte{os.PathSeparator}),-1)
               Goose.New.Logf(2,"Checking %s for static handlers has full exported  %s", fld.Name, exported)
            default:
               continue
            }

            proto = strings.Split(fld.Tag.Get("proto"),",")
            if len(proto) == 0 || proto[0]=="" {
               proto = svcProto
            }

            for _, p := range proto {
               if p == "https" {
                  resp.SecureStatic[static] = exported
               } else if p == "http" {
                  resp.PlainStatic[static] = exported
               }
            }
         }
      }
   }

//   Goose.New.Logf(6,"Operations matcher: %s\n",reAllOps[1:])
   Goose.New.Logf(6,"Operations %#v\n",resp.Svc)
   if len(reAllOps) > 0 {
      resp.Matcher = regexp.MustCompile(reAllOps[1:]) // Cutting the leading '|'
   } else {
      resp.Matcher = regexp.MustCompile(reAllOps)
   }

   if extAuth, ok = resp.Authorizer.(ExtAuthT); ok {
      resp.ch = make(chan ExtAuthorizeIn)
      go extAuth.StartExtAuthorizer(resp.ch)
   }


   return resp, nil
}
