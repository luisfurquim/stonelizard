package certkit


import (
   "os"
   "fmt"
   "net/http"
   "io/ioutil"
   "crypto/rsa"
   "crypto/tls"
   "crypto/x509"
   "crypto/ecdsa"
   "encoding/pem"
   "path/filepath"
)

//Load in memory and decodes the microservice certificate from the PemPath field of Service struct
func (crtkit CertKit) ReadCert(pembuf *[]byte, cert **x509.Certificate, path string) error {
   var err error

   *pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Cert %s",err)
      return err
   }

   block, _  := pem.Decode(*pembuf)
   *cert, err  = x509.ParseCertificate(block.Bytes)
   if err != nil {
      Goose.Logf(1,"Failed decoding Cert %s",err)
      return err
   }

   return nil
}

//Load in memory and decodes the certificate from the PemPath field of Service struct
func (crtkit CertKit) ReadCertificate(cert **x509.Certificate, path string) error {
   var err error
   var pembuf []byte

   pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Cert %s",err)
      return err
   }

   block, _  := pem.Decode(pembuf)
   *cert, err  = x509.ParseCertificate(block.Bytes)
   if err != nil {
      Goose.Logf(1,"Failed decoding Cert %s",err)
      return err
   }

   return nil
}

//Load in memory the Certificate Revogation List from the PemPath field of Service struct
func (crtkit CertKit) ReadCRL(buf *[]byte, path string) error {
   var err error

   *buf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading CRL %s",err)
      return err
   }

   return nil
}

func (crtkit CertKit) ReadEcdsaKey(pembuf *[]byte, key **ecdsa.PrivateKey, path string) error {
   var err error

   *pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Key %s",err)
      return err
   }

   block, _  := pem.Decode(*pembuf)
   *key, err  = x509.ParseECPrivateKey(block.Bytes)
   if err != nil {
      Goose.Logf(1,"Failed decoding key %s",err)
      return err
   }

   return nil
}

func (crtkit CertKit) ReadRsaKey(pembuf *[]byte, key **rsa.PrivateKey, path string) error {
   var err error

   *pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Key %s",err)
      return err
   }

   block, _  := pem.Decode(*pembuf)
   *key, err  = x509.ParsePKCS1PrivateKey(block.Bytes)
   if err != nil {
      Goose.Logf(1,"Failed decoding key %s",err)
      return err
   }

   return nil
}

func (crtkit CertKit) ReadRsaPrivKey(key **rsa.PrivateKey, path string) error {
   var err      error
   var pembuf []byte

   pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Key %s",err)
      return err
   }

   block, _  := pem.Decode(pembuf)
   *key, err  = x509.ParsePKCS1PrivateKey(block.Bytes)
   if err != nil {
      Goose.Logf(1,"Failed decoding key %s",err)
      return err
   }

   return nil
}

func (svc CertKit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   if r.URL.Path == "/rootCA.crl" {
      w.WriteHeader(http.StatusOK)
      w.Header().Set("Content-Type", "application/pkix-crl")
      w.Write([]byte(svc.CACRL))
      return
   }

   w.WriteHeader(http.StatusNotFound)
}

func Load(path, ServerCert, CACert, ServerKey, CAKey string) (*CertKit, error) {
   var ck  CertKit
   var err error

   ck = CertKit{}

   if !((ServerCert=="") && (ServerKey=="")) {
      if (ServerCert=="") || (ServerKey=="") {
         Goose.Logf(1,"Error loading server certificate/key pair: %s",ErrorCertsMustHaveKeys)
         return nil, ErrorCertsMustHaveKeys
      }

      certpath := fmt.Sprintf("%s%c%s",path, os.PathSeparator,ServerCert)
      err = ck.ReadCertificate(&ck.ServerCert, certpath)
      if err != nil {
         Goose.Logf(1,"Error loading server certificate: %s",err)
         return nil, err
      }

      keypath := fmt.Sprintf("%s%c%s",path, os.PathSeparator,ServerKey)
      err = ck.ReadRsaPrivKey(&ck.ServerKey, keypath)
      if err != nil {
         Goose.Logf(1,"Error loading server private key: %s",err)
         return nil, err
      }

      ck.ServerX509KeyPair, err = tls.LoadX509KeyPair(certpath, keypath)
      if err != nil {
         Goose.Logf(1,"Failed reading server certificate/key pair: %s",err)
         return nil, err
      }
   }

   if !((CACert=="") && (CAKey=="")) {
      if (CACert=="") || (CAKey=="") {
         Goose.Logf(1,"Error loading CA certificate/key pair: %s",ErrorCertsMustHaveKeys)
         return nil, ErrorCertsMustHaveKeys
      }

      certpath := fmt.Sprintf("%s%c%s",path, os.PathSeparator,CACert)
      err = ck.ReadCertificate(&ck.CACert, certpath)
      if err != nil {
         Goose.Logf(1,"Error loading CA certificate: %s",err)
         return nil, err
      }

      keypath := fmt.Sprintf("%s%c%s",path, os.PathSeparator,CAKey)
      err = ck.ReadRsaPrivKey(&ck.CAKey, keypath)
      if err != nil {
         Goose.Logf(1,"Error loading CA private key: %s",err)
         return nil, err
      }
   }

   err = filepath.Walk(fmt.Sprintf("%s%c%s",path, os.PathSeparator,"client"), func (path string, f os.FileInfo, err error) error {
      var ClientCert *x509.Certificate

      if (len(path)<4) || (path[len(path)-4:]!=".crt") {
         return nil
      }

      err = ck.ReadCertificate(&ClientCert, path)
      if err != nil {
         Goose.Logf(1,"Failed reading %s file: %s", path, err)
         return err
      }

      if ck.UserCerts == nil {
         ck.UserCerts = map[string]*x509.Certificate{ClientCert.Subject.CommonName:ClientCert}
      } else {
         ck.UserCerts[ClientCert.Subject.CommonName] = ClientCert
      }

     return nil
   })



   return &ck, nil
}

