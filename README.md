Stonelizard is a REST Services middleware golang package.

[![GoDoc](https://godoc.org/github.com/luisfurquim/stonelizard?status.png)](http://godoc.org/github.com/luisfurquim/stonelizard)

From the HTTP request, Stonelizard determines which method implements the operation. It handles parameter extraction and decoding providing data to the method call.

Setting up Stonelizard is done by calling stonelizard.New(), providing parameters of a type that satisfies the EndPointHandler interface.
The stonelizard.New() is a variadic function, allowing you to provide multiple EndPointHandler parameters.
Each EndPointHandler parameter must also be defined as a struct type whose fields conforms to the following definition:


```Go
type Service struct {
   // Required field. Reserved name. Defines the root of this service, and its meta data. Tags:
   // 1. root: defines the base root path of the service
   // 2. consumes: defines default accepted input mime types
   // 3. produces: defines default provided output mime types
   // 4. allowGzip: if true and the client request compression, stonelizard will automatically compress HTTP responses
   // 5. enableCORS: if true makes stonelizard provide CORS headers when the client requests CORS INFO through the OPTIONS HTTP method
   // 6. proto: either use http or https, other protocols, like ws and wss are in my TODO list, but don't expect it soon
   // 7. access: determines the level of authentication required, but it is not handled by Stonelizard itself. Instead, it is passed to the AuthT interface to handle it
   root    stonelizard.Void `root:"/myservice/" consumes:"application/json" produces:"application/json" allowGzip:"true" enableCORS:"true" proto:"https" access:"verifyauthinfo"`

   // Optional fields. Reserved names. Provides service information, the tags are used by the automatic swagger.json generator.
   info    stonelizard.Void `title:"MyGreatBot" description:"This is my great service bot!" tos:"The use of this bot is regulated by my great terms of service." version:"0.1"`
   contact stonelizard.Void `contact:"John Doe" url:"http://example.com" email:"john@doe"`
   license stonelizard.Void `license:"The title of the license terms chosen" url:"http://example.com/license"`

   // Operations to create/update/delete/fetch researches (the presence of the 'method' tag tells Stonelizard that this field is an operation definition)
   // 1. The field name MUST start with a lowercase letter and the method to be called when the operation is requested MUST have the exact same
   //    name, just changing the first letter to uppercase.
   // 2. The field type defines the type of the returned data. Use stonelizard.Void when there is no data to return. The returned data does not include the HTTP status code.
   // 3. Possible tags for operation definition:
   //    3.1. method: defines the HTTP method
   //    3.2. path: define the REST path, parameter variables are defined by enclosing the variable name between braces '{' and '}'
   //    3.3. header: optional tag to define parameter variables sent through HTTP headers, multiple variable names are defined as a comma separated list
   //    3.4. query:  optional tag to define parameter variables sent through path query part, multiple variable names are defined as a comma separated list
   //    3.4. postdata: optional tag to define parameter variables sent through HTTP POST body, multiple variable names are defined as a comma separated list
   //    3.5. accepted, created, ok: optional tags to define the corresponding custom HTTP status message
   //    3.6. doc: a textual description of what the operation does
   //    3.7. tags named after the parameter variable names: a textual description of the parameters
   // 4. All parameter variables are then passed to the corresponding methods. The signature of these methods must conform to which is defined by the tags
   //    4.1. The parameter order is important:
   //         4.1.1. You have to declare path parameters first and in the order they appear in the path.
   //         4.1.2. Then declare the header parameters, if there is multiple header variables, they have to be in the order that appears in the comma separated list
   //         4.1.3. Then declare the query parameters, if there is multiple query variables, they have to be in the order that appears in the comma separated list
   //         4.1.4. Then declare the post body parameters, if there is multiple body variables, they have to be in the order that appears in the comma separated list
   //         4.1.5. Lastly, you may optionally declare a parameter to receive authentication information. The type of this parameter must conform to whatever is returned
   //                by the Authorize method of the authorization system chosen by you
   newResearch int `method:"POST" path:"/research/{ResearchType}/user/{User}" header:"X-Trackid" postdata:"files" accepted:"Research registered" doc:"Use this operation to register new researchs on my great system. It returns the new research ID." X-Trackid:"A requester defined unique token to include in log messages making debugging easier." ResearchType:"The type of the research, valid values are 'x', 'y' and 'z'." User:"The user ID of the researcher." files:"Any uploaded documents related to the research."`
   dropResearch stonelizard.Void `method:"DELETE" path:"/research/{id}" header:"X-Trackid" id:"ID retornado por newResearch" ok:"Ok" X-Trackid:"ID único por request, para acompanhamento/debug via log" doc:"Removes the specified research from my great system"`
   getResearch ResearchT `method:"GET" path:"/research/{id}" header:"X-Trackid" query:"full" id:"ID retornado por newResearch" ok:"Ok" X-Trackid:"ID único por request, para acompanhamento/debug via log" doc:"Retrieves data from the specified research from my great system"`


   // Ignored by Stonelizard because they do not have a reserved name, nor the method tag, they make sense and are only useful to your application, not to Stonelizard
   someData int
   OtherData int
}
```

An earlier version deprecated requirement was json configuration files.
Now the EndPointHandler interface just defines a GetConfig() method that returns a Shaper interface compliant data.
This Shaper interface defines methods that retrieves the configuration needed by Stonelizard but left undefined in the EndPointHandler interface.
This separation allows applications to choose the place to store this configuration data
(the filesystem, command line parameters, environment variables, databases, whatever the application decides it is best for them).

In earlier versions, authentication was included in Stonelizard itself. Now there is an AuthT interface and an included PublicAccessT type that satisfies this interface.
As the name implies, this object makes no authentication at all, allowing anyone to access the services. In the subdirectories, you will find 2 packages, certkit and certkitetcd,
that provide authentication using x509 certificates. The first one, certkit, handles the certificates stored in the filesystem. The second, certkitetcd, handles certificates
stored in an etcd database. Feel free to use them or develop your own authentication system.


## Example:


```Go
.
.
.

import (
   ...
   "github.com/luisfurquim/stonelizard"
   "github.com/luisfurquim/stonelizard/certkit"
   ...
)


type Service struct {
   root    stonelizard.Void `root:"/myservice/" consumes:"application/json" produces:"application/json" allowGzip:"true" enableCORS:"true" proto:"https" access:"verifyauthinfo"`

   info    stonelizard.Void `title:"MyGreatBot" description:"This is my great service bot!" tos:"The use of this bot is regulated by my great terms of service." version:"0.1"`
   contact stonelizard.Void `contact:"John Doe" url:"http://example.com" email:"john@doe"`
   license stonelizard.Void `license:"The title of the license terms chosen" url:"http://example.com/license"`

   newResearch int `method:"POST" path:"/research/{ResearchType}/user/{User}" header:"X-Trackid" postdata:"files" accepted:"Research registered" doc:"Use this operation to register new researchs on my great system. It returns the new research ID." X-Trackid:"A requester defined unique token to include in log messages making debugging easier." ResearchType:"The type of the research, valid values are 'x', 'y' and 'z'." User:"The user ID of the researcher." files:"Any uploaded documents related to the research."`
   dropResearch stonelizard.Void `method:"DELETE" path:"/research/{id}" header:"X-Trackid" id:"ID retornado por newResearch" ok:"Ok" X-Trackid:"ID único por request, para acompanhamento/debug via log" doc:"Removes the specified research from my great system"`
   getResearch ResearchT `method:"GET" path:"/research/{id}" header:"X-Trackid" id:"ID retornado por newResearch" ok:"Ok" X-Trackid:"ID único por request, para acompanhamento/debug via log" doc:"Retrieves data from the specified research from my great system"`

   // Let's authenticate through the certkit interface
   ck *certkit.CertKit

   someData int
   OtherData int
}

type ResearchT struct {
   A int `json:"a"`
   B string `json:"b"`
}




// For example simplicity, we made the Service struct satisfy both the EndPointHandler and
// Shaper interfaces...
func (s Service) GetConfig() (stonelizard.Shaper, error) {
   return s, nil
}

func (s Service) PageNotFound() []byte {
   return []byte("<html><body>Page not found!</body></html>")
}

func (s Service) ListenAddress() string {
   return ":5000"
}

func (s Service) CRLListenAddress() string {
   return ":5001"
}


func (s Service) CertKit() stonelizard.AuthT {
   return s.ck
}



// The Service.NewResearch method is called when the Service.newResearch operation is requested.
// According to what was declared in the Service struct, 'ResearchType' and 'User' are parameters
// passed in the REST path, 'trackId' is a parameter passed through HTTP header, 'files' holds any
// upload file and authinfo contains the certificate provided by the authenticated client (as the
// access tag was set with the value verifyauthinfo, the certificate CA chain was already verified)
func (s *Service) NewResearch(ResearchType int, User string, trackId string, files []*multipart.FileHeader, authinfo *x509.Certificate) stonelizard.Response {

   // Do whatever your application needs to do in order to create a new research in your system,
   // for example:
   // a) store data in a persistent storage
   // b) set the newId local variable with the ID of the recently created research
   // c) authinfo may be used to log who created the research
   // d) trackId may be used to help tracking log messages from the same request for debugging

   if (some_error_occured) {
      // do some error handling
      return stonelizard.Response{
         Status: http.StatusInternalServerError, // HTTP status code to return
         Body: "My error message",
      }
   }

   return stonelizard.Response{
      // This HTTP status code to return when successful has its custom message defined by
      // the 'accepted' tag
      Status: http.StatusAccepted,
      Body: newId, // an int, as defined by the data type of the Service.newResearch field
   }
}

// The Service.DropResearch method is called when the Service.dropResearch operation is requested.
// According to what was declared in the Service struct, 'Id' is a parameter passed in the REST
// path, 'trackId' is a parameter passed through HTTP header and authinfo contains the certificate
// provided by the authenticated client (as the access tag was set with the value verifyauthinfo,
// the certificate CA chain was already verified)
func (s *Service) DropResearch(Id int, trackId string, authinfo *x509.Certificate) stonelizard.Response {

   // Do whatever your application needs to do in order to remove a research from your system,
   // for example:
   // a) remove data from a persistent storage
   // b) authinfo may be used to log who removed the research
   // c) trackId may be used to help tracking log messages from the same request for debugging

   if (some_error_occured) {
      // do some error handling
      return stonelizard.Response{
         Status: http.StatusInternalServerError, // HTTP status code to return
         Body: "My error message",
      }
   }

   return stonelizard.Response{
      Status: http.StatusOK,
   }

}


// The Service.GetResearch method is called when the Service.getResearch operation is requested.
// According to what was declared in the Service struct, 'Id' is a parameter passed in the REST
// path, 'trackId' is a parameter passed through HTTP header and, this time, no authinfo parameter
// was declared just to illustrate it is optional
func (s *Service) GetResearch(Id int, trackId string) stonelizard.Response {

   // Do whatever your application needs to do in order to retrieve a research from your system,
   // for example:
   // a) get data from a persistent storage
   // b) trackId may be used to help tracking log messages from the same request for debugging

   if (some_error_occured) {
      // do some error handling
      return stonelizard.Response{
         Status: http.StatusInternalServerError, // HTTP status code to return
         Body: "My error message",
      }
   }

   return stonelizard.Response{
      Status: http.StatusOK,
      Body: ResearchT{
         A: someRetrievedValue,
         B: anotherRetrievedValue,
      },
   }

}


func MyApp() {

   .
   .
   .

   ck, err = certkit.NewFromCK("/path/to/cert/dir")
   if err != nil {
      // handle the error
   }

   err = ck.LoadUserData(map[string]interface{}{})
   if err != nil {
      // handle the error
   }

   ws, err := stonelizard.New(&Service{someData:6})
   if err != nil {
      // handle the error
   }

   // Code execution is hold here as the ListenAndServeTLS method enters listen mode
   err = ws.ListenAndServeTLS()
   if err != nil {
      // handle the error
   }

   .
   .
   .

}


```

