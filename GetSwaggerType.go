package stonelizard

import (
   "strings"
   "reflect"
)

// Returns a swagger definition of the type received
// Scalar types are just a simple type string and maybe a format specification
// Nils are nils :P
// Pointers get de-referenced and then we recurse
// Aggregated types (array/slice, map and struct) needs more burocracy, but they all end up recursing until we reach the scalar types
func GetSwaggerType(parm reflect.Type) (*SwaggerParameterT, error) {
   var item      *SwaggerParameterT
   var field      reflect.StructField
   var err        error
   var i          int
   var schema     SwaggerSchemaT
   var schemaPtr *SwaggerSchemaT
   var subItem   *SwaggerParameterT
   var fldName    string
   var prm       *SwaggerParameterT
   var title      string
   var jsonName   string
   var ignore		string
   var ok			bool

   Goose.Swagger.Logf(6,"Parameter type : %d: %s",parm.Kind(),parm)

   if parm == voidType {
      return nil, nil
   }

   prm = &SwaggerParameterT{
      Name: parm.Name(),
      Schema: &SwaggerSchemaT{},
   }

   if parm.Kind() == reflect.Interface {
      prm.Schema.Type = "*"
      return prm, nil
   }

   if parm.Kind()==reflect.Ptr {
      return GetSwaggerType(parm.Elem())
   }

   if parm.Kind()==reflect.String {
      prm.Schema.Type   = "string"
      prm.Schema.Title  = "string"
      return prm, nil
   }

   if (parm.Kind() >= reflect.Bool && parm.Kind() <= reflect.Float64) {
      if parm.Kind() == reflect.Bool {
         prm.Schema.Type = "boolean"
      }

      if (parm.Kind()>=reflect.Int) && (parm.Kind()<=reflect.Int32) {
         prm.Schema.Type   = "integer"
         prm.Schema.Format = "int32"
      }

      if parm.Kind()==reflect.Int64 {
         prm.Schema.Type   = "integer"
         prm.Schema.Format = "int64"
      }

      if (parm.Kind()>=reflect.Uint) && (parm.Kind()<=reflect.Uint64) {
         prm.Schema.Type   = "integer"
      }

      if parm.Kind()==reflect.Float32 {
         prm.Schema.Type   = "number"
         prm.Schema.Format = "float"
      }

      if parm.Kind()==reflect.Float64 {
         prm.Schema.Type   = "number"
         prm.Schema.Format = "double"
      }

      prm.Schema.Title  = prm.Schema.Format
      return prm, nil
   }

   // Array/slice register as array using csv format, then recurse getting the type definition of its elements
   if (parm.Kind()==reflect.Array) || (parm.Kind()==reflect.Slice) {
      item, err = GetSwaggerType(parm.Elem())
      Goose.Swagger.Logf(5,"array elem type item=%#v, err:%s", item, err)
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         Goose.Swagger.Logf(6,"Error get array elem type item=%#v, err:%s", item, err)
         if (item!=nil) && (err==nil) && (item.Schema==nil) {
            Goose.Swagger.Logf(6,"And also error get array elem type item.Schema=%#v",item.Schema)
         }
         return nil, err
      }


      prm.Type          = "array"
      prm.Schema.Title  = arrayName(parm)
      prm.Schema.Type   = "array"
      prm.Schema.Items  = &SwaggerSchemaT{
         Title:            item.Schema.Title,
         Type:             item.Schema.Type,
         Format:           item.Schema.Format,
         Items:            item.Schema.Items,
         Required:         item.Schema.Required,
         Properties:       item.Schema.Properties,
      }
      // TODO: allow more collection formats
      prm.CollectionFormat = "csv"

      return prm, nil
   }

   // Maps are registered as array using csv format and we define swagger extensions to
   // define the key types and a key/value collection format (cskv - comma separated key/value),
   // then we recurse getting the type definition of its elements
   if parm.Kind()==reflect.Map {
      item, err = GetSwaggerType(parm.Elem())
      Goose.Swagger.Logf(6,"map elem type item=%#v, err:%s", item, err)
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         Goose.Swagger.Logf(6,"Error get map elem type item=%#v, err:%s", item, err)
         if (item!=nil) && (err==nil) && (item.Schema==nil) {
            Goose.Swagger.Logf(6,"And also error get map elem type item.Schema=%#v",item.Schema)
         }
         return nil, err
      }

      kname   := parm.Key().Name()
      ktype   := ""
      kformat := ""
      if kname == "string" {
         ktype = "string"
      } else {
         ktype   = "integer"
         if kname[len(kname)-2:] == "64" {
            kformat = "int64"
         } else {
            kformat = "int32"
         }
      }

      prm.Type                     = "array"
      prm.Schema.Title             = mapName(parm)
      prm.Schema.Type              = "array"
      prm.Schema.XKeyType          = ktype
      prm.Schema.XKeyFormat        = kformat
      prm.Schema.XCollectionFormat = "cskv"
      prm.Schema.Items  = &SwaggerSchemaT{
         Title:            item.Schema.Title,
         Type:             item.Schema.Type,
         Format:           item.Schema.Format,
         Items:            item.Schema.Items,
         Required:         item.Schema.Required,
         Properties:       item.Schema.Properties,
      }
      prm.XKeyType          = ktype
      prm.XKeyFormat        = kformat
      // TODO: allow more collection formats
      prm.CollectionFormat  = "csv"
      prm.XCollectionFormat = "cskv"

      return prm, nil
   }

   // Structs are defined as objects and we recurse each field to get its types
   if parm.Kind()==reflect.Struct {
      item = &SwaggerParameterT{
         Name: parm.Name(),
         Type:"object",
         Schema: &SwaggerSchemaT{
            Title: parm.Name(),
            Type:"object",
            Required: []string{},
            Properties: map[string]SwaggerSchemaT{},
//            Description: description,
         },
      }
      Goose.Swagger.Logf(6,"Got struct: %#v",item)
      for i=0; i<parm.NumField(); i++ {
         field = parm.Field(i)

			if ignore, ok = field.Tag.Lookup("swagger"); ok && ignore=="-" {
				continue
			}

         if field.Name[:1] == strings.ToLower(field.Name[:1]) {
            continue // Unexported field
         }

         title = field.Tag.Get("title")
         Goose.New.Logf(2,"Title: [%s=%s]", field.Name, title)
         if title != "" {
            item.Schema.Title = title
            Goose.New.Logf(2,"Title found: [%s=%#v]", field.Name, item.Schema)
            continue
         }

         if field.Anonymous {
            // Embedded fields must handle the field type
            if field.Type.Kind() == reflect.Struct {
               // In case of embedded structs we have to promote the promotable fields
               // We will test if they do not conflict with any field already in item.Schema.Properties
               // But it may happen that we promote a not promotable field because the conflicted
               // is not in properties yet. This will be naturally fixed when we process the conflicted field
               // adding (overlapping) it to the properties map
               subItem, err = GetSwaggerType(field.Type)
               if err != nil {
                  return nil, err
               }
               for fldName, schema = range subItem.Schema.Properties {
                  item.Schema.Required = append(item.Schema.Required,schema.Required...)
                  if _, ok := item.Schema.Properties[fldName]; !ok {
                     item.Schema.Properties[fldName] = schema
                  }
               }
            } else {

               if field.Type.Kind() == reflect.Ptr {

                  schemaPtr, item.Schema.Required, err = fieldHandle(field.Type.Elem().Name(), field)
                  if err != nil {
                     return nil, err
                  }
                  if schemaPtr == nil {
                     continue
                  }
                  item.Schema.Properties[field.Type.Elem().Name()] = *schemaPtr
               } else {

                  schemaPtr, item.Schema.Required, err = fieldHandle(field.Type.Name(), field)
                  if err != nil {
                     return nil, err
                  }
                  if schemaPtr == nil {
                     continue
                  }
                  item.Schema.Properties[field.Type.Name()] = *schemaPtr
               }
            }
         } else {

            jsonName = strings.Split(field.Tag.Get("json"),",")[0]
            if jsonName != "" {
               schemaPtr, item.Schema.Required, err = fieldHandle(jsonName, field)
            } else {
               schemaPtr, item.Schema.Required, err = fieldHandle(field.Name, field)
            }
            if err != nil {
               return nil, err
            }
            if schemaPtr == nil {
               continue
            }

            if jsonName != "" {
/*
               if schemaPtr.Type == "object" {
                  schemaPtr.Title = jsonName
               } else if schemaPtr.Type == "array" || schemaPtr.Type == "map" {
                  aggrIndentifierRE
               }
*/
               item.Schema.Properties[jsonName] = *schemaPtr
            } else {
               item.Schema.Properties[field.Name] = *schemaPtr
            }
         }

      }

      Goose.Swagger.Logf(2,"Got final struct: %#v",item)//6
      Goose.Swagger.Logf(2,"Got final struct: %#v",item.Schema)//6
      return item, nil
   }

   return nil, ErrorInvalidParameterType
}


func fieldHandle(fldName string, field reflect.StructField) (*SwaggerSchemaT, []string, error) {
   var err            error
   var doc            string
   var description   *string
   var required     []string
   var subItem       *SwaggerParameterT
   var fieldType      string

   Goose.Swagger.Logf(6,"Struct field: %s",fldName)
   doc   = field.Tag.Get("doc")
   if doc != "" {
      description    = new(string)
      (*description) = doc
   } else {
      description = nil
   }

//   required = append(required,fldName)

   subItem, err = GetSwaggerType(field.Type)
   Goose.Swagger.Logf(6,"struct subitem=%#v, err:%s", subItem, err)
   if (subItem==nil) || (err != nil) || (subItem.Schema==nil) {
      if err == ErrorInvalidParameterType {
         Goose.Swagger.Logf(5,"%s on subitem %s, just ignoring", err, fldName)
         return nil, nil, nil
      }
      Goose.Swagger.Logf(1,"Error getting type of subitem %s: %s",fldName,err)
      return nil, nil, err
   }

   if subItem.Type != "" {
      fieldType = subItem.Type
   } else {
      fieldType = subItem.Schema.Type
   }

   return &SwaggerSchemaT{
      Title:            subItem.Schema.Title,
      Type:             fieldType,
      Format:           subItem.Schema.Format,
      Items:            subItem.Schema.Items,
      Description:      description,
      Required:         append(subItem.Schema.Required,fldName),
      Properties:       subItem.Schema.Properties,
   }, required, nil

}


func mapName(p reflect.Type) string {
   var n string

   n = p.Elem().Name()
   if n != "" {
      return n + "{}"
   }

   if p.Kind() == reflect.Map {
      return mapName(p.Elem()) + "{}"
   }

   if p.Kind() == reflect.Array {
      return arrayName(p.Elem()) + "{}"
   }

   return p.Name() + "{}"
}

func arrayName(p reflect.Type) string {
   var n string

   n = p.Elem().Name()
   if n != "" {
      return n + "[]"
   }

   if p.Kind() == reflect.Map {
      return mapName(p.Elem()) + "[]"
   }

   if p.Kind() == reflect.Array {
      return arrayName(p.Elem()) + "[]"
   }

   return p.Name() + "[]"
}
