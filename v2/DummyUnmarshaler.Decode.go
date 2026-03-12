package stonelizard


import (
   "io/ioutil"
)

// Fetches the next value from the HTTP POST vars
func (b *DummyUnmarshaler) Decode(v interface{}) error {
   var buf  []byte
   var err    error

   Goose.OpHandle.Logf(8,"DummyUnmarshaler: v=%v",v)

   buf, err = ioutil.ReadAll(b.r)
   if err != nil {
      Goose.OpHandle.Logf(1,"%s: %s", ErrorDecodeError, err)
      return err
   }

   switch target := v.(type) {
      case *string: *target = string(buf)
      case *[]byte: *target = buf
      default:
         Goose.OpHandle.Logf(1,"%s", WrongParameterType)
         return WrongParameterType
   }

   return nil
}


