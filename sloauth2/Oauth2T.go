package sloauth2

import (
   "io"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
   "golang.org/x/oauth2"
   "github.com/luisfurquim/stonelizard"
)


func (oa *Oauth2T) Set(key string, val interface{}) error {
   return nil
}

func (oa *Oauth2T) Get(key string) (interface{}, error) {
   return nil, nil
}

func (oa *Oauth2T) Configure(key string, setter map[string]func(interface{})) {
   setter[key] = func(val interface{}) {
      var v map[string]interface{}

      v = val.(map[string]interface{})
      setter[key + "/RegEndPoint"](v["RegEndPoint"])
      setter[key + "/TokInfEndPoint"](v["TokInfEndPoint"])
      setter[key + "/UsrInfEndPoint"](v["UsrInfEndPoint"])
      setter[key + "/OIDMetaEndPoint"](v["OIDMetaEndPoint"])
      setter[key + "/RevokeEndPoint"](v["RevokeEndPoint"])
      setter[key + "/JSONWKSEndPoint"](v["JSONWKSEndPoint"])

      v = v["Config"].(map[string]interface{})
            setter[key + "/Config/ClientID"](v["ClientID"])
      setter[key + "/Config/ClientSecret"](v["ClientSecret"])
      setter[key + "/Config/RedirectURL"](v["RedirectURL"])
      setter[key + "/Config/Scopes"](v["Scopes"])

      v = v["Endpoint"].(map[string]interface{})
      setter[key + "/Config/Endpoint/AuthURL"](v["AuthURL"])
      setter[key + "/Config/Endpoint/TokenURL"](v["TokenURL"])
      setter[key + "/Config/Endpoint/AuthStyle"](v["AuthStyle"])
   }

   setter[key + "/RegEndPoint"] = func(val interface{}) {
      oa.RegEndPoint = val.(string)
   }

   setter[key + "/TokInfEndPoint"] = func(val interface{}) {
      oa.TokInfEndPoint = val.(string)
   }

   setter[key + "/UsrInfEndPoint"] = func(val interface{}) {
      oa.UsrInfEndPoint = val.(string)
   }

   setter[key + "/OIDMetaEndPoint"] = func(val interface{}) {
      oa.OIDMetaEndPoint = val.(string)
   }

   setter[key + "/RevokeEndPoint"] = func(val interface{}) {
      oa.RevokeEndPoint = val.(string)
   }

   setter[key + "/JSONWKSEndPoint"] = func(val interface{}) {
      oa.JSONWKSEndPoint = val.(string)
   }

   setter[key + "/Config/ClientID"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.ClientID = val.(string)
   }

   setter[key + "/Config/ClientSecret"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.ClientSecret = val.(string)
   }

   setter[key + "/Config/Endpoint/AuthURL"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.Endpoint.AuthURL = val.(string)
   }

   setter[key + "/Config/Endpoint/TokenURL"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.Endpoint.TokenURL = val.(string)
   }

   setter[key + "/Config/Endpoint/AuthStyle"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.Endpoint.AuthStyle = val.(oauth2.AuthStyle)
   }

   setter[key + "/Config/RedirectURL"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.RedirectURL = val.(string)
   }

   setter[key + "/Config/Scopes"] = func(val interface{}) {
      if val == nil {
         return
      }

      if oa.Config == nil {
         oa.Config = &oauth2.Config{}
      }
      oa.Config.Scopes = val.([]string)
   }


}

func (o *Oauth2T) StartCRLServer(listenAddress string, listener *stonelizard.StoppableListener) error {
   return o.CertKit.StartCRLServer(listenAddress, listener)
}

func (o *Oauth2T) GetDNSNames() []string {
   return o.CertKit.GetDNSNames()
}

func (o *Oauth2T) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   return o.CertKit.Authorize(path, parms, RemoteAddr, TLS, SavePending)
}

func (o *Oauth2T) GetServerCert() *x509.Certificate {
   return o.CertKit.GetServerCert()
}

func (o *Oauth2T) GetServerKey() *rsa.PrivateKey {
   return o.CertKit.GetServerKey()
}

func (o *Oauth2T) GetCACert() *x509.Certificate {
   return o.CertKit.GetCACert()
}

func (o *Oauth2T) GetCAKey() *rsa.PrivateKey {
   return o.CertKit.GetCAKey()
}

func (o *Oauth2T) GetServerX509KeyPair() tls.Certificate {
   return o.CertKit.GetServerX509KeyPair()
}

func (o *Oauth2T) GetCertPool() *x509.CertPool {
   return o.CertKit.GetCertPool()
}

func (o *Oauth2T) ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error) {
   return o.CertKit.ReadCertFromReader(r)
}

func (o *Oauth2T) ReadCertificate(fname string) (*x509.Certificate, []byte, error) {
   return o.CertKit.ReadCertificate(fname)
}

func (o *Oauth2T) ReadRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return o.CertKit.ReadRsaPrivKeyFromReader(r)
}

func (o *Oauth2T) ReadRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return o.CertKit.ReadRsaPrivKey(fname)
}

func (o *Oauth2T) ReadDecryptRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return o.CertKit.ReadDecryptRsaPrivKeyFromReader(r)
}

func (o *Oauth2T) ReadDecryptRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return o.CertKit.ReadDecryptRsaPrivKey(fname)
}

func (o *Oauth2T) Setup(udata map[string]interface{}) error {
   return o.CertKit.Setup(udata)
}

func (o *Oauth2T) LoadUserData() error {
   return o.CertKit.LoadUserData()
}

func (o *Oauth2T) AddUserData(usrKey string, ClientCert *x509.Certificate) error {
   return o.CertKit.AddUserData(usrKey, ClientCert)
}

func (o *Oauth2T) Trust(id string) error {
   return o.CertKit.Trust(id)
}

func (o *Oauth2T) Reject(id string) error {
   return o.CertKit.Reject(id)
}

func (o *Oauth2T) Drop(id string) error {
   return o.CertKit.Drop(id)
}

func (o *Oauth2T) Delete(tree, id string) error {
   return o.CertKit.Delete(tree, id)
}

func (o *Oauth2T) GetPending() (map[string]interface{}, error) {
   return o.CertKit.GetPending()
}

func (o *Oauth2T) GetTrusted() (map[string]interface{}, error) {
	Goose.Auth.Logf(4,"o.CertKit: %#v", o.CertKit)
   return o.CertKit.GetTrusted()
}

