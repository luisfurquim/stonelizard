package stonelizard

import (
   "io"
   "os"
   "net/http"
   "crypto/tls"
   "crypto/rsa"
   "crypto/x509"
)

/*
 * Defines a dummy authorization type that satisfies the AuthT interface but makes no authentication at all
*/

func (pa PublicAccessT) Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error) {
   return http.StatusOK, nil, nil
}

func (pa PublicAccessT)  GetTLSConfig(Access uint8) (*tls.Config, error) {
   return nil, nil
}

func (pa PublicAccessT) StartCRLServer(listenAddress string, listener *StoppableListener) error {
   return nil
}


func (pa PublicAccessT) GetDNSNames() []string {
   var hn  string
   var err error
   hn, err = os.Hostname()
   if err != nil {
      return []string{"localhost"}
   }
   return []string{hn}
}

func (pa PublicAccessT) GetServerCert() *x509.Certificate {
   return nil
}

func (pa PublicAccessT) GetServerKey() *rsa.PrivateKey {
   return nil
}

func (pa PublicAccessT) GetCACert() *x509.Certificate {
   return nil
}

func (pa PublicAccessT) GetCAKey() *rsa.PrivateKey {
   return nil
}

func (pa PublicAccessT) GetServerX509KeyPair() tls.Certificate {
   return tls.Certificate{}
}

func (pa PublicAccessT) GetCertPool() *x509.CertPool {
   return nil
}

func (pa PublicAccessT) ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadCertificate(fname string) (*x509.Certificate, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadDecryptRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) ReadDecryptRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error) {
   return nil, nil, nil
}

func (pa PublicAccessT) Setup(udata map[string]interface{}) error {
   return nil
}

func (pa PublicAccessT) LoadUserData() error {
   return nil
}

func (pa PublicAccessT) Trust(id string) error {
   return nil
}

func (pa PublicAccessT) GetPending() (map[string]interface{}, error) {
   return map[string]interface{}{}, nil
}

func (pa PublicAccessT) GetTrusted() (map[string]interface{}, error) {
   return map[string]interface{}{}, nil
}

func (pa PublicAccessT) Reject(id string) error {
   return nil
}

func (pa PublicAccessT) Drop(id string) error {
   return nil
}

func (pa PublicAccessT) Delete(tree, id string) error {
   return nil
}

