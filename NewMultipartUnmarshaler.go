package stonelizard

import (
   "net/http"
)

func NewMultipartUnmarshaler(r *http.Request, fields []string) (*MultipartUnmarshaler, error) {
   var err  error
   var m    MultipartUnmarshaler

   err = r.ParseMultipartForm(MaxUploadMemory)
   if err != nil {
      Goose.Serve.Logf(1,"Error Parsing multipart field: %s",err)
      return nil, err
   }

   m.form = r.MultipartForm
   m.fields = fields

   return &m, nil
}

