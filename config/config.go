package config

type ServerConfig struct {
	Units []*Unit `mapstructure:"units"`
}
type Unit struct {
	Exec    string `mapstructure:"exec"`
	WorkDir string `mapstructure:"work_dir"`
	Disable bool   `mapstructure:"disable"`
}
