package snowflake

import (
	"Plow/plow/targets/common"
	"database/sql"
	"errors"
	"fmt"
	"github.com/noirbizarre/gonja"
)

// Manages temporary usage grants to any role identified as an owner within the target system and identified by the owning operation
// This is to allow execution of specs/scopes under an assumed role of the owner while having rights to access
// the change mgmt processes warehouse.
//
// To use initialize the coordinator with a set of owner(s) role names, ONLY ROLES CAN BE USED IN THIS MANNER,
// due to the system ever assuming a specific users identity
//
// Prior to any other spec/scope command execution, Activate is called and provided with the connection to the database.
// A grant on usage for the configured warehouse for this solution is applied for each owner role in the list
//
// When complete the owning operation deactivates the coordinator which in turn revokes all previous grants
// this should ideally be utilized with a deferred method call to assure execution occurs, due to this if an error
// is encountered during deactivation a panic is thrown
// .........................................................

type WarehouseUnitCoordinator struct {
	connection *sql.DB
	warehouse  string
	userole    string
	owners     []string
	activated  bool
}

func newWarehouseUnitCoordinator(warehouseName string, role string) *WarehouseUnitCoordinator {
	return &WarehouseUnitCoordinator{connection: nil, warehouse: warehouseName, activated: false, userole: role}
}

func (whc *WarehouseUnitCoordinator) AddOwner(owner string) {
	if whc.owners == nil {
		whc.owners = make([]string, 0)
	}
	whc.owners = append(whc.owners, owner)
}

func (whc *WarehouseUnitCoordinator) Activate(connection *sql.DB) error {
	whc.connection = connection
	err := whc.runStatementForOwners(GrantUsageOnWarehouseToRoleSQL)
	if err != nil {
		return errors.New("unable to activate warehouse unit coordinator: " + err.Error())
	}
	whc.activated = true
	return nil
}

func (whc *WarehouseUnitCoordinator) DeActivate() {
	if whc.activated {
		err := whc.runStatementForOwners(RevokeUsageOnWarehouseToRoleSQL)
		if err != nil {
			panic(errors.New("error deactivating warehouse unit coordinator: " + err.Error()))
		}
		whc.activated = false
	}
}

func (whc *WarehouseUnitCoordinator) runStatementForOwners(template string) error {
	_, err := whc.connection.Exec(fmt.Sprintf("USE ROLE %s;", whc.userole))
	if err != nil {
		return err
	}

	for _, owner := range whc.owners {
		stmt, err := common.RenderStatement(template, &gonja.Context{"NAME": whc.warehouse, "ROLE": owner})
		if err != nil {
			return err
		}

		_, err = whc.connection.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
