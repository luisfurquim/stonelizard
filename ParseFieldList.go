package stonelizard

import (
   "strings"
   "reflect"
)

// Consider this example tag: `query:"field1,field2" field1:"This is the first parameter for this operation"`
// Searches for a listEncoding ('query' in this example) and returns a swagger list of parameters with 2 parameters ("field1" and "field2")
// the parameter "field1" will have a documentation based on the text provided by the tag "field1", the parameter "field2" will have no documentation
func ParseFieldList(listEncoding string, parmcountIn int, fld reflect.StructField, method reflect.Method, methodName string, swaggerParametersIn []SwaggerParameterT) (list []string, parmcount int, pt reflect.Type, swaggerParameters  []SwaggerParameterT, err error) {
   var lstFlds            string
   var lstFld             string
   var doc                string
   var SwaggerParameter  *SwaggerParameterT

   parmcount         = parmcountIn
   swaggerParameters = swaggerParametersIn

   list = []string{}
   lstFlds = fld.Tag.Get(listEncoding)
   if lstFlds != "" {
      for _, lstFld = range strings.Split(lstFlds,",") {
         parmcount++
         if (parmcount+1) > method.Type.NumIn() {
            Goose.ParseFields.Logf(1,"%s (with query) at method %s", ErrorWrongParameterCount, methodName)
            err = ErrorWrongParameterCount
            return
         }
         pt = method.Type.In(parmcount)
         SwaggerParameter, err = GetSwaggerType(pt)
         if err != nil {
            return
         }

         if SwaggerParameter == nil {
            err = ErrorInvalidNilParam
            return
         }

/*
         if SwaggerParameter.Schema.Required != nil {
            Goose.ParseFields.Logf(1,"%s: %s",ErrorInvalidParameterType,lstFld)
            Goose.ParseFields.Logf(1,"SwaggerParameter: %#v",SwaggerParameter)
            err = ErrorInvalidParameterType
            return
         }
*/

         doc = fld.Tag.Get(lstFld)
         if doc != "" {
            SwaggerParameter.Description = doc
         }

         SwaggerParameter.Name     = lstFld
         SwaggerParameter.In       = listEncoding
         SwaggerParameter.Required = true
         SwaggerParameter.Schema   = nil

         if pt.Kind() == reflect.Map {
            SwaggerParameter.Schema   = &SwaggerSchemaT{
               Items: &SwaggerSchemaT{},
            }
            kname := pt.Key().Name()
            if kname == "string" {
               SwaggerParameter.Schema.Items.XKeyType = "string"
            } else {
               SwaggerParameter.Schema.Items.XKeyType   = "integer"
               if kname[len(kname)-2:] == "64" {
                  SwaggerParameter.Schema.Items.XKeyFormat = "int64"
               } else {
                  SwaggerParameter.Schema.Items.XKeyFormat = "int32"
               }
            }
         }

         swaggerParameters = append(swaggerParameters,*SwaggerParameter)
         list              = append(list,lstFld)
      }
   }

   Goose.ParseFields.Logf(6,"parm: %s, count: %d, met.in:%d",methodName, parmcount,method.Type.NumIn()) // 3, 4
   return
}

