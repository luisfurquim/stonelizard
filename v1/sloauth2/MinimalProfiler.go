package sloauth2

type MinimalProfiler struct{
	name, email string
}


func (pf MinimalProfiler) New() Profiler {
	return pf
}

func (pf MinimalProfiler) Init(data map[string]interface{}) MinimalProfiler {
	var kprof string
	var vprof interface{}
	var ok bool
	var svprof string

	for kprof, vprof = range data {
		if svprof, ok = vprof.(string); ok {
			if reEMail.MatchString(svprof) {
				pf.email = svprof
				continue
			}

			if reName.MatchString(kprof) {
				pf.name = svprof
			}
		}
	}

	return pf
}

func (pf MinimalProfiler) Id() string {
   return ""
}

func (pf MinimalProfiler) Email() string {
   return pf.email
}

func (pf MinimalProfiler) Nick() string {
   return ""
}

func (pf MinimalProfiler) Login() string {
   return ""
}

func (pf MinimalProfiler) Name() string {
   return pf.name
}

func (pf MinimalProfiler) SurName() string {
   return ""
}

func (pf MinimalProfiler) Avatar() string {
   return ""
}

