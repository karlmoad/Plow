package snowflake

import (
	"Plow/plow/objects"
	"Plow/plow/targets/common"
	"Plow/plow/utility"
	"database/sql"
	"errors"
	"github.com/noirbizarre/gonja"
	"strings"
)

type columnInformation struct {
	Name             sql.NullString
	Position         sql.NullInt64
	Nullable         sql.NullString
	DataType         sql.NullString
	MaxCharLen       sql.NullInt64
	NumericPrecision sql.NullInt64
	NumericScale     sql.NullInt64
}

type columnVerify struct {
	Origin   columnInformation
	Validate columnInformation
}

type SnowflakeTableStructureValidator struct {
	db           *sql.DB
	databaseName string
}

func newSnowflakeTableStructureValidator(snowflake *SnowflakeTarget) *SnowflakeTableStructureValidator {
	return &SnowflakeTableStructureValidator{db: snowflake.connection, databaseName: snowflake.config.Database}
}

func (tsv *SnowflakeTableStructureValidator) Init() error {
	return nil
}

func (tsv *SnowflakeTableStructureValidator) Destroy() error {
	return nil
}

func (tsv *SnowflakeTableStructureValidator) Designation() string {
	return "TableStructureValidator"
}

func (tsv *SnowflakeTableStructureValidator) Validate(change *objects.ChangeItem) error {

	if change.ObjectType == "table" {
		if change.Item.Options.Validate && change.Item.Options.CheckExists {
			if change.ExistsFlag {

				var spec sfDefaultSpecification
				err := utility.UnmarshalYamlSubObject(change.Item.Spec, &spec)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical, false, errors.New("spec definition invalid,"+err.Error()), tsv.Designation())
				}

				//verify the change item has a change and init spec block, if not validation cannot be performed
				//if validation cant be performed due to this condition return false success but add warning to change do not set error state
				if utility.IsStringEmpty(&spec.Change) {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical, false, errors.New("spec definition, lacks change scope [change]"), tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				if utility.IsStringEmpty(&spec.Init) {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical, false, errors.New("spec definition, lacks init scope [init]"), tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				//prep values needed to render moc& validation object statements
				//moc clone table
				paramsMocTable := gonja.Context{
					"CHG_MGMT_DB": tsv.databaseName,
					"DATABASE":    strings.ToUpper(change.Item.Object.Database),
					"SCHEMA":      strings.ToUpper(change.Item.Object.Schema),
					"NAME":        strings.ToUpper(change.Item.Object.Name)}

				//moc create new from init for validation against variables
				paramsValidateTable := gonja.Context{
					"DATABASE":    tsv.databaseName,
					"CHG_MGMT_DB": tsv.databaseName,
					"SCHEMA":      "VALIDATE",
					"NAME":        strings.ToUpper(change.Item.Object.Name)}

				//validation logic variables
				paramValidationSql := gonja.Context{"NAME": strings.ToUpper(change.Item.Object.Name),
					"SCHEMA":      "ORIGIN",
					"DATABASE":    tsv.databaseName,
					"CHG_MGMT_DB": tsv.databaseName,
				}

				prepAndCleanup, err := common.RenderStatement(TableStructureCLeanUpSQL, &paramsMocTable)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("rendering failed [prep]:", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				prepAndCleanUpCmds := common.SegmentScopeCommands(prepAndCleanup)
				//run in event prior run created but never cleaned up after itself
				for _, cmd := range prepAndCleanUpCmds {
					_, err := tsv.db.Exec(cmd)
					if err != nil {
						change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
							false,
							utility.WrapError("failed to execute prep command:", err),
							tsv.Designation())
						return nil //  validator did not fail but has completed its task
					}
				}

				moc, err := common.RenderStatement(CreateMocTableSQL, &paramsMocTable)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("rendering failed [moc]", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				verifyStmt, err := common.RenderStatement(TableStructureVerifySQL, &paramValidationSql)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("rendering failed [verify statement]", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				initScope, err := renderSpecStatement(spec.Init, "init", &paramsValidateTable)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("rendering failed [init]", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				changeScope, err := renderSpecStatement(spec.Change, "change", &paramValidationSql)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("rendering failed [change]", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				//create origin and validate object structures
				_, err = tsv.db.Exec(moc)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("failed creating moc objects:", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				for _, cmdStmt := range initScope.Commands {
					_, err = tsv.db.Exec(cmdStmt)
					if err != nil {
						change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
							false,
							utility.WrapError("failed executing init command(s):", err),
							tsv.Designation())
						return nil //  validator did not fail but has completed its task
					}
				}

				for _, cmdStmt := range changeScope.Commands {
					_, err = tsv.db.Exec(cmdStmt)
					if err != nil {
						change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
							false,
							utility.WrapError("failed executing change command(s):", err),
							tsv.Designation())
						return nil //  validator did not fail but has completed its task
					}
				}

				//execute verification sql
				rows, err := tsv.db.Query(verifyStmt)
				if err != nil {
					change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
						false,
						utility.WrapError("failed executing verification commands:", err),
						tsv.Designation())
					return nil //  validator did not fail but has completed its task
				}

				passed := false // at least one column must be validated to pass

				for rows.Next() {
					var cv columnVerify
					err = rows.Scan(&cv.Origin.Name,
						&cv.Origin.Position,
						&cv.Origin.Nullable,
						&cv.Origin.DataType,
						&cv.Origin.MaxCharLen,
						&cv.Origin.NumericPrecision,
						&cv.Origin.NumericScale,
						&cv.Validate.Name,
						&cv.Validate.Position,
						&cv.Validate.Nullable,
						&cv.Validate.DataType,
						&cv.Validate.MaxCharLen,
						&cv.Validate.NumericPrecision,
						&cv.Validate.NumericScale)

					if !cv.Origin.Name.Valid ||
						!cv.Validate.Name.Valid ||
						strings.Compare(strings.ToUpper(cv.Origin.Name.String), strings.ToUpper(cv.Validate.Name.String)) != 0 {

						change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical, false, errors.New("column failed validation"), tsv.Designation())

						return nil //  validator did not fail but has completed its task
					}

					if !cv.Origin.DataType.Valid || !cv.Validate.DataType.Valid ||
						strings.Compare(strings.ToUpper(cv.Origin.DataType.String), strings.ToUpper(cv.Validate.DataType.String)) != 0 {
						change.Validation.AddValidationStepInfo(objects.ValidationErrorCritical,
							false,
							errors.New("column failed validation"),
							tsv.Designation())

						return nil //  validator did not fail but has completed its task
					}
					passed = true
				}

				if passed {
					// if we have gotten to here all is good the object passes structural validation
					change.Validation.AddValidationStepInfo(objects.ValidationErrorNone,
						true,
						nil,
						tsv.Designation())

				}

				// run cleanup commands, if error dont fail validator, next pass will cleanup in prep stage
				for _, cmd := range prepAndCleanUpCmds {
					_, _ = tsv.db.Exec(cmd)
				}

				return nil //  validator did not fail but has completed its task
			} else {
				change.Validation.AddValidationStepInfo(objects.ValidationErrorWarn, false, errors.New("structural validation skipped object does not exist"), tsv.Designation())
				return nil
			}
		} else {
			err := errors.New("object settings for validation and checking are not enabled")
			change.Validation.AddValidationStepInfo(objects.ValidationErrorInfo,
				false,
				err,
				tsv.Designation())

			return err
		}
	}

	return nil //no errors occurred
}
