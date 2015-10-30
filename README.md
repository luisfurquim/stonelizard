Stonelizard is a golang package that provides functionality to process HTTP requests based on a Microservice architecture using REST technologies, Swagger and JSON to communicate between services.

[![GoDoc](https://godoc.org/github.com/luisfurquim/stonelizard?status.png)](http://godoc.org/github.com/luisfurquim/stonelizard)




From the HTTP request, the stonelizard discover which service must be activated and make the call the function 
that implements the service passing the parameters required by the function. 
To use stonelizard is necessary that the Service struct set the root field and at least one endpoint with the <method> tag.
If the Service struct does not have these fields the service is ignored, because it is from this pair, using regular expressions, 
the package defines what service / function should call.


In addition, the stonelizard package adopts digital certificate servers to ensure security and authentication between services.
To store the certificates must be manually created the working directory. The path of this directory is stored in PemPath field of Service struct.
The working directory must contain a client folder that stores the certificates of clients allowed to connect to the stonelizard.



Another requirement to use the stonelizard is to create a configuration file. 
This file must be in JSON format and set at least four basic attributes: id, host, workdir, pem and listen.


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



