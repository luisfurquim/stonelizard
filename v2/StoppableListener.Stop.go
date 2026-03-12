package stonelizard

//Stop the service and releases the channel
func (sl *StoppableListener) Stop() {
   close(sl.stop)
}


