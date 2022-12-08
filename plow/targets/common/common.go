package common

import (
	"Plow/plow/objects"
	"context"
	"errors"
	"strings"
)

func isStringEmpty(val *string) bool {
	if len(strings.TrimSpace(*val)) > 0 {
		return false
	}
	return true
}

var (
	ErrInvalidTargetType        = errors.New("invalid target type")
	ErrNotImplemented           = errors.New("not implemented")
	ErrNoChangesProvided        = errors.New("change log is empty")
	ErrInternalError            = errors.New("something very bad happened")
	ErrInvalidTrackingStructure = errors.New("invalid or missing objects structure found on target")
	ErrNoChangeHistory          = errors.New("no change history found on target")
)

type Command int
type TargetType int

const (
	CommandStart = iota
	CommandHalt
	CommandExit
	CommandSkip
)

type Target interface {
	Open(context *objects.PlowContext) error
	PersistTrackingLogDetail(detail *objects.LogItemEntry) error
	PersistTrackingLogEntry(entry *objects.LogEntry) error
	GetTrackingHistory(depth int) (*objects.TrackingLog, error)
	GetTrackingLogDetail(entry objects.LogEntry) ([]objects.LogItemEntry, error)
	ValidateChangeLog(changes *objects.ChangeLog) error
	RenderChangeLog(changes *objects.ChangeLog) ([]*RenderedChange, error)
	ApplyChangeLog(context context.Context, changes *objects.ChangeLog) error
	Close() error
	GetObjectTypeTranslator() objects.ObjectTypeTranslator
	GetObjectTypeExecutionOrder() []int64
}
