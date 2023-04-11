package cloauth2

import (
   "crypto/tls"
   "github.com/luisfurquim/stonelizard/certkit"
   "github.com/luisfurquim/stonelizard/certkitetcd"
)

func (ck *Oauth2T) GetTLSConfig(Access uint8) (*tls.Config, error) {
   var tlsConfig *tls.Config

   tlsConfig = &tls.Config{
      ClientAuth: tls.NoClientCert,
//      InsecureSkipVerify: true,
      Certificates: make([]tls.Certificate, 1),
   }

   switch c := ck.CertKit.(type) {
   case *certkit.CertKit:
      tlsConfig.Certificates[0] = c.ServerX509KeyPair
   case *certkitetcd.CertKit:
      tlsConfig.Certificates[0] = c.ServerX509KeyPair
   }

   Goose.Auth.Logf(5,"X509KeyPair used: %#v",tlsConfig.Certificates[0])
   tlsConfig.BuildNameToCertificate()

   return tlsConfig, nil
}

