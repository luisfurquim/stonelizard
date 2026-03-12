package cloauth2

import (
   "os"
   "io"
   "runtime"
   "crypto/sha1"
   "crypto/rand"
)


type RandomReader struct {
   rd io.Reader
}

func reverseXor(u uint64) uint64 {
   return u<<56 ^ u<<48 ^ u<<40 ^ u<<32 ^ u<<24 ^u<<16 ^ u<<8 ^ u
}

func toBytes(u uint64) [sha1.Size]byte {
   var i int
   var buf []byte
   var sha [sha1.Size]byte
   buf = make([]byte,8)

   for ; i<8; i++ {
      buf[i] = byte((u>>(i<<3))&0xff)
   }

   sha = sha1.Sum(buf)

   return sha
}

func sumBuf(a, b [sha1.Size]byte) {
   var i int

   for i=0; i<sha1.Size; i++ {
      a[i] += b[i]
   }
}


func xorBuf(a, b [sha1.Size]byte) {
   var i int

   for i=0; i<sha1.Size; i++ {
      a[i] ^= b[i]
   }
}


func (rd RandomReader) Read(p []byte) (n int, err error) {
   var i, j int
   var m runtime.MemStats
   var sum [sha1.Size]byte

   runtime.ReadMemStats(&m)
   sum = toBytes(m.Alloc)
   sumBuf(sum, toBytes(m.TotalAlloc))
   xorBuf(sum, toBytes(m.Sys))

   sumBuf(sum, toBytes(m.Lookups))
   xorBuf(sum, toBytes(m.Mallocs))

   sumBuf(sum, toBytes(m.TotalAlloc))
   xorBuf(sum, toBytes(m.HeapAlloc))

   sumBuf(sum, toBytes(m.Frees))
   xorBuf(sum, toBytes(m.HeapSys))

   sumBuf(sum, toBytes(m.HeapIdle))
   xorBuf(sum, toBytes(m.HeapInuse))

   sumBuf(sum, toBytes(m.HeapReleased))
   xorBuf(sum, toBytes(m.HeapObjects))

   sumBuf(sum, toBytes(m.StackInuse))
   xorBuf(sum, toBytes(m.StackSys))

   sumBuf(sum, toBytes(m.MSpanInuse))
   xorBuf(sum, toBytes(m.MSpanSys))

   sumBuf(sum, toBytes(m.MCacheInuse))
   xorBuf(sum, toBytes(m.MCacheSys))

   sumBuf(sum, toBytes(m.BuckHashSys))
   xorBuf(sum, toBytes(m.GCSys))

   sumBuf(sum, toBytes(m.OtherSys))
   xorBuf(sum, toBytes(m.NextGC))

   sumBuf(sum, toBytes(m.LastGC))
   xorBuf(sum, toBytes(m.PauseTotalNs))

   sumBuf(sum, toBytes(uint64(len(os.Environ()))))
   xorBuf(sum, toBytes(uint64(os.Getpid())))

   sumBuf(sum, toBytes(uint64(os.Getppid())))

   n, err = rd.rd.Read(p)
   if err != nil {
      return n, err
   }

   j = 0
   for i=0; i<n; i++ {
      p[i] ^= sum[j]
      j++
      if j>=sha1.Size {
         j = 0
      }
   }

   return n, nil
}

func (r RandomReader) Prob() float32 {
   var buf []byte

   buf = make([]byte,4)
   r.Read(buf)

   return float32(
      uint(buf[0]) |
      (uint(buf[1]) << 8) |
      (uint(buf[2]) <<16) |
      (uint(buf[3]) <<24)) /
      float32(uint(0xffffffff))

}

func NewRandReader() RandomReader {
   return RandomReader{
      rd: rand.Reader,
   }
}

func RandReader() io.Reader {
   return rand.Reader
}