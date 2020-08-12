package stonelizard

import (
   "net"
   "reflect"
)

// From the endpoint, defined in the Service struct, init the variables and the server listeners.
// Besides, load de configuration file to start basic data required for the proposed solution.
func initSvc(svcElem EndPointHandler) (*Service, error) {
   var err     error
   var i       int
   var resp   *Service
   var cfg     Shaper
   var ls      net.Listener
//   var met     reflect.Method
   var met     reflect.Value
//   var ok      bool
   var typ     reflect.Type
   var fld     reflect.StructField
   var auth    AuthT

   resp = &Service{
//      AuthRequired: false,
      MatchedOps: map[int]int{},
   }

//   Goose.Initialize.Fatalf(0,"svcElem: %#v",svcElem)
   cfg, err = svcElem.GetConfig()
   if err != nil {
      Goose.Initialize.Logf(1,"Failed opening config: %s", err)
      return nil, err
   }

//   Goose.Initialize.Fatalf(0,"cfg: %#v",cfg)

   if cfg == nil {
      return nil, nil
   }

/*
   //TODO: shaper -> remover
   err = json.NewDecoder(cfg).Decode(&resp)

   if (err!=nil) && (err!=io.EOF) {
      Goose.Initialize.Logf(1,"Failed parsing config file: %s", err)
      return nil, err
   }
*/

   resp.Config = cfg

   ls, err = net.Listen("tcp", cfg.ListenAddress())
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.Listener, err = NewListener(ls)
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }


   ls, err = net.Listen("tcp", cfg.CRLListenAddress())
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating listener: %s", err)
      return nil, err
   }

   resp.CRLListener, err = NewListener(ls)
   if err != nil {
      Goose.Initialize.Logf(1,"Failed creating stoppable listener: %s", err)
      return nil, err
   }

   typ = reflect.ValueOf(svcElem).Type()
   if typ.Kind()==reflect.Ptr {
      typ = typ.Elem()
   }


//   met, ok = typ.MethodByName("SavePending")
//   if ok {
   met = reflect.ValueOf(resp.Config).MethodByName("SavePending")
   if !met.IsZero() {
      resp.SavePending = func(info interface{}) error {
         var err error
/*
         var this reflect.Value

         this = reflect.ValueOf(resp.Config)

         if this.Kind() == reflect.Ptr {
            Goose.Auth.Logf(1,"Dereferencing SavePending: %#v",this)
            this = this.Elem()
         }
*/
//         errIFace := met.Func.Call([]reflect.Value{this,reflect.ValueOf(info)})[0].Interface()
         errIFace := met.Call([]reflect.Value{reflect.ValueOf(info)})[0].Interface()
         switch errIFace.(type) {
            case error:
               err = errIFace.(error)
         }

         if err != nil {
            Goose.Auth.Logf(1,"Internal server error saving unauthorized access attempt info: %s",err)
            Goose.Auth.Logf(5,"Dump of info on internal server error saving unauthorized access attempt info: %#v",info)
         }
         return err
      }
   } else {
      resp.SavePending = func(interface{}) error {
         return nil
      }
   }

   Goose.Auth.Logf(5,"cfg: %#v",cfg)
   Goose.Auth.Logf(5,"cfg.CertKit(): %#v",cfg.CertKit())
   Goose.Auth.Logf(5,"auth: %#v",auth)
   Goose.Auth.Logf(5,"auth: %#v",reflect.TypeOf((*AuthT)(nil)).Elem())
   if cfg.CertKit()!=nil && reflect.TypeOf(cfg.CertKit()).Implements(reflect.TypeOf((*AuthT)(nil)).Elem()) {
      resp.Authorizer = cfg.CertKit()
   } else {
      resp.Authorizer = PublicAccessT{}
   }

   for i=0; i<typ.NumField(); i++ {
      fld = typ.Field(i)
      if fld.Anonymous && fld.Type.Implements(reflect.TypeOf((*AuthT)(nil)).Elem()) {
         resp.Authorizer = reflect.ValueOf(svcElem).Field(i).Interface().(AuthT)
         break
      }
   }

   return resp, err
}

