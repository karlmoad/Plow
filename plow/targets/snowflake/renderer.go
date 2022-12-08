package snowflake

import (
	"Plow/plow/objects"
	"Plow/plow/targets/common"
	"Plow/plow/utility"
	"github.com/noirbizarre/gonja"
	"regexp"
)

var (
	regexUseRoleCommand     = regexp.MustCompile(`(?ims)USE\s+ROLE\s+([a-zA-Z0-9\_\-]+)+\;?`)
	regexDisallowedCommands = [...]*regexp.Regexp{regexUseRoleCommand}
)

type SnowflakeRenderer struct {
	defaultRole          string
	warehouseCoordinator *WarehouseUnitCoordinator
}

func evalAllowedCommands(input string) bool {
	for _, rgex := range regexDisallowedCommands {
		if rgex.MatchString(input) {
			return false
		}
	}
	return true
}

func newSnowflakeRenderer(role string, warehouseName string) *SnowflakeRenderer {
	return &SnowflakeRenderer{
		defaultRole:          role,
		warehouseCoordinator: newWarehouseUnitCoordinator(warehouseName, role),
	}
}

func (sfr *SnowflakeRenderer) GetWarehouseCoordinator() *WarehouseUnitCoordinator {
	return sfr.warehouseCoordinator
}

func (sfr *SnowflakeRenderer) RenderWithContext(change *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {

	switch StringToSnowflakeObjectType(change.Item.Type) {
	case Role:
		{
			spec := &sfRoleSpecification{}
			err := utility.UnmarshalYamlSubObject(change.Item.Spec, spec)
			if err != nil {
				return nil, err
			}

			return sfr.renderRoleSpec(spec, change, params)
		}
	case Database:
		{
			spec := &sfDatabaseSpecification{}
			err := utility.UnmarshalYamlSubObject(change.Item.Spec, spec)
			if err != nil {
				return nil, err
			}

			return sfr.renderDatabaseSpec(spec, change, params)

		}
	case Schema:
		{
			spec := &sfSchemaSpecification{}
			err := utility.UnmarshalYamlSubObject(change.Item.Spec, spec)
			if err != nil {
				return nil, err
			}

			return sfr.renderSchemaSpec(spec, change, params)
		}
	case Warehouse:
		{
			spec := &sfWarehouseSpecification{}
			err := utility.UnmarshalYamlSubObject(change.Item.Spec, spec)
			if err != nil {
				return nil, err
			}

			return sfr.renderWarehouseSpec(spec, change, params)
		}
	default:
		{
			spec := &sfDefaultSpecification{}
			err := utility.UnmarshalYamlSubObject(change.Item.Spec, spec)
			if err != nil {
				return nil, err
			}
			return sfr.renderDefaultSpec(spec, change, params)
		}
	}
}

func (sfr *SnowflakeRenderer) Render(change *objects.ChangeItem, context *objects.PlowContext) ([]*objects.ApplyScope, error) {
	params, err := change.Item.ExtractVariableValues(context)
	if err != nil {
		return nil, err
	}
	return sfr.RenderWithContext(change, params)
}

func (sfr *SnowflakeRenderer) renderRoleSpec(spec *sfRoleSpecification, item *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {
	stmts := make([]string, 0)
	stmts = append(stmts, generateUseRoleStmt(string(SECURITYADMIN)))

	for _, role := range spec.Roles {
		stmt, err := common.RenderStatement(CreateRoleSQL, &gonja.Context{"ROLE": role})
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)

		stmt, err = common.RenderStatement(GrantRoleToRoleSQL, &gonja.Context{"ROLE": role, "ROLENAME": sfr.defaultRole})
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	//process role grants
	for _, grant := range spec.Grants {
		variables := make(map[string]interface{})
		if !IsProtectedSystemRole(grant.Role) {
			variables["ROLE"] = grant.Role
			var stmt string
			switch StringToSnowflakeObjectType(grant.Object.ObjectType) {
			case Role:
				{
					stmt = GrantRoleToRoleSQL
					variables["ROLENAME"] = grant.Object.Identifier
					break
				}
			case User:
				{
					stmt = GrantRoleToUserSQL
					variables["USER"] = grant.Object.Identifier
					break
				}
			}
			stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}
	}

	//process role revokes
	for _, revoke := range spec.Revokes {
		variables := make(map[string]interface{})
		if !IsProtectedSystemRole(revoke.Role) {
			variables["ROLE"] = revoke.Role
			var stmt string
			switch StringToSnowflakeObjectType(revoke.Object.ObjectType) {
			case Role:
				{
					stmt = RevokeRoleToRoleSQL
					variables["ROLENAME"] = revoke.Object.Identifier
					break
				}
			case User:
				{
					stmt = RevokeRoleToUserSQL
					variables["USER"] = revoke.Object.Identifier
					break
				}
			}
			stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}
	}

	stmts = append(stmts, generateUseRoleStmt(sfr.defaultRole)) //set role back to default for good measure

	return []*objects.ApplyScope{common.NewScope("role", stmts)}, nil
}

func (sfr *SnowflakeRenderer) renderDatabaseSpec(spec *sfDatabaseSpecification, item *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {
	stmts := make([]string, 0)
	vars := params

	sfr.addOwnerToWarehouseCoordinator(spec.Owner)

	if item.Item.Options.Drop {
		if StringToSnowflakeObjectType(spec.Owner.ObjectType) == Role && !utility.IsStringEmpty(&spec.Owner.Identifier) {
			vars["ROLE"] = spec.Owner.Identifier
			useRoleStmt := generateUseRoleStmt(spec.Owner.Identifier)
			dropStmt, err := common.RenderStatement(DropDatabaseSQL, (*gonja.Context)(&vars))
			if err != nil {
				return nil, err
			}

			stmts = append(stmts, useRoleStmt)
			stmts = append(stmts, dropStmt)

		}
	} else {
		if !item.ExistsFlag {
			//create if not exists
			stmt, err := common.RenderStatement(CreateDatabaseSQL, (*gonja.Context)(&vars))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)

			stmt, err = common.RenderStatement(DropDefaultPublicSchemaSQL, (*gonja.Context)(&vars))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)

			//apply ownership
			if StringToSnowflakeObjectType(spec.Owner.ObjectType) == Role && !utility.IsStringEmpty(&spec.Owner.Identifier) {
				vars["ROLE"] = spec.Owner.Identifier
				stmt, err := common.RenderStatement(GrantDatabaseOwnershipSQL, (*gonja.Context)(&vars))
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, stmt)

				//render command to use new owner
				stmts = append(stmts, generateUseRoleStmt(spec.Owner.Identifier))

				//if owner has been modified
				//add default usage grant to default change mgmt role for future needs
				vars["ROLE"] = sfr.defaultRole
				stmt, err = common.RenderStatement(GrantUsageToRoleDatabaseSQL, (*gonja.Context)(&vars))
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, stmt)
			}
		}

		//process grants
		for _, grant := range spec.Usage.Grants {
			variables := utility.DeepMapCopy(params)

			var stmt string
			switch StringToSnowflakeObjectType(grant.ObjectType) {
			case Role:
				{
					stmt = GrantUsageToRoleDatabaseSQL
					variables["ROLE"] = grant.Identifier
					break
				}
			case User:
				{
					stmt = GrantUsageToUserDatabaseSQL
					variables["USER"] = grant.Identifier
					break
				}
			}
			stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}

		//process revokes
		for _, revoke := range spec.Usage.Revokes {
			variables := utility.DeepMapCopy(params)

			var stmt string
			switch StringToSnowflakeObjectType(revoke.ObjectType) {
			case Role:
				{
					stmt = RevokeUsageToRoleDatabaseSQL
					variables["ROLE"] = revoke.Identifier
					break
				}
			case User:
				{
					stmt = RevokeUsageToUserDatabaseSQL
					variables["USER"] = revoke.Identifier
					break
				}
			}
			stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}
	}
	return []*objects.ApplyScope{common.NewScope("database", stmts)}, nil
}

func (sfr *SnowflakeRenderer) renderSchemaSpec(spec *sfSchemaSpecification, item *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {
	stmts := make([]string, 0)
	vars := params

	sfr.addOwnerToWarehouseCoordinator(spec.Owner)

	if item.Item.Options.Drop {
		if StringToSnowflakeObjectType(spec.Owner.ObjectType) == Role && !utility.IsStringEmpty(&spec.Owner.Identifier) {
			vars["ROLE"] = spec.Owner.Identifier
			useRoleStmt := generateUseRoleStmt(spec.Owner.Identifier)
			dropStmt, err := common.RenderStatement(DropSchemaSQL, (*gonja.Context)(&vars))
			if err != nil {
				return nil, err
			}

			stmts = append(stmts, useRoleStmt)
			stmts = append(stmts, dropStmt)
		}
	} else {

		//if owner is defined switch to owner role
		if StringToSnowflakeObjectType(spec.Owner.ObjectType) == Role && !utility.IsStringEmpty(&spec.Owner.Identifier) {
			//render command to use new owner
			stmts = append(stmts, generateUseRoleStmt(spec.Owner.Identifier))
		}

		//create if not exists
		stmt, err := common.RenderStatement(CreateSchemaSQL, (*gonja.Context)(&vars))
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)

		//add usage grant to default change mgmt role for future needs
		vars["ROLE"] = sfr.defaultRole
		stmt, err = common.RenderStatement(GrantUsageToRoleSchemaSQL, (*gonja.Context)(&vars))
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)

		//process grants
		for _, grant := range spec.Usage.Grants {
			variables := utility.DeepMapCopy(params)

			var stmt string
			switch StringToSnowflakeObjectType(grant.ObjectType) {
			case Role:
				{
					stmt = GrantUsageToRoleSchemaSQL
					variables["ROLE"] = grant.Identifier
					break
				}
			case User:
				{
					stmt = GrantUsageToUserSchemaSQL
					variables["USER"] = grant.Identifier
					break
				}
			}
			stmt, err = common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}

		//process revokes
		for _, revoke := range spec.Usage.Revokes {
			variables := utility.DeepMapCopy(params)

			var stmt string
			switch StringToSnowflakeObjectType(revoke.ObjectType) {
			case Role:
				{
					stmt = RevokeUsageToRoleSchemaSQL
					variables["ROLE"] = revoke.Identifier
					break
				}
			case User:
				{
					stmt = RevokeUsageToUserSchemaSQL
					variables["USER"] = revoke.Identifier
					break
				}
			}
			stmt, err = common.RenderStatement(stmt, (*gonja.Context)(&variables))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)
		}
	}

	return []*objects.ApplyScope{common.NewScope("schema", stmts)}, nil
}

func (sfr *SnowflakeRenderer) renderDefaultSpec(spec *sfDefaultSpecification, item *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {
	out := make([]*objects.ApplyScope, 0)
	var err error
	var scope *objects.ApplyScope

	if spec.Metadata.Owner != nil {
		sfr.addOwnerToWarehouseCoordinator(*spec.Metadata.Owner)
	}

	//apply role based on config to this rendered block of work
	//role can not be a protected role. if protected role, error is produced
	//if protected role is required create custom spec type/handler to protect from injection of unintended role escalation
	if spec.Metadata.Owner != nil {
		if StringToSnowflakeObjectType(spec.Metadata.Owner.ObjectType) == Role && !utility.IsStringEmpty(&spec.Metadata.Owner.Identifier) {
			if !IsProtectedSystemRole(spec.Metadata.Owner.Identifier) {
				out = append(out, common.NewScope("security", []string{generateUseRoleStmt(spec.Metadata.Owner.Identifier)}))
			} else {
				return nil, ErrDisallowedPrivilegedRole
			}
		}
	}

	//pre scope statements are always applied if present in the spec
	if !utility.IsStringEmpty(&spec.Pre) {
		gc := gonja.Context(params)
		if scope, err = renderSpecStatement(spec.Pre, "pre", &gc); err == nil {
			out = append(out, scope)
		} else {
			return nil, err
		}
	}

	//init and change scope statements execution depend on if the object exists, which was determined during validation
	//if change scope is present a corresponding init scope must also present, change scope can be omitted
	//if the objects exists change scope is applied, otherwise the init scope is applied

	initPresent := !utility.IsStringEmpty(&spec.Init)
	changePresent := !utility.IsStringEmpty(&spec.Change)

	if initPresent {
		if item.ExistsFlag {
			if changePresent {
				gc := gonja.Context(params)
				if scope, err = renderSpecStatement(spec.Change, "change", &gc); err == nil {
					out = append(out, scope)
				} else {
					return nil, err
				}
			}
		} else {
			gc := gonja.Context(params)
			if scope, err = renderSpecStatement(spec.Init, "init", &gc); err == nil {
				out = append(out, scope)
			} else {
				return nil, err
			}
		}
	}

	//post scope statements are always applied if present in the spec
	if !utility.IsStringEmpty(&spec.Post) {
		gc := gonja.Context(params)
		if scope, err = renderSpecStatement(spec.Post, "post", &gc); err == nil {
			out = append(out, scope)
		} else {
			return nil, err
		}
	}

	return out, nil
}

func (sfr *SnowflakeRenderer) renderWarehouseSpec(spec *sfWarehouseSpecification, item *objects.ChangeItem, params map[string]interface{}) ([]*objects.ApplyScope, error) {
	stmts := make([]string, 0)

	sfr.addOwnerToWarehouseCoordinator(spec.Owner)

	//render statement setting role to default change mgmt role
	stmts = append(stmts, generateUseRoleStmt(sfr.defaultRole))

	//process usage grants/revokes
	//process grants
	for _, grant := range spec.Usage.Grants {
		variables := utility.DeepMapCopy(params)

		var stmt string
		switch StringToSnowflakeObjectType(grant.ObjectType) {
		case Role:
			{
				stmt = GrantUsageOnWarehouseToRoleSQL
				variables["ROLE"] = grant.Identifier
				break
			}
		case User:
			{
				stmt = GrantUsageOnWarehouseToUserSQL
				variables["USER"] = grant.Identifier
				break
			}
		}
		stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	//process revokes
	for _, revoke := range spec.Usage.Revokes {
		variables := utility.DeepMapCopy(params)

		var stmt string
		switch StringToSnowflakeObjectType(revoke.ObjectType) {
		case Role:
			{
				stmt = RevokeUsageOnWarehouseToRoleSQL
				variables["ROLE"] = revoke.Identifier
				break
			}
		case User:
			{
				stmt = RevokeUsageOnWarehouseToUserSQL
				variables["USER"] = revoke.Identifier
				break
			}
		}
		stmt, err := common.RenderStatement(stmt, (*gonja.Context)(&variables))
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	return []*objects.ApplyScope{common.NewScope("warehouse", stmts)}, nil
}

func (sfr *SnowflakeRenderer) addOwnerToWarehouseCoordinator(owner objects.ObjectDesignation) {
	//only add owners who are of type role, this system can not assume the identity of a specific user in process
	//so using the coordinator to grant usage to the warehouse for a user is not advised
	if StringToSnowflakeObjectType(owner.ObjectType) == Role {
		sfr.warehouseCoordinator.AddOwner(owner.Identifier)
	}
}

func renderSpecStatement(input string, name string, params *gonja.Context) (*objects.ApplyScope, error) {
	stmt, err := common.RenderStatement(input, params)
	if err != nil {
		return nil, err
	}
	commands := common.SegmentScopeCommands(stmt)
	if !evaluateCommands(commands) {
		return nil, ErrInvalidUnapprovedCommand
	}

	return common.NewScope(name, commands), nil
}

func evaluateCommands(commands []string) bool {
	return utility.All(commands, evalAllowedCommands)
}

func generateUseRoleStmt(role string) string {
	if stmt, err := common.RenderStatement(UseRoleSQL, &gonja.Context{"ROLE": role}); err == nil {
		return stmt
	}
	return ""
}
