package model

type Cfg struct {
	Scope      []string `json:"scope"`
	NmapArgs   []string `json:"nmap_args"`
	SqlmapArgs []string `json:"sqlmap_args"`
	DBPath     string   `json:"db_path"`
	Workers    int      `json:"workers"`
	OutDir     string   `json:"out_dir"`
}

type Host struct {
	ID    int64
	Cyc   int64
	IP    string
	State string
	Ports []PortSvc
}

type PortSvc struct {
	ID      int64
	HostID  int64
	Port    int
	Proto   string
	Service string
	Version string
	State   string
}

type WebEP struct {
	ID     int64
	HostID int64
	URL    string
}

type Finding struct {
	ID   int64
	EPID int64
	Typ  string
	Det  string
	Raw  string
}

type Change struct {
	ID  int64
	Cyc int64
	Ent string
	Key string
	Typ string
	Det string
}
