package server

type Deps struct {
	Backup         map[string]interface{} `json:"backup,omitempty"`
	PMM            map[string]interface{} `json:"pmm,omitempty"`
	ProxySQL       map[string]interface{} `json:"proxy_sql,omitempty"`
	Haproxy        map[string]interface{} `json:"haproxy,omitempty"`
	LogCollector   map[string]interface{} `json:"logCollector,omitempty"`
	PgBackrest     map[string]interface{} `json:"pgbackrest,omitempty"`
	PgBackrestRepo map[string]interface{} `json:"pgbackrest_repo,omitempty"`
	Pgbadger       map[string]interface{} `json:"pgbadger,omitempty"`
	Pgbouncer      map[string]interface{} `json:"pgbouncer,omitempty"`
	Orchestrator   map[string]interface{} `json:"orchestrator,omitempty"`
	Router         map[string]interface{} `json:"router,omitempty"`
	Toolkit        map[string]interface{} `json:"toolkit,omitempty"`
}
