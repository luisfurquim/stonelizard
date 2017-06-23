package stonelizard

import (
   "reflect"
)

// Checks if the value provided is the zero value of its type
func isZero(v reflect.Value) bool {
   return reflect.Zero(v.Type()) == v
}

