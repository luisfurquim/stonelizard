Stonelizard is a golang package that provides features to process HTTP requests on a Microservice architecture using REST, Swagger and JSON technologies to communicate between services.

[![GoDoc](https://godoc.org/github.com/luisfurquim/stonelizard?status.png)](http://godoc.org/github.com/luisfurquim/stonelizard)




From the HTTP request, the stonelizard finds which service must be activated and make the call the function 
that implements the service passing the parameters required by the function. 

To use stonelizard is necessary that the Service struct set the root field and at least one endpoint with the <method> tag.
If the Service struct does not have these fields, the service is ignored because it is from this pair, using regular expressions, 
the package defines what service / function should call.


In addition, the stonelizard package adopts digital certificate servers to ensure security and authentication between services.
To store the certificates,  must be manually created the working directory. The path of this directory is stored in PemPath field of Service struct.
The work directory must contain a client folder that stores the certificates of clients that ared allowed to connect to the stonelizard.


Another requirement to use the stonelizard is to create a configuration file. 
This file must be in JSON format and set at least five basic attributes: id, host, workdir, pem, listen and crllisten.
id: microservice identifier
host: server name where the microservice running
workdir: work directory path which are stored the certificates and keys
listen: port number on which a microservice connects with stonelizard (encrypted) 
crllisten: port number on which a microservice accesses the Certificate Revogation List - CRL (unencrypted)

Example of the configuration file:


```Go
{
  "id": "lizardry", 
  "host": ["localhost"], 
  "workdir": "/home/lizardry/work", 
  "pageNotFound": "/home/lizardry/work/notfound.html", //Directory with not found page are stored
  "pem": "/home/lizardry/work/auth", //Subdirectory from work directory which stored the keys and certificates
  "listen": ":1500",
  "crllisten": ":1501"
}
```



The package offers some methods to implements the basic infrastructure Microservice architecture:


```Go
func (svc *Service) ListenAndServeTLS() error
```

Init a webserver and wait for http requests.


```Go
func initSvc(svcElem EndPointHandler) (*Service, error)
```

From the endpoint, defined in the Service struct, init the variables, the certificates infrastructure and the server listeners.
Besides, load de configuration file to start basic data required for the proposed solution. 


```Go
func NewListener(l net.Listener) (*StoppableListener, error)
```


Init the listener for the service.



```Go
func (sl *StoppableListener) Accept() (net.Conn, error)
```

Implements a wrapper on the system accept.




```Go
func (sl *StoppableListener) Stop()
```

Stop the service and releases the chanel




```Go
func (svc Service) Close()
```

Close the webservices and the listeners




## Example:

In your microservice main code:



```Go
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

```


Set your microservice Service struct



```Go
type Service struct {
   // defines the root of this service, and its meta data.
   root bool `root:"/concurso/" consumes:"application/json" produces:"application/json" allowGzip:"true" enableCORS:"true"`

   // endpoint for create new researches 
   research bool `method:"POST" path:"/research/{researchnum}/xy/{User}" postdata:"RegisterRequestT"`

   //ignored because do not have root tag or field method
   yy int
}
```


Define your research struct



```Go
type ResearchRequestT struct {
   A int `json:"a"`
   B string `json:"b"`
}
```


Define variable to stored the work directory path



```Go
var configPath string
```

In your func main


```Go
func main() {
   var err        error
   var svc        Service


   flag.StringVar(&configPath,"conf", "./config.json", "Path to JSON configuration file")
   flag.Parse()
```

If your use Goose package, define the log level necessary




```Go
   Goose = goose.Alert(1)
   stonelizard.Goose = goose.Alert(1)
```

Initialize the webservice using stonelizard.New() and ListenAndServerTLS() methods. 
The main flow is blocked waiting for new requests



```Go
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
```


Outside your main code, you must set the function that implements the endpoint service.



```Go
func (s *Service) Research(researchnum int, user string, data ResearchRequestT) stonelizard.Response {
   return stonelizard.Response{
      Status: http.StatusOK,
      Body: fmt.Sprintf("Called research with %d and %s. Data:A=%d/B=%s. Internal object data: %d\n",researchnum,user,data.A,data.B,s.yy),
   }
}

```
