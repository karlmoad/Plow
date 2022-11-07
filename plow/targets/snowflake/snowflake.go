package snowflake

import (
	"Plow/plow/objects"
	"Plow/plow/secrets"
	"Plow/plow/targets/common"
	"context"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/noirbizarre/gonja"
	sf "github.com/snowflakedb/gosnowflake"
	"github.com/youmark/pkcs8"
	"os"
	"strconv"
	"time"
)

type SnowflakeTarget struct {
	connection  *sql.DB
	config      sf.Config
	secretStore secrets.SecretStore
	validation  *common.ValidationHandler
	options     *objects.Options
	renderer    *SnowflakeRenderer
}

func (s *SnowflakeTarget) Open(config SnowflakeConfiguration, options *objects.Options, secretStore secrets.SecretStore) error {
	if IsProtectedSystemRole(config.Role) {
		return ErrDisallowedPrivilegedRole
	}

	s.renderer = newSnowflakeRenderer(config.Role, config.Warehouse)
	s.options = options
	pkeyFile := config.PublicKeyFile
	s.secretStore = secretStore
	bytes, err := os.ReadFile(pkeyFile)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(bytes)

	pwd, err := secretStore.GetSecret(config.KeyPasswordSecret)
	if err != nil {
		return err
	}

	pkey, err := pkcs8.ParsePKCS8PrivateKeyRSA(block.Bytes, []byte(pwd))
	if err != nil {
		return err
	}

	s.config = sf.Config{
		Authenticator: sf.AuthTypeJwt,
		PrivateKey:    pkey,
		User:          config.UserId,
		Account:       config.Account,
		Region:        config.Region,
		Database:      config.Database,
		Warehouse:     config.Warehouse,
		Role:          config.Role}

	dsn, err := sf.DSN(&s.config)
	if err != nil {
		return err
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return err
	}

	s.connection = db
	return nil
}
func (s *SnowflakeTarget) GetTrackingHistory(depth int) (*objects.TrackingLog, error) {
	stmt, err := common.RenderStatement(TrackingHistorySQL, &gonja.Context{"DATABASE": s.config.Database})
	if err != nil {
		return nil, err
	}
	rez := objects.NewTrackingLog()

	if depth > 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, depth)
	}

	rows, err := s.connection.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	count := 0

	for rows.Next() {
		var entry objects.LogEntry
		err := rows.Scan(&entry.TrackingId,
			&entry.Message,
			&entry.Start,
			&entry.End,
			&entry.AppliedBy,
			&entry.TotalChanges,
			&entry.SuccessfulChanges,
			&entry.FailedChanges,
			&entry.Completed,
			&entry.FastForward)

		if err != nil {
			return nil, err
		}
		count += 1
		rez.Add(entry)
	}

	if count == 0 {
		rez.Empty = true
	}

	return rez, nil
}

func (s *SnowflakeTarget) GetTrackingLogDetail(entry objects.LogEntry) ([]objects.LogItemEntry, error) {
	return nil, common.ErrNotImplemented
}

func (s *SnowflakeTarget) RenderChangeLog(changes *objects.ChangeLog) ([]*common.RenderedChange, error) {
	if changes == nil {
		return nil, common.ErrNoChangesProvided
	}
	renderedChanges := make([]*common.RenderedChange, 0)

	validationDisabled := s.options.OptionFlags.Has(objects.SkipValidationSetting)
	//for each bundle in order
	for _, bundle := range changes.Bundles {

		for _, objType := range SnowflakeProcessingOrder {

			items, err := bundle.GetChangesOfType(int64(objType))
			if err != nil {
				return nil, err //can't get a list of objects in the correct order type, time to error and halt
			}
			if len(items) > 0 {
				for _, item := range items {
					//check the bundle header see if validation was run,
					//if any changes are configured for validation, we need to stop and not go on for this item
					if (!bundle.Validated || validationDisabled) && item.Item.Options.Validate {
						item.ApplyInformation.Executed = false
						item.ApplyInformation.Completed = false
						item.ApplyInformation.Error = errors.New("validation not performed and object has validation enabled")
						continue
					}
					//also check if this item failed validation then should be skipped
					if !validationDisabled && !item.Validation.PassedValidation() {
						item.ApplyInformation.Executed = false
						item.ApplyInformation.Completed = false
						item.ApplyInformation.Error = errors.New("validation not performed and object has validation enabled")
						continue
					}

					if rezult := s.renderChange(item); rezult != nil {
						renderedChanges = append(renderedChanges, rezult)
					}
				}
			}
		}
	}

	return renderedChanges, nil
}

func (s *SnowflakeTarget) ApplyChangeLog(context context.Context, changes *objects.ChangeLog) error {
	if changes == nil {
		return common.ErrNoChangesProvided
	}

	//set initial role as default
	if err := s.ResetActiveRole(); err != nil {
		return err
	}

	rendered, err := s.RenderChangeLog(changes)
	if err != nil {
		return err
	}

	warehouseCoordinator := s.renderer.GetWarehouseCoordinator()
	if err := warehouseCoordinator.Activate(s.connection); err != nil {
		return err
	}
	defer warehouseCoordinator.DeActivate()

	//apply rendered changes in order, if error occurs in application halt
	timeStart := time.Now()
	for _, renderedChg := range rendered {
		if err := s.applyChangeToTarget(renderedChg); err != nil {
			return err
		}
	}

	//set initial role as default
	if err := s.ResetActiveRole(); err != nil {
		return err
	}

	//for each bundle gather statistics and log
	for _, bundle := range changes.Bundles {
		//for each object type in processing order, get list of changelog items of type
		total := 0
		success := 0
		failed := 0

		//for _, objType := range SnowflakeProcessingOrder {
		//
		//	items, err := bundle.GetChangesOfType(int64(objType)) // log in order of type order to match rendering order
		//	if err != nil {
		//		return err //can't get a list of objects in the correct order type, time to error and halt
		//	}
		//collect and track item level apply metrics
		for _, item := range bundle.Items {
			total += 1
			successful, partial, err := item.ApplyInformation.IsSuccess()

			if successful {
				success += 1
			} else {
				failed += 1
			}

			var msg string
			if err != nil {
				msg = err.Error()
			}

			logEntry := &objects.LogItemEntry{TrackingId: bundle.Ref.Hash,
				FileName:  item.Metadata.Name,
				Reference: item.Metadata.GitHash,
				Hash:      item.Metadata.IdentifierHash,
				Status:    successful,
				ApplyDate: time.Now(),
				Message:   msg,
				Partial:   partial}
			//dont error main process for logging issue of an individual item
			_ = s.PersistTrackingLogDetail(logEntry)
		}
		//}
		timeEnd := time.Now()

		//gather bundle metrics
		log := &objects.LogEntry{TrackingId: bundle.Ref.Hash, Message: bundle.Ref.Message,
			Start:             timeStart,
			End:               timeEnd,
			AppliedBy:         s.config.User,
			TotalChanges:      total,
			SuccessfulChanges: success,
			FailedChanges:     failed,
			Completed:         true,
			FastForward:       s.options.OptionFlags.Has(objects.FastForwardSetting),
		}

		// log bundle to tracking
		err := s.PersistTrackingLogEntry(log)
		if err != nil {
			// something occurred saving log to DB, because this is critical to the logic of this solution that these entries exist
			// error and halt
			return errors.New(fmt.Sprintf("failed to save to comit log :%s", err.Error()))
		}
	}
	return nil
}

func (s *SnowflakeTarget) renderChange(item *objects.ChangeItem) *common.RenderedChange {

	scopes := []*objects.ApplyScope{common.NewScope("header", []string{generateUseRoleStmt(s.config.Role)})}
	//render scopes for the item, if error: add error info to item and return
	rScopes, err := s.renderer.Render(item)
	if err != nil {
		item.ApplyInformation.Executed = false
		item.ApplyInformation.Completed = false
		item.ApplyInformation.Error = err
		return nil
	} else {
		//append rendered scope to final result
		scopes = append(scopes, rScopes...)
	}

	return common.NewRenderedChange(item, scopes)
}

func (s *SnowflakeTarget) applyChangeToTarget(renderedChange *common.RenderedChange) error {
	// apply scopes
	appliedScopes := 0
	item := renderedChange.Item()
	item.ApplyInformation.Executed = true
	renderedChange.TimeApplied = time.Now()

	//set role to default role
	if err := s.ResetActiveRole(); err != nil {
		item.ApplyInformation.Error = err
		return err
	}

	for _, scope := range item.ApplyInformation.GetScopes() {
		appliedCmds := 0
		for _, cmd := range scope.Commands {
			_, err := s.connection.Exec(cmd)
			if err != nil {
				scope.SetEffectInfo(true, false, appliedCmds > 0 || appliedScopes > 0, err)
				item.ApplyInformation.Error = err
				return err
			}
			appliedCmds += 1
		}
		scope.SetEffectInfo(true, true, false, nil)
		appliedScopes += 1
	}

	item.ApplyInformation.Completed = len(item.ApplyInformation.GetScopes()) == appliedScopes
	return nil
}

func (s *SnowflakeTarget) ValidateChangeLog(changes *objects.ChangeLog) error {
	if changes != nil {
		//initialize the validation handler
		s.validation = common.NewValidationHandler(StringToSnowflakeObjectTypeInt64)
		s.validation.RegisterGlobalValidator(newSnowflakeObjectExistsValidator(s, changes))
		s.validation.RegisterTypeValidator(int64(Table), newSnowflakeTableStructureValidator(s))
		if err := s.validation.Initialize(); err != nil {
			return err
		}

		for _, bundle := range changes.Bundles {
			if err := s.validateBundle(bundle); err != nil {
				return err
			}
			bundle.Validated = true
		}
	}

	return nil
}

func (s *SnowflakeTarget) Close() error {
	return s.connection.Close()
}

func (s *SnowflakeTarget) validateBundle(bundle *objects.ChangeLogBundle) error {
	if s.validation == nil {
		return errors.New("ASSERT Validation handler is null")
	}

	for _, item := range bundle.Items {
		if err := s.validation.Validate(item); err != nil {
			return err
		}
	}
	return nil
}

func (s *SnowflakeTarget) PersistTrackingLogDetail(detail *objects.LogItemEntry) error {
	gc := gonja.Context{"DATABASE": s.config.Database,
		"COMMIT": detail.TrackingId,
		"FILE":   detail.FileName,
		"REF":    detail.Reference,
		"HASH":   detail.Hash,
		"STATUS": strconv.FormatBool(detail.Status),
		"TIME":   detail.ApplyDate.UTC().Format("2006-01-02 15:04:05"),
		"MSG":    common.MakeStringDatabaseSafe(detail.Message)}

	stmt, err := common.RenderStatement(InsertTrackingDetailSQL, &gc)
	if err != nil {
		return err
	}

	_, err = s.connection.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}
func (s *SnowflakeTarget) PersistTrackingLogEntry(entry *objects.LogEntry) error {
	gc := gonja.Context{"DATABASE": s.config.Database,
		"COMMIT":       entry.TrackingId,
		"MSG":          common.MakeStringDatabaseSafe(entry.Message),
		"START":        entry.Start.UTC().Format("2006-01-02 15:04:05"),
		"END":          entry.End.UTC().Format("2006-01-02 15:04:05"),
		"WHO":          entry.AppliedBy,
		"TOTAL":        strconv.Itoa(entry.TotalChanges),
		"SUCCESS":      strconv.Itoa(entry.SuccessfulChanges),
		"FAIL":         strconv.Itoa(entry.FailedChanges),
		"COMPLETED":    strconv.FormatBool(entry.Completed),
		"FAST_FORWARD": strconv.FormatBool(entry.FastForward)}

	stmt, err := common.RenderStatement(InsertTrackingInfoSQL, &gc)
	if err != nil {
		return err
	}

	_, err = s.connection.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func (s *SnowflakeTarget) GetObjectTypeTranslator() objects.ObjectTypeTranslator {
	return StringToSnowflakeObjectTypeInt64
}

func (s *SnowflakeTarget) ResetActiveRole() error {
	if _, err := s.connection.Exec(generateUseRoleStmt(s.config.Role)); err != nil {
		return err
	}
	return nil
}

func (s *SnowflakeTarget) GetObjectTypeExecutionOrder() []int64 {
	rv := make([]int64, len(SnowflakeProcessingOrder))
	for i, v := range SnowflakeProcessingOrder {
		rv[i] = v.ToInt64()
	}
	return rv
}
