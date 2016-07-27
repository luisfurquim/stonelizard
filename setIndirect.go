package stonelizard

/*
import (
   "reflect"
   "mime/multipart"
)

func setIndirect(vptr interface{}, val interface{}) error {
   var typ reflect.Type

   switch val.(type) {
      case bool:       *(vptr.(*bool))       = val.(bool)
      case int:        *(vptr.(*bool))       = val.(bool)
      case int8:       *(vptr.(*int8))       = val.(int8)
      case int16:      *(vptr.(*int16))      = val.(int16)
      case int32:      *(vptr.(*int32))      = val.(int32)
      case int64:      *(vptr.(*int64))      = val.(int64)
      case uint:       *(vptr.(*uint))       = val.(uint)
      case uint8:      *(vptr.(*uint8))      = val.(uint8)
      case uint16:     *(vptr.(*uint16))     = val.(uint16)
      case uint32:     *(vptr.(*uint32))     = val.(uint32)
      case uint64:     *(vptr.(*uint64))     = val.(uint64)
      case uintptr:    *(vptr.(*uintptr))    = val.(uintptr)
      case float32:    *(vptr.(*float32))    = val.(float32)
      case float64:    *(vptr.(*float64))    = val.(float64)
      case complex64:  *(vptr.(*complex64))  = val.(complex64)
      case complex128: *(vptr.(*complex128)) = val.(complex128)
      case string:     *(vptr.(*string))     = val.(string)
      case []*multipart.FileHeader: *(vptr.(*[]*multipart.FileHeader)) = val.([]*multipart.FileHeader)
      default:
         typ = reflect.TypeOf(val)
         switch typ.Kind() {
            case reflect.Array, reflect.Slice:
            case reflect.Map:
            case reflect.Struct:
            default: return ErrorInvalidParameterType
         }
   }

   return nil
}

*/