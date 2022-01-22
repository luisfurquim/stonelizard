package stonelizard

import (
   "io"
   "net"
   "bytes"
   "errors"
   "regexp"
   "reflect"
   "net/http"
   "crypto/rsa"
   "crypto/tls"
   "crypto/x509"
   "mime/multipart"
   "crypto/x509/pkix"
   "golang.org/x/net/websocket"
   "github.com/luisfurquim/goose"
   "github.com/luisfurquim/strtree"
)

type Void struct{}
type Mod struct{}

type ExtAuthorizeIn struct {
   Path string
   Parms map[string]interface{}
   Resp http.ResponseWriter
   Req *http.Request
   SavePending func(interface{}) error
   Out chan ExtAuthorizeOut
}

type ExtAuthorizeOut struct {
   Stat int
   Data interface{}
   Err error
}


type Static struct {
   w io.Writer
}

type StoppableListener struct {
  *net.TCPListener           // Wrapped listener
   stop             chan int // Channel used only to indicate listener should shutdown
}

type Unmarshaler interface {
   Decode(v interface{}) error
}

type Marshaler interface {
   Encode(v interface{}) error
}

type MultipartUnmarshaler struct {
   form     *multipart.Form
   fields  []string
   index     int
}

type Base64Unmarshaler struct {
   r        io.Reader
}

type DummyUnmarshaler struct {
   r        io.Reader
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

type WSocketEventRegistry []*websocket.Conn

type HandleHttpFn func ([]interface{}, Unmarshaler, interface{}, string, string) Response // Operation handler function, it calls the method defined by the application
type HandleWsFn   func ([]interface{}, Unmarshaler) Response // Operation handler function, it calls the method defined by the application

type WSocketOperation struct {
   ParmNames     []string          // List of parameter names
   CallByRef       bool            // Flags methods whoose receiver is a pointer
   Method          reflect.Method  // Method to handle the operation
}

type WSEventTrigger struct {
   EventData chan interface{}
   Status bool
}

type UrlNode struct {
   Path               string       // The URL path with input parameters
   produces           string       // The possible mime-types produced by the operation for output
   consumes           string       // The possible mime-types required by the operation for input
   Proto            []string       // Valid protocols: http, https, ws, wss
   allowGzip          bool         // If true AND the client requests with the HTTP header, then we gzip the communication
   Headers          []string       // List of HTTP header parameters required by the operation for input
   Query            []string       // List of query parameters required by the operation for input
   Body             []string       // List of HTTP body parameters required by the operation for input
   ParmNames        []string       // Complete list of parameter names (from path, headers, query and body)
   Handle             HandleHttpFn
   WSocketOperations *strtree.Node // If this.Proto contains ws | wss, then store here the suboperations defined by the methods of the type returned by the main handlers
   WSocketEvents     *strtree.Node // If this.Proto contains ws | wss, then store here the events defined by the event fields of the type returned by the main handlers
/*
   Access          uint8   // Defined in the access tag.
                           // Possible values:
                           //    'none'=public unauthenticated access
                           //    'auth'=authenticated access (through the AuthT interface)
                           //    'authinfo'=authenticated access (through the AuthT interface) and pass the data output parameter generated by the authorizer to the operation method
*/
}

// AuthT interface defines an authorizer function
// To use it put an anonymous field that satisfies the AuthT interface in the EndPointHandler provided to stonelizard.New
// Input (
//    @method: HTTP method
//    @path: the path part of the URL of the operation as it appears in your EndPointHandler definition
//    @parms: the key is the parameter name as defined in your EndPointHandler definition, the value is the one sent by the requesting client
//    @RemoteAddr: the remote client address
//    @TLS: the TLS connection information received from the client, may be nil
//    @SavePending: function to save pending authorization information. It receives one interface{} argument with the info to be saved for later 3rd party analysis.
//                  It may be just NOOP function, Authorize implementations SHOULD NOT trust that the info is really being saved.
// )
// Output (
//    @httpstat: HTTP status code (used only if error is not nil)
//    @data: if the operation has a tag 'access' with value 'authinfo', this data is passed to the operation
//           method as the last parameter (remember to declare it as interface{}). This setting is not
//           added to the swagger.json generated or any other generated service description.
//    @err: error status (if it is not nil, the operation method is not called and the error message is sent to the client, along with the http status code)
// )
type AuthT interface {
   GetTLSConfig(Access uint8) (*tls.Config, error)
   StartCRLServer(listenAddress string, listener *StoppableListener) error
   GetDNSNames() []string
   Authorize(path string, parms map[string]interface{}, RemoteAddr string, TLS *tls.ConnectionState, SavePending func(interface{}) error) (httpstat int, data interface{}, err error)
   GetServerCert() *x509.Certificate
   GetServerKey() *rsa.PrivateKey
   GetCACert() *x509.Certificate
   GetCAKey() *rsa.PrivateKey
   GetServerX509KeyPair() tls.Certificate
   GetCertPool() *x509.CertPool
   ReadCertFromReader(r io.Reader) (*x509.Certificate, []byte, error)
   ReadCertificate(fname string) (*x509.Certificate, []byte, error)
   ReadRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error)
   ReadRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error)
   ReadDecryptRsaPrivKeyFromReader(r io.Reader) (*rsa.PrivateKey, []byte, error)
   ReadDecryptRsaPrivKey(fname string) (*rsa.PrivateKey, []byte, error)
   Setup(udata map[string]interface{}) error
   LoadUserData() error
   AddUserData(usrKey string, ClientCert *x509.Certificate) error
   Trust(id string) error
   Reject(id string) error
   Drop(id string) error
   Delete(tree, id string) error
   GetPending() (map[string]interface{}, error)
   GetTrusted() (map[string]interface{}, error)
}

// ExtAuthT interface defines an extended authorizer function.
// If stonelizard detects that your authorizer also satisfies this interface, then
// ExtAuthorize will be used INSTEAD of Authorize.
// Input (
//    @ch: channel to send authorizer request data
//    @path: the path part of the URL of the operation as it appears in your EndPointHandler definition
//    @parms: the key is the parameter name as defined in your EndPointHandler definition, the value is the one sent by the requesting client
//    @request: the entire http request object
//    @SavePending: function to save pending authorization information. It receives one interface{} argument with the info to be saved for later 3rd party analysis.
//                  It may be just NOOP function, Authorize implementations SHOULD NOT trust that the info is really being saved.
// )
// Output (
//    @httpstat: HTTP status code (used only if error is not nil)
//    @data: if the operation has a tag 'access' with value 'authinfo', this data is passed to the operation
//           method as the last parameter (remember to declare it as interface{}). This setting is not
//           added to the swagger.json generated or any other generated service description.
//    @err: error status (if it is not nil, the operation method is not called and the error message is sent to the client, along with the http status code)
// )
type ExtAuthT interface {
   ExtAuthorize(ch chan ExtAuthorizeIn, path string, parms map[string]interface{}, resp http.ResponseWriter, req *http.Request, SavePending func(interface{}) error) (httpstat int, data interface{}, err error)
   StartExtAuthorizer(authReq chan ExtAuthorizeIn)
}

type EndPointHandler interface {
   GetConfig() (Shaper, error)
}

// Shaper interface has an optional method: SavePending(cert *x509.Certificate) error
type Shaper interface {
   PageNotFound()     []byte
   ListenAddress()      string
   CRLListenAddress()   string
   CertKit()            AuthT
}

type Service struct {
   Matcher           *regexp.Regexp
   MatchedOps map[int]int
   Svc              []UrlNode
   Config             Shaper
   AuthRequired       bool
   AllowGzip          bool
   EnableCORS         bool
   Proto            []string
   Listener          *StoppableListener
   CRLListener       *StoppableListener
   Swagger           *SwaggerT
   Access             uint8
   Authorizer         AuthT
   SavePending        func(interface{}) error
   PlainStatic     map[string]string
   SecureStatic    map[string]string
   ch            chan ExtAuthorizeIn
   SwaggerPath        string
}

type Pki interface{
   GenerateClientCSR(subject pkix.Name, email string) ([]byte, error)
   GenerateClient(asn1Data []byte) (*x509.Certificate, *rsa.PublicKey, error)
   NewPemCertReqFromReader(rd io.Reader) error
   NewPemCertFromMemory(buf []byte) error
   NewPemCertFromReader(rd io.Reader) error
   NewPemCertFromFile(fname string) error
   NewPemKeyFromMemory(buf []byte, password string) error
   NewPemKeyFromReader(rd io.Reader, password string) error
   NewPemKeyFromFile(fname string, password string) error
   PemKey(password string) ([]byte, error)
   PemKeyToFile(fname, password string) error
   PemCsr(der []byte, fname string) error
   NewPemCert(fname string) error
   Sign(msg string) ([]byte, error)
   Verify(msg string, signature []byte) error
   Encrypt(msg []byte) ([]byte, error)
   Decrypt(secret []byte) ([]byte, error)
   Challenge() ([]byte, []byte, error)
   QrKeyId(keyId string, challenge []byte) ([]byte, error)
   FindCertificate(keyId string) (Pki, string, error)
   Certificate() []byte
}

type ServiceWithPKI struct {
   Service
   PK Pki
}

type ReadCloser struct {
   Rd *bytes.Reader
}

type ResponseT struct {
   Description string
   TypeReturned interface{}
}

type FileServerHandlerT struct {
   hnd http.Handler
   svc *Service
   path string
}

type PublicAccessT struct {}

const (
   AccessNone uint8 = iota
   AccessAuth
   AccessAuthInfo
   AccessVerifyAuth
   AccessVerifyAuthInfo
   AccessInfo
)

const StatusTrigEvent = 275
const StatusTrigEventDescription = "Trig Event"

var voidType = reflect.TypeOf(Void{})
var ModType = reflect.TypeOf(Mod{})
var float64Type = reflect.TypeOf(float64(0))
var MaxUploadMemory int64 = 16 * 1024 * 1024
var gorootRE *regexp.Regexp
var gosrcRE *regexp.Regexp
var gosrcFNameRE *regexp.Regexp
var tagRE *regexp.Regexp
var aggrIndentifierRE *regexp.Regexp = regexp.MustCompile(`[^\[\]\{\}]+([\[\]\{\}]+)`)

var ErrorStopped = errors.New("Stop signal received")
var ErrorParmListSyntax = errors.New("Syntax error on parameter list")
var ErrorDescriptionSyntax = errors.New("Syntax error on response description")
var ErrorInvalidNilParam = errors.New("Syntax error nil parameter not allowed in this context")
var ErrorWrongParameterCount = errors.New("Wrong parameter count")
var ErrorInvalidParameterType = errors.New("Invalid parameter type")
var ErrorMissingRequiredHTTPHeader = errors.New("Missing required HTTP header")
var ErrorMissingRequiredQueryField = errors.New("Error missing required query field")
var ErrorMissingRequiredPostBodyField = errors.New("Error missing required post body field")
var ErrorWrongAuthorizerReturnValues = errors.New("Error wrong authorizer return values")
var ErrorInvalidProtocol = errors.New("Invalid protocol")
var ErrorMixedProtocol = errors.New("Mixed http/https with ws/wss")
var ErrorWrongHandlerKind = errors.New("Wrong handler kind")
var WrongParameterLength = errors.New("Wrong parameter length")
var WrongParameterType = errors.New("Wrong parameter type")
var MapParameterEncodingError = errors.New("Map parameter encoding error")
var ErrorInvalidType = errors.New("Invalid type")
var ErrorConversionOverflow = errors.New("Conversion overflow")
var ErrorDecodeError = errors.New("Decode error")
var ErrorStopEventTriggering = errors.New("Stop event triggering")
var ErrorEndEventTriggering = errors.New("End event triggering")
var ErrorMissingWebsocketInTagSyntax = errors.New("Syntax error missing websocket 'in' tag")
var ErrorNoRoot = errors.New("No service root specified")
var ErrorServiceSyntax = errors.New("Syntax error on service definition")
var ErrorCannotWrapListener = errors.New("Cannot wrap listener")
var ErrorFieldIsOfWSEventTriggerTypeButUnexported = errors.New("Field is of Event Trigger type, but it is not exported")
var ErrorUndefMod = errors.New("Undefined module")
var ErrorUndefPropType = errors.New("Undefined property type")

type StonelizardG struct {
   Listener     goose.Alert
   Swagger      goose.Alert
   Initialize   goose.Alert
   OpHandle     goose.Alert
   ParseFields  goose.Alert
   New          goose.Alert
   InitServe    goose.Alert
   Serve        goose.Alert
   Auth         goose.Alert
}

var Goose StonelizardG
var WebSocketResponse Response
var dummyWSEventTrigger *WSEventTrigger
var typeWSEventTrigger reflect.Type = reflect.TypeOf(dummyWSEventTrigger)
var isBase64DataURL *regexp.Regexp = regexp.MustCompile(`^data:[a-zA-Z0-9]+/[a-zA-Z0-9]+;base64,(.*)$`)

