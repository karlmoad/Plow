package snowflake

import "strings"

type SnowflakeObjectType int64

const (
	UnknownType SnowflakeObjectType = iota
	Warehouse
	Database
	Schema
	Table
	View
	Procedure
	UserDefinedFunction
	Role
	Security
	ResourceMonitor
	Stage
	Pipe
	Stream
	Task
	Sequence
	User
)

func (s SnowflakeObjectType) ToInt64() int64 {
	return int64(s)
}

var SnowflakeProcessingOrder = [...]SnowflakeObjectType{Role, Warehouse, Database, Schema, Table, View, Procedure, UserDefinedFunction, ResourceMonitor, Stage, Pipe, Stream, Task, Sequence}

func StringToSnowflakeObjectTypeInt64(s string) int64 {
	return int64(StringToSnowflakeObjectType(s))
}

func StringToSnowflakeObjectType(s string) SnowflakeObjectType {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "warehouse":
		return Warehouse
	case "database":
		return Database
	case "schema":
		return Schema
	case "table":
		return Table
	case "view":
		return View
	case "procedure":
		return Procedure
	case "sproc", "storedprocedure":
		return Procedure
	case "udf", "userdefinedfunction":
		return UserDefinedFunction
	case "role":
		return Role
	case "security":
		return Security
	case "resourcemonitor":
		return ResourceMonitor
	case "stage":
		return Stage
	case "pipe":
		return Pipe
	case "stream":
		return Stream
	case "task":
		return Task
	case "sequence", "seq":
		return Sequence
	case "user":
		return User
	default:
		return UnknownType
	}
}
