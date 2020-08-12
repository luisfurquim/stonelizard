package stonelizard

import (
   "os"
   "net"
   "sync"
   "net/http"
   "crypto/tls"
//   "github.com/kr/pretty"
)

//Init a webserver and wait for http requests.
func (svc *Service) ListenAndServeTLS() error {
   var err      error
   var wg       sync.WaitGroup
   var crypls   net.Listener
   var hn       string
   var useHttps bool
   var useHttp  bool
   var unode    UrlNode
   var p        string
   var auth     AuthT
   var tc      *tls.Config
   var mux     *http.ServeMux
   var path     string
   var exported string

   // If the host/IP was not provided, get the hostname as provided by operating system
   if svc.Config.ListenAddress()[0] == ':' {
      hn, err = os.Hostname()
      if err!=nil {
         Goose.InitServe.Logf(1,"Error checking hostname: %s", err)
         return err
      }
   }

   // Checks if we use plain text, encryption or both
FindEncLoop:
   for _, unode = range svc.Svc {
      for _, p = range unode.Proto {
         if p == "https" || p == "wss" {
            useHttps = true
         }

         if p == "http" || p == "ws" {
            useHttp = true
         }

         if useHttp && useHttps {
            break FindEncLoop
         }
      }
   }


   // In case of encryption
   if useHttps {
      // Grab the AuthT compliant authorizer provided by the caller.
      auth = svc.Authorizer

      // Start controling how many goroutines will run.
      // We will need it later to wait all of them stop before we end the execution.
      wg.Add(2)

      // Optionally start a Certificate Revocation List server.
      // Used when we want to use self signed certificates.
      go func() {
         var errcrl error
         defer wg.Done()
         if svc.Config.CRLListenAddress() != "" {
            errcrl = auth.StartCRLServer(hn + svc.Config.CRLListenAddress(), svc.CRLListener)
            if errcrl != nil {
               Goose.InitServe.Logf(5,"Error starting CRL listener: %s", errcrl)
               err = errcrl
            } else {
               Goose.InitServe.Logf(5,"CRL Listen started listening")
            }
         }
      }()


      // Start the encrypted REST server
      go func() {
         var errsrv error

         defer wg.Done()

         // Get the configuration options for the encrypted service
         // svc.Access defines the level of authentication required, ranging from none to Edna Mode ;P
         tc, errsrv = auth.GetTLSConfig(svc.Access)
         if errsrv != nil {
            Goose.InitServe.Logf(1,"Failed setting up listener encryption parameters: %s",err)
            err = errsrv
            return
         }

         Goose.InitServe.Logf(2,"Adding http web service handler")
         mux = http.NewServeMux()
         mux.Handle("/",svc)

         for path, exported = range svc.SecureStatic {
            if path[len(path)-1] != '/' {
               path += "/"
            }
            Goose.InitServe.Logf(2,"Adding http file server handler on %s: %s", path, exported)
            mux.Handle(path,FileServerHandlerT{
               hnd:http.StripPrefix(path, http.FileServer(http.Dir(exported))),
               svc:svc,
               path:path,
               })
         }

         // Configure the server
         srv := &http.Server{
            Addr: hn + svc.Config.ListenAddress(),
            Handler: mux,
            TLSConfig: tc,
         }

         // Configure the listener
//         Goose.InitServe.Fatalf(0,"svc %# v: \n\n\n %# v", pretty.Formatter(svc), pretty.Formatter(svc.Listener))
//         Goose.InitServe.Fatalf(0,"svc.Listener %# v: \n\n\n srv.TLSConfig: %# v", pretty.Formatter(svc.Listener), pretty.Formatter(srv.TLSConfig))
         crypls = tls.NewListener(svc.Listener,srv.TLSConfig)

         // Start the server
         errsrv = srv.Serve(crypls)
         if errsrv != nil {
            Goose.InitServe.Logf(1,"Error starting service listener: %s", errsrv)
            err = errsrv
         } else {
            Goose.InitServe.Logf(2,"Service started listening")
         }
      }()

      // Synchronization point: wait for both servers (CRL and REST) to end
      wg.Wait()

   } else if useHttp { // TODO: provide 2 listeners and change this XOR to an OR
      // Start the plain text REST server

      Goose.InitServe.Logf(2,"Adding http web service handler")
      mux = http.NewServeMux()
      mux.Handle("/",svc)

      for path, exported = range svc.PlainStatic {
         if path[len(path)-1] != '/' {
            path += "/"
         }
         Goose.InitServe.Logf(2,"Adding http file server handler on %s: %s", path, exported)
         mux.Handle(path,FileServerHandlerT{
            hnd:http.StripPrefix(path, http.FileServer(http.Dir(exported))),
            svc:svc,
            path:path,
            })
      }

      // Configure the server
      srv := &http.Server{
         Addr: hn + svc.Config.ListenAddress(),
         Handler: mux,
      }

      // Start the server
      err = srv.Serve(svc.Listener)
      if err != nil {
         Goose.InitServe.Logf(5,"Error starting service listener: %s", err)
      } else {
         Goose.InitServe.Logf(5,"Service started listening")
      }
   }

   Goose.InitServe.Logf(1,"Ending listening")

   return err
}


