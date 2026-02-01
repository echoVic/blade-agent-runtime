package git

import (
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

func CurrentHEAD(repoRoot string) (string, string, error) {
	repo, err := gogit.PlainOpen(repoRoot)
	if err != nil {
		return "", "", err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", "", err
	}
	branch := ""
	if ref.Name().IsBranch() {
		branch = strings.TrimPrefix(ref.Name().String(), "refs/heads/")
	}
	return ref.Hash().String(), branch, nil
}
