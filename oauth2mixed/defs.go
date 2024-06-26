package oauth2mixed

import (
   "errors"
   "regexp"
   "crypto/x509"
   "golang.org/x/oauth2"
   "github.com/luisfurquim/goose"
   "github.com/luisfurquim/stonelizard"
)


type Profiler interface{
   New() Profiler
   Id() string
   Email() string
   Nick()  string
   Login() string
   Name() string
   SurName() string
   Avatar() string
}

type OptionsT struct {
	Secure bool
}


type Oauth2T struct {
   CertKit            stonelizard.AuthT `json:"CertKit"`
   RegEndPoint        string            `json:"RegEndPoint"`
   TokInfEndPoint     string            `json:"TokInfEndPoint"`
   UsrInfEndPoint     string            `json:"UsrInfEndPoint"`
   OIDMetaEndPoint    string            `json:"OIDMetaEndPoint"`
   RevokeEndPoint     string            `json:"RevokeEndPoint"`
   IntrospectEndPoint string            `json:"IntrospectEndPoint"`
   JSONWKSEndPoint    string            `json:"JSONWKSEndPoint"`
   Config            *oauth2.Config     `json:"Config"`
   UserProfileModel   Profiler          `json:"-"`
   Session            map[string]map[string]interface{} `json:"-"`
   SavePending        func(cert *x509.Certificate, parms ...interface{}) error
	Secure			    bool              `json:"secure"`
	ClientCredScope  []string            `json:"client_credentials_scope"`
}

type BearerT struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int `json:"expires_in"`
	Scope string `json:"scope"`
}

type Oauth2G struct {
   Auth      goose.Alert `json:"Auth"`
}

var Goose  Oauth2G

var ErrorUnauthorized      = errors.New("Unauthorized access attempt")
var ErrorDuplicateFile     = errors.New("Error duplicate file")


var ErrCreateHttpToken error = errors.New("Error creating token http request")
var ErrFetchingHttpToken error = errors.New("Error fetching token")
var ErrParsingToken error = errors.New("Error parsing token")


var ckidchars []byte = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

var reEMail *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
var reName *regexp.Regexp = regexp.MustCompile(`[nN][aAoO][mM][eEoO][sS]?`)

