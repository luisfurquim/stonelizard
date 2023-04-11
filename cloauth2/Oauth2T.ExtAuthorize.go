package cloauth2

import (
   "os"
   "io"
   "bytes"
   "strings"
   "context"
   "net/http"
   "crypto/x509"
//   "io/ioutil"
   "encoding/pem"
   "encoding/json"
   "crypto/x509/pkix"
   "golang.org/x/oauth2"
   "github.com/luisfurquim/stonelizard"
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
//   var ck *http.Cookie
   var err error
   var oid string
   var state string
   var cliState string
   var ok bool
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
//   var certIface interface{}
   var in stonelizard.ExtAuthorizeIn

//   var path string
   var parms map[string]interface{}
   var resp http.ResponseWriter
   var req *http.Request
//   var SavePending func(interface{}) error

   var hname string

   hname, _ = os.Hostname()


main:
   for in = range authReq {
      Goose.Auth.Logf(0,"1")
//      path = in.Path
      parms = in.Parms
      resp = in.Resp
      req = in.Req
//      SavePending = in.SavePending

      ctx = context.Background()
      if oa.Session == nil {
         oa.Session = map[string]map[string]interface{}{}
      }

      cliCode, ok = parms["code"].(string)
      if !ok || cliCode=="" {
			oa.CheckCookie(req, resp, in.Out, hname)
			Goose.Auth.Logf(0,"4B parms: %#v", parms)
         continue
      }

      Goose.Auth.Logf(0,"5")

      // preventing CSRF
      state, ok = oa.Session[oid]["state"].(string)
      if !ok {
         state = ""
      }

      Goose.Auth.Logf(0,"6")

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

      Goose.Auth.Logf(0,"7")


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

      Goose.Auth.Logf(0,"9")
      oa.SetCookie(oid, hname, resp)
      oa.Session[oid]["client"] = oa.Config.Client(ctx, tok)


      rq, err = http.NewRequest("GET", oa.UsrInfEndPoint, nil)
      rq.Header.Add("Authorization", `Bearer ` + tok.AccessToken)
      oaResp, err = oa.Session[oid]["client"].(*http.Client).Do(rq)

   //   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.UsrInfEndPoint)
      if err != nil {
         Goose.Auth.Logf(0,"Error contacting user information endpoint: %s", err)
         in.Out<- stonelizard.ExtAuthorizeOut{
            Stat: 0,
            Data: nil,
            Err: ErrorUnauthorized,
         }
         continue
      }


      buf := new(bytes.Buffer)
      io.Copy(buf, oaResp.Body)


      var msgMapTemplate interface{}
      var msgMap map[string]interface{}
      Goose.Auth.Logf(0,"################################## Profile: --------------------------------------------")
      err = json.Unmarshal(buf.Bytes(), &msgMapTemplate)
      if err == nil {
         msgMap = msgMapTemplate.(map[string]interface{})
         for kprof, vprof := range msgMap {
            Goose.Auth.Logf(0,"Profile: %s -> %s", kprof, vprof)
         }
         Goose.Auth.Logf(0,"Profile: %s", buf)
      } else {
         Goose.Auth.Logf(0,"Profile error: %s", err)
      }

      pf = oa.UserProfileModel.New()
      err = json.NewDecoder(buf).Decode(pf)
//       defer oaResp.Body.Close()
//      err = json.NewDecoder(oaResp.Body).Decode(pf)
   //

      email = strings.ToLower(pf.Email()) + "_"
      trusted, err = oa.GetTrusted()
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

      certpem, _, err = oa.CertKit.(*certkitetcd.CertKit).GenerateClient(
         pkix.Name{CommonName: pf.Name() + " " + pf.SurName() + ":" + pf.Id()},
         pf.Email(),
         "")

      block, _  := pem.Decode(certpem)
      cert, err  = x509.ParseCertificate(block.Bytes)
      if err == nil {
         err = oa.CertKit.(*certkitetcd.CertKit).SavePending(cert)
      }

      if err != nil {
         Goose.Auth.Logf(1,"Err generating certificate for user logged with oauth2? %2", err)
      }

   /*
      cert = &x509.Certificate{
         Issuer: pkix.Name{
            CommonName: "StoneLizard OAuth2 Authorizer",
         },
         Subject: pkix.Name{
            CommonName: pf.Name() + " " + pf.SurName() + ":" + pf.Id(),
         },
         EmailAddresses: []string{pf.Email()},
      }
   */
   //   Goose.Auth.Fatalf(0,"UsrInfEndPoint: %#v", pf)
   //   Goose.Auth.Fatalf(0,"success: %#v", oaResp)
      in.Out<- stonelizard.ExtAuthorizeOut{
         Stat: 0,
         Data: nil,
         Err: ErrorUnauthorized,
      }
      continue
   }
}



func (oa *Oauth2T) CheckCookie(req *http.Request, resp http.ResponseWriter, out chan stonelizard.ExtAuthorizeOut, hname string) {
   var ck *http.Cookie
   var err error
   var oid string
   var ok bool
   var cert *x509.Certificate
   var certIface interface{}
   var state string

	Goose.Auth.Logf(0,"2")

	ck, err = req.Cookie("OID")
	if err != nil || ck.Value == "" {
		Goose.Auth.Logf(0,"2A ck=%#v, err=%s, oa=%#v", ck, err, oa)
		oa.NewSession(hname, resp)
		Goose.Auth.Logf(0,"2A1 ck=%#v, err=%s, oa=%#v", ck, err, oa)
		out<- stonelizard.ExtAuthorizeOut{
			Stat: http.StatusFound,
			Data: nil,
			Err: ErrorUnauthorized,
		}
		return
	}

	oid = ck.Value

	Goose.Auth.Logf(0,"4A oid=%s, session=%#v", oid, oa.Session)

	if _, ok = oa.Session[oid]; !ok {
		oa.ReNewSession(oid, hname, resp)
		Goose.Auth.Logf(0,"4A1 oid=%s, session=%#v", oid, oa.Session)
		out<- stonelizard.ExtAuthorizeOut{
			Stat: http.StatusFound,
			Data: nil,
			Err: ErrorUnauthorized,
		}
		return
	}

	if certIface, ok = oa.Session[oid]["cert"]; ok {
		if cert, ok = certIface.(*x509.Certificate); ok {
			out<- stonelizard.ExtAuthorizeOut{
				Stat: 0,
				Data: cert,
				Err: nil,
			}
			return
		}
	}

	state = MkCookieId()
	Goose.Auth.Logf(0,"Location: %s", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
	resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
	oa.Session[oid]["state"] = state
	out<- stonelizard.ExtAuthorizeOut{
		Stat: http.StatusFound,
		Data: nil,
		Err: ErrorUnauthorized,
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
      Secure: true,
//      SameSite: http.SameSiteStrictMode,
//      SameSite: http.SameSiteLaxMode,
      SameSite: http.SameSiteNoneMode,
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

   Goose.Auth.Logf(0,"3a2 oa=%#v", oa)
}


func (oa *Oauth2T) NewSession(hname string, resp http.ResponseWriter) {
   oa.ReNewSession(MkCookieId(), hname, resp)
}
