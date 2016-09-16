package certkit

import (
   "os"
   "fmt"
   "bytes"
   "net/http"
   "io/ioutil"
   "crypto/tls"
   "crypto/x509"
   "github.com/luisfurquim/stonelizard"
)

func (ck *CertKit) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   var cert       *x509.Certificate
   var UserCert   *x509.Certificate
   var found       bool
   var ok          bool
   var commonName  string

   Goose.Auth.Logf(5,"Peer certificates")
   found = false
   for _, cert = range TLS.PeerCertificates {
      Goose.Auth.Logf(6,"Peer certificate: %#v",cert)
      Goose.Auth.Logf(5,"Peer certificate: #%s, ID: %s, Issuer: %s, Subject: %s, \n\n\n",cert.SerialNumber,cert.SubjectKeyId,cert.Issuer.CommonName,cert.Subject.CommonName)
      if UserCert, ok = ck.UserCerts[cert.Subject.CommonName]; ok {
         if bytes.Equal(UserCert.Raw,cert.Raw) {
            found = true
            break
         } else {
            Goose.Auth.Logf(4,"certs diff: %s",cert.Subject.CommonName)
         }
      }
   }

   Goose.Auth.Logf(5,"Ending peer certificate scan")

   if found {
      Goose.Auth.Logf(7,"TLS:%#v",TLS)
      return http.StatusOK, cert, nil
   }

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

   Goose.Auth.Logf(6,"authtype: %#v",atype)
   Goose.Auth.Logf(6,"CAs: %#v",roots)

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
   Goose.Auth.Logf(5,"X509KeyPair used: %#v",tlsConfig.Certificates[0])
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
   return ck.ServerCert.DNSNames
}
