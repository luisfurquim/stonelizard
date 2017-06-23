package stonelizard

import (
   "sync"
   "strings"
   "reflect"
   "golang.org/x/net/websocket"
)

// Creates the handles for each websocket event
func wsEventHandle(ws *websocket.Conn, codec websocket.Codec, obj interface{}, wg sync.WaitGroup, evHandlers map[string]*WSEventTrigger) {
   var i, j int
   var ev string
   var tag []string
   var tags [][]string
   var t reflect.Type
   var v reflect.Value
   var parmtypes []reflect.Type

   // De-reference the object, if needed
   v = reflect.ValueOf(obj)
   if v.Kind() == reflect.Ptr {
      v = v.Elem()
   }

   t = v.Type()
   Goose.Serve.Logf(2,"Looking for event channel for type %#v",t)
   // Scans the websocket definer struct object for definitions of event triggers
EventTriggerScan:
   for i=0; i<t.NumField(); i++ {
      Goose.Serve.Logf(2,"Looking for event channel on field %s",t.Field(i).Name)

      // Rules to define an event trigger:
      // a) the field MUST be of type WSEventTrigger
      // b) the field name MUST be exportable (UPPERCASE first letter)
      if t.Field(i).Type.AssignableTo(typeWSEventTrigger) {
         ev = t.Field(i).Name
         if ev[0:1] == strings.ToLower(ev[0:1]) {
            Goose.Serve.Logf(1,"Warning %s on %s, ignoring it",ErrorFieldIsOfWSEventTriggerTypeButUnexported, ev)
            continue
         }
         Goose.Serve.Logf(2,"Event setup for %s",ev)
         // stores the event handler in a global mapping
         evHandlers[ev] = v.Field(i).Interface().(*WSEventTrigger)

         // Scans for the TYPES of the event trigger parameters.
         // Manual tag parsing (no reflect.StructTag facilities) because the tag names are not known.
         tags = tagRE.FindAllStringSubmatch(string(t.Field(i).Tag),-1)
         parmtypes = make([]reflect.Type,len(tags))
         for j, tag = range tags {
            switch tag[2] {
            case "string":
               parmtypes[j] = reflect.TypeOf("")
            case "number":
               parmtypes[j] = reflect.TypeOf(float64(0.0))
            case "integer":
               parmtypes[j] = reflect.TypeOf(int64(0))
            case "boolean":
               parmtypes[j] = reflect.TypeOf(false)
            case "array":
               parmtypes[j] = reflect.TypeOf([]interface{}{})
            default:
               Goose.Serve.Logf(1,"Error %s (%s), trigger not configured",WrongParameterType,tags[2])
               continue EventTriggerScan
            }
         }

         // Waitgroup control needed to avoid closing the websocket before all data is sent to the client.
         wg.Add(1)
         go func(c reflect.Value, name string, types []reflect.Type) {
            var ok bool
            var v reflect.Value
            var t reflect.Type
            var i int

ExpectTrigger:
            for {
               // Wait for data sent by the websocket application layer (the event triggering)
               Goose.Serve.Logf(3,"Event comm loop will wait on channel")
               v, ok = c.Recv()
               Goose.Serve.Logf(4,"Event comm loop test %#v received %#v",ok, v.Interface())
               if !ok {
                  // End event triggering if wsStopEvent has closed the channel
                  wg.Done()
                  return
               }

               // Check for compliance: the websocket application MUST send data as defined (length and type) in the StrucTag
               if v.Len() != len(types) {
                  Goose.Serve.Logf(1,"Error %s: %d",WrongParameterLength, v.Len())
                  continue
               }

               for i, t = range types {
                  if t != v.Index(i).Type() {
                     Goose.Serve.Logf(1,"Error %s (@%d), expected %s caught %s, ignoring this trigger",WrongParameterType,i,t,v.Index(i).Type)
                     continue ExpectTrigger
                  }
               }

               Goose.Serve.Logf(1,"Event comm Recv: %#v",v.Interface())
               // callID , event data
               codec.Send(ws, []interface{}{name, v.Interface()})
            }

         }(reflect.ValueOf(evHandlers[ev].EventData), ev, parmtypes)
      }
   }
   Goose.Serve.Logf(1,"Event channels for type %#v all configured",t.Name)
}


