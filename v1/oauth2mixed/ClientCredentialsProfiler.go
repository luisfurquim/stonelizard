package oauth2mixed

type ClientCredentialsProfiler struct{
	email string
}


func (pf ClientCredentialsProfiler) New() Profiler {
	return pf
}

func (pf ClientCredentialsProfiler) Init(data map[string]interface{}) ClientCredentialsProfiler {
	pf.email, _ = data["client_id"].(string)
	return pf
}

func (pf ClientCredentialsProfiler) Id() string {
   return ""
}

func (pf ClientCredentialsProfiler) Email() string {
   return pf.email
}

func (pf ClientCredentialsProfiler) Nick() string {
   return ""
}

func (pf ClientCredentialsProfiler) Login() string {
   return ""
}

func (pf ClientCredentialsProfiler) Name() string {
   return ""
}

func (pf ClientCredentialsProfiler) SurName() string {
   return ""
}

func (pf ClientCredentialsProfiler) Avatar() string {
   return ""
}

