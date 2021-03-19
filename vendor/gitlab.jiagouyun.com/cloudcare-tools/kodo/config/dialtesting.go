package config

var (
	DialtestingCfg = DialtestingConfig{
		Database: DatabaseCfg{
			Dialect: "mysql",
		},

		LogConfig: LogCfg{
			LogFile:    `/logdata/dialtesting.log`,
			Level:      `info`,
			GinLogFile: `/logdata/dialtesing-gin.log`,
		},

		Global: GlobalCfg{
			Listen: `:9531`,
		},

		Regions: map[string]bool{},
	}
)

type DialtestingConfig struct {
	Database  DatabaseCfg     `yaml:"database"`
	LogConfig LogCfg          `yaml:"log"`
	Global    GlobalCfg       `yaml:"global"`
	Regions   map[string]bool `yaml:"regions"`
}
