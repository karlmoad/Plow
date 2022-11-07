package snowflake

import "errors"

var (
	ErrAllHellHasBrokenLoose    = errors.New("snowflake target critical error")
	ErrDisallowedPrivilegedRole = errors.New("execution not approved using privileged role")
	ErrInvalidUnapprovedCommand = errors.New("invalid or unapproved command")
	ErrUnableSetRoleContext     = errors.New("unable to establish execution role context")
)
