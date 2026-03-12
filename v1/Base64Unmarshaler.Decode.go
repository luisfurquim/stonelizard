package stonelizard


import (
   "io/ioutil"
)

// Fetches the next value from the HTTP POST vars
func (b *Base64Unmarshaler) Decode(v interface{}) error {
   var buf  []byte
   var err    error

   buf, err = ioutil.ReadAll(b.r)
   Goose.OpHandle.Logf(1,"Base64Unmarshaler error checking: %s [%s]", err, buf)
   if err != nil {
      Goose.OpHandle.Logf(1,"%s: %s", ErrorDecodeError, err)
      return err
   }

   switch target := v.(type) {
      case *string:
         *target = string(buf)
      case *[]byte:
         *target = buf
      default:
         Goose.OpHandle.Logf(1,"%s", WrongParameterType)
         return WrongParameterType
   }

   return nil
}
