package types

type ListenCfg = struct {
	Listen    string `toml:"listen"`
	SslListen string `toml:"ssl_listen"`
	SslCert   string `toml:"ssl_cert"`
	SslKey    string `toml:"ssl_key"`
}

type GlobalCfg = struct {
	DataDir string   `toml:"data_dir"`
	TxsDirs []string `toml:"txs_dirs"`
	Network int64    `toml:"network"`
	Debug   int64    `toml:"debug"`
}
