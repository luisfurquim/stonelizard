package stonelizard


//Close the webservices and the listeners
func (svc Service) Close() {
   Goose.Listener.Logf(2,"Closing listeners")
   svc.Listener.Stop()
   svc.Listener.Close()
   svc.CRLListener.Stop()
   svc.CRLListener.Close()
   Goose.Listener.Logf(3,"All listeners closed")
}

