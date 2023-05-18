package stonelizard

import (
   "strings"
   "reflect"
   "encoding/json"
)

// Creates an array []reflect.Value to be used as parameter list when calling the method that handles the operation invoked by the network client
// It checks the method signature and converts the parameters received from the client to the types expected by the method
// The conversion is done constructing a json string and then using json.Unmarshal to get the specific type (some improvement must be done here...)
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

   Goose.OpHandle.Logf(5,"parsing params %#v",parms)

   for i, iface = range parms {
      parmType = met.Type.In(i+1)
      Goose.OpHandle.Logf(6,"parsing parm %#v",iface)

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
         case []interface{}:
            parmTypeName = "[]interface{}"
            buf, err = json.Marshal(iface)
            if err != nil {
               Goose.OpHandle.Logf(1,"marshal error.1: %s",err)
               Goose.OpHandle.Logf(1,"Internal server error parsing.1 %s: %s",buf,err)
               return nil, err
            }

         case bool, map[string]interface{}:
            buf, err = json.Marshal(iface)
            if err != nil {
               Goose.OpHandle.Logf(1,"marshal error.1: %s",err)
               Goose.OpHandle.Logf(1,"Internal server error parsing.1 %s: %s",buf,err)
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
      Goose.OpHandle.Logf(4,"adding parm %s",buf)
      err = json.Unmarshal(buf,parm.Interface())
      if err != nil {
         Goose.OpHandle.Logf(1,"unmarshal error: %s",err)
         Goose.OpHandle.Logf(1,"Internal server error parsing [%s]: %s",buf,err)
         Goose.OpHandle.Logf(2,"parms %#v",parms)
         return nil, err
      }

      Goose.OpHandle.Logf(4,"added parm %#v",reflect.Indirect(parm).Interface())

      ins = append(ins,reflect.Indirect(parm))
      Goose.OpHandle.Logf(5,"ins: %d:%s",len(ins),ins)
   }

   return ins, nil
}
