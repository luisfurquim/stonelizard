package sloauth2

import (
   "strings"
   "context"
   "net/http"
   "crypto/x509"
//   "io/ioutil"
   "encoding/pem"
   "encoding/json"
   "crypto/x509/pkix"
   "golang.org/x/oauth2"
   "github.com/luisfurquim/stonelizard/certkitetcd"
)

func (oa *Oauth2T) ExtAuthorize(path string, parms map[string]interface{}, resp http.ResponseWriter, req *http.Request, SavePending func(interface{}) error) (int, interface{}, error) {
   var ck *http.Cookie
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

   Goose.Auth.Logf(0,"1")

   ctx = context.Background()
   if oa.Session == nil {
      oa.Session = map[string]map[string]interface{}{}
   }

   Goose.Auth.Logf(0,"2")

   ck, err = req.Cookie("OID")
   if err != nil || ck.Value == "" {
      Goose.Auth.Logf(0,"2A ck=%#v, err=%s, oa=%#v", ck, err, oa)
      oa.NewSession(resp)
      Goose.Auth.Logf(0,"2A1 ck=%#v, err=%s, oa=%#v", ck, err, oa)
      return http.StatusFound, nil, ErrorUnauthorized
   }

   oid = ck.Value

   Goose.Auth.Logf(0,"4A oid=%s, session=%#v", oid, oa.Session)

   if _, ok = oa.Session[oid]; !ok {
      oa.NewSession(resp)
      Goose.Auth.Logf(0,"4A1 oid=%s, session=%#v", oid, oa.Session)
      return http.StatusFound, nil, ErrorUnauthorized
   }

   Goose.Auth.Logf(0,"4B parms: %#v", parms)


   cliCode, ok = parms["code"].(string)
   if !ok || cliCode=="" {
      state = MkCookieId()
      Goose.Auth.Logf(0,"Location: %s", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
      resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile")
      oa.Session[oid]["state"] = state
      return http.StatusFound, nil, ErrorUnauthorized
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

      return http.StatusUnauthorized, nil, ErrorUnauthorized
   }

   Goose.Auth.Logf(0,"7")


// claims_supported":["add","modify","delete","read","website","birthdate","gender","profile","preferred_username","given_name","middle_name","locale","picture","zone_info","updated_at","nickname","name","family_name","address","phone_number_verified","phone_number"]

   // Ok, let's get the token
   tok, err = oa.Config.Exchange(ctx, cliCode, oauth2.AccessTypeOffline)
   if err != nil {
      if state != "" {
         oa.Session[oid]["state"] = ""
      }

      return http.StatusUnauthorized, nil, ErrorUnauthorized
   }

   oa.Session[oid]["client"] = oa.Config.Client(ctx, tok)

/*
   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.OIDMetaEndPoint)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }

   buf, err = ioutil.ReadAll(oaResp.Body)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }


   Goose.Auth.Logf(0,"OIDMetaEndPoint: %s", buf)


   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.JSONWKSEndPoint)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }

   buf, err = ioutil.ReadAll(oaResp.Body)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }


   Goose.Auth.Logf(0,"JSONWKSEndPoint: %s", buf)


   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.TokInfEndPoint)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }

   buf, err = ioutil.ReadAll(oaResp.Body)
   if err != nil {
      Goose.Auth.Fatalf(0,"Erro: %s", err)
   }



   Goose.Auth.Logf(0,"TokInfEndPoint: %s", buf)
//   Goose.Auth.Fatalf(0,"success: %#v", oaResp)
*/



   rq, err = http.NewRequest("GET", oa.UsrInfEndPoint, nil)
   rq.Header.Add("Authorization", `Bearer ` + tok.AccessToken)
   oaResp, err = oa.Session[oid]["client"].(*http.Client).Do(rq)

//   oaResp, err = oa.Session[oid]["client"].(*http.Client).Get(oa.UsrInfEndPoint)
   if err != nil {
      Goose.Auth.Logf(0,"Error contacting user information endpoint: %s", err)
      return 0, nil, ErrorUnauthorized
   }

   pf = oa.UserProfileModel.New()
   err = json.NewDecoder(oaResp.Body).Decode(pf)
//

   email = pf.Email() + "_"
   trusted, err = oa.GetTrusted()
   if err == nil {
      Goose.Auth.Logf(0,"find cert")
      for key, certdata = range trusted {
         Goose.Auth.Logf(0,"cert-key: %s -- %s", key, email)
         if strings.HasPrefix(key, email) {
            block, _  := pem.Decode([]byte(certdata.(map[string]interface {})["cert"].(string)))
            cert, err  = x509.ParseCertificate(block.Bytes)
            if err == nil {
               return 0, cert, nil
            }
         }
      }
   }

   Goose.Auth.Logf(0,"no trusted! %s", err)

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
   return 0, nil, ErrorUnauthorized
}

// var suicida int = 3

func (oa *Oauth2T) NewSession(resp http.ResponseWriter) {
   var ck *http.Cookie
   var oid string
   var state string

/*
   Goose.Auth.Logf(0,"3 suicida=%d, oa=%#v", suicida, oa)
   suicida--
   if suicida==0 {
      Goose.Auth.Fatalf(0,"suicidou-se")
   }
*/

   oid = MkCookieId()
   state = MkCookieId()
   ck = &http.Cookie{
      Name: "OID",
      Value: oid,
//         Path       string    // optional
//         Domain     string    // optional
//         Expires    time.Time // optional
      HttpOnly: true,
      MaxAge: 3600 * 24 * 365,
   }
   resp.Header().Add("Set-Cookie", ck.String())
//   Goose.Auth.Fatalf(0,"oa: %#v", *oa)
   resp.Header().Add("Location", oa.Config.AuthCodeURL(state, oauth2.AccessTypeOffline) + "&scope=profile+cpf+website+birthdate+gender+preferred_username+given_name+middle_name+locale+picture+zone_info+updated_at+nickname+name+family_name+address+phone_number_verified+phone_number")

   oa.Session[oid] = map[string]interface{}{
      "state": state,
   }

   Goose.Auth.Logf(0,"3a oa=%#v", oa)
}