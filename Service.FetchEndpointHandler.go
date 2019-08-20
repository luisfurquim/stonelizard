package stonelizard

import (
   "fmt"
   "strings"
   "net/http"
)

func (svc *Service) FetchEndpointHandler(proto, method, path string) (*UrlNode, []interface{}, map[string]interface{}, int) {
   var i, j int
   var pathId string
   var match []string
   var parms  []interface{}
   var authparms map[string]interface{}
   var endpoint UrlNode

   pathId = fmt.Sprintf("%s+%s:%s", strings.Split(proto,"/")[0], method, path)

   Goose.Serve.Logf(6,"Matcher will look on %#v", svc.Matcher)
   Goose.Serve.Logf(6,"Matcher will look for [%s]", pathId)
   match = svc.Matcher.FindStringSubmatch(pathId)
   Goose.Serve.Logf(0,"Matcher found this %#v\n", match) //6
   if len(match) == 0 {
      Goose.Serve.Logf(1,"Invalid service handler : %s", pathId)
      return nil, nil, nil, http.StatusBadRequest
   }

   parms = []interface{}{}
   authparms = map[string]interface{}{}

//   for _, endpoint = range svc.Svc {
   for i=1; i<len(match); i++ {
      Goose.Serve.Logf(6,"trying %s with endpoint:  %s",pathId,svc.Svc[svc.MatchedOps[i-1]].Path)
      if len(match[i]) > 0 {
         Goose.Serve.Logf(0,"Found endpoint %s for: %s",svc.Svc[svc.MatchedOps[i-1]].Path,pathId)//4
         endpoint = svc.Svc[svc.MatchedOps[i-1]]
         for j=i+1; (j<len(match)) && (len(match[j])>0); j++ {
            Goose.Serve.Logf(0,"j=%d, i=%d, endpoint.ParmNames: %#v, authparms: %#v", j, i, endpoint.ParmNames, authparms)
            authparms[endpoint.ParmNames[j-i-1]] = match[j]
         }
         for k := i+1; k<j; k++ { // parms = []interface{}(match[i+1:j])
            parms = append(parms,match[k])
         }
         j -= i + 1
         break
      }
   }

   return &endpoint, parms, authparms, 0
}

