package objects

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strings"
)

type Flags uint64

const (
	SkipValidationSetting = 1 << iota
	SkipApplySetting
	TerminateOnValidationFailureSetting
	FastForwardSetting
	FullManifestSetting
	SingleFileChangeSetting
	RenderChangesSetting
	UseLocalRepositorySetting
)

func (f *Flags) Set(flag Flags)      { *f |= flag }
func (f *Flags) Clear(flag Flags)    { *f &= ^flag }
func (f *Flags) Toggle(flag Flags)   { *f ^= flag }
func (f *Flags) Has(flag Flags) bool { return *f&flag != 0 }

type FileInfo struct {
	Name  string
	Bytes []byte
}

type Options struct {
	OptionFlags    Flags
	BranchOverride *string
	CommitId       *string
	File           *FileInfo
}

func (o *Options) EvaluateTargetCommit(commits []*object.Commit) (*object.Commit, error) {
	//using repo provided commits and input options determine which commit to use as target
	//assert we have a list of commits provided by caller
	if commits == nil || len(commits) == 0 {
		return nil, errors.New("commit list not provided, or empty")
	}

	if o.CommitId != nil {
		//verify that the commit provided is in the list provided, if not error
		found := false
		idx := 0
		for i, c := range commits {
			if strings.Compare(c.String(), *o.CommitId) == 0 {
				found = true
				idx = i
				break
			}
		}
		if found {
			return commits[idx], nil
		} else {
			return nil, errors.New(fmt.Sprintf("provided commit [%s] not found", *o.CommitId))
		}
	} else {
		//return the top most commit from the list
		return commits[0], nil
	}
}

func (o *Options) IsFileProvided() bool {
	return o.File != nil
}
