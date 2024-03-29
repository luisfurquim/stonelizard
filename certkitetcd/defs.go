package certkitetcd

import (
   "time"
   "errors"
   "regexp"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
   "github.com/luisfurquim/goose"
//   etcd "github.com/coreos/etcd/client"
//   etcd "github.com/etcd-io/etcd/client/v2"
//   etcd "github.com/etcd-io/etcd/client/v3"
   etcd "go.etcd.io/etcd/client/v2"
)

type UserDB struct {
   Cert *x509.Certificate
}

type CertKit struct {
   Etcdcli                    etcd.Client
   Etcdkey                    string
   Path                       string
   ServerCertPem, CACertPem []byte
   ServerCert,    CACert     *x509.Certificate
   ServerKeyPem,  CAKeyPem  []byte
   ServerKey,     CAKey      *rsa.PrivateKey
   CACRL                    []byte
   CertPool                  *x509.CertPool
   UserCerts      map[string]*UserDB
   PendingCerts   map[string]*UserDB
   ServerX509KeyPair          tls.Certificate
   etcdCertKeyRE             *regexp.Regexp
   etcdDeleteKeyRE           *regexp.Regexp
   notAfterCA                 time.Time
   notAfterClient             time.Time
   notAfterServer             time.Time
}


type CertkitG struct {
   Generator goose.Alert `json:"Generator"`
   Loader    goose.Alert `json:"Loader"`
   Serve     goose.Alert `json:"Serve"`
   Auth      goose.Alert `json:"Auth"`
}

var Goose  CertkitG

var ErrorCertsMustHaveKeys = errors.New("Either provide both certificate and key or none of them")
var ErrorUnauthorized      = errors.New("Unauthorized access attempt")
var ErrorNoEtcdHandler     = errors.New("No etcd handler provided")
var ErrorNoEtcdKey         = errors.New("No etcd key provided")
var ErrorBadEtcdHandler    = errors.New("Bad etcd handler provided")
var ErrorBadEtcdKey        = errors.New("Bad etcd key provided")
var ErrorBadPEMBlock       = errors.New("Bad PEM block")
var ErrorValidDate         = errors.New("Failed certificate has expired or not yet valid date")



var ServerTime time.Duration = 365*24*time.Hour
var ClientTime time.Duration = 3650*24*time.Hour
var CaTime time.Duration = 365*24*20*time.Hour


