package stonelizard


import (
   "io"
   "reflect"
   "encoding/json"
   "mime/multipart"
)

// converts json encoded data to the type 'typ'
func toType(elem string, typ reflect.Type) (reflect.Value, error) {
   var val reflect.Value
   var err error

   val = reflect.New(typ)

   err = json.Unmarshal([]byte(elem),val.Interface())
   if err != nil {
      return reflect.Zero(typ), err
   }

   return val.Elem(), nil
}

// Extracts one field from the HTTP POST vars
func (mp *MultipartUnmarshaler) getField(fldName string, vtype reflect.Type) (interface{}, error) {
   var a     []string
   var f    []*multipart.FileHeader
   var ok      bool
   var res     reflect.Value
//   var sz      int
   var s       string
   var h      *multipart.FileHeader
   var tmpval  reflect.Value
   var err     error
   var fhtype  reflect.Type

//   sz = len(mp.form.Value[fldName]) + len(mp.form.File[fldName])
   fhtype = reflect.TypeOf(&multipart.FileHeader{})

   switch vtype.Kind() {
      case reflect.Array, reflect.Slice :
         res = reflect.MakeSlice(vtype, 0, 0)

         // Extracts data values
         if a, ok = mp.form.Value[fldName]; ok {
            if (vtype.Elem().Kind() == reflect.String) || (vtype.Elem().Kind() == reflect.Interface) {
               // No need of conversion, just add to the array
               for _, s = range a {
                  res = reflect.Append(res,reflect.ValueOf(s))
               }
            } else {
               // Conversion needed based on the kind of target storage provided
               for _, s = range a {
                  tmpval, err = toType(s,vtype.Elem())
                  if err != nil {
                     return nil, err
                  }
                  res = reflect.Append(res,tmpval)
               }
            }
         }

         // Extract files
         if f, ok = mp.form.File[fldName]; ok {
            if (vtype.Elem()==fhtype) || (vtype.Elem().Kind() == reflect.Interface) {
               for _, h = range f {
                  res = reflect.Append(res,reflect.ValueOf(h))
               }
            } else {
               return nil, ErrorInvalidParameterType
            }
         }

         return res.Interface(), nil
      default:
         // Extract a single variable from post form
         if a, ok = mp.form.Value[fldName]; ok {
            if (vtype.Kind() == reflect.String) || (vtype.Kind() == reflect.Interface) {
               return a[0], nil
            }

            res, err = toType(a[0],vtype.Elem())
            if err != nil {
               return nil, err
            }

            return res.Interface(), nil
         }

         // Extract a single file
         if f, ok = mp.form.File[fldName]; ok {
            if (vtype==fhtype) || (vtype.Kind() == reflect.Interface) {
               return f[0], nil
            }
            Goose.OpHandle.Logf(5,"Um arquivo vtype:%q", vtype)
            return nil, ErrorInvalidParameterType
         }
   }
   return nil, ErrorMissingRequiredPostBodyField
}

// Fetches the next value from the HTTP POST vars
func (mp *MultipartUnmarshaler) Decode(v interface{}) error {
   var val    reflect.Value
   var vtype  reflect.Type
   var tmp    interface{}
   var err    error

   Goose.OpHandle.Logf(6,"MultipartUnmarshaler: fields=%q v=%#v form=%q",mp.fields,v,mp.form)

   if mp.index >= len(mp.fields) {
      return io.EOF
   }

   val   = reflect.ValueOf(v).Elem()
//   val   = reflect.ValueOf(*v)
   vtype = val.Type()

   tmp, err = mp.getField(mp.fields[mp.index],vtype)
   if err != nil {
      return err
   }

   val.Set(reflect.ValueOf(tmp))

   mp.index++

   return nil
}


