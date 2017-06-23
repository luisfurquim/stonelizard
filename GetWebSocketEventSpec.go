package stonelizard

import (
   "reflect"
)

// Returns a swagger definition of all the events defined by type of the struct field "field"
func GetWebSocketEventSpec(field reflect.StructField, WSMethodName string, WSMethod reflect.Method) (*SwaggerWSEventT, error) {
   var i                     int
   var docndx                int
   var parmcount             int
   var event                *SwaggerWSEventT
   var tags              [][]string
   var tag                 []string
   var desc1, desc2          string
   var fld                   reflect.StructField

   if field.Type.Kind() != reflect.Struct {
      Goose.Swagger.Logf(1,"Error %s scanning for websocket events defined by %s, it must be of struct kind",ErrorInvalidType,WSMethodName)
      return nil, ErrorInvalidType
   }

   desc1  = "Data sent to the client when the event triggers"
   desc2  = "Event data"
   docndx = -1

   // Scans the struct for event trigger definitions
   // Event trigger definitions are fields satisfying these rules:
   // a) MUST be of type WSEventTrigger
   // b) MUST be exported symbol (uppercase first letter)
   for i=0; i<field.Type.NumField(); i++ {
      fld = field.Type.Field(i)

      if fld.Name[0:1] == strings.ToLower(fld.Name[0:1]) || !fld.Type.AssignableTo(typeWSEventTrigger) {
         continue
      }

      event = &SwaggerWSEventT{
         Description: fld.Tag.Get("doc"),
         EventId: WSMethodName,
         Parameters: []SwaggerEventParameterT{
            SwaggerEventParameterT{
               Name: "EventId",
               Description: &desc1,
               Type: "string",
               Required: true,
            },
            SwaggerEventParameterT{
               Name: "EventData",
               Description: &desc2,
               Type: "array",
               Required: true,
            },
         },
      }


      // Parse the fld tag scanning for parameter definitions.
      // All struct tags are parameters definitions, except the "doc" tag which is used for description/documentation of the event trigger
      tags = tagRE.FindAllStringSubmatch(string(fld.Tag),-1)
      event.Parameters[1].Items = make([]SwaggerEventParameterT,len(tags))
      for parmcount, tag = range tags {
         if tag[1] == "doc" {
            docndx = parmcount
            continue
         }
         switch tag[2] {
         // allowed types are these listed below
         case "string", "number", "integer", "boolean", "array":
            event.Parameters[1].Items[parmcount] = SwaggerEventParameterT{
               Name: tag[1],
               Type: tag[2],
               Required: true,
            }

            if tag[2] == "array" {
               event.Parameters[1].Items[parmcount].Items = []SwaggerEventParameterT{
                  SwaggerEventParameterT{
                     Name: tag[1] + "...",
                     Type: "any",
                     Required: true,
                  },
               }
            }
         default:
            Goose.Swagger.Logf(1,"Error %s @%d (%s)",WrongParameterType,parmcount,tags[2])
            return nil, WrongParameterType
         }
      }

      // We allocated the event.Parameters[1].Items array based on the number of tags defined
      // if a "doc" tag was defined, we have to delete its position from the array
      if docndx >= 0 {
         copy(event.Parameters[1].Items[docndx:],event.Parameters[1].Items[docndx+1:])
         event.Parameters[1].Items = event.Parameters[1].Items[:len(event.Parameters[1].Items)-1]
      }
   }

   return event, nil
}

