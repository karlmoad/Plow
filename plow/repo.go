package plow

import (
	"Plow/plow/objects"
	"Plow/plow/secrets"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"io"
	"os"
	"strings"
)

var (
	ErrNoCommitsToProcess  = errors.New("no commits to process")
	ErrNoTargetCommitFound = errors.New("unable to acquire target commit reference")
	ErrNoLastCommitFound   = errors.New("fast forwarding is not set, and unable to acquire last processed commit reference")
)

type Repo struct {
	sshKey     *ssh.PublicKeys
	config     *Configuration
	repo       *git.Repository
	primaryRef *plumbing.Reference
	options    *objects.Options
}

func (r *Repo) memoryClone() (*git.Repository, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  r.config.GitConfig.Url,
		Auth: r.sshKey,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repo) GetPrimaryCommitReference() (*plumbing.Reference, error) {
	if !r.options.OptionFlags.Has(objects.UseLocalRepositorySetting) {
		return r.repo.Head()
	} else {
		return r.primaryRef, nil
	}
}

func (r *Repo) Clone(directory string) (*git.Repository, error) {
	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:  r.config.GitConfig.Url,
		Auth: r.sshKey,
	})

	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repo) open(directory string) (*git.Repository, error) {
	return git.PlainOpen(directory)
}

func (r *Repo) Branches() ([]*plumbing.Reference, error) {
	out := make([]*plumbing.Reference, 0)
	iter, err := r.repo.References()
	if err != nil {
		return nil, err
	}

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			out = append(out, ref)
		}
		return nil
	})

	return out, nil
}

func (r *Repo) SetBranchReference(name string) error {
	branches, err := r.Branches()
	if err != nil {
		return err
	}
	for _, v := range branches {
		if strings.Contains(v.String(), name) {
			r.primaryRef = v
			return nil
		}
	}
	return fmt.Errorf("reference not founc")
}

func (r *Repo) GetCommitHistory(reference *plumbing.Reference) ([]*object.Commit, error) {
	out := make([]*object.Commit, 0)
	if reference == nil {
		reference = r.primaryRef
	}

	err := r.repo.Storer.SetReference(reference)
	if err != nil {
		return nil, err
	}

	iter, err := r.repo.Log(&git.LogOptions{From: reference.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	err = iter.ForEach(func(c *object.Commit) error {
		out = append(out, c)
		return nil
	})

	return out, nil
}

func newLocalRepo(config *Configuration, options *objects.Options, secrets secrets.SecretStore) (*Repo, error) {
	r := &Repo{config: config, options: options}
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repo, err := r.open(dir)
	if err != nil {
		return nil, err
	}
	r.repo = repo
	return r, nil
}

func newMemoryRepo(config *Configuration, options *objects.Options, secrets secrets.SecretStore) (*Repo, error) {
	//establishes a new fresh memory based clone of github repo

	r := &Repo{config: config, options: options}

	sshpwd, err := secrets.GetSecret(config.GitConfig.KeyPasswordSecret)
	if err != nil {
		return nil, err
	}

	sshkey, _ := os.ReadFile(r.config.GitConfig.SSHKeyFile)
	keys, err := ssh.NewPublicKeys("git", []byte(sshkey), sshpwd) //pwd form secrets
	if err != nil {
		return nil, err
	}
	r.sshKey = keys

	repo, err := r.memoryClone()

	if err != nil {
		return nil, err
	}
	r.repo = repo

	err = r.SetBranchReference(config.GitConfig.Branch)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repo) GetDiffChanges(from *object.Commit, to *object.Commit) ([]*object.Change, error) {

	treeFrom, err := from.Tree()
	if err != nil {
		return nil, err
	}

	treeTo, err := to.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := treeFrom.Diff(treeTo)
	if err != nil {
		return nil, err
	}

	return changes, nil
}

func (r *Repo) ReadBlob(file *object.File) ([]byte, error) {
	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	buffer := make([]byte, file.Size)
	_, err = reader.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (r *Repo) BuildChangeLog(log *objects.TrackingLog, translator objects.ObjectTypeTranslator) (*objects.ChangeLog, error) {
	rez := objects.NewChangeLog(translator)

	var last *object.Commit

	//if this is a fast forward or single file execution ignore tracking history
	if !r.options.OptionFlags.Has(objects.SingleFileChangeSetting) &&
		!r.options.OptionFlags.Has(objects.FastForwardSetting) {

		lastTracked := log.GetLastProcessed()
		if lastTracked == nil {
			return nil, errors.New("unable to get latest commit from objects log")
		}

		ltCommit, err := r.getCommit(lastTracked.TrackingId)
		if err != nil {
			return nil, err
		}

		last = ltCommit
	}

	commits, err := r.buildWorkList(last)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, ErrNoCommitsToProcess
	}

	if !r.options.OptionFlags.Has(objects.FastForwardSetting) {
		if last == nil {
			//just in case this has not been evaluated prior
			return nil, ErrNoLastCommitFound
		}
		//step through commit list and load changes,
		//changes will consist of only modified objects between two commits, not the full digest of the repo
		//reference to hold previous commit to enable diff changelist extraction
		prev := last

		//reverse through the array, changes need applied in order they arrive in repo
		for i := len(commits) - 1; i >= 0; i-- {

			bundle := rez.AddBundle(commits[i])

			changes, err := r.GetDiffChanges(prev, commits[i])
			if err != nil {
				return nil, err
			}

			commit := commits[i]
			for _, change := range changes {
				action, err := change.Action()
				if err != nil {
					return nil, errors.New("unable to determine git change action")
				}
				if action != merkletrie.Delete {
					tree, err := commit.Tree()
					if err != nil {
						return nil, err
					}

					file, err := tree.File(change.To.Name)
					if err != nil {
						return nil, err
					}

					bytes, err := r.ReadBlob(file)
					if err != nil {
						return nil, err
					}

					err = bundle.AddItem(bytes, objects.NewChangeMetaFromGitChange(change))
					if err != nil {
						return nil, err
					}
				}
			}
			prev = commit
		}
	} else {
		//fast forwarding is active worklist will be
		// based on the full digest of the repo tree at that commit point
		// we only care about the first commit, and there will be only one change bundle

		bundle := rez.AddBundle(commits[0])
		tree, err := commits[0].Tree()
		if err != nil {
			return nil, err
		}

		fIter := tree.Files()
		file, err := fIter.Next()
		for err != io.EOF {
			bytes, err_inr := r.ReadBlob(file)
			if err != nil {
				return nil, err_inr
			}
			err_inr = bundle.AddItem(bytes, objects.NewChangeMetaFromGitFileTreeItem(file))
			if err_inr != nil {
				return nil, err_inr
			}
			file, err = fIter.Next()
		}
	}
	return rez, nil
}

func (r *Repo) buildWorkList(last *object.Commit) ([]*object.Commit, error) {
	worklist := make([]*object.Commit, 0)

	ref, err := r.GetPrimaryCommitReference()
	if err != nil {
		return nil, err
	}

	commits, err := r.GetCommitHistory(ref)

	//var last *object.Commit
	//
	////assert tracking last processed entry is actually a commit in the tree
	//if lastProcessed != nil {
	//	last, err = r.repo.CommitObject(plumbing.NewHash(lastProcessed.TrackingId))
	//	if err != nil {
	//		return nil, errors.New("tracking identifier not found in commit history")
	//	}
	//}

	if err != nil {
		return nil, err
	}

	targetCommit, err := r.options.EvaluateTargetCommit(commits)
	if err != nil {
		return nil, err
	}

	if targetCommit == nil {
		return nil, ErrNoTargetCommitFound
	}

	//IF NOT FAST FORWARD
	if !r.options.OptionFlags.Has(objects.FastForwardSetting) {
		if last == nil {
			return nil, ErrNoLastCommitFound
		}

		//check if the last commit is the target commit, if so there is nothing to do
		if strings.Compare(last.Hash.String(), targetCommit.Hash.String()) == 0 {
			return worklist, nil
		}

		//loop commit list until target is found
		//once target is found, add it and all following to worklist until last processed commit is found
		found := false
		for _, c := range commits {
			if !found {
				if strings.Compare(c.Hash.String(), targetCommit.Hash.String()) == 0 {
					worklist = append(worklist, c)
					found = true
				}
				continue
			} else { //target was found and added, keep adding commits until last processed is encountered, break when this occurs
				if strings.Compare(c.Hash.String(), last.Hash.String()) == 0 {
					break
				} else {
					worklist = append(worklist, c)
				}
			}
		}
	} else {
		// IS FAST FORWARD JUST PUT THE TARGET COMMIT IN THE WORKLIST
		worklist = append(worklist, targetCommit)
	}

	return worklist, nil
}

func (r *Repo) getCommit(hash string) (*object.Commit, error) {
	return r.repo.CommitObject(plumbing.NewHash(hash))
}
