package stonelizard

import (
   "io"
   "strings"
   "net/http"
   "encoding/base64"
)

func (svc *ServiceWithPKI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   var err error
   var usertype string
   var buf []byte
   var signature []byte
   var msgToVerify string
   var upk Pki

   r.Header.Del("X-Request-Signer-Certificate")
   r.Header.Del("X-Request-Signer-Type")

   if r.URL.Path != (svc.SwaggerPath+"/swagger.json") {
      Goose.InitServe.Logf(0,"wrapped")

      signature, err = base64.StdEncoding.DecodeString(r.Header.Get("X-Request-Signature"))
      if err == nil {
         upk, usertype, err = svc.PK.FindCertificate(r.Header.Get("X-Request-Signer"))
         if err == nil {
            msgToVerify = strings.ToUpper(r.Method) + "+" + r.URL.String()

            buf, err = io.ReadAll(r.Body)
            if (err == nil) && (len(buf) > 0) {
               msgToVerify += "\n" + string(buf)
            }

            r.Body = NewReadCloser(buf)

            if upk.Verify(msgToVerify, signature) == nil {
               Goose.InitServe.Logf(0,"Verified")
               r.Header.Set("X-Request-Signer-Certificate", base64.StdEncoding.EncodeToString(upk.Certificate()))
               r.Header.Set("X-Request-Signer-Type", usertype)
            }
         }
      }
   }
   svc.Service.ServeHTTP(w,r)
}


