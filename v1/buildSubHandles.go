package stonelizard

import (
   "strings"
   "reflect"
   "github.com/luisfurquim/strtree"
)

// Creates and stores the handlers and the swagger specifications for each websocket suboperation.
func buildSubHandles(OpType reflect.Type, produces []string, consumes []string) (*strtree.Node, map[string]*SwaggerWSOperationT, error) {
   var fld reflect.StructField
   var method reflect.Method
   var methodName string
   var callByRef, ok bool
   var i int
   var t strtree.Node
   var spec map[string]*SwaggerWSOperationT
   var operation *SwaggerWSOperationT
   var err error
   var inTag string
   var parmNames []string

   // OpType is the service struct field type that specifies a websocket start operation
   // It MUST BE a struct which implements the websocket's semantic
   //
   if OpType.Kind() != reflect.Struct {
      Goose.New.Logf(1,"Error parsing object %s.%s to handle websocket: %s",OpType.PkgPath(),OpType.Name(), ErrorWrongHandlerKind)
      return nil, nil, ErrorWrongHandlerKind
   }

   spec = map[string]*SwaggerWSOperationT{}

   for i=0; i<OpType.NumField(); i++ {
      fld = OpType.Field(i)
      // BIND/UNBIND are reserved names: they will be used to control if there is at least one event consumer
      // No data will be sent across the network if there is no one to consume the data
      // WARNING: THIS FEATURE IS NOT IMPLEMENTED YET, CURRENTLY WE ARE JUST ASSURING NO ONE IMPLEMENTS
      // OPERATIONS USING THESE NAMES TO AVOID FUTURE COLISIONS
      if strings.ToUpper(fld.Name) == "BIND" || strings.ToUpper(fld.Name) == "UNBIND" {
         Goose.New.Logf(1,"Reserved word conflict: field %s ignored",fld.Name)
         continue
      }

      // The 'in' tag defines the websocket operation parameters
      // The Lookup method demands newer GO compilers
      // Sorry about the requisite, but otherwise it would not be possible to define operations with no parameters...
      // WARNING: it means that if your operation has no parameters you still MUST define the 'in' tag like this `in:""`
      if inTag, ok = fld.Tag.Lookup("in"); !ok {
         continue // not a websocket operation because input parameters not defined ...
      }

      // The first identifier letter must be lowercase in struct and uppercase in method name ...
      if fld.Name[:1] != strings.ToLower(fld.Name[:1]) {
         continue // not a websocket operation (first identifier letter must be lowercase in struct and uppercase in method name) ...
      }

      // Determines:
      // a) if the method has a pointer receiver or not
      // b) the name of the method
      // c) the method in the reflect.Value format
      callByRef, methodName, method = GetMethodInfo(fld, OpType)
      if methodName == "" {
         continue // Unexported field but not a websocket operation ...
      }

      // Get the swagger definition for this websocket operation
      operation, err = GetWebSocketSpec(fld, methodName, method)
      if err != nil {
         return nil, nil, err
      }

      // Split returns an array with one empty string element if the original string is an empty string
      // We want an empty array if there is no parameters defined
      if inTag == "" {
         parmNames = []string{}
      } else {
         parmNames = strings.Split(inTag,",")
      }

      // Stores the websocket operation identified by the method name
      // Method names are organized in a tree of its runes to speed up the lookup
      t.Set(
         methodName,
         &WSocketOperation{
            ParmNames: parmNames,
            CallByRef: callByRef,
            Method:    method,
         })

      spec[methodName] = operation
   }

   return &t, spec, nil
}




