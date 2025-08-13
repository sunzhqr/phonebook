package config

func (p Postgres) GetURL() string     { return p.URL }
func (p Postgres) GetMaxConns() int32 { return p.MaxConns }
func (p Postgres) GetMinConns() int32 { return p.MinConns }
