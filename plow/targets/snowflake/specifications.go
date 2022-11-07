package snowflake

import "Plow/plow/objects"

type sfSpecMetadata struct {
	Owner *objects.ObjectDesignation `yaml:"owner,omitempty"`
}

type sfRoleAssoc struct {
	Role   string                    `yaml:"role"`
	Object objects.ObjectDesignation `yaml:"object"`
}
type sfRoleSpecification struct {
	Roles   []string      `yaml:"roles"`
	Grants  []sfRoleAssoc `yaml:"grants"`
	Revokes []sfRoleAssoc `yaml:"revoke"`
}

type sfDefaultSpecification struct {
	Metadata sfSpecMetadata `yaml:"meta"`
	Pre      string         `yaml:"pre"`
	Init     string         `yaml:"init"`
	Change   string         `yaml:"change"`
	Post     string         `yaml:"post"`
}

type sfUsageSpecification struct {
	Grants  []objects.ObjectDesignation `yaml:"grants"`
	Revokes []objects.ObjectDesignation `yaml:"revoke"`
}

type sfDatabaseSpecification struct {
	Owner objects.ObjectDesignation `yaml:"owner"`
	Usage sfUsageSpecification      `yaml:"usage"`
}

type sfSchemaSpecification struct {
	Owner objects.ObjectDesignation `yaml:"owner"`
	Usage sfUsageSpecification      `yaml:"usage"`
}

type sfWarehouseSpecification struct {
	Owner objects.ObjectDesignation `yaml:"owner"`
	Usage sfUsageSpecification      `yaml:"usage"`
}
