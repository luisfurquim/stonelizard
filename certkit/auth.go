package certkit

import (
   "os"
   "fmt"
   "bytes"
   "strings"
   "net/http"
   "io/ioutil"
   "crypto/rsa"
   "crypto/tls"
   "crypto/x509"
   "encoding/pem"
   "github.com/luisfurquim/stonelizard"
)

func (ck *CertKit) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   var cert       *x509.Certificate
   var UserCert   *x509.Certificate
   var ok          bool
   var commonName  string

   Goose.Auth.Logf(0,"Peer certificates: %#v", Goose)
   for _, cert = range TLS.PeerCertificates {
      Goose.Auth.Logf(5,"ck.UserCerts: %#v",ck.UserCerts)
      Goose.Auth.Logf(7,"Peer certificate: %#v",cert)
      Goose.Auth.Logf(5,"Peer certificate: #%s, ID: %s, Issuer: %s, Subject: %s, \n\n\n",cert.SerialNumber,cert.SubjectKeyId,cert.Issuer.CommonName,cert.Subject.CommonName)
      if UserCert, ok = ck.UserCerts[cert.Subject.CommonName]; ok {
         Goose.Auth.Logf(5,"1")
         if bytes.Equal(UserCert.Raw,cert.Raw) {
            Goose.Auth.Logf(5,"found")
            Goose.Auth.Logf(7,"TLS:%#v",TLS)
            return http.StatusOK, cert, nil
         } else {
            Goose.Auth.Logf(4,"certs diff: %s",cert.Subject.CommonName)
         }
      }
   }

   Goose.Auth.Logf(5,"Ending peer certificate scan")

   Goose.Auth.Logf(1,"Unauthorized access attempt from %s to path %s", RemoteAddr, path)
   Goose.Auth.Logf(6,"Unauthorized access attempt, certs: %#v",TLS.PeerCertificates)
   for _, cert = range TLS.PeerCertificates {
      Goose.Auth.Logf(4,"Sent access cert (needs authorization): %#v",cert.Subject)
   }
   for commonName, _ = range ck.UserCerts {
      Goose.Auth.Logf(4,"access grantable to: %s",commonName)
   }

   if cert != nil {
      // Shaper interface has an optional method: SavePending(cert *x509.Certificate) error
      err = SavePending(cert)
      if err != nil {
         Goose.Auth.Logf(1,"Internal server error saving unauthorized certificate for %s: %s",cert.Subject.CommonName,err)
      }
   }

   return http.StatusUnauthorized, nil, ErrorUnauthorized
//   w.WriteHeader(http.StatusUnauthorized)
//   w.Write(svc.Config.PageNotFound())

}

//func (ck *CertKit) GetTLSConfig(AuthRequired bool) (*tls.Config, error) {
func (ck *CertKit) GetTLSConfig(Access uint8) (*tls.Config, error) {
   var atype      tls.ClientAuthType
   var tlsConfig *tls.Config
   var roots     *x509.CertPool

   switch Access {
      case stonelizard.AccessNone:
         atype = tls.NoClientCert
      case stonelizard.AccessAuth, stonelizard.AccessAuthInfo:
         atype = tls.RequestClientCert
      case stonelizard.AccessVerifyAuth, stonelizard.AccessVerifyAuthInfo:
         atype = tls.RequireAndVerifyClientCert

         // Code adapted from crypto/x509/root_unix.go
         roots = x509.NewCertPool()

         for _, directory := range CertDirectories {
            fis, err := ioutil.ReadDir(directory)
            if err != nil {
               Goose.Auth.Logf(5,"Error scanning certificate directory %s: %s",directory,err)
               continue
            }
            for _, fi := range fis {
               data, err := ioutil.ReadFile(fmt.Sprintf("%s%c%s",directory,os.PathSeparator,fi.Name()))
               if err != nil {
                  Goose.Auth.Logf(5,"Error load CA certificate from %s%c%s: %s",directory,os.PathSeparator,fi.Name(),err)
                  continue
               }
               Goose.Auth.Logf(5,"Loaded CA certificate from %s%c%s: %s",directory,os.PathSeparator,fi.Name(),err)

               roots.AppendCertsFromPEM(data)
            }
         }
   }

   Goose.Auth.Logf(4,"authtype: %#v",atype)
   Goose.Auth.Logf(5,"CAs: %#v",roots)

   tlsConfig = &tls.Config{
      ClientAuth: atype,
      ClientCAs: roots,
//      InsecureSkipVerify: true,
      Certificates: make([]tls.Certificate, 1),
   }

/*
   srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(svc.PemPath + "/server.crt", svc.PemPath + "/server.key")
   if err != nil {
      Goose.InitServe.Logf(1,"Failed reading server certificates: %s",err)
      return err
   }
*/

   tlsConfig.Certificates[0] = ck.ServerX509KeyPair
   Goose.Auth.Logf(0,"X509KeyPair used: %#v",tlsConfig.Certificates[0])//5
   tlsConfig.BuildNameToCertificate()

   return tlsConfig, nil
}

func (ck *CertKit) StartCRLServer(listenAddress string, listener *stonelizard.StoppableListener) (error) {
   srvcrl := &http.Server{
      Addr: listenAddress,
      Handler: ck,
   }

   Goose.Auth.Logf(5,"CRL Listen Address: %s",listenAddress)
   return srvcrl.Serve(listener)

//   Goose.InitServe.Logf(5,"CRL Listen is serving")
//   err = http.ListenAndServe(svc.CRLListenAddress,svc.Auth)
//   if err != nil {
//      Goose.InitServe.Fatalf(1,"Error serving CRL: %s",err)
//   }

}

func (ck *CertKit) GetDNSNames() []string {
   if ck==nil || ck.ServerCert == nil {
      return []string{}
   }
   return ck.ServerCert.DNSNames
}

func (ck *CertKit) GetServerCert() *x509.Certificate {
   return ck.ServerCert
}

func (ck *CertKit) GetServerKey() *rsa.PrivateKey {
   return ck.ServerKey
}

func (ck *CertKit) GetCACert() *x509.Certificate {
   return ck.CACert
}

func (ck *CertKit) GetCAKey() *rsa.PrivateKey {
   return ck.CAKey
}

func (ck *CertKit) GetServerX509KeyPair() tls.Certificate {
   return ck.ServerX509KeyPair
}

func (ck *CertKit) GetCertPool() *x509.CertPool {
   return ck.CertPool
}




func (ck *CertKit) Trust(id string) error {
   return os.Rename(fmt.Sprintf("%s%cpending%c%s.crt", ck.Path, os.PathSeparator, os.PathSeparator, id),fmt.Sprintf("%s%cclient%c%s.crt", ck.Path, os.PathSeparator, os.PathSeparator, id))
}

func (ck *CertKit) GetPending() (map[string]interface{}, error) {
   var err error
   var resp map[string]interface{}
   var names []string
   var fh *os.File
   var buf []byte

   resp = map[string]interface{}{}

   fh, err = os.Open(fmt.Sprintf("%s%cpending", ck.Path, os.PathSeparator))
   if err != nil {
      Goose.Auth.Logf(1,"Error opening pending directory (%s%cpending): %s",ck.Path,os.PathSeparator,err)
      return nil, err
   }

   names, err = fh.Readdirnames(-1)
   if err != nil {
      Goose.Auth.Logf(1,"Error reading pending directory (%s%cpending): %s",ck.Path,os.PathSeparator,err)
      return nil, err
   }

   for _, k := range names {
      if (len(k)>4) && (k[len(k)-4:]==".crt") {
         buf, err = ioutil.ReadFile(fmt.Sprintf("%s%cpending%c%s",ck.Path,os.PathSeparator,os.PathSeparator,k))
         if err != nil {
            Goose.Auth.Logf(1,"Error reading pending certificate (%s%cpending%c%s): %s",ck.Path,os.PathSeparator,os.PathSeparator,k,err)
            return nil, err
         }
         resp[k[:len(k)-4]] = string(buf)
      }
   }

   return resp, nil
}

func (ck *CertKit) SavePending(cert *x509.Certificate) error {
   var err error
   var fname string

   fname = fmt.Sprintf("%s%cpending%c%s.crt", ck.Path, os.PathSeparator, os.PathSeparator, strings.Replace(cert.Subject.CommonName," ","_",-1))

   _, err = os.Stat(fname)
   if !os.IsNotExist(err) {
      Goose.Auth.Logf(1,"Can't write file %s: %s", fname, err)
      return ErrorDuplicateFile
   }

   err = ioutil.WriteFile(fname, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}), 0600)
   if err != nil {
      Goose.Auth.Logf(1,"Error writing file %s: %s", fname, err)
      return err
   }

   return nil
}

func (ck *CertKit) GetTrusted() (map[string]interface{}, error) {
   var resp map[string]interface{}

   resp = map[string]interface{}{}

   for k, v := range ck.UserCerts {
      resp[k] = v
   }

   return resp, nil
}

func (ck *CertKit) Reject(id string) error {
   return ck.Delete("pending",id)
}

func (ck *CertKit) Drop(id string) error {
   delete(ck.UserCerts,id)
   return ck.Delete("client",id)
}

func (ck *CertKit) Delete(tree, id string) error {
   return os.Remove(fmt.Sprintf("%s%c%s%c%s.crt", ck.Path, os.PathSeparator, tree, os.PathSeparator, id))
}


func (ck *CertKit) AddUserData(usrKey string, ClientCert *x509.Certificate) error {
   var err error

   if err == nil {
      if ck.UserCerts == nil {
         ck.UserCerts = map[string]*x509.Certificate{usrKey : ClientCert}
      } else {
         ck.UserCerts[usrKey] = ClientCert
      }
   } else {
      Goose.Loader.Logf(1,"Error decoding certificate for %s: %s", usrKey, err)
   }

   return err
}
