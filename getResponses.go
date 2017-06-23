package stonelizard

import (
   "fmt"
   "reflect"
   "net/http"
)

// Get the specifications for the successful execution responses
// Responses may be defined in 2 ways:
// 1) defining a tag with one of these names:
//    a) ok
//    b) created
//    c) accepted
// 2) specifiying a method with a name equal to the operation name append with "Responses"
//    a) the method signature must be func () map[string]ResponseT
//    b) the map[string]ResponseT keys MUST be defined as HTTP status code
//       and the correspondent value must be a swagger (data schema) response definition
func getResponses(fld reflect.StructField, typ reflect.Type) (map[string]SwaggerResponseT, reflect.Type, error) {
   var responseOk              string
   var responseCreated         string
   var responseAccepted        string
   var responses    map[string]SwaggerResponseT
   var fldType                 reflect.Type
   var SwaggerParameter       *SwaggerParameterT
   var err                     error
   var doc                     string
   var responseFunc            reflect.Method
   var ok                      bool
   var responseList map[string]ResponseT

   responses = map[string]SwaggerResponseT{}

   responseOk = fld.Tag.Get("ok")
   responseCreated = fld.Tag.Get("created")
   responseAccepted = fld.Tag.Get("accepted")
   if responseOk != "" || responseCreated != "" || responseAccepted != "" {
      if responseOk != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
         }

         if SwaggerParameter == nil {
            responses[fmt.Sprintf("%d",http.StatusNoContent)] = SwaggerResponseT{
               Description: responseOk,
            }
         } else {

            doc = fld.Tag.Get(fld.Name)

            if doc != "" {
               SwaggerParameter.Schema.Description    = new(string)
               (*SwaggerParameter.Schema.Description) = doc
            }

            if SwaggerParameter.Schema == nil {
               SwaggerParameter.Schema = &SwaggerSchemaT{}
            }

            if (SwaggerParameter.Schema.Type=="") && (SwaggerParameter.Type!="") {
               SwaggerParameter.Schema.Type = SwaggerParameter.Type
            }

            responses[fmt.Sprintf("%d",http.StatusOK)] = SwaggerResponseT{
               Description: responseOk,
               Schema:      SwaggerParameter.Schema,
            }
            //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
            //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusOK)].Schema))
         }
      }
      if responseCreated != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
         }

         if SwaggerParameter == nil {
            responses[fmt.Sprintf("%d",http.StatusCreated)] = SwaggerResponseT{
               Description: responseCreated,
            }
         } else {

            doc = fld.Tag.Get(fld.Name)

            if doc != "" {
               SwaggerParameter.Schema.Description    = new(string)
               (*SwaggerParameter.Schema.Description) = doc
            }

            if SwaggerParameter.Schema == nil {
               SwaggerParameter.Schema = &SwaggerSchemaT{}
            }

            if (SwaggerParameter.Schema.Type=="") && (SwaggerParameter.Type!="") {
               SwaggerParameter.Schema.Type = SwaggerParameter.Type
            }

            responses[fmt.Sprintf("%d",http.StatusCreated)] = SwaggerResponseT{
               Description: responseCreated,
               Schema:      SwaggerParameter.Schema,
            }
            //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
            //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusCreated)].Schema))
         }
      }
      if responseAccepted != "" {
         fldType = fld.Type
         if fldType.Kind() == reflect.Ptr {
            fldType = fldType.Elem()
         }

         SwaggerParameter, err = GetSwaggerType(fldType)
         if err != nil {
            return nil, nil, err
         }

         if SwaggerParameter == nil {
            responses[fmt.Sprintf("%d",http.StatusAccepted)] = SwaggerResponseT{
               Description: responseAccepted,
            }
         } else {

            doc = fld.Tag.Get(fld.Name)

            if doc != "" {
               SwaggerParameter.Schema.Description    = new(string)
               (*SwaggerParameter.Schema.Description) = doc
            }

            if SwaggerParameter.Schema == nil {
               SwaggerParameter.Schema = &SwaggerSchemaT{}
            }

            if (SwaggerParameter.Schema.Type=="") && (SwaggerParameter.Type!="") {
               SwaggerParameter.Schema.Type = SwaggerParameter.Type
            }

            responses[fmt.Sprintf("%d",http.StatusAccepted)] = SwaggerResponseT{
               Description: responseAccepted,
               Schema:      SwaggerParameter.Schema,
            }
            //(*responses[fmt.Sprintf("%d",http.StatusOK)].Schema) = *SwaggerParameter.Schema
            //ioutil.WriteFile("debug.txt", []byte(fmt.Sprintf("%#v",responses)), os.FileMode(0770))
            Goose.New.Logf(6,"====== %#v",*(responses[fmt.Sprintf("%d",http.StatusAccepted)].Schema))
         }
      }
   } else if responseFunc, ok = typ.MethodByName(fld.Name + "Responses"); ok {
      responseList = responseFunc.Func.Call([]reflect.Value{})[0].Interface().(map[string]ResponseT)
      for responseStatus, responseSchema := range responseList {
         SwaggerParameter, err = GetSwaggerType(reflect.TypeOf(responseSchema.TypeReturned))
         if err != nil {
            return nil, nil, err
         }
         if SwaggerParameter == nil {
            responses[responseStatus] = SwaggerResponseT{
               Description: responseSchema.Description,
            }
         } else {
            responses[responseStatus] = SwaggerResponseT{
               Description: responseSchema.Description,
               Schema:      SwaggerParameter.Schema,
            }
         }
      }
   }

   return responses, fldType, nil
}


