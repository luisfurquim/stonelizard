package stonelizard

func NewWSEventTrigger() *WSEventTrigger {
   var wset WSEventTrigger
   wset = WSEventTrigger{
      ch: make(chan interface{}),
      stat: true,
   }
   return &wset
}
