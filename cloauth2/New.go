package cloauth2

import (
   "golang.org/x/oauth2"
)

func New(cliId, cliSec, authURL, tokURL string) (*Oauth2T, error) {
   var oa Oauth2T
//   var e  error

   oa = Oauth2T{
      Config: &oauth2.Config{
         ClientID:     cliId,
         ClientSecret: cliSec,
//         Scopes:       []string{"profile"},
         Scopes:       []string{"profile","website","birthdate","gender","preferred_username","given_name","middle_name","locale","picture","zone_info","updated_at","nickname","name","family_name","address","phone_number_verified","phone_number"},
//         Scopes:       []string{"urn:netiq.com:nam:scope:oauth:registration:full","profile","address","phone","urn:netiq.com:nam:scope:oauth:registration:read","openid"},
         Endpoint: oauth2.Endpoint{
            AuthURL:  authURL,
            TokenURL: tokURL,
         },
      },
      Session: map[string]map[string]interface{}{},
   }

   return &oa, nil
}

