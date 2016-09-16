package certkitetcd


import (
   "io"
   "os"
   "fmt"
   "net/http"
   "io/ioutil"
   "crypto/rsa"
   "crypto/tls"
   "archive/zip"
   "crypto/x509"
   "encoding/pem"
   "path/filepath"
//   "github.com/luisfurquim/etcdconfig"
//   etcd "github.com/coreos/etcd/client"
)

//Load in memory and decodes the certificate from the reader
func (crtkit CertKit) ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error) {
   var err      error
   var pembuf []byte
   var cert    *x509.Certificate

   pembuf, err = ioutil.ReadAll(r)
   if err != nil {
      Goose.Loader.Logf(1,"Failed reading cert %s",err)
      return nil, nil, err
   }

   block, _  := pem.Decode(pembuf)
   cert, err  = x509.ParseCertificate(block.Bytes)
   if err != nil {
      Goose.Loader.Logf(1,"Failed parsing cert %s",err)
      return nil, nil, err
   }

   return cert, pembuf, nil
}

// Open certificate file and call the reader
func (crtkit CertKit) ReadCertificate(fname string) (*x509.Certificate, []byte, error) {
   var err       error
   var f        *os.File

   if fname[0] == os.PathSeparator {
      f, err = os.Open(fname)
   } else {
      f, err = os.Open(fmt.Sprintf("%s%c%s",crtkit.Path, os.PathSeparator,fname))
   }
   if err != nil {
      Goose.Loader.Logf(1,"Failed opening Cert %s",err)
      return nil, nil, err
   }

   return crtkit.ReadCertFromReader(f)
}

func (crtkit CertKit) ReadRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   var err      error
   var pembuf []byte
   var key     *rsa.PrivateKey

   pembuf, err = ioutil.ReadAll(r)
   if err != nil {
      Goose.Loader.Logf(0,"Failed reading Key %s",err)
      return nil, nil, err
   }

   block, _  := pem.Decode(pembuf)
   key, err  = x509.ParsePKCS1PrivateKey(block.Bytes)
   if err != nil {
      Goose.Loader.Logf(0,"Failed parsing key %s",err)
      return nil, nil, err
   }

   return key, pembuf, nil
}

func (crtkit CertKit) ReadRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   var err      error
   var f       *os.File

   if fname[0] == os.PathSeparator {
      f, err = os.Open(fname)
   } else {
      f, err = os.Open(fmt.Sprintf("%s%c%s",crtkit.Path, os.PathSeparator,fname))
   }
   if err != nil {
      Goose.Loader.Logf(1,"Failed opening Cert %s",err)
      return nil, nil, err
   }

   return crtkit.ReadRsaPrivKeyFromReader(f)
}

func (crtkit CertKit) ReadDecryptRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   var err        error
   var pembuf   []byte
   var key       *rsa.PrivateKey
   var plainkey []byte
   var block     *pem.Block

   pembuf, err = ioutil.ReadAll(r)
   if err != nil {
      Goose.Loader.Logf(1,"Failed reading Key %s",err)
      return nil, nil, err
   }

   block, _ = pem.Decode(pembuf)

   plainkey, err = x509.DecryptPEMBlock(block,[]byte{})
   Goose.Loader.Logf(1,"DecryptPEMBlock: %s",plainkey)
   if err != nil {
      return nil, nil, err
   }

   key, err  = x509.ParsePKCS1PrivateKey(plainkey)
   if err != nil {
      Goose.Loader.Logf(1,"Failed parsing key %s",err)
      return nil, nil, err
   }

   return key, pem.EncodeToMemory(&pem.Block{
      Type:   "RSA PRIVATE KEY",
//    Headers map[string]string // Optional headers.
      Bytes:   plainkey,
   }), nil
}

func (crtkit CertKit) ReadDecryptRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   var err      error
   var f       *os.File

   if fname[0] == os.PathSeparator {
      f, err = os.Open(fname)
   } else {
      f, err = os.Open(fmt.Sprintf("%s%c%s",crtkit.Path, os.PathSeparator,fname))
   }
   if err != nil {
      Goose.Loader.Logf(1,"Failed opening Cert %s",err)
      return nil, nil, err
   }

   return crtkit.ReadDecryptRsaPrivKeyFromReader(f)
}

//Load in memory the Certificate Revogation List from the PemPath field of Service struct
func (crtkit CertKit) ReadCRL(fname string) ([]byte, error) {
   var err error
   var buf []byte

   buf, err = ioutil.ReadFile(fmt.Sprintf("%s%c%s",crtkit.Path, os.PathSeparator,fname))
   if err != nil {
      Goose.Loader.Logf(1,"Failed reading CRL %s",err)
      return nil, err
   }

   return buf, nil
}

func (svc CertKit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   Goose.Serve.Logf(3,"Certificate revocation list requested by %s",r.RemoteAddr)
   if r.URL.Path == "/rootCA.crl" {
      w.WriteHeader(http.StatusOK)
      w.Header().Set("Content-Type", "application/pkix-crl")
      w.Write([]byte(svc.CACRL))
      Goose.Serve.Logf(3,"Certificate revocation list sent for %s",r.RemoteAddr)
      return
   }

   Goose.Serve.Logf(3,"Can only serve certificate revocation list %s",r.RemoteAddr)
   w.WriteHeader(http.StatusNotFound)
}

func NewFromCK(path string) (*CertKit, error) {
   var ck  CertKit
   var err error
   var hn  string

   ck = CertKit{
      Path:    path,
   }

   hn, err = os.Hostname()
   if err != nil {
      Goose.Loader.Logf(1,"Error checking hostname: %s",err)
      return nil, err
   }

   r, err := zip.OpenReader(fmt.Sprintf("%s%c%s.ck",path, os.PathSeparator,hn))
   if err != nil {
      Goose.Loader.Logf(1,"Error decompressing certificate archive: %s",err)
      return nil, err
   }
   defer r.Close()

   // Iterate through the files in the archive.
   for _, f := range r.File {
      rc, err := f.Open()
      if err != nil {
         Goose.Loader.Logf(1,"Error opening %s: %s",f.Name,err)
         return nil, err
      }

      switch f.Name {
         case "server.crt":
            ck.ServerCert, ck.ServerCertPem, err = ck.ReadCertFromReader(rc)
         case "server.key":
            ck.ServerKey, ck.ServerKeyPem, err = ck.ReadRsaPrivKeyFromReader(rc)
         case "rootCA.crt":
            ck.CACert, ck.CACertPem, err = ck.ReadCertFromReader(rc)
         case "rootCA.key":
            ck.CAKey, ck.CAKeyPem, err = ck.ReadRsaPrivKeyFromReader(rc)
      }
      rc.Close()

      if err != nil {
         Goose.Loader.Logf(1,"Error loading %s: %s",f.Name,err)
         return nil, err
      }
   }
   ck.ServerX509KeyPair, err = tls.X509KeyPair(ck.ServerCertPem, ck.ServerKeyPem)
   return &ck, nil
}

//func (ck *CertKit) LoadUserData(etcdcli etcd.Client, key string) error {
func (ck *CertKit) LoadUserData(udata map[string]interface{}) error {
   var err error
   err = filepath.Walk(fmt.Sprintf("%s%c%s",ck.Path, os.PathSeparator,"client"), func (path string, f os.FileInfo, err error) error {
      var ClientCert *x509.Certificate

      if (len(path)<4) || (path[len(path)-4:]!=".crt") {
         return nil
      }

      ClientCert, _, err = ck.ReadCertificate(path)
      if err != nil {
         Goose.Loader.Logf(1,"Failed reading %s file: %s", path, err)
         return err
      }

      if ck.UserCerts == nil {
         ck.UserCerts = map[string]*x509.Certificate{ClientCert.Subject.CommonName:ClientCert}
      } else {
         ck.UserCerts[ClientCert.Subject.CommonName] = ClientCert
      }

     return nil
   })

   return err
}

