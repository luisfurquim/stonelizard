package stonelizard

import (
   "io"
   "net"
   "errors"
   "regexp"
   "net/http"
   "crypto/x509"
   "crypto/rsa"
   "github.com/luisfurquim/goose"
)

type StoppableListener struct {
  *net.TCPListener           // Wrapped listener
   stop             chan int // Channel used only to indicate listener should shutdown
}

type ServerCert struct {
   ServerCertPem, CACertPem []byte
   ServerCert,    CACert     *x509.Certificate
   ServerKeyPem,  CAKeyPem  []byte
   ServerKey                 *rsa.PrivateKey
   CAKey                     *rsa.PrivateKey
   CACRL                    []byte
}

type Unmarshaler interface {
   Decode(v interface{}) error
}

type Marshaler interface {
   Encode(v interface{}) error
}

type EndPointHandler interface {
   GetConfig() (io.Reader, error)
}

type Response struct {
   Status            int
   Header map[string]string
   Body              interface{}
}

type ResponseWriter interface {
   Header() http.Header
   WriteHeader(int)
}

type gzHttpResponseWriter struct {
   io.Writer
   ResponseWriter
}


type UrlNode struct {
   Path      string
   produces  string
   consumes  string
   allowGzip bool
   Matcher  *regexp.Regexp
   Handle    func ([]string, Unmarshaler) Response
}

type Service struct {
   PageNotFoundPath   string `json:"pageNotFound"`
   PageNotFound     []byte
   ListenAddress      string `json:"listen"`
   CRLListenAddress   string `json:"crllisten"`
   Svc              []UrlNode
   PemPath            string `json:"pem"`
   AuthRequired       bool
   AllowGzip          bool
   EnableCORS         bool
   Auth               ServerCert
   CertPool          *x509.CertPool
   UserCerts map[int]*x509.Certificate
   Listener          *StoppableListener
   CRLListener       *StoppableListener
}

var Goose goose.Alert
var ErrorStopped = errors.New("Stop signal received")