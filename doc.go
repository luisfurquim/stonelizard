/*
Stonelizard is a golang package that provides features to process HTTP requests on a Microservice architecture using REST, 
* Swagger and JSON technologies to communicate between services.


Example


package main

import (
   "os"
   "io"
   "fmt"
   "flag"
   "net/http"
   "github.com/luisfurquim/stonelizard"
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
   Goose = goose.Alert(1)
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

//Set the function that implements the endpoint service.
func (s *Service) Research(researchnum int, user string, data ResearchRequestT) stonelizard.Response {
   return stonelizard.Response{
      Status: http.StatusOK,
      Body: fmt.Sprintf("Called research with %d and %s. Data:A=%d/B=%s. Internal object data: %d\n",researchnum,user,data.A,data.B,s.yy),
   }
}
*/
package stonelizard
