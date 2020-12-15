package server

type Deps struct {
	Backup       map[string]interface{} `json:"backup,omitempty"`
	PMM          map[string]interface{} `json:"pmm,omitempty"`
	ProxySQL     map[string]interface{} `json:"proxy_sql,omitempty"`
	Haproxy      map[string]interface{} `json:"haproxy,omitempty"`
	LogCollector map[string]interface{} `json:"logCollector,omitempty"`
}
