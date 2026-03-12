package stonelizard

import (
   "reflect"
)

// Converts an interface{} containing a float64 number to a float or integer specified by 't'
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

