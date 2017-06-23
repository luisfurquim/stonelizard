package stonelizard

import (
   "os"
   "regexp"
)

func init() {
   // Used to remove paths from type names obtained using the reflect packages
   gorootRE = regexp.MustCompile("^\\s+" + os.Getenv("GOROOT") + "[^\\s]+\\.go:[0-9]+")
   gosrcRE = regexp.MustCompile("^\\s+[^\\s]+\\.go:[0-9]+")

   // Used to manually parse the reflect.StructTag to list all tags defined
   tagRE = regexp.MustCompile(`([\pL\pN\x21\x23-\x2f\x3b-\x40\x5b-\x60\x7b-\x7e]+):"([^"]+)"`)
}

