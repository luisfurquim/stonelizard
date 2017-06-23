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
   var item, subItem *SwaggerParameterT
   var field          reflect.StructField
   var doc            string
   var description   *string
   var err            error
   var i              int
   var fieldType      string

   Goose.Swagger.Logf(6,"Parameter type : %d: %s",parm.Kind(),parm)

   if parm == voidType {
      return nil, nil
   }

   if parm.Kind() == reflect.Interface {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"*"}}, nil
   }

   if parm.Kind() == reflect.Bool {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"boolean"}}, nil
   }

   if (parm.Kind()>=reflect.Int) && (parm.Kind()<=reflect.Int32) {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer", Format: "int32"}}, nil
   }

   if parm.Kind()==reflect.Int64 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer", Format: "int64"}}, nil
   }

   if (parm.Kind()>=reflect.Uint) && (parm.Kind()<=reflect.Uint64) {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"integer"}}, nil
   }

   if parm.Kind()==reflect.Float32 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"number", Format: "float"}}, nil
   }

   if parm.Kind()==reflect.Float64 {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"number", Format: "double"}}, nil
   }

   if parm.Kind()==reflect.String {
      return &SwaggerParameterT{Schema: &SwaggerSchemaT{Type:"string"}}, nil
   }

   if parm.Kind()==reflect.Ptr {
      return GetSwaggerType(parm.Elem())
   }

   // Array/slice register as array using csv format, then recurse getting the type definition of its elements
   if (parm.Kind()==reflect.Array) || (parm.Kind()==reflect.Slice) {
      item, err = GetSwaggerType(parm.Elem())
      if (item==nil) || (err!=nil) || (item.Schema==nil) {
         Goose.Swagger.Logf(6,"Error get array elem type item=%#v, err:%s", item, err)
         if (item!=nil) && (err==nil) && (item.Schema==nil) {
            Goose.Swagger.Logf(6,"And also error get array elem type item.Schema=%#v",item.Schema)
         }
         return nil, err
      }
      return &SwaggerParameterT{
         Type:"array",
         Items: &SwaggerItemT{
            Type:             item.Schema.Type,
            Format:           item.Schema.Format,
            Items:            item.Schema.Items,
         },
         Schema: &SwaggerSchemaT{
            Type: item.Schema.Type,
         },
         CollectionFormat: "csv",
      }, nil // TODO: allow more collection formats
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

      return &SwaggerParameterT{
         Type:"array",
         Items: &SwaggerItemT{
            Type:              item.Schema.Type,
            Format:            item.Schema.Format,
            Items:             item.Schema.Items,
         },
         Schema: &SwaggerSchemaT{
            Type: item.Schema.Type,
         },
         XKeyType:             ktype,
         XKeyFormat:           kformat,
         CollectionFormat:     "csv",
         XCollectionFormat:    "cskv",
      }, nil // TODO: allow more collection formats
   }

   // Structs are defined as objects and we recurse each field to get its types
   if parm.Kind()==reflect.Struct {
      item = &SwaggerParameterT{
         Type:"object",
         Schema: &SwaggerSchemaT{
            Required: []string{},
            Properties: map[string]SwaggerSchemaT{},
//            Description: description,
         },
      }
      Goose.Swagger.Logf(6,"Got struct: %#v",item)
      for i=0; i<parm.NumField(); i++ {
         field = parm.Field(i)

         if field.Name[:1] == strings.ToLower(field.Name[:1]) {
            continue // Unexported field
         }


         Goose.Swagger.Logf(6,"Struct field: %s",field.Name)
         doc   = field.Tag.Get("doc")
         if doc != "" {
            description    = new(string)
            (*description) = doc
         } else {
            description = nil
         }

         item.Schema.Required = append(item.Schema.Required,field.Name)

         subItem, err = GetSwaggerType(field.Type)
         Goose.Swagger.Logf(6,"struct subitem=%#v, err:%s", subItem, err)
         if (subItem==nil) || (err != nil) || (subItem.Schema==nil) {
            if err == ErrorInvalidParameterType {
               Goose.Swagger.Logf(5,"%s on subitem %s, just ignoring", err, field.Name)
               continue
            }
            Goose.Swagger.Logf(1,"Error getting type of subitem %s: %s",field.Name,err)
            return nil, err
         }

         if subItem.Type != "" {
            fieldType = subItem.Type
         } else {
            fieldType = subItem.Schema.Type
         }

         item.Schema.Properties[field.Name] = SwaggerSchemaT{
            Type:             fieldType,
            Format:           subItem.Format,
            Items:            subItem.Items,
            Description:      description,
            Required:         subItem.Schema.Required,
            Properties:       subItem.Schema.Properties,
         }
      }

      Goose.Swagger.Logf(6,"Got final struct: %#v",item)
      return item, nil
   }

   return nil, ErrorInvalidParameterType
}

