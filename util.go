package stonelizard


import (
   "fmt"
   "io/ioutil"
   "crypto/rsa"
   "crypto/x509"
   "crypto/ecdsa"
   "encoding/pem"
)

func (svc Service) String() string {
   var s string

   if svc.PageNotFound != nil {
      s = "404: " + string(svc.PageNotFound) + "\n"
   }

   s += fmt.Sprintf("%s",svc.Svc)
   return s
}

func (u UrlNode) String() string {
   var s string

   if u.Handle != nil {
      s = fmt.Sprintf("Method: %#v\n",u.Handle)
   }

   return s
}

func (svc Service) readPem(pembuf *[]byte, path string) error {
   var err error
/*
   if (path[0] != '/') && (path[0] != '.') {
      path = svc.PemPath + "/" + path
   }
*/
   *pembuf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading Pem %s",err)
      return err
   }

   return nil
}

func (svc Service) readCert(pembuf *[]byte, cert **x509.Certificate, path string) error {
   var err error

   err = svc.readPem(pembuf,path)
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

func (svc Service) readCRL(buf *[]byte, path string) error {
   var err error

   if (path[0] != '/') && (path[0] != '.') {
      path = svc.PemPath + "/" + path
   }

   *buf, err = ioutil.ReadFile(path)
   if err != nil {
      Goose.Logf(1,"Failed reading CRL %s",err)
      return err
   }

   return nil
}

func (svc Service) readEcdsaKey(pembuf *[]byte, key **ecdsa.PrivateKey, path string) error {
   var err error

   err = svc.readPem(pembuf,path)
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

func (svc Service) readRsaKey(pembuf *[]byte, key **rsa.PrivateKey, path string) error {
   var err error

   err = svc.readPem(pembuf,path)
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


