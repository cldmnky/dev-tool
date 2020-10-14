package git

import (
	"github.com/go-git/go-git/v5"
)

// GetCommit return the current commit
func GetCommit(path string) (string, error) {
	opts := git.PlainOpenOptions{
		DetectDotGit: true,
	}
	r, err := git.PlainOpenWithOptions(path, &opts)
	if err != nil {
		return "", err
	}
	ref, err := r.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String()[0:8], nil
}
