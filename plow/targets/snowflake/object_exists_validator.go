package snowflake

import (
	"Plow/plow/objects"
	"Plow/plow/targets/common"
	"database/sql"
	"github.com/noirbizarre/gonja"
	"strings"
)

type SnowflakeObjectExistsValidator struct {
	meta        *common.Metadata
	db          *sql.DB
	target      *SnowflakeTarget
	changes     *objects.ChangeLog
	initialized bool
}

func newSnowflakeObjectExistsValidator(snowflake *SnowflakeTarget, changes *objects.ChangeLog) *SnowflakeObjectExistsValidator {
	return &SnowflakeObjectExistsValidator{
		db:      snowflake.connection,
		target:  snowflake,
		meta:    common.NewMetadata(StringToSnowflakeObjectTypeInt64),
		changes: changes,
	}
}

func (sfev *SnowflakeObjectExistsValidator) Init() error {
	if !sfev.initialized {
		err := sfev.loadMeta(sfev.identifyChangeDatabases(sfev.changes), sfev.meta)
		if err != nil {
			return err
		}
		sfev.initialized = true
	}
	return nil
}

func (sfev *SnowflakeObjectExistsValidator) Destroy() error {
	return nil
}

func (sfev *SnowflakeObjectExistsValidator) Designation() string {
	return "ObjectExistsValidator"
}

func (sfev *SnowflakeObjectExistsValidator) Validate(change *objects.ChangeItem) error {
	metaobj, err := sfev.meta.FindObjectFromSpec(change.Item)
	if err != nil {
		change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical, false, err, sfev.Designation())
		return err
	}

	change.Validation.AddValidationStepInfo(objects.ValidationErrorNone, true, nil, sfev.Designation())

	if metaobj != nil && err == nil {
		change.ExistsFlag = true //set the exists flag so downstream validators can consume
	}

	return nil
}

func (sfev *SnowflakeObjectExistsValidator) identifyChangeDatabases(changes *objects.ChangeLog) map[string]bool {
	databases := make(map[string]bool)
	//loop changes, eval object spec pull database names to extract meta information for
	for _, bundle := range changes.Bundles {
		for _, change := range bundle.Items {
			sfType := StringToSnowflakeObjectType(change.Item.Type)
			switch sfType {
			case UnknownType:
				break
			case Database:
				databases[strings.ToUpper(change.Item.Object.Name)] = true
				break
			default:
				if len(strings.TrimSpace(change.Item.Object.Database)) > 0 {
					databases[strings.ToUpper(change.Item.Object.Database)] = true
				}
				break
			}
		}
	}
	return databases
}

func (sfev *SnowflakeObjectExistsValidator) loadMeta(databases map[string]bool, meta *common.Metadata) error {
	prunedDbList, err := sfev.loadDatabasesMeta(meta, databases)
	if err != nil {
		return err
	}
	for _, database := range prunedDbList {
		err = sfev.loadDatabaseMeta(database, meta)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sfev *SnowflakeObjectExistsValidator) loadDatabasesMeta(meta *common.Metadata, databases map[string]bool) ([]string, error) {
	output := make([]string, 0)
	stmt, err := common.RenderStatement(GetDatabasesSQL, &gonja.Context{"DATABASE": sfev.target.config.Database})
	if err != nil {
		return nil, err
	}

	rows, err := sfev.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	var dbname string
	for rows.Next() {
		var err error
		if err = rows.Scan(&dbname); err == nil {
			if _, ok := databases[strings.ToUpper(dbname)]; ok {
				output = append(output, dbname)
				if metaObject, err := common.NewMetadataObject(int64(Database), common.Property{Name: "name", Value: dbname, IsKey: true}); err == nil {
					meta.AddObject(metaObject)
				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}
	return output, nil
}

func (sfev *SnowflakeObjectExistsValidator) loadDatabaseMeta(database string, meta *common.Metadata) error {
	steps := []func(string, *common.Metadata) error{sfev.loadSchemaMeta, sfev.loadTableViewMeta}
	for _, step := range steps {
		err := step(database, meta)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sfev *SnowflakeObjectExistsValidator) loadSchemaMeta(database string, meta *common.Metadata) error {
	stmt, err := common.RenderStatement(GetSchemasSQL, &gonja.Context{"DATABASE": database})
	if err != nil {
		return err
	}

	rows, err := sfev.db.Query(stmt)
	if err != nil {
		return err
	}

	var schemaname string
	for rows.Next() {
		var err error
		if err = rows.Scan(&schemaname); err == nil {
			if metaObject, err := common.NewMetadataObject(int64(Schema),
				common.Property{Name: "name", Value: schemaname, IsKey: true},
				common.Property{Name: "database", Value: database}); err == nil {
				meta.AddObject(metaObject)
			} else {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (sfev *SnowflakeObjectExistsValidator) loadTableViewMeta(database string, meta *common.Metadata) error {
	stmt, err := common.RenderStatement(GetTablesViewsSQL, &gonja.Context{"DATABASE": database})
	if err != nil {
		return err
	}

	rows, err := sfev.db.Query(stmt)
	if err != nil {
		return err
	}

	var schemaname, tablename, tbltype string
	for rows.Next() {
		var err error
		if err = rows.Scan(&schemaname, &tablename, &tbltype); err == nil {
			otype := int64(Table) //type is table unless view indicated
			if strings.Contains(strings.ToUpper(tbltype), "VIEW") {
				otype = int64(View)
			}

			if metaObject, err := common.NewMetadataObject(otype,
				common.Property{Name: "name", Value: tablename, IsKey: true},
				common.Property{Name: "schema", Value: schemaname},
				common.Property{Name: "database", Value: database}); err == nil {
				meta.AddObject(metaObject)
			} else {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
