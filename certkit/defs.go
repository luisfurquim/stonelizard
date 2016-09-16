package certkit

import (
   "errors"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
   "github.com/luisfurquim/goose"
)

/*
type CAServerT      []byte

type Shaper interface {
   CertPool()          *x509.CertPool
   UserCerts() map[int]*x509.Certificate
   ServerCert()        *x509.Certificate
   CACert()            *x509.Certificate
   ServerKey()         *rsa.PrivateKey
   CAKey()             *rsa.PrivateKey
   CACRL()              CAServerT // []byte
   ServerX509KeyPair()  tls.Certificate
}
*/

type CertKit struct {
   ServerCertPem, CACertPem []byte
   ServerCert,    CACert     *x509.Certificate
   ServerKeyPem,  CAKeyPem  []byte
   ServerKey,     CAKey      *rsa.PrivateKey
   CACRL                    []byte
   CertPool                  *x509.CertPool
   UserCerts      map[string]*x509.Certificate
   ServerX509KeyPair          tls.Certificate
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

var CertDirectories = []string{
   "/etc/ssl/certs",
}


