package objects

import (
	"Plow/plow/utility"
	"errors"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	ValidationErrorNone ValidationErrorSeverity = iota
	ValidationErrorInfo
	ValidationErrorWarn
	ValidationErrorCritical
)

const (
	UndeterminedChangeAction ChangeAction = iota
	UpdateChangeAction
	AddChangeAction
)

var (
	ErrUndefinedObjectTypeTranslator = errors.New("undefined object type translator")
	ErrDoesNotExist                  = errors.New("key or value does nto exist")
)

type ChangeAction int
type ValidationErrorSeverity int

type ApplyScopeEffect struct {
	Executed bool
	Success  bool
	Partial  bool
	Error    error
}

type ApplyScope struct {
	Name     string
	Commands []string
	effect   ApplyScopeEffect
}

type ApplyEffectInformation struct {
	Executed  bool
	Completed bool
	scopes    []*ApplyScope
	Error     error
}

func (a *ApplyEffectInformation) IsSuccess() (bool, bool, error) {
	if a.Executed && a.Completed && a.Error == nil {
		return a.Completed, false, a.Error // completed, not a partial and nil error
	} else {
		// we may have a partial failure scan scope info and determine
		for _, scope := range a.scopes {
			if scope.effect.Executed && !scope.effect.Success {
				return false, scope.effect.Partial, scope.effect.Error // find first executed scope with failed state, return partial and error
			}
		}
		return false, false, errors.New("undefined error state") // we shouldn't hit this point but just in case
	}
}

func (a *ApplyEffectInformation) AddScope(scope *ApplyScope) {
	if a.scopes == nil {
		a.scopes = make([]*ApplyScope, 0)
	}

	a.scopes = append(a.scopes, scope)
}

func (a *ApplyEffectInformation) GetScopes() []*ApplyScope {
	return a.scopes
}

func (as *ApplyScope) GetEffectInfo() ApplyScopeEffect {
	return as.effect
}

func (as *ApplyScope) SetEffectInfo(executed bool, success bool, partial bool, err error) {
	as.effect = ApplyScopeEffect{Executed: executed, Success: success, Partial: partial, Error: err}
}

type ValidationStepInfo struct {
	ValidatorName string
	Success       bool
	Error         error
	Severity      ValidationErrorSeverity
}

type ChangeMetadata struct {
	Action         ChangeAction `yaml:"action"`
	Name           string       `yaml:"path"`
	IdentifierHash string       `yaml:"idHash"`
	GitHash        string       `yaml:"gitHash"`
}

type ValidationInfo struct {
	Critical int
	Warning  int
	Success  int
	Steps    map[string]ValidationStepInfo
}

func (vi *ValidationInfo) PassedValidation() bool {
	if vi.Success > 0 && vi.Critical == 0 {
		return true
	}
	return false
}

func (vi *ValidationInfo) addValidationStepInfo(info ValidationStepInfo) {
	if vi.Steps == nil {
		vi.Steps = make(map[string]ValidationStepInfo)
	}
	vi.Steps[info.ValidatorName] = info

	switch info.Severity {
	case ValidationErrorCritical:
		vi.Critical += 1
		break
	case ValidationErrorWarn:
		vi.Warning += 1
		break
	case ValidationErrorNone:
		vi.Success += 1
		break
	}
}

func (vi *ValidationInfo) AddValidationStepInfo(severity ValidationErrorSeverity, success bool, err error, name string) {
	vi.addValidationStepInfo(ValidationStepInfo{Severity: severity,
		Success:       success,
		Error:         err,
		ValidatorName: name})
}

type ChangeItem struct {
	ObjectType       string                 `yaml:"type"`
	Item             *CodeBlockSpec         `yaml:"code"`
	Metadata         ChangeMetadata         `yaml:"meta"`
	ExistsFlag       bool                   `yaml:"-"`
	Validation       ValidationInfo         `yaml:"-"`
	ApplyInformation ApplyEffectInformation `yaml:"-"`
	Bundle           *ChangeLogBundle       `yaml:"-"`
}

type ChangeLogBundle struct {
	typeIndex map[int64][]int `yaml:"-"`
	Ref       ChangeReference `yaml:"ref"`
	Items     []*ChangeItem   `yaml:"items"`
	parent    *ChangeLog      `yaml:"-"`
	Validated bool            `yaml:"-"`
}

type ChangeReference struct {
	Hash    string `yaml:"hash"`
	Message string `yaml:"msg"`
}

func isPossibleCodeFile(meta *ChangeMetadata) bool {
	if name := strings.ToUpper(strings.TrimSpace(meta.Name)); len(name) > 0 {
		if strings.Compare(name[len(name)-4:], "YAML") == 0 {
			return true
		}
	}
	return false
}

func (clb *ChangeLogBundle) GetChangesOfType(objtype int64) ([]*ChangeItem, error) {
	if clb.parent == nil || clb.parent.translator == nil {
		return nil, ErrUndefinedObjectTypeTranslator
	}

	output := make([]*ChangeItem, 0)
	if idx, ok := clb.typeIndex[objtype]; ok {
		for _, i := range idx {
			if i < len(clb.Items) {
				item := clb.Items[i]
				if clb.parent.translator(item.Item.Type) == objtype {
					output = append(output, item)
				}
			}
		}
	}
	return output, nil
}

func (clb *ChangeLogBundle) AddItem(bytes []byte, meta ChangeMetadata) error {
	if clb.parent == nil || clb.parent.translator == nil {
		return ErrUndefinedObjectTypeTranslator
	}
	//check if item is actually a code item. if not, do not add to bundle but do not error, ignore it
	//first determine if a .yaml file initially
	if !isPossibleCodeFile(&meta) {
		return nil
	}

	spec, err := clb.bytesToSpec(bytes)
	if err != nil {
		return err
	}

	//secondary code file validity test, make sure the objet type is set, if not ignore it do not add to bundle
	//do not need to validate type designator is valid this will be handled later, just that something is there
	if len(strings.TrimSpace(spec.Type)) == 0 {
		return nil
	}

	item := &ChangeItem{Metadata: meta, Item: spec, ObjectType: spec.Type, Bundle: clb}

	//translate obj type string to int64 representation
	objType := clb.parent.translator(item.ObjectType)

	if _, ok := clb.typeIndex[objType]; !ok {
		clb.typeIndex[objType] = make([]int, 0)
	}

	idx := len(clb.Items)
	clb.Items = append(clb.Items, item)
	clb.typeIndex[objType] = append(clb.typeIndex[objType], idx)

	return nil
}

func (clb *ChangeLogBundle) bytesToSpec(bytes []byte) (*CodeBlockSpec, error) {
	var spec CodeBlockSpec
	err := yaml.Unmarshal(bytes, &spec)
	if err != nil {
		return nil, err
	}
	return &spec, err
}

type ChangeLog struct {
	Bundles    []*ChangeLogBundle `yaml:"bundles"`
	translator ObjectTypeTranslator
}

func (cl *ChangeLog) AddBundle(commit *object.Commit) *ChangeLogBundle {
	if cl.Bundles == nil {
		cl.Bundles = make([]*ChangeLogBundle, 0)
	}

	bundle := &ChangeLogBundle{
		Items:     make([]*ChangeItem, 0),
		Ref:       ChangeReference{Hash: commit.Hash.String(), Message: commit.Message},
		typeIndex: make(map[int64][]int),
		parent:    cl,
	}

	cl.Bundles = append(cl.Bundles, bundle)
	return bundle
}

func (cl *ChangeLog) AddManualBundle() *ChangeLogBundle {
	if cl.Bundles == nil {
		cl.Bundles = make([]*ChangeLogBundle, 0)
	}

	bundle := &ChangeLogBundle{
		Items:     make([]*ChangeItem, 0),
		Ref:       ChangeReference{Hash: utility.Sha256Hash("manual"), Message: "manual"},
		typeIndex: make(map[int64][]int),
		parent:    cl,
	}

	cl.Bundles = append(cl.Bundles, bundle)
	return bundle
}

func NewChangeMetaFromGitChange(change *object.Change) ChangeMetadata {
	action := UndeterminedChangeAction

	a, err := change.Action()
	if err == nil {
		switch a {
		case merkletrie.Modify:
			{
				action = UpdateChangeAction
				break
			}
		case merkletrie.Insert:
			{
				action = AddChangeAction
				break
			}
		default:
			action = UndeterminedChangeAction
		}
	}

	return ChangeMetadata{Action: action,
		Name:           change.To.Name,
		GitHash:        change.To.TreeEntry.Hash.String(),
		IdentifierHash: utility.Sha256Hash(change.To.Name)}
}

func NewChangeMetaFromGitFileTreeItem(file *object.File) ChangeMetadata {
	return ChangeMetadata{Action: UndeterminedChangeAction,
		Name:           file.Name,
		GitHash:        file.Hash.String(),
		IdentifierHash: utility.Sha256Hash(file.Name)}
}

func NewChangeMetaFromOptions(options *Options) ChangeMetadata {
	meta := ChangeMetadata{}
	meta.Name = options.File.Name
	meta.IdentifierHash = utility.Sha256HashBytes(options.File.Bytes)
	meta.Action = UndeterminedChangeAction
	return meta
}

func NewChangeLog(typeTranslator ObjectTypeTranslator) *ChangeLog {
	return &ChangeLog{Bundles: make([]*ChangeLogBundle, 0), translator: typeTranslator}
}
