package oauth2mixed

import (
   "os"
   "io"
   "fmt"
   "bytes"
   "strings"
   "context"
   "net/url"
   "net/http"
   "crypto/x509"
//   "io/ioutil"
   "encoding/pem"
   "encoding/json"
   "crypto/x509/pkix"
   "golang.org/x/oauth2"
   "github.com/golang-jwt/jwt/v5"
   "github.com/luisfurquim/stonelizard"
   "github.com/luisfurquim/stonelizard/certkit"
   "github.com/luisfurquim/stonelizard/certkitetcd"
)

func (oa *Oauth2T) ExtAuthorize(ch chan stonelizard.ExtAuthorizeIn, path string, parms map[string]interface{}, resp http.ResponseWriter, req *http.Request, SavePending func(interface{}) error) (int, interface{}, error) {
   var in stonelizard.ExtAuthorizeIn
   var out stonelizard.ExtAuthorizeOut

   in = stonelizard.ExtAuthorizeIn{
      Path: path,
      Parms: parms,
      Resp: resp,
      Req: req,
      SavePending: SavePending,
      Out: make(chan stonelizard.ExtAuthorizeOut),
   }

   ch<- in
   out = <-in.Out
   return out.Stat, out.Data, out.Err
}


func (oa *Oauth2T) StartExtAuthorizer(authReq chan stonelizard.ExtAuthorizeIn) {
   var ck *http.Cookie
   var err error
   var oid string
   var state string
   var cliState string
   var ok bool
   var InstrospectFlow bool
   var cliCode string
   var tok *oauth2.Token
   var ctx context.Context
   var oaResp *http.Response
//   var buf []byte
   var rq *http.Request
   var pf Profiler
   var cert *x509.Certificate
   var key string
   var email string
   var certdata interface{}
   var trusted map[string]interface{}
   var certpem []byte
   var certIface interface{}
   var in stonelizard.ExtAuthorizeIn
	var msgMapTemplate interface{}

//   var path string
   var parms map[string]interface{}
   var parm interface{}
   var sparm string
   var resp http.ResponseWriter
   var req *http.Request
   var httpsStat int
//   var SavePending func(interface{}) error
	var bearer BearerT
   var hname string
   var body string

   hname, _ = os.Hostname()


main:
   for in = range authReq {
//      Goose.Auth.Logf(0,"1")
//      path = in.Path
      parms = in.Parms
      resp = in.Resp
      req = in.Req
//      SavePending = in.SavePending

		Goose.Auth.Logf(0,"1")

		ctx = context.Background()
		if oa.Session == nil {
			oa.Session = map[string]map[string]interface{}{}
		}

		Goose.Auth.Logf(0,"1a")

		parm, InstrospectFlow = parms["Authorization"]
		if InstrospectFlow {
			Goose.Auth.Logf(0,"1b")
			sparm, InstrospectFlow = parm.(string)
			if InstrospectFlow {
				Goose.Auth.Logf(0,"1c")
				if InstrospectFlow = strings.HasPrefix(sparm, "Bearer "); ok {
					Goose.Auth.Logf(0,"1d")
					sparm = sparm[7:]
				}
			}
		}

		Goose.Auth.Logf(0,"1e")

		if InstrospectFlow {
			Goose.Auth.Logf(0,"1f")
			tok = &oauth2.Token{AccessToken: sparm}
			oid = "__APP__"

			ck, err = req.Cookie("OID")
			if err != nil || ck.Value == "" {
				oa.NewSession(hname, resp)
			}

			if oa.Session[oid] == nil {
				oa.Session[oid] = map[string]interface{}{}
			}
			
		} else {

	      Goose.Auth.Logf(0,"2")
			ck, err = req.Cookie("OID")
			if err != nil || ck.Value == "" {
	//         Goose.Auth.Logf(0,"2A ck=%#v, err=%s, oa=%#v", ck, err, oa)
				oa.NewSession(hname, resp)
	//         Goose.Auth.Logf(0,"2A1 ck=%#v, err=%s, oa=%#v", ck, err, oa)
				in.Out<- stonelizard.ExtAuthorizeOut{
					Stat: http.StatusFound,
					Data: nil,
					Err: ErrorUnauthorized,
				}
				continue
			}

			oid = ck.Value

	//      Goose.Auth.Logf(0,"4A oid=%s, session=%#v", oid, oa.Session)

			if _, ok = oa.Session[oid]; !ok {
				oa.ReNewSession(oid, hname, resp)
	//         Goose.Auth.Logf(0,"4A1 oid=%s, session=%#v", oid, oa.Session)
				in.Out<- stonelizard.ExtAuthorizeOut{
					Stat: http.StatusFound,
					Data: nil,
					Err: ErrorUnauthorized,
				}
				continue
			}

	//      Goose.Auth.Logf(0,"4B parms: %#v", parms)

			if certIface, ok = oa.Session[oid]["cert"]; ok {
				if cert, ok = certIface.(*x509.Certificate); ok {
					in.Out<- stonelizard.ExtAuthorizeOut{
						Stat: 0,
						Data: cert,
						Err: nil,
					}
					continue
				}
			}

			cliCode, ok = parms["code"].(string)
			if !ok || cliCode=="" {
				state = MkCookieId()
	//         Goose.Auth.Logf(0,"Location: %s", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
				resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
				oa.Session[oid]["state"] = state
				in.Out<- stonelizard.ExtAuthorizeOut{
					Stat: http.StatusFound,
					Data: nil,
					Err: ErrorUnauthorized,
				}
				continue
			}

	//      Goose.Auth.Logf(0,"5")

			// preventing CSRF
			state, ok = oa.Session[oid]["state"].(string)
			if !ok {
				state = ""
			}

	//      Goose.Auth.Logf(0,"6")

			cliState, ok = parms["state"].(string)
			if !ok || cliState=="" || (state!="" && cliState!=state) {
				if state != "" {
					oa.Session[oid]["state"] = ""
				}

				in.Out<- stonelizard.ExtAuthorizeOut{
					Stat: http.StatusUnauthorized,
					Data: nil,
					Err: ErrorUnauthorized,
				}
				continue
			}

	//      Goose.Auth.Logf(0,"7")


		// claims_supported":["add","modify","delete","read","website","birthdate","gender","profile","preferred_username","given_name","middle_name","locale","picture","zone_info","updated_at","nickname","name","family_name","address","phone_number_verified","phone_number"]

			// Ok, let's get the token
			tok, err = oa.Config.Exchange(ctx, cliCode, oauth2.AccessTypeOffline)
			if err != nil {
				Goose.Auth.Logf(1,"oa.Config.Exchange error: %s", err)
				if state != "" {
					oa.Session[oid]["state"] = ""
				}

				in.Out<- stonelizard.ExtAuthorizeOut{
					Stat: http.StatusUnauthorized,
					Data: nil,
					Err: ErrorUnauthorized,
				}
				continue
			}
		}

//      Goose.Auth.Logf(0,"9")
		oa.SetCookie(oid, hname, resp)
		oa.Session[oid]["client"] = oa.Config.Client(ctx, tok)

		if InstrospectFlow {
			// token_type_hint": {"access_token"}IntrospectEndPoint
			rq, err = http.NewRequest("POST", strings.Split(oa.Config.Endpoint.TokenURL,"?")[0], bytes.NewReader([]byte(
				`client_id=` + oa.Config.ClientID +
				`&client_secret=` + oa.Config.ClientSecret +
				`&grant_type=client_credentials&scope=` + strings.Join(oa.Config.Scopes,","))))
			if err != nil {
				fmt.Printf("%s:%s\n", ErrCreateHttpToken, err)
				continue
			}
			rq.Header.Add("Content-Type", `application/x-www-form-urlencoded`)

			Goose.Auth.Logf(1,"--------------- TS 3 body:%s\n", `client_id=` + oa.Config.ClientID +
				`&client_secret=` + oa.Config.ClientSecret +
				`&grant_type=client_credentials&scope=` + strings.Join(oa.Config.Scopes,","))

			Goose.Auth.Logf(1,"--------------- TS 4 scopes: %#v\n", oa.Config.Scopes)

			oaResp, err = oa.Session[oid]["client"].(*http.Client).Do(rq)
			if err != nil {
				Goose.Auth.Logf(1,"%s:%s\n", ErrFetchingHttpToken, err)
				continue
			}
			defer oaResp.Body.Close()

			Goose.Auth.Logf(1,"--------------- TS 5\n")

			err = json.NewDecoder(oaResp.Body).Decode(&bearer)
			if err != nil {
				Goose.Auth.Logf(1,"%s:%s\n", ErrParsingToken, err)
				continue
			}

			body = `token=` + url.PathEscape(tok.AccessToken[7:])

			Goose.Auth.Logf(0,"bearer: %#v\n", bearer)
			Goose.Auth.Logf(0,"token: %#v\n", body)



			token, err := jwt.Parse(tok.AccessToken[7:], func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
//				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return oa.Config.ClientSecret, nil
			})
			if err != nil {
				Goose.Auth.Fatalf("Error decoding token: %s", err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				Goose.Auth.Logf(0,"%s: %s", claims["foo"], claims["nbf"])
			} else {
				Goose.Auth.Logf(0,"Error getting claims: %s", err)
			}


			rq, err = http.NewRequest("POST", oa.IntrospectEndPoint, strings.NewReader(body))
			if err != nil {
				Goose.Auth.Logf(1,"%s:%s\n", ErrCreateHttpToken, err)
				continue
			}
			rq.Header.Add("Content-Type", `application/x-www-form-urlencoded`)

		} else {
			rq, err = http.NewRequest("GET", oa.UsrInfEndPoint, nil)
		}

		rq.Header.Add("Authorization", `Bearer ` + bearer.AccessToken)
      oaResp, err = oa.Session[oid]["client"].(*http.Client).Do(rq)

		Goose.Auth.Logf(0,"request: %#v", rq)
		Goose.Auth.Logf(0,"request url: %s", rq.URL)

   //   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.UsrInfEndPoint)
      if err != nil || oaResp.Status[0] != '2' {
			Goose.Auth.Logf(0,"oaResp: %#v [%s]", *oaResp, err)
//         Goose.Auth.Logf(0,"Error contacting user information endpoint: %s", err)
			fmt.Sscanf(oaResp.Status, "%d", &httpsStat)
         in.Out<- stonelizard.ExtAuthorizeOut{
            Stat: httpsStat,
            Data: nil,
            Err: ErrorUnauthorized,
         }
         continue
      }

      buf := new(bytes.Buffer)
      io.Copy(buf, oaResp.Body)

		if oa.UserProfileModel == nil {
			err = json.Unmarshal(buf.Bytes(), &msgMapTemplate)
			if err == nil {
				pf = MinimalProfiler{}.Init(msgMapTemplate.(map[string]interface{}))
			}
		} else {
			pf = oa.UserProfileModel.New()
			err = json.NewDecoder(buf).Decode(pf)
		}
//       defer oaResp.Body.Close()
//      err = json.NewDecoder(oaResp.Body).Decode(pf)
   //

//		Goose.Auth.Logf(0, "Email: --%s--", pf.Email())

      email = strings.ToLower(pf.Email()) + "_"
//      Goose.Auth.Logf(0,"email: [%s]", email)
      trusted, err = oa.GetTrusted()
//      Goose.Auth.Logf(0,"trusted: [%#v]", trusted)
      if err == nil {
         Goose.Auth.Logf(4,"find cert")
         for key, certdata = range trusted {
            Goose.Auth.Logf(4,"cert-key: %s -- %s", key, email)
            if strings.HasPrefix(strings.ToLower(key), email) {
               block, _  := pem.Decode([]byte(certdata.(map[string]interface {})["cert"].(string)))
               cert, err  = x509.ParseCertificate(block.Bytes)
               if err == nil {
                  oa.Session[oid]["cert"] = cert
                  in.Out<- stonelizard.ExtAuthorizeOut{
                     Stat: 0,
                     Data: cert,
                     Err: nil,
                  }
                  continue main
               }
            }
         }
      }

      Goose.Auth.Logf(1,"%s not trusted! %s", email, err)

		switch ck := oa.CertKit.(type) {
		case *certkitetcd.CertKit:
			certpem, _, err = ck.GenerateClient(
				pkix.Name{CommonName: pf.Name() + " " + pf.SurName() + ":" + pf.Id()},
				pf.Email(),
				"")
		case *certkit.CertKit:
			certpem, _, err = ck.GenerateClient(
				pkix.Name{CommonName: pf.Name() + " " + pf.SurName() + ":" + pf.Id()},
				pf.Email(),
				"")
		}

      block, _  := pem.Decode(certpem)
      cert, err  = x509.ParseCertificate(block.Bytes)
      if err == nil {
			switch ck := oa.CertKit.(type) {
			case *certkitetcd.CertKit:
				if oa.SavePending != nil {
					err = oa.SavePending(cert, ck)
				} else {
					err = ck.SavePending(cert)
				}
			case *certkit.CertKit:
				if oa.SavePending != nil {
					err = oa.SavePending(cert, ck)
				} else {
					err = ck.SavePending(cert)
				}
			}
      }

      if err != nil {
         Goose.Auth.Logf(1,"Err generating certificate for user logged with oauth2? %s", err)
			in.Out<- stonelizard.ExtAuthorizeOut{
				Stat: 0,
				Data: nil,
				Err: ErrorUnauthorized,
			}
			continue
      }

		oa.Session[oid]["cert"] = cert
      in.Out<- stonelizard.ExtAuthorizeOut{
         Stat: 0,
         Data: cert,
         Err: nil,
      }
      continue
   }
}

func (oa *Oauth2T) SetCookie(oid string, hname string, resp http.ResponseWriter) {
   var ck *http.Cookie

   ck = &http.Cookie{
      Name: "OID",
      Value: oid,
      Path: "/",
//      Domain: hname,
//         Expires    time.Time // optional
      HttpOnly: true,
      MaxAge: 3600 * 24,
      Secure: oa.Secure,
	}

	if oa.Secure {
//      ck.SameSite = http.SameSiteStrictMode
//      ck.SameSite = http.SameSiteLaxMode
		ck.SameSite = http.SameSiteNoneMode
	}
	
   resp.Header().Add("Set-Cookie", ck.String())
}


func (oa *Oauth2T) ReNewSession(oid string, hname string, resp http.ResponseWriter) {
   var state string

   oa.SetCookie(oid, hname, resp)

   state = MkCookieId()
   oa.Session[oid] = map[string]interface{}{
      "state": state,
   }
   resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile+cpf+website+birthdate+gender+preferred_username+given_name+middle_name+locale+picture+zone_info+updated_at+nickname+name+family_name+address+phone_number_verified+phone_number")

//   Goose.Auth.Logf(0,"3a2 oa=%#v", oa)
}


func (oa *Oauth2T) NewSession(hname string, resp http.ResponseWriter) {
   var oid string
//   var state string

   oid = MkCookieId()
   oa.ReNewSession(oid, hname, resp)

/*
   oa.SetCookie(oid, hname, resp)

   state = MkCookieId()
   oa.Session[oid] = map[string]interface{}{
      "state": state,
   }
   resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile+cpf+website+birthdate+gender+preferred_username+given_name+middle_name+locale+picture+zone_info+updated_at+nickname+name+family_name+address+phone_number_verified+phone_number")

   Goose.Auth.Logf(0,"3a1 oa=%#v", oa)
*/

}
