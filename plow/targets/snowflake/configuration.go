package snowflake

type SnowflakeConfiguration struct {
	PublicKeyFile     string `mapstructure:"publicKeyFile"`
	UserId            string `mapstructure:"userId"`
	Account           string `mapstructure:"account"`
	Region            string `mapstructure:"region"`
	Database          string `mapstructure:"database"`
	Warehouse         string `mapstructure:"warehouse"`
	Role              string `mapstructure:"role"`
	KeyPasswordSecret string `mapstructure:"passwordSecret"`
}
