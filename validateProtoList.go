package stonelizard

// Checks if an array of protocols contains only valid protocols and verifies if it has mixed uses
func validateProtoList(proto []string) (bool, bool, error) {
   var plain, crypt, http, ws bool
   var p string

   for _, p = range proto {
      switch p {
         case "http" :
            plain = true
            http  = true
         case "https" :
            crypt = true
            http  = true
         case "ws" :
            plain = true
            ws    = true
         case "wss" :
            crypt = true
            ws    = true
         default:
            return false, false, ErrorInvalidProtocol
      }
   }

   // verifies if it has mixed uses
   return http && ws, plain && crypt, nil
}

