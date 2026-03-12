package stonelizard

func NewWithPKI(pk Pki, svcs ...EndPointHandler) (*ServiceWithPKI, error) {
   var svc  *Service
   var err   error

   svc, err = New(svcs ...)
   if err != nil {
      return nil, err
   }

   return &ServiceWithPKI{
      Service: *svc,
      PK: pk,
   }, nil
}
