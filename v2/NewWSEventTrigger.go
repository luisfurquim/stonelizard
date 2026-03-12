package stonelizard

// Initializes a websocket event trigger
func NewWSEventTrigger() *WSEventTrigger {
   var wset WSEventTrigger
   wset = WSEventTrigger{
      EventData: make(chan interface{}),
      Status: false,
   }
   return &wset
}
