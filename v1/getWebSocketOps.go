package stonelizard

import (
   "strings"
   "reflect"
)

// Scans the websocket definition struct, identifies which fields defines websocket operations
// and returns its specifications in extended swagger format.
// Websocket operations are defined by:
// a) MUST be an unexported symbol (names starting with LOWERCASE letter)
// b) having the "in" tag defined
// c) Its type MUST not be WSEventTrigger (reserved for events)
// d) MUST exist an exported method whose name is equal to the field (except for the first letter which MUST be UPPERCASE)
func getWebSocketOps(field reflect.StructField) (map[string]*SwaggerWSOperationT, error) {
   var ops map[string]*SwaggerWSOperationT
   var err error
   var retv reflect.Type
   var fld reflect.StructField
   var WSMethodName string
   var WSMethod reflect.Method
   var ok bool
   var i int
   var webSocketSpec *SwaggerWSOperationT

   retv = field.Type

   for i=0; i<retv.NumField(); i++ {
      fld = retv.Field(i)

      // WSEventTrigger fields defines events, not operations (rule 'c')
      if fld.Type == typeWSEventTrigger {
         continue
      } else {
         // Checks the "in" tag and if it is unexported field (rules 'a' and 'b')
         if fld.Tag.Get("in") == "" || len(fld.Name)==0 || strings.ToLower(fld.Name[:1])!=fld.Name[:1] {
            continue
         }

         // Checks the correspondent method (rule 'd')
         WSMethodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]
         if WSMethod, ok = retv.MethodByName(WSMethodName); !ok {
            continue
         }

         // Get the swagger specification
         webSocketSpec, err = GetWebSocketSpec(fld, WSMethodName, WSMethod)
         if err != nil {
            return nil, err
         }

         // stores it in the mapping key is operation name, value is swagger operation
         ops[WSMethodName] = webSocketSpec
      }
   }

   return ops, nil
}


