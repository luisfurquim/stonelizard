package stonelizard

type SwaggerContactT struct {
   // The identifying name of the contact person/organization.
   Name  string              `json:"name"`

   // The URL pointing to the contact information. MUST be in the format of a URL.
   Url   string              `json:"url"`

   // The email address of the contact person/organization. MUST be in the format of an email address.
   Email string              `json:"email"`
}

type SwaggerLicenseT struct {
   // The license name used for the API.
   Name  string              `json:"name"`

   // A URL to the license used for the API. MUST be in the format of a URL.
   Url   string              `json:"url,omitempty"`
}

type SwaggerInfoT struct {
   // The title of the application.
   Title          string             `json:"title"`

   // A short description of the application.
   // GFM syntax can be used for rich text representation.
   Description    string             `json:"description,omitempty"`

   // The Terms of Service for the API.
   TermsOfService string             `json:"termsOfService,omitempty"`

   // The contact information for the exposed API.
   Contact        SwaggerContactT    `json:"contact,omitempty"`

   // The license information for the exposed API.
   License        SwaggerLicenseT    `json:"license,omitempty"`

   // Provides the version of the application API (not to be confused with the specification version).
   Version        string             `json:"version"`
}

type SwaggerWSOperationT struct {
   // A list of tags for API documentation control.
   // Tags can be used for logical grouping of operations by resources or any other qualifier.
   Tags  []string `json:"tags,omitempty"`

   // A short summary of what the operation does.
   // For maximum readability in the swagger-ui, this field SHOULD be less than 120 characters.
   Summary  string `json:"summary,omitempty"`

   // A verbose explanation of the operation behavior. GFM syntax can be used for rich text representation.
   Description string `json:"description,omitempty"`

   // Additional external documentation for this operation.
   ExternalDocs   *SwaggerExternalDocsT `json:"externalDocs,omitempty"`

   // Required. Unique string used to identify the websocket suboperation.
   // The operation id combined with the suboperation id MUST be unique among
   // all operations described in the API. Tools and libraries MAY use the
   // operationId/suboperationId combination to uniquely identify a websocket
   // suboperation, therefore, it is recommended to follow common programming naming
   // conventions.
   SuboperationId string `json:"suboperationId"`

   // A list of parameters that are applicable for this websocket suboperation.
   // The list MUST NOT include duplicated parameters.
   Parameters           []SwaggerParameterT `json:"parameters,omitempty"`

   // Required. The list of possible responses as they are returned from executing this operation.
   Responses   map[string]SwaggerResponseT `json:"responses"`

   // Declares this operation to be deprecated.
   // Usage of the declared operation should be refrained.
   // Default value is false.
   Deprecated  bool `json:"deprecated,omitempty"`

   // A declaration of which security schemes are applied for this operation.
   // The list of values describes alternative security schemes that can be used
   // (that is, there is a logical OR between the security requirements).
   // This definition overrides any declared top-level security.
   // To remove a top-level security declaration, an empty array can be used.
   Security map[string]SwaggerSecurityT `json:"security,omitempty"`
}

type SwaggerEventParameterT struct {
   Name string `json:"name"`

   // GFM syntax can be used for rich text representation
   Description *string `json:"description,omitempty"`

   // The type of the parameter / The internal type of the array.
   // Since the parameter is not located at the request body, it is limited to simple types (that is, not an object).
   // The value MUST be one of "string", "number", "integer", "boolean", "array" or "file" (Files and models are not allowed in arrays).
   // If type is "file", the consumes MUST be either "multipart/form-data" or " application/x-www-form-urlencoded" and the parameter MUST be in "formData".
   Type              string         `json:"type"`

   // The extending format for the previously mentioned type. See Data Type Formats for further details.
   Format            string         `json:"format,omitempty"`

   // If Type is "array", this must show a list of its items.
   // The list MUST NOT include duplicated parameters.
   // If type
   Items        []SwaggerEventParameterT `json:"items,omitempty"`

   // Declares this operation to be deprecated.
   // Usage of the declared operation should be refrained.
   // Default value is false.
   Deprecated  bool `json:"deprecated,omitempty"`

   // Determines whether this parameter is mandatory.
   // If the parameter is in "path", this property is required and its value MUST be true.
   // Otherwise, the property MAY be included and its default value is false.
   Required          bool  `json:"required,omitempty"`
}

type SwaggerWSEventT struct {
   // A list of tags for API documentation control.
   // Tags can be used for logical grouping of operations by resources or any other qualifier.
   Tags  []string `json:"tags,omitempty"`

   // A short summary of when the event fires.
   // For maximum readability in the swagger-ui, this field SHOULD be less than 120 characters.
   Summary  string `json:"summary,omitempty"`

   // A verbose explanation of the event behavior. GFM syntax can be used for rich text representation.
   Description string `json:"description,omitempty"`

   // Additional external documentation for this event.
   ExternalDocs   *SwaggerExternalDocsT `json:"externalDocs,omitempty"`

   // Required. Unique string used to identify the websocket event.
   // The operation id combined with the event id MUST be unique among
   // all operations+events described in the API. Tools and libraries MAY use the
   // operationId/eventId combination to uniquely identify a websocket
   // event, therefore, it is recommended to follow common programming naming
   // conventions.
   EventId string `json:"eventId"`

   // A list of parameters that are applicable for this websocket event.
   // The list MUST NOT include duplicated parameters.
   Parameters        []SwaggerEventParameterT `json:"parameters,omitempty"`

   // Declares this event to be deprecated.
   // Usage of the declared event should be refrained.
   // Default value is false.
   Deprecated  bool `json:"deprecated,omitempty"`

   // A declaration of which security schemes are applied for this operation.
   // The list of values describes alternative security schemes that can be used
   // (that is, there is a logical OR between the security requirements).
   // This definition overrides any declared top-level security.
   // To remove a top-level security declaration, an empty array can be used.
   Security map[string]SwaggerSecurityT `json:"security,omitempty"`
}

type SwaggerOperationT struct {
   // A list of tags for API documentation control.
   // Tags can be used for logical grouping of operations by resources or any other qualifier.
   Tags  []string `json:"tags,omitempty"`

   // A short summary of what the operation does.
   // For maximum readability in the swagger-ui, this field SHOULD be less than 120 characters.
   Summary  string `json:"summary,omitempty"`

   // A verbose explanation of the operation behavior. GFM syntax can be used for rich text representation.
   Description string `json:"description,omitempty"`

   // Additional external documentation for this operation.
   ExternalDocs   *SwaggerExternalDocsT `json:"externalDocs,omitempty"`

   // Unique string used to identify the operation.
   // The id MUST be unique among all operations described in the API.
   // Tools and libraries MAY use the operationId to uniquely identify an operation, therefore,
   // it is recommended to follow common programming naming conventions.
   OperationId string `json:"operationId,omitempty"`

   // A list of MIME types the operation can consume.
   // This overrides the consumes definition at the Swagger Object.
   // An empty value MAY be used to clear the global definition.
   // Value MUST be as described under Mime Types.
   Consumes []string `json:"consumes,omitempty"`

   // A list of MIME types the operation can produce.
   // This overrides the produces definition at the Swagger Object.
   // An empty value MAY be used to clear the global definition.
   // Value MUST be as described under Mime Types.
   Produces []string `json:"produces,omitempty"`

   // A list of parameters that are applicable for this operation.
   // If a parameter is already defined at the Path Item, the new definition will override it, but can never remove it.
   // The list MUST NOT include duplicated parameters.
   // A unique parameter is defined by a combination of a name and location.
   // The list can use the Reference Object to link to parameters that are defined at the Swagger Object's parameters.
   // There can be one "body" parameter at most.
   Parameters           []SwaggerParameterT `json:"parameters,omitempty"`

   //  Required. The list of possible responses as they are returned from executing this operation.
   Responses   map[string]SwaggerResponseT `json:"responses"`

   // The transfer protocol for the operation.
   // Values MUST be from the list: "http", "https", "ws", "wss".
   // The value overrides the Swagger Object schemes definition.
   Schemes  []string `json:"schemes,omitempty"`

   // Declares this operation to be deprecated.
   // Usage of the declared operation should be refrained.
   // Default value is false.
   Deprecated  bool `json:"deprecated,omitempty"`

   // A declaration of which security schemes are applied for this operation.
   // The list of values describes alternative security schemes that can be used
   // (that is, there is a logical OR between the security requirements).
   // This definition overrides any declared top-level security.
   // To remove a top-level security declaration, an empty array can be used.
   Security map[string]SwaggerSecurityT `json:"security,omitempty"`

   // Custom stonelizard extension. Specifies suboperations. Used for websocket operations.
   // The HTTP connection is upgraded to websocket connection only if the operation returns a 2XX HTTP status code.
   XWSOperations map[string]*SwaggerWSOperationT `json:"x-websocketoperations,omitempty"`

   // Custom stonelizard extension. Specifies events. Used for websocket operations.
   // Events only fire if the HTTP connection is upgraded to websocket connection and subjected to application specific
   // semantics/rules.
   XWSEvents map[string]*SwaggerWSEventT `json:"x-websocketevents,omitempty"`

   // Custom stonelizard extension. A list of websocket subprotocols the operation can consume.
   // At this moment, it only accepts the stonelizard's non-standard 'sam+json', which stands for 'JSON encoded simple array messaging'.
   XWSConsumes []string `json:"x-websocketconsumes,omitempty"`

   // Custom stonelizard extension. It doesn't affect the web service operation.
   // It is intended to provide developers a way to organize their web service client code.
   // Our suggestion is that operations belonging to the same module be accesssed using
   // the same class. But it is not mandatory.
   XModule string `json:"x-module,omitempty"`

   // Custom stonelizard extension. It doesn't affect the web service operation.
   // It is intended to provide developers a way to handle the returned data.
   // Our suggestion is to use this info to give clients a hint on WHERE to store
   // the return data
   XOutputVar string `json:"x-outputvar,omitempty"`

   // Custom stonelizard extension. It doesn't affect the web service operation.
   // It is intended to provide developers a way to handle the returned data.
   // Our suggestion is to use this info to give clients a hint on HOW to store
   // the return data
   XOutput string `json:"x-output,omitempty"`
}

type SwaggerPathT map[string]*SwaggerOperationT
/*
struct {
   // A definition of a GET operation on this path.
   Get         SwaggerOperationT `json:"get"`

   // A definition of a PUT operation on this path.
   Put         SwaggerOperationT `json:"put"`

   // A definition of a POST operation on this path.
   Post        SwaggerOperationT `json:"post"`

   // A definition of a DELETE operation on this path.
   Delete      SwaggerOperationT `json:"delete"`

   // A definition of a OPTIONS operation on this path.
   Options     SwaggerOperationT `json:"options"`

   // A definition of a HEAD operation on this path.
   Head        SwaggerOperationT `json:"head"`

   // A definition of a PATCH operation on this path.
   Patch       SwaggerOperationT `json:"patch"`

   // A list of parameters that are applicable for all the operations described under this path.
   // These parameters can be overridden at the operation level, but cannot be removed there.
   // The list MUST NOT include duplicated parameters.
   // A unique parameter is defined by a combination of a name and location.
   // The list can use the Reference Object to link to parameters that are defined at the Swagger Object's parameters.
   // There can be one "body" parameter at most.
   // SL: Not used in stonelizard...
   //Parameters         []SwaggerParameterT `json:"parameters"`
}
*/


type SwaggerXmlT struct {
   // Replaces the name of the element/attribute used for the described schema property.
   // When defined within the Items Object (items), it will affect the name of the individual XML elements within the list.
   // When defined alongside type being array (outside the items), it will affect the wrapping element and only if wrapped is true.
   // If wrapped is false, it will be ignored.
   Name  string `json:"name,omitempty"`

   // The URL of the namespace definition. Value SHOULD be in the form of a URL.
   Namespace   string `json:"namespace,omitempty"`

   // The prefix to be used for the name.
   Prefix   string `json:"prefix,omitempty"`

   // Declares whether the property definition translates to an attribute instead of an element.
   // Default value is false.
   Attribute   bool `json:"attribute,omitempty"`

   // MAY be used only for an array definition.
   // Signifies whether the array is wrapped (for example, <books><book/><book/></books>) or unwrapped (<book/><book/>).
   // Default value is false. The definition takes effect only when defined alongside type being array (outside the items).
   Wrapped  bool `json:"wrapped,omitempty"`
}

type SwaggerSchemaT struct {
   Title string `json:"title,omitempty"`

   // GFM syntax can be used for rich text representation
   Description *string `json:"description,omitempty"`

   // The type of the parameter / The internal type of the array.
   // Since the parameter is not located at the request body, it is limited to simple types (that is, not an object).
   // The value MUST be one of "string", "number", "integer", "boolean", "array" or "file" (Files and models are not allowed in arrays).
   // If type is "file", the consumes MUST be either "multipart/form-data" or " application/x-www-form-urlencoded" and the parameter MUST be in "formData".
   Type              string         `json:"type"`

   // Required if type is "array". Describes the type of items in the array.
//   Items            *SwaggerItemT   `json:"items,omitempty"`
   Items            *SwaggerSchemaT   `json:"items,omitempty"`

   // The extending format for the previously mentioned type. See Data Type Formats for further details.
   Format            string         `json:"format,omitempty"`

   // Declares the value of the parameter that the server will use if none is provided,
   // for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request.
   // (Note: "default" has no meaning for required parameters.)
   // See http://json-schema.org/latest/json-schema-validation.html#anchor101.
   // Unlike JSON Schema this value MUST conform to the defined type for this parameter.
   Default  interface{}             `json:"default,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   Maximum          string          `json:"maximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   ExclusiveMaximum bool            `json:"exclusiveMaximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   Minimum          string          `json:"minimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   ExclusiveMinimum bool            `json:"exclusiveMinimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor26.
   MaxLength        int64           `json:"maxLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor29.
   MinLength        int64           `json:"minLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor33.
   Pattern          string          `json:"pattern,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor42.
   MaxItems         int64           `json:"maxItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor45.
   MinItems         int64           `json:"minItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor49.
   UniqueItems      bool            `json:"uniqueItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor76.
   Enum           []interface{}     `json:"enum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor14.
   MultipleOf       string          `json:"multipleOf,omitempty"`

   // The value of this keyword MUST be an integer. This integer MUST be greater than, or equal to, 0.
   // An object instance is valid against "maxProperties" if its number of properties is less than, or equal to, the value of this keyword.
   MaxProperties uint64 `json:"maxProperties,omitempty"`

   // The value of this keyword MUST be an integer. This integer MUST be greater than, or equal to, 0.
   // An object instance is valid against "maxProperties" if its number of properties is more than, or equal to, the value of this keyword.
   MinProperties uint64 `json:"minProperties,omitempty"`

   // Lists which of the objetct properties are mandatory.
   Required          []string  `json:"required,omitempty"`

   // This keyword's value MUST be an array. This array MUST have at least one element.
   // Elements of the array MUST be objects. Each object MUST be a valid Swagger JSON Schema.
   // An instance validates successfully against this keyword if it validates successfully against all schemas defined by this keyword's value.
   AllOf                 []*SwaggerSchemaT `json:"allOf,omitempty"`

   Properties  map[string]SwaggerSchemaT `json:"properties,omitempty"`

   // The value of "additionalProperties" MUST be a boolean or an object.
   // If it is an object, it MUST also be a valid JSON Schema.
   // SL: Always a schema...
   AdditionalProperties *SwaggerSchemaT `json:"additionalProperties,omitempty"`

   // Adds support for polymorphism.
   // The discriminator is the schema property name that is used to differentiate between other schema that inherit this schema.
   // The property name used MUST be defined at this schema and it MUST be in the required property list.
   // When used, the value MUST be the name of this schema or any schema that inherits it.
   Discriminator  string `json:"discriminator,omitempty"`

   // Relevant only for Schema "properties" definitions.
   // Declares the property as "read only".
   // This means that it MAY be sent as part of a response but MUST NOT be sent as part of the request.
   // Properties marked as readOnly being true SHOULD NOT be in the required list of the defined schema.
   // Default value is false.
   ReadOnly bool `json:"readOnly,omitempty"`

   // This MAY be used only on properties schemas.
   // It has no effect on root schemas.
   // Adds Additional metadata to describe the XML representation format of this property.
   Xml   *SwaggerXmlT `json:"xml,omitempty"`

   // External Documentation Object Additional external documentation for this schema.
   ExternalDocs *SwaggerExternalDocsT `json:"externalDocs,omitempty"`

   // A free-form property to include a an example of an instance for this schema.
   Example  interface{} `json:"example,omitempty"`

   // Custom stonelizard extension. Currently it only accepts cskv (comma separated key-values: k1:v1,...,kn:vn)
   XCollectionFormat  string         `json:"x-collectionFormat,omitempty"`

   // Custom stonelizard extension. Specifies the type of the key. Used for key-value data types
   XKeyType    string `json:"x-keytype,omitempty"`

   // Custom stonelizard extension. Specifies the format of the key. Used for key-value data types
   XKeyFormat  string `json:"x-keyformat,omitempty"`
}

type SwaggerItemT struct {
   // The type of the parameter / The internal type of the array.
   // Since the parameter is not located at the request body, it is limited to simple types (that is, not an object).
   // The value MUST be one of "string", "number", "integer", "boolean", "array" or "file" (Files and models are not allowed in arrays).
   // If type is "file", the consumes MUST be either "multipart/form-data" or " application/x-www-form-urlencoded" and the parameter MUST be in "formData".
   Type              string         `json:"type"`

   // Required if type is "array". Describes the type of items in the array.
   Items            *SwaggerItemT   `json:"items,omitempty"`

   // The extending format for the previously mentioned type. See Data Type Formats for further details.
   Format            string         `json:"format,omitempty"`

   // Declares the value of the parameter that the server will use if none is provided,
   // for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request.
   // (Note: "default" has no meaning for required parameters.)
   // See http://json-schema.org/latest/json-schema-validation.html#anchor101.
   // Unlike JSON Schema this value MUST conform to the defined type for this parameter.
   Default  interface{}             `json:"default,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   Maximum          string          `json:"maximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   ExclusiveMaximum bool            `json:"exclusiveMaximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   Minimum          string          `json:"minimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   ExclusiveMinimum bool            `json:"exclusiveMinimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor26.
   MaxLength        int64           `json:"maxLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor29.
   MinLength        int64           `json:"minLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor33.
   Pattern          string          `json:"pattern,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor42.
   MaxItems         int64           `json:"maxItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor45.
   MinItems         int64           `json:"minItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor49.
   UniqueItems      bool            `json:"uniqueItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor76.
   Enum           []interface{}     `json:"enum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor14.
   MultipleOf       string          `json:"multipleOf,omitempty"`

   // Determines the format of the array if type array is used. Possible values are:
   //   csv - comma separated values foo,bar.
   //   ssv - space separated values foo bar.
   //   tsv - tab separated values foo\tbar.
   //   pipes - pipe separated values foo|bar.
   //   multi - corresponds to multiple parameter instances instead of multiple values for a single instance foo=bar&foo=baz. This is valid only for parameters in "query" or "formData".
   //   Default value is csv.
   CollectionFormat  string         `json:"collectionFormat,omitempty"`

   // Custom stonelizard extension. Specifies the type of the key. Used for key-value data types
   XKeyType    string `json:"x-keytype,omitempty"`

   // Custom stonelizard extension. Specifies the format of the key. Used for key-value data types
   XKeyFormat  string `json:"x-keyformat,omitempty"`
}

type SwaggerParameterT struct {
   // If in is "path", the name field MUST correspond to the associated path segment from the path field in the Paths Object.
   // See Path Templating for further information.
   // For all other cases, the name corresponds to the parameter name used based on the in property.
   // The name of the parameter. Parameter names are case sensitive.
   Name              string   `json:"name"`

   // The location of the parameter.
   // Possible values are "query", "header", "path", "formData" or "body".
   In                string   `json:"in"`

   // A brief description of the parameter.
   // This could contain examples of use.
   // GFM syntax can be used for rich text representation.
   Description       string   `json:"description,omitempty"`

   // Determines whether this parameter is mandatory.
   // If the parameter is in "path", this property is required and its value MUST be true.
   // Otherwise, the property MAY be included and its default value is false.
   Required          bool  `json:"required"`

   // The schema defining the type used for the body parameter.
   Schema            *SwaggerSchemaT  `json:"schema,omitempty"` // required if in=="body"

   // Sets the ability to pass empty-valued parameters.
   // This is valid only for either query or formData parameters and allows you to send a parameter with a name only or an empty value.
   // Default value is false.
   AllowEmptyValue   bool  `json:"allowEmptyValue,omitempty"`

   // The type of the parameter / The internal type of the array.
   // Since the parameter is not located at the request body, it is limited to simple types (that is, not an object).
   // The value MUST be one of "string", "number", "integer", "boolean", "array" or "file" (Files and models are not allowed in arrays).
   // If type is "file", the consumes MUST be either "multipart/form-data" or " application/x-www-form-urlencoded" and the parameter MUST be in "formData".
   Type              string         `json:"type,omitempty"`

   // Required if type is "array". Describes the type of items in the array.
//   Items            *SwaggerItemT   `json:"items,omitempty"`

   // The extending format for the previously mentioned type. See Data Type Formats for further details.
   Format            string         `json:"format,omitempty"`

   // Declares the value of the parameter that the server will use if none is provided,
   // for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request.
   // (Note: "default" has no meaning for required parameters.)
   // See http://json-schema.org/latest/json-schema-validation.html#anchor101.
   // Unlike JSON Schema this value MUST conform to the defined type for this parameter.
   Default  interface{}             `json:"default,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   Maximum          string          `json:"maximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   ExclusiveMaximum bool            `json:"exclusiveMaximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   Minimum          string          `json:"minimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   ExclusiveMinimum bool            `json:"exclusiveMinimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor26.
   MaxLength        int64           `json:"maxLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor29.
   MinLength        int64           `json:"minLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor33.
   Pattern          string          `json:"pattern,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor42.
   MaxItems         int64           `json:"maxItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor45.
   MinItems         int64           `json:"minItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor49.
   UniqueItems      bool            `json:"uniqueItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor76.
   Enum           []interface{}     `json:"enum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor14.
   MultipleOf       string          `json:"multipleOf,omitempty"`

   // Determines the format of the array if type array is used. Possible values are:
   //   csv - comma separated values foo,bar.
   //   ssv - space separated values foo bar.
   //   tsv - tab separated values foo\tbar.
   //   pipes - pipe separated values foo|bar.
   //   multi - corresponds to multiple parameter instances instead of multiple values for a single instance foo=bar&foo=baz. This is valid only for parameters in "query" or "formData".
   //   Default value is csv.
   CollectionFormat  string         `json:"collectionFormat,omitempty"`

   // Custom stonelizard extension. Currently it only accepts cskv (comma separated key-values: k1:v1,...,kn:vn)
   XCollectionFormat  string         `json:"x-collectionFormat,omitempty"`

   // Custom stonelizard extension. Specifies the type of the key. Used for key-value data types
   XKeyType    string `json:"x-keytype,omitempty"`

   // Custom stonelizard extension. Specifies the format of the key. Used for key-value data types
   XKeyFormat  string `json:"x-keyformat,omitempty"`
}

type SwaggerHeaderT struct {
   // A short description of the header.
   description string

   // The type of the parameter / The internal type of the array.
   // Since the parameter is not located at the request body, it is limited to simple types (that is, not an object).
   // The value MUST be one of "string", "number", "integer", "boolean", "array" or "file" (Files and models are not allowed in arrays).
   // If type is "file", the consumes MUST be either "multipart/form-data" or " application/x-www-form-urlencoded" and the parameter MUST be in "formData".
   Type              string         `json:"type"`

   // Required if type is "array". Describes the type of items in the array.
   Items            *SwaggerItemT   `json:"items,omitempty"`

   // The extending format for the previously mentioned type. See Data Type Formats for further details.
   Format            string         `json:"format,omitempty"`

   // Declares the value of the parameter that the server will use if none is provided,
   // for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request.
   // (Note: "default" has no meaning for required parameters.)
   // See http://json-schema.org/latest/json-schema-validation.html#anchor101.
   // Unlike JSON Schema this value MUST conform to the defined type for this parameter.
   Default  interface{}             `json:"default,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   Maximum          string          `json:"maximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor17.
   ExclusiveMaximum bool            `json:"exclusiveMaximum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   Minimum          string          `json:"minimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor21.
   ExclusiveMinimum bool            `json:"exclusiveMinimum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor26.
   MaxLength        int64           `json:"maxLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor29.
   MinLength        int64           `json:"minLength,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor33.
   Pattern          string          `json:"pattern,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor42.
   MaxItems         int64           `json:"maxItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor45.
   MinItems         int64           `json:"minItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor49.
   UniqueItems      bool            `json:"uniqueItems,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor76.
   Enum           []interface{}     `json:"enum,omitempty"`

   // See http://json-schema.org/latest/json-schema-validation.html#anchor14.
   MultipleOf       string          `json:"multipleOf,omitempty"`
}

type SwaggerResponseT struct {
   // A short description of the response. GFM syntax can be used for rich text representation.
   Description string   `json:"description"`

   // A definition of the response structure.
   // It can be a primitive, an array or an object.
   // If this field does not exist, it means no content is returned as part of the response.
   // As an extension to the Schema Object, its root type value may also be "file".
   // This SHOULD be accompanied by a relevant produces mime-type.
   Schema   *SwaggerSchemaT `json:"schema,omitempty"`

   // A list of headers that are sent with the response.
   Headers  map[string]SwaggerHeaderT `json:"headers,omitempty"`

   // An example of the response message.
   Examples map[string]interface{} `json:"examples,omitempty"`
}

type SwaggerSecDefsT struct {
   // Validity: Any
   // Required. The type of the security scheme. Valid values are "basic", "apiKey" or "oauth2".
   Type  string `json:"type"`

   // Validity: Any
   // A short description for security scheme.
   Description string `json:"description,omitempty"`

   // Validity:  apiKey   Required.
   // The name of the header or query parameter to be used.
   Name  string `json:"name"`

   // Validity:  apiKey   Required
   // The location of the API key. Valid values are "query" or "header".
   In string `json:"in"`

   // Validity:  oauth2   Required.
   // The flow used by the OAuth2 security scheme. Valid values are "implicit", "password", "application" or "accessCode".
   Flow  string `json:"flow"`

   // Validity:  oauth2 ("implicit", "accessCode")   Required.
   // The authorization URL to be used for this flow. This SHOULD be in the form of a URL.
   AuthorizationUrl  string `json:"authorizationUrl"`

   // Validity:  oauth2 ("password", "application", "accessCode")   Required.
   // The token URL to be used for this flow. This SHOULD be in the form of a URL.
   TokenUrl string `json:"tokenUrl"`

   // Validity:  oauth2   Required.
   // The available scopes for the OAuth2 security scheme.
   // Maps between a name of a scope to a short description of it (as the value of the property).
   Scopes   map[string]string `json:"scopes"`
}

type SwaggerSecurityT []string

type SwaggerTagT struct {
   // Required. The name of the tag.
   Name  string `json:"name"`

   // A short description for the tag. GFM syntax can be used for rich text representation.
   Description string `json:"description,omitempty"`

   // Additional external documentation for this tag.
   ExternalDocs   *SwaggerExternalDocsT `json:"externalDocs,omitempty"`
}

type SwaggerExternalDocsT struct {
   // A short description of the target documentation.
   // GFM syntax can be used for rich text representation.
   Description string `json:"description,omitempty"`

   // The URL for the target documentation.
   // Value MUST be in the format of a URL.
   Url   string `json:"url"`
}


// All the fields comments are copied from swagger specification at http://swagger.io/specification/
// Stonelizard is not entirely compliant with the specification right now.
// We tried to add notes (preceded by the 'SL:' notation) pointing the differences, but it is possible that
// we have missed some details here or there.
type SwaggerT struct {
   // Required. Specifies the Swagger Specification version being used.
   // It can be used by the Swagger UI and other clients to interpret the API listing.
   // The value MUST be "2.0".
   Version             string                 `json:"swagger"`

   // Required. Provides metadata about the API.
   // The metadata can be used by the clients if needed.
   Info                SwaggerInfoT           `json:"info"`

   // The host (name or ip) serving the API.
   // This MUST be the host only and does not include the scheme nor sub-paths.
   // It MAY include a port. If the host is not included, the host serving the documentation is to be used (including the port).
   // The host does not support path templating.
   Host                string                 `json:"host,omitempty"`

   // The base path on which the API is served, which is relative to the host.
   // If it is not included, the API is served directly under the host.
   // The value MUST start with a leading slash (/).
   // The basePath does not support path templating.
   BasePath            string                 `json:"basePath,omitempty"`

   // The transfer protocol of the API.
   // Values MUST be from the list: "http", "https", "ws", "wss".
   // If the schemes is not included, the default scheme to be used is the one used to access the Swagger definition itself.
   // SL: only "https" supported right now.
   Schemes           []string                 `json:"schemes,omitempty"`

   // A list of MIME types the APIs can consume.
   // This is global to all APIs but can be overridden on specific API calls.
   // Value MUST be as described under Mime Types.
   Consumes          []string                 `json:"consumes,omitempty"`

   // A list of MIME types the APIs can produce.
   // This is global to all APIs but can be overridden on specific API calls.
   // Value MUST be as described under Mime Types.
   Produces          []string                 `json:"produces,omitempty"`

   // The available paths and operations for the API.
   Paths    map[string]SwaggerPathT           `json:"paths"`

   // An object to hold data types produced and consumed by operations.
   Definitions         map[string]SwaggerSchemaT    `json:"definitions,omitempty"`

   // An object to hold parameters that can be used across operations.
   // This property does not define global parameters for all operations.
   Parameters        []SwaggerParameterT     `json:"parameters,omitempty"`

   // An object to hold responses that can be used across operations.
   // This property does not define global responses for all operations.
   Responses map[string]SwaggerResponseT      `json:"responses,omitempty"`

   // Security scheme definitions that can be used across the specification.
   SecurityDefinitions map[string]SwaggerSecDefsT        `json:"securityDefinitions,omitempty"`

   // A declaration of which security schemes are applied for the API as a whole.
   // The list of values describes alternative security schemes that can be used (that is, there is a logical OR between the security requirements).
   // Individual operations can override this definition.
   Security          []SwaggerSecurityT       `json:"security,omitempty"`

   // A list of tags used by the specification with additional metadata.
   // The order of the tags can be used to reflect on their order by the parsing tools.
   // Not all tags that are used by the Operation Object must be declared.
   // The tags that are not declared may be organized randomly or based on the tools' logic.
   // Each tag name in the list MUST be unique.
   Tags              []SwaggerTagT            `json:"tags,omitempty"`

   // Additional external documentation.
   ExternalDocs       *SwaggerExternalDocsT   `json:"externalDocs,omitempty"`
}

