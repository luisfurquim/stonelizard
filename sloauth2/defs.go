package sloauth2

import (
   "errors"
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


type Oauth2T struct {
   CertKit          stonelizard.AuthT `json:"CertKit"`
   RegEndPoint      string            `json:"RegEndPoint"`
   TokInfEndPoint   string            `json:"TokInfEndPoint"`
   UsrInfEndPoint   string            `json:"UsrInfEndPoint"`
   OIDMetaEndPoint  string            `json:"OIDMetaEndPoint"`
   RevokeEndPoint   string            `json:"RevokeEndPoint"`
   JSONWKSEndPoint  string            `json:"JSONWKSEndPoint"`
   Config          *oauth2.Config     `json:"Config"`
   UserProfileModel Profiler          `json:"-"`
   Session          map[string]map[string]interface{} `json:"-"`
}

type Oauth2G struct {
   Auth      goose.Alert `json:"Auth"`
}

var Goose  Oauth2G

var ErrorUnauthorized      = errors.New("Unauthorized access attempt")
var ErrorDuplicateFile     = errors.New("Error duplicate file")

var ckidchars []byte = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")
