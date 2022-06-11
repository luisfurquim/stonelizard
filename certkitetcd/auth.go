package certkitetcd

import (
//   "os"
   "fmt"
   "hash"
   "bytes"
   "errors"
   "net/http"
//   "io/ioutil"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
   "crypto/sha1"
   "encoding/pem"
   "golang.org/x/net/context"
   "github.com/luisfurquim/stonelizard"
   "github.com/luisfurquim/etcdconfig"
   etcd "github.com/coreos/etcd/client"
)

func certKey(cert *x509.Certificate) string {
   var k string

   if len(cert.EmailAddresses)>0 {
      k = cert.EmailAddresses[0]
   } else {
      k = cert.Subject.CommonName
   }

   return fmt.Sprintf("%s_%2x_%2x", k, sha1.Sum(cert.AuthorityKeyId), sha1.Sum(cert.RawSubjectPublicKeyInfo))
}

func keyHash(issuer []byte) string {
   var h hash.Hash

   h = sha1.New()
   h.Write(issuer)

   return string(h.Sum(nil))
}

func (ck *CertKit) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   var cert       *x509.Certificate
   var User       *UserDB
   var issuer      string
   var subj        string
   var CertKey     string
   var found       bool
   var ok          bool
   var commonName  string
   var issuers     map[string]struct{}
   var users       map[string]*x509.Certificate

   Goose.Auth.Logf(5,"Peer certificates")
   found   = false
   issuers = map[string]struct{}{}
   users   = map[string]*x509.Certificate{}

   Goose.Auth.Logf(6,"Peer certificates: %#v",TLS.PeerCertificates)
   for _, cert = range TLS.PeerCertificates {
      if cert.IsCA {
         continue
      }
      Goose.Auth.Logf(6,"Peer certificate: %#v",cert)
      Goose.Auth.Logf(6,"Peer certificate: #%s, ID: %s, Issuer: %s, Subject: %s",cert.SerialNumber,cert.SubjectKeyId,cert.Issuer.CommonName,cert.Subject.CommonName)
      issuer = keyHash(cert.AuthorityKeyId)
      subj = keyHash(cert.SubjectKeyId)

      if _, ok = users[issuer]; ok {
         delete(users,issuer)
      }

      issuers[issuer] = struct{}{}

      if _, ok = issuers[subj]; !ok {
         users[subj] = cert
      }
   }

   Goose.Auth.Logf(6,"Peer certificates: %#v",TLS.PeerCertificates)
   for _, cert = range users {
      Goose.Auth.Logf(6,"Peer certificate.1: %#v",cert)
      Goose.Auth.Logf(6,"Peer certificate.1: #%s, ID: %s, Issuer: %s, Subject: %s",cert.SerialNumber,cert.SubjectKeyId,cert.Issuer.CommonName,cert.Subject.CommonName)
      CertKey = certKey(cert)
      if CertKey == "" {
         continue
      }
      if User, ok = ck.UserCerts[CertKey]; ok {
         if bytes.Equal(User.Cert.Raw,cert.Raw) {
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

   Goose.Auth.Logf(6,"Unauthorized access attempt, cert: %#v",cert)
   if cert == nil {
      Goose.Auth.Logf(1,"Unauthorized access attempt from nil@%s to path %s", RemoteAddr, path)
   } else {
      Goose.Auth.Logf(1,"Unauthorized access attempt from %s@%s to path %s", cert.Subject.CommonName, RemoteAddr, path)
      Goose.Auth.Logf(4,"Sent access cert (needs authorization): %#v",cert.Subject)
   }

   for commonName, _ = range ck.UserCerts {
      Goose.Auth.Logf(4,"access grantable to: %s",commonName)
   }

   if cert != nil {
      Goose.Auth.Logf(1,"SavePending :%#v",SavePending)
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

   // Code adapted from crypto/x509/root_unix.go
   roots = x509.NewCertPool()
   roots.AppendCertsFromPEM(ck.CACertPem)

   switch Access {
      case stonelizard.AccessNone:
         atype = tls.NoClientCert
      case stonelizard.AccessAuth, stonelizard.AccessAuthInfo:
         atype = tls.RequestClientCert
      case stonelizard.AccessVerifyAuth, stonelizard.AccessVerifyAuthInfo:
         atype = tls.RequireAndVerifyClientCert

/*
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
*/
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
   Goose.Auth.Logf(7,"X509KeyPair used: %#v",tlsConfig.Certificates[0])
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


// Stores the certificate in the authorization pending subtree
func (ck *CertKit) SavePending(cert *x509.Certificate) error {
   var err error
   var CertKey string
   var Pem string
   var tgtpath string

   CertKey = certKey(cert)
   Goose.Auth.Logf(3,"User certificate of %s not authorized", CertKey)
   Goose.Auth.Logf(7,"Certificate is %#v", cert)

   tgtpath = ck.Etcdkey + "/pending/" + CertKey

   _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), tgtpath, "", &etcd.SetOptions{Dir:true})
   if err != nil {
      Goose.Auth.Logf(1,"Error creating diretory for pending certificate (%s): %s",tgtpath,err)
      return err
   }

   Pem = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
   Goose.Auth.Logf(7,"Pem Certificate is %#v", Pem)
   err = etcdconfig.SetKey(ck.Etcdcli, tgtpath + "/cert", Pem)
   if err != nil {
      Goose.Auth.Logf(1,"Error saving pending certificate (%s): %s",tgtpath,err)
      return err
   }

   return err
}

//Transfer a user certificate from the pending subtree to the trusted subtree (so, enabling this user accesses)
func (ck *CertKit) Trust(id string) error {
   var err error
   var srcpath string
   var tgtpath string
   var etcdData interface{}

   srcpath = ck.Etcdkey + "/pending/" + id
   tgtpath = ck.Etcdkey + "/trusted/" + id

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, srcpath + "/cert")
   if err != nil {
      Goose.Auth.Logf(1,"Error retrieving pending user certificate for %s: %s", id, err)
      return err
   }

   Goose.Auth.Logf(7,"etcddata %s: %#v", id, etcdData)

   _, err = etcd.NewKeysAPI(ck.Etcdcli).Set(context.Background(), tgtpath, "", &etcd.SetOptions{Dir:true})
   if err != nil {
      Goose.Auth.Logf(1,"Error setting configuration, creating directory (%s): %s",tgtpath,err)
      return err
   }

   err = etcdconfig.SetKey(ck.Etcdcli, tgtpath + "/cert", etcdData.(string))
   if err != nil {
      Goose.Auth.Logf(1,"Error saving pending user certificate on trusted subtree for %s: %s", id, err)
      return err
   }

   err = etcdconfig.DeleteConfig(ck.Etcdcli, srcpath)
   if err != nil {
      Goose.Auth.Logf(1,"Error deleting pending user certificate for %s: %s", id, err)
      return err
   }

   return nil
}


//Remove a user certificate from the pending subtree (so, rejecting this user accesses)
func (ck *CertKit) Reject(id string) error {
   return ck.Delete("pending",id)
}

//Remove a user certificate from the trusted subtree (so, rejecting this user accesses)
func (ck *CertKit) Drop(id string) error {
   return ck.Delete("trusted",id)
}

//Remove a user certificate from the trusted subtree (so, rejecting this user accesses)
func (ck *CertKit) Delete(tree, id string) error {
   var err error
   var srcpath string

   srcpath = ck.Etcdkey + "/" + tree + "/" + id

   err = etcdconfig.DeleteConfig(ck.Etcdcli, srcpath)
   if err != nil {
      Goose.Auth.Logf(1,"Error deleting pending user certificate for %s: %s", id, err)
      return err
   }

   return nil
}


//List certificates from the pending subtree
func (ck *CertKit) GetPending() (map[string]interface{}, error) {
   var err error
   var etcdData interface{}

   if ck.Etcdcli == nil {
      err = errors.New("Error no etcd client initialized")
      Goose.Auth.Logf(1,"%s",err)
      return nil, err
   }

   if ck.Etcdkey == "" {
      err = errors.New("Error no etcd key provided")
      Goose.Auth.Logf(1,"%s", err)
      return nil, err
   }

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey)
   Goose.Auth.Logf(6,"etcdkey: %#v", etcdData)

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey + "/pending")
   if err != nil {
      Goose.Auth.Logf(1,"Error retrieving pending users certificates: %s", err)
      return nil, err
   }

   if etcdData == nil {
      return map[string]interface{}{}, nil
   }

   return etcdData.(map[string]interface{}), nil
}



//List certificates from the trusted subtree
func (ck *CertKit) GetTrusted() (map[string]interface{}, error) {
   var err error
   var etcdData interface{}

   if ck.Etcdcli == nil {
      err = errors.New("Error no etcd client initialized")
      Goose.Auth.Logf(1,"%s",err)
      return nil, err
   }

   if ck.Etcdkey == "" {
      err = errors.New("Error no etcd key provided")
      Goose.Auth.Logf(1,"%s", err)
      return nil, err
   }

   _, etcdData, err = etcdconfig.GetConfig(ck.Etcdcli, ck.Etcdkey + "/trusted")
   Goose.Auth.Logf(5,"etcdkey: %#v", etcdData)
   if err != nil {
      Goose.Auth.Logf(1,"Error retrieving trusted users certificates: %s", err)
      return nil, err
   }

   if etcdData == nil {
      return map[string]interface{}{}, nil
   }

   return etcdData.(map[string]interface{}), nil
}





