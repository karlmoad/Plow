package targets

import (
	"Plow/plow/objects"
	"Plow/plow/targets/common"
	sf "Plow/plow/targets/snowflake"
	"strings"
)

func NewTarget(target string, context *objects.PlowContext) (common.Target, error) {
	switch strings.TrimSpace(strings.ToUpper(target)) {
	case "SNOWFLAKE":
		{
			snowflake := &sf.SnowflakeTarget{}
			err := snowflake.Open(context)
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
