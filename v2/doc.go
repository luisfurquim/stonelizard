/*
Stonelizard v1 is a golang package that provides features to process HTTP requests on a Microservice architecture using REST,
Swagger and JSON technologies to communicate between services.

# Breaking change from v0

In v1, every method that implements a service operation MUST have context.Context as its first parameter
(after the receiver), followed by the parameters formally described in the struct field tags.

The context carries the request environment:
  - stonelizard.CertificateFromContext(ctx) — client X.509 certificate (*x509.Certificate), or nil
  - stonelizard.HostFromContext(ctx)         — HTTP Host header value (string)
  - stonelizard.RemoteAddrFromContext(ctx)   — client remote address (string)

This replaces the optional Env struct and the optional (host, remoteAddr string) trailing parameters
of v0, which caused runtime errors when the optional parameters were omitted or mis-ordered.


Example


package main

import (
   "os"
   "fmt"
   "flag"
   "context"
   "net/http"
   "github.com/luisfurquim/stonelizard/v2"
   "github.com/luisfurquim/goose"
)


//Set your microservice Service struct
type Service struct {
   // defines the root of this service, and its meta data.
   root bool `root:"/concurso/" consumes:"application/json" produces:"application/json" allowGzip:"true" enableCORS:"true"`

   // endpoint for create new researches
   research bool `method:"POST" path:"/research/{researchnum}/xy/{User}" postdata:"RegisterRequestT"`

   //ignored because do not have root tag or field method
   yy int
}

//Define your research struct
type ResearchRequestT struct {
   A int `json:"a"`
   B string `json:"b"`
}

//Define variable to stored the work directory path
var configPath string
```

func main() {
   var err        error
   var svc        Service


   flag.StringVar(&configPath,"conf", "./config.json", "Path to JSON configuration file")
   flag.Parse()

//If your use Goose package, define the log level necessary
   stonelizard.Goose = goose.Alert(1)


//Initialize the webservice using stonelizard.New() and ListenAndServerTLS() methods. The main flow is blocked waiting for new requests
   svc = Service{yy:6}
   ws, err := stonelizard.New(&svc)
   if err != nil {
      fmt.Printf("Error: %s\n",err)
      os.Exit(1)
   }

   err = ws.ListenAndServeTLS()
   if err != nil {
      fmt.Println(err)
   }

   os.Exit(1)
}

// In v1, context.Context is the mandatory first parameter of every service method.
// Use the helper functions to access the request environment from the context.
func (s *Service) Research(ctx context.Context, researchnum int, user string, data ResearchRequestT) stonelizard.Response {
   host       := stonelizard.HostFromContext(ctx)
   remoteAddr := stonelizard.RemoteAddrFromContext(ctx)
   cert       := stonelizard.CertificateFromContext(ctx) // nil if unauthenticated
   _ = cert
   return stonelizard.Response{
      Status: http.StatusOK,
      Body: fmt.Sprintf("Called research with %d and %s from %s (host %s). Data:A=%d/B=%s. Internal object data: %d\n",researchnum,user,remoteAddr,host,data.A,data.B,s.yy),
   }
}
*/
package stonelizard
