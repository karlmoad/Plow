package common

import (
	"Plow/plow/objects"
	"github.com/noirbizarre/gonja"
	"strings"
	"time"
)

type RenderedChange struct {
	item        *objects.ChangeItem
	TimeApplied time.Time
}

func NewRenderedChange(item *objects.ChangeItem, rendered []*objects.ApplyScope) *RenderedChange {
	for _, scope := range rendered {
		item.ApplyInformation.AddScope(scope)
	}
	return &RenderedChange{item: item}
}

func (rc *RenderedChange) Item() *objects.ChangeItem {
	return rc.item
}

func NewScopeFromMultilineStatement(name string, statement *string) *objects.ApplyScope {
	return &objects.ApplyScope{Name: name, Commands: SegmentScopeCommands(*statement)}
}

func NewScope(name string, commands []string) *objects.ApplyScope {
	return &objects.ApplyScope{Name: name, Commands: commands}
}

func RenderStatement(statement string, context *gonja.Context) (string, error) {
	template, err := gonja.FromString(statement)
	if err != nil {
		return "", err
	}
	return template.Execute(*context)
}

type Renderer interface {
	Render(change *objects.ChangeItem) ([]*objects.ApplyScope, error)
	RenderWithContext(change *objects.ChangeItem, params *map[string]interface{}) ([]*objects.ApplyScope, error)
}

func NewRenderContextFromObjectInfo(obj objects.ObjectSpec) *map[string]interface{} {
	return &map[string]interface{}{"NAME": strings.TrimSpace(strings.ToUpper(obj.Name)),
		"DATABASE": strings.TrimSpace(strings.ToUpper(obj.Database)),
		"SCHEMA":   strings.TrimSpace(strings.ToUpper(obj.Schema))}
}
