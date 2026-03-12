package stonelizard

import (
   "os"
   "fmt"
   "regexp"
)

func init() {
   // Used to remove paths from type names obtained using the reflect packages
   gorootRE = regexp.MustCompile(fmt.Sprintf(`^\s+%s%csrc[^\s]+\.go:[0-9]+`,os.Getenv("GOROOT"),os.PathSeparator))
   gosrcRE = regexp.MustCompile("^\\s+[^\\s]+\\.go:[0-9]+")
   gosrcFNameRE = regexp.MustCompile(`([^\s/]+/[^\s/]+\.go:[0-9]+)\s`) // path

   // Used to manually parse the reflect.StructTag to list all tags defined
   tagRE = regexp.MustCompile(`([\pL\pN\x21\x23-\x2f\x3b-\x40\x5b-\x60\x7b-\x7e]+):"([^"]+)"`)
}

