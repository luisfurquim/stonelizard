package sloauth2

func MkCookieId() string {
   var i, j int
   var buf []byte
   var id []byte

   buf = make([]byte, 48)
   id = make([]byte, 64)

   NewRandReader().Read(buf)

   for i<48 {
      id[j] = ckidchars[buf[i] & 0x3f]
      j++
      id[j] = ckidchars[(buf[i]>>6) | ((buf[i+1] & 0xf)<<2)]
      j++
      i++
      id[j] = ckidchars[(buf[i]>>4) | ((buf[i+1] & 0x3)<<4)]
      j++
      i++
      id[j] = ckidchars[buf[i] >> 2]
      j++
      i++
   }

   return string(id)
}
