package certkitetcd


import (
   "io"
   "os"
   "fmt"
   "time"
   "regexp"
   "strings"
   "net/http"
   "io/ioutil"
   "crypto/rsa"
   "crypto/tls"
   "archive/zip"
   "crypto/x509"
   "encoding/pem"
   "golang.org/x/net/context"
   "github.com/luisfurquim/etcdconfig"
   etcd "github.com/coreos/etcd/client"
)

//Load in memory and decodes the certificate from the reader
func (crtkit CertKit) ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error) {
   var err      error
   var pembuf []byte
   var cert    *x509.Certificate

   Goose.Loader.Logf(5,"Will read certificate")

   pembuf, err = ioutil.ReadAll(r)
   if err != nil {
      Goose.Loader.Logf(1,"Failed reading cert %s",err)
      return nil, nil, err
   }

   Goose.Loader.Logf(5,"Certificate was read into memory buffer")

   block, _  := pem.Decode(pembuf)

   if block == nil {
      Goose.Loader.Logf(6,"%s: [%s]",ErrorBadPEMBlock,pembuf)
      return nil, nil, ErrorBadPEMBlock
   }

   Goose.Loader.Logf(5,"PEM was decoded")

   cert, err  = x509.ParseCertificate(block.Bytes)
   if err != nil {
      Goose.Loader.Logf(1,"Failed parsing cert %s",err)
      return nil, nil, err
   }

   if time.Now().Before(cert.NotBefore) || time.Now().After(cert.NotAfter) {
      Goose.Loader.Logf(1,"%s",ErrorValidDate)
      return nil, nil, ErrorValidDate
   }

   Goose.Loader.Logf(5,"Certificate was parsed")

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
      Goose.Loader.Logf(1,"Failed reading Key %s",err)
      return nil, nil, err
   }

   block, _  := pem.Decode(pembuf)
   key, err  = x509.ParsePKCS1PrivateKey(block.Bytes)
   if err != nil {
      Goose.Loader.Logf(1,"Failed parsing key %s",err)
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

   if block == nil {
      Goose.Loader.Logf(1,"%s: [%s]",ErrorBadPEMBlock,pembuf)
      return nil, nil, ErrorBadPEMBlock
   }

   plainkey, err = x509.DecryptPEMBlock(block,[]byte{})
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

func (ck *CertKit) Setup(udata map[string]interface{}) error {
   if _, ok := udata["etcd"]; !ok {
      Goose.Loader.Logf(1,"%s", ErrorNoEtcdHandler)
      return ErrorNoEtcdHandler
   }

   if _, ok := udata["etcdkey"]; !ok {
      Goose.Loader.Logf(1,"%s", ErrorNoEtcdKey)
      return ErrorNoEtcdKey
   }

   switch udata["etcd"].(type) {
      case etcd.Client:
         ck.Etcdcli = udata["etcd"].(etcd.Client)
      default:
         Goose.Loader.Logf(1,"%s", ErrorBadEtcdHandler)
         return ErrorBadEtcdHandler
   }

   switch udata["etcdkey"].(type) {
      case string:
         ck.Etcdkey = udata["etcdkey"].(string)
      default:
         Goose.Loader.Logf(1,"%s", ErrorBadEtcdKey)
         return ErrorBadEtcdKey
   }


   ck.etcdCertKeyRE   = regexp.MustCompile("^/" + ck.Etcdkey + "/trusted/" + "(.+)_([0-9a-fA-F]+)/cert$")
   ck.etcdDeleteKeyRE = regexp.MustCompile("^/" + ck.Etcdkey + "/trusted/" + "(.+)_([0-9a-fA-F]+)$")

   return nil
}


func (ck *CertKit) LoadUserData() error {
   var err error
   var ClientCert *x509.Certificate
   var ClientCertRaw interface{}
   var etcdData interface{}
   var usrKey string
   var usrData interface{}

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey)
   if err != nil {
      _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), ck.Etcdkey, "", &etcd.SetOptions{Dir:true})
      if err != nil {
         Goose.Loader.Logf(1,"Error creating creating base diretory (%s): %s",ck.Etcdkey,err)
         return err
      }

      _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), ck.Etcdkey + "/trusted", "", &etcd.SetOptions{Dir:true})
      if err != nil {
         Goose.Loader.Logf(1,"Error creating creating trusted users base diretory (%s): %s",ck.Etcdkey + "/trusted",err)
         return err
      }

      _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), ck.Etcdkey + "/pending", "", &etcd.SetOptions{Dir:true})
      if err != nil {
         Goose.Loader.Logf(1,"Error creating creating pending users base diretory (%s): %s",ck.Etcdkey + "/pending",err)
         return err
      }
   }

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey + "/pending")
   if err != nil {
      _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), ck.Etcdkey + "/pending", "", &etcd.SetOptions{Dir:true})
      if err != nil {
         Goose.Loader.Logf(1,"Error creating creating pending users base diretory (%s): %s",ck.Etcdkey + "/pending",err)
         return err
      }
   }

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey + "/trusted")
   if err != nil {
      _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), ck.Etcdkey + "/trusted", "", &etcd.SetOptions{Dir:true})
      if err != nil {
         Goose.Loader.Logf(1,"Error creating creating trusted users base diretory (%s): %s",ck.Etcdkey + "/trusted",err)
         return err
      }
   }

   ck.UserCerts = map[string]*UserDB{}

   if etcdData != nil {
      for usrKey, usrData = range etcdData.(map[string]interface{}) {
         switch usrData.(type) {
            case (map[string]interface{}):
               ClientCertRaw = usrData.(map[string]interface{})["cert"]
               switch ClientCertRaw.(type) {
                  case string:
                     ClientCert, _, err = ck.ReadCertFromReader(strings.NewReader(ClientCertRaw.(string)))
                     if err == nil {
                        ck.UserCerts[usrKey] = &UserDB{Cert: ClientCert}
                     } else {
                        Goose.Loader.Logf(1,"Error decoding certificate for %s: %s", usrKey, err)
                        Goose.Loader.Logf(1,"cert:%s", ClientCertRaw.(string))
                     }
                  default:
                     Goose.Loader.Logf(1,"Error reading certificate for %s: %s", usrKey, err)
               }
            default:
               Goose.Loader.Logf(1,"Error reading user data for %s: %s", usrKey, err)
         }
      }
   }

   etcdconfig.OnUpdateTree(ck.Etcdcli, ck.Etcdkey + "/trusted", func(key string, val interface{}, action string) {
      Goose.Loader.Logf(2,"Update (%s) on %s|", action, key)
      Goose.Loader.Logf(5,"Update (%s) on %s: %#v", action, key, val)
      //key = key[len(ck.Etcdkey) + 10:]
      Goose.Loader.Logf(2,"Update (%s) on map key %s|", action, key)
      switch action {
         case "set":
            idParts := ck.etcdCertKeyRE.FindStringSubmatch(key)
            if len(idParts) > 0 {
               key = idParts[1] + "_" + idParts[2]
               ClientCert, _, err = ck.ReadCertFromReader(strings.NewReader(val.(string)))
               Goose.Loader.Logf(2,"certificate for %s was decoded", idParts[1])
               if err == nil {
                  if _, ok := ck.UserCerts[key]; ok {
                     ck.UserCerts[key].Cert = ClientCert
                     Goose.Loader.Logf(2,"certificate for %s was updated", idParts[1])
                  } else {
                     ck.UserCerts[key] = &UserDB{Cert: ClientCert}
                     Goose.Loader.Logf(2,"certificate for %s was created", idParts[1])
                  }
               } else {
                  Goose.Loader.Logf(1,"Error decoding certificate for %s: %s", key, err)
               }
            }

         case "delete":
            idParts := ck.etcdDeleteKeyRE.FindStringSubmatch(key)
            if len(idParts) > 0 {
               key = idParts[1] + "_" + idParts[2]
               delete(ck.UserCerts,key)
               Goose.Loader.Logf(2,"certificate for %s was deleted", key)
            }
      }
   })

/*
   _, etcdData, err = etcdconfig.GetConfig(etcdhandle, etcdkey + "/pending")
   if err != nil {
      Goose.Loader.Logf(1,"Error retrieving trusted users certificates: %s", err)
      return err
   }

   ck.PendingCerts = map[string]UserDB{}
   for usrKey, usrData = range etcdData {
      switch usrData.(type) {
         case (map[string]interface{}):
            ClientCertRaw = usrData.(map[string]interface{})["cert"]
            switch ClientCertRaw.(type) {
               case string:
                  ClientCert, _, err = ck.ReadCertFromReader(strings.NewReader(ClientCertRaw.(string)))
                  if err == nil {
                     ck.PendingCerts[usrKey] = UserDB{Cert: ClientCert}
                  } else {
                     Goose.Loader.Logf(1,"Error decoding pending certificate for %s: %s", usrKey, err)
                  }
               default:
                  Goose.Loader.Logf(1,"Error reading pending certificate for %s: %s", usrKey, err)
            }
         default:
            Goose.Loader.Logf(1,"Error reading pending user data for %s: %s", usrKey, err)
      }
   }
*/

   return err
}


func (ck *CertKit) AddUserData(usrKey string, ClientCert *x509.Certificate) error {
   var err error

   if err == nil {
      if ck.UserCerts == nil {
         ck.UserCerts = map[string]*UserDB{usrKey : &UserDB{Cert: ClientCert}}
      } else {
         ck.UserCerts[usrKey] = &UserDB{Cert: ClientCert}
      }
   } else {
      Goose.Loader.Logf(1,"Error decoding certificate for %s: %s", usrKey, err)
   }

   return err
}

func (ck *CertKit) SetServerYearValidity(notBefore time.Time, validityServerCert int) {
   //If not set by application use default
   if validityServerCert == 0 {
      ck.notAfterServer =  notBefore.Add(ServerTime)
   } else { //Apply validate time defined by application
      ck.notAfterServer = notBefore.AddDate(validityServerCert,0,0)
   }
}

func (ck *CertKit) SetClientYearValidity(notBefore time.Time, validityClientCert int)  {
   //If not set by application use default
   if validityClientCert == 0 {
      ck.notAfterClient = notBefore.Add(ClientTime)
   } else { //Apply validate time defined by application
      ck.notAfterClient = notBefore.AddDate(validityClientCert,0,0)
   }
}

func (ck *CertKit) SetCAYearValidity(notBefore time.Time, validityCACert int) {
   //If not set by application use default
   if validityCACert == 0 {
      ck.notAfterCA = notBefore.Add(CaTime)
   }  else { //Apply validate time defined by application
      ck.notAfterCA =  notBefore.AddDate(validityCACert,0,0)
   }
}
