package targets

import (
	"Plow/plow/objects"
	"Plow/plow/secrets"
	"Plow/plow/targets/common"
	sf "Plow/plow/targets/snowflake"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func NewTarget(target string, config map[string]interface{}, options *objects.Options, secrets secrets.SecretStore) (common.Target, error) {
	switch strings.TrimSpace(strings.ToUpper(target)) {
	case "SNOWFLAKE":
		{
			var sfconfig sf.SnowflakeConfiguration
			err := mapstructure.Decode(config, &sfconfig)
			if err != nil {
				return nil, err
			}

			snowflake := &sf.SnowflakeTarget{}
			err = snowflake.Open(sfconfig, options, secrets)
			if err != nil {
				return nil, err
			}

			return snowflake, nil
		}
	default:
		{
			return nil, common.ErrInvalidTargetType
		}
	}
}
