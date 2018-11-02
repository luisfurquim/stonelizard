package stonelizard

import (
   "strings"
   "reflect"
)

// Returns de websocket swagger specification
func GetWebSocketSpec(field reflect.StructField, WSMethodName string, WSMethod reflect.Method) (*SwaggerWSOperationT, error) {
   var operation            *SwaggerWSOperationT
   var err                   error
   var inTag                 string
   var tags                []string
   var parmName              string
   var SwaggerParameter     *SwaggerParameterT
   var parmcount             int
   var responses  map[string]SwaggerResponseT
   var xkeytype              string
   var parmTmp               SwaggerParameterT
   var parmList            []string

   // Parse the 'in' tag looking for websocket operation parameters
   inTag = field.Tag.Get("in")
   tags  = strings.Split(field.Tag.Get("tags"),",")
   if len(tags) == 0 {
      tags = nil
   }

   if (len(tags) == 1) && (tags[0]=="") {
      tags = nil
   }

   // Gets the response of the websocket operation
   // The syntax is the same for normal RESTful operations
   // tags "ok", "accepted", etc defines a successful operation and
   // the data returned is defined by the field type
   responses, _, err = getResponses(field, field.Type)
   if err != nil {
      return nil, err
   }

   operation = &SwaggerWSOperationT{
      Tags:        tags,
      Description: field.Tag.Get("doc"),
      SuboperationId: WSMethodName,
      Parameters: []SwaggerParameterT{},
      //Responses:   map[string]SwaggerResponseT{},
      Responses: responses, // callID , status, response
   }

   if inTag != "" {
      parmList = strings.Split(strings.Trim(inTag,""),",")
      if len(parmList) != (WSMethod.Type.NumIn()-1) {
         Goose.Swagger.Logf(1,"%s",ErrorParmListSyntax)
         return nil, ErrorWrongParameterCount
      }

      // Scans for parameters defined in the "in" tag
      for parmcount, parmName = range parmList {
         if parmName=="" {
            Goose.Swagger.Logf(1,"%s",ErrorParmListSyntax)
            return nil, ErrorInvalidType
         }

         // Get the swagger definition for the 'parmcount'th parameter
         SwaggerParameter, err = GetSwaggerType(WSMethod.Type.In(parmcount+1))
         if err != nil {
            return nil, err
         }

         if SwaggerParameter == nil {
            return nil, ErrorInvalidNilParam
         }

         xkeytype = ""
         if SwaggerParameter.Schema != nil {
            xkeytype = SwaggerParameter.Schema.XKeyType
         }

         parmTmp = SwaggerParameterT{
            Name: parmName,
            In: "body",
            Required: true,
            Type: SwaggerParameter.Schema.Type,
            XKeyType: xkeytype,
            Format: SwaggerParameter.Format,
         }

         if (SwaggerParameter.CollectionFormat != "") || (len(SwaggerParameter.Schema.Required)>0) {
//            Goose.New.Logf(1,"%s: %s -> sch_req:%#v %#v",ErrorInvalidParameterType,parmName,SwaggerParameter.Schema.Required,SwaggerParameter)
//            return nil, ErrorInvalidParameterType
            parmTmp.CollectionFormat = SwaggerParameter.CollectionFormat
            parmTmp.Schema           = SwaggerParameter.Schema
         }

         operation.Parameters = append(operation.Parameters,parmTmp)

      }

   }

   return operation, nil
}

