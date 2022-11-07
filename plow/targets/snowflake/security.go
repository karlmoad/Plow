package snowflake

import (
	"strings"
)

type SystemRole string

const (
	SYSADMIN      SystemRole = "SYSADMIN"
	ACCOUNTADMIN  SystemRole = "ACCOUNTADMIN"
	SECURITYADMIN SystemRole = "SECURITYADMIN"
	USERADMIN     SystemRole = "USERADMIN"
)

var protectedRoles = [...]SystemRole{SYSADMIN, ACCOUNTADMIN, SECURITYADMIN, USERADMIN}

func IsProtectedSystemRole(role string) bool {
	tmp := strings.ToUpper(strings.TrimSpace(role))
	for _, pRole := range protectedRoles {
		if strings.Compare(string(pRole), tmp) == 0 {
			return true
		}
	}
	return false
}
