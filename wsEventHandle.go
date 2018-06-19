package stonelizard

import (
   "sync"
   "strings"
   "reflect"
   "golang.org/x/net/websocket"
)

// Creates the handles for each websocket event
func wsEventHandle(ws *websocket.Conn, codec websocket.Codec, obj interface{}, wg sync.WaitGroup, evHandlers map[string]*WSEventTrigger) {
   var i int
   var ev string
   var tag []string
   var tags [][]string
   var t reflect.Type
   var v reflect.Value
   var parmtypes []reflect.Type
   var object struct{}
   var evHandle *WSEventTrigger

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
      Goose.Serve.Logf(3,"Looking for event channel on field %s",t.Field(i).Name)

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
         evHandle = v.Field(i).Interface().(*WSEventTrigger)
         if evHandle == nil {
            Goose.Serve.Logf(1,"Warning event %s is nil, ignoring it", ev)
            continue
         }
         evHandlers[ev] = evHandle

         // Scans for the TYPES of the event trigger parameters.
         // Manual tag parsing (no reflect.StructTag facilities) because the tag names are not known.
         tags = tagRE.FindAllStringSubmatch(string(t.Field(i).Tag),-1)
         parmtypes = []reflect.Type{}
         for _, tag = range tags {
            if tag[1] != "doc" {
               switch tag[2] {
               case "string":
                  parmtypes = append(parmtypes,reflect.TypeOf(""))
               case "number":
                  parmtypes = append(parmtypes,reflect.TypeOf(float64(0.0)))
               case "integer":
                  parmtypes = append(parmtypes,reflect.TypeOf(int64(0)))
               case "boolean":
                  parmtypes = append(parmtypes,reflect.TypeOf(false))
               case "array":
                  parmtypes = append(parmtypes,reflect.TypeOf([]interface{}{}))
               case "object":
                  parmtypes = append(parmtypes,reflect.TypeOf(object))
               default:
                  Goose.Serve.Logf(1,"Error %s (%s), trigger not configured",WrongParameterType,tag[2])
                  continue EventTriggerScan
               }
            }
         }

         Goose.Serve.Logf(2,"Event %s handled by %#v", ev, evHandlers[ev])
         Goose.Serve.Logf(2,"Event %s data %#v", ev, evHandlers[ev].EventData)

         // Waitgroup control needed to avoid closing the websocket before all data is sent to the client.
         wg.Add(1)
         go func(c reflect.Value, name string, types []reflect.Type) {
            var ok bool
            var v reflect.Value
            var vtype reflect.Type
            var t reflect.Type
            var i int
            var err error

            defer wg.Done()

ExpectTrigger:
            for {
               // Wait for data sent by the websocket application layer (the event triggering)
               Goose.Serve.Logf(3,"Event comm loop will wait on channel")
               v, ok = c.Recv()
               Goose.Serve.Logf(4,"Event comm loop test %#v received %#v",ok, v.Interface())
               if !ok {
                  // End event triggering if wsStopEvent has closed the channel
                  return
               }

               // Check for compliance: the websocket application MUST send data as defined (length and type) in the StrucTag
               Goose.Serve.Logf(4,"Debug %T - %#v, %T - %#v", v, v, types, types)
               if len(v.Interface().([]interface{})) != len(types) {
                  Goose.Serve.Logf(1,"Error %s: len(param) == %d, len(types)==%d",WrongParameterLength, len(v.Interface().([]interface{})), len(types))
                  continue
               }

               Goose.Serve.Logf(4,"Will check trigger parameter types")
               for i, t = range types {
                  if t.Kind() == reflect.Ptr {
                     t = t.Elem()
                  }
                  Goose.Serve.Logf(4,"Got t: %#v",t)
                  vtype = v.Elem().Index(i).Elem().Type()
                  if vtype.Kind() == reflect.Ptr {
                     vtype = vtype.Elem()
                  }
                  Goose.Serve.Logf(4,"Got vtype: %#v",vtype)
                  if t != vtype {
                     Goose.Serve.Logf(4,"types differ (@%d), expected %s (Kind=%s) caught %s (Kind=%s)", i, t, t.Kind(), vtype, vtype.Kind())
                     if (t.Kind() != reflect.Array && t.Kind() != reflect.Slice) ||
                        (vtype.Kind() != reflect.Array && vtype.Kind() != reflect.Slice) ||
                         !vtype.Elem().Implements(t.Elem()) {
                        if !(t.Kind() == reflect.Struct && vtype.Kind() == reflect.Struct) &&
                           !(t.Kind() == reflect.Map && vtype.Kind() == reflect.Map) {
                           Goose.Serve.Logf(1,"Error %s (@%d), expected %s caught %s, ignoring this trigger",WrongParameterType,i,t,v.Elem().Index(i).Elem().Type())
                           continue ExpectTrigger
                        }
                     }
                  }
               }

               Goose.Serve.Logf(4,"Event comm Recv: %#v",v.Interface())
               // event name , event data
               err = codec.Send(ws, []interface{}{0, name, v.Interface()})
               if err != nil {
                  Goose.Serve.Logf(4,"Event trigger channel was closed when sending output on event %s: %#v", name, v.Interface())
//                  c.Close()
//                  return
               }
            }

         }(reflect.ValueOf(evHandlers[ev].EventData), ev, parmtypes)
      }
   }
   Goose.Serve.Logf(2,"Event channels for type %#v all configured",t.Name)
}


