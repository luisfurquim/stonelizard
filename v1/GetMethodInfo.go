package stonelizard

import (
   "strings"
   "reflect"
)


// Returns if a method name, if it is pointer referenced and its reflect definition
// It receives the unexported field from the service definition struct and searches for a
// method with same name but with an uppercase first letter
func GetMethodInfo(fld reflect.StructField, OpType reflect.Type) (bool, string, reflect.Method) {
   var i int
   var callByRef, ok bool
   var methodName string
   var method reflect.Method
   var Op reflect.Type

   callByRef = false
   methodName = strings.ToUpper(fld.Name[:1]) + fld.Name[1:]
   if method, ok = OpType.MethodByName(methodName); !ok {
      Op = reflect.PtrTo(OpType)
      if method, ok = Op.MethodByName(methodName); !ok {
         Goose.New.Logf(5,"|wsmethods|=%d",Op.NumMethod())
         Goose.New.Logf(5,"wstype=%s.%s",Op.PkgPath(),Op.Name())
         for i=0; i<Op.NumMethod(); i++ {
            mt := Op.Method(i)
            Goose.New.Logf(5,"%d: %s",i,mt.Name)
         }

         Goose.New.Logf(1,"Method not found: %s, Data: %#v",methodName,Op)
         return false, "", method // Unexported field but no correspondent method ...
      }

      Goose.New.Logf(3,"Pointer method %s found, type of operation: %s",methodName,OpType)
      callByRef = true
   }

   return callByRef, methodName, method
}
