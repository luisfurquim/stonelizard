package stonelizard

import (
   "io"
   "strings"
   "net/http"
   "encoding/base64"
)
//2022/11/22 02:28:06 {paracentric}[paracentric.go]<(*PkiT).Verify>(414): Error from verification: crypto/rsa: verification error

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
				Goose.InitServe.Logf(0,"Certificate found: %#v", upk)
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
			} else {
				Goose.InitServe.Logf(0,"Certificate not found: %s", r.Header.Get("X-Request-Signer"))
				Goose.InitServe.Logf(0,"Certificates: %#v", svc.PK)
         }
      } else {
			Goose.InitServe.Logf(0,"Signature not found: %s", r.Header.Get("X-Request-Signature"))
		}
   }
   svc.Service.ServeHTTP(w,r)
}


