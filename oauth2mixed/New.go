package oauth2mixed

import (
   "golang.org/x/oauth2"
)

func New(cliId, cliSec, authURL, tokURL, Introspect string, opt ...OptionsT) (*Oauth2T, error) {
   var oa Oauth2T
//   var e  error
	var secure bool
	var o OptionsT

	if len(opt) == 0 {
		secure = true
	} else {
		secure = o.Secure
	}

   oa = Oauth2T{
		IntrospectEndPoint: Introspect,
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
      Secure: secure,
   }

   return &oa, nil
}

