package git

import (
	"github.com/go-git/go-git/v5"
)

// GetCommit return the current commit
func GetCommit() (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}
	ref, err := r.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}
