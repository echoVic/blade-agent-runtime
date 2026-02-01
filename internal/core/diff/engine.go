package diff

import (
	"strconv"
	"strings"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
)

type Engine struct {
	Git *gitadapter.Runner
}

type Result struct {
	Files     int
	Additions int
	Deletions int
	FileList  []string
	Patch     []byte
}

func NewEngine(git *gitadapter.Runner) *Engine {
	return &Engine{Git: git}
}

func (e *Engine) Generate(workspacePath string, baseRef string) (*Result, error) {
	patch, err := e.Git.Run(workspacePath, "diff", baseRef)
	if err != nil {
		return nil, err
	}
	stat, err := e.Git.Run(workspacePath, "diff", "--shortstat", baseRef)
	if err != nil {
		return nil, err
	}
	nameOnly, err := e.Git.Run(workspacePath, "diff", "--name-only", baseRef)
	if err != nil {
		return nil, err
	}
	files, adds, dels := parseShortStat(stat)
	fileList := parseFileList(nameOnly)
	return &Result{
		Files:     files,
		Additions: adds,
		Deletions: dels,
		FileList:  fileList,
		Patch:     []byte(patch),
	}, nil
}

func parseFileList(nameOnly string) []string {
	lines := strings.Split(strings.TrimSpace(nameOnly), "\n")
	result := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

func parseShortStat(stat string) (int, int, int) {
	if strings.TrimSpace(stat) == "" {
		return 0, 0, 0
	}
	fields := strings.Fields(stat)
	files := 0
	adds := 0
	dels := 0
	for i, f := range fields {
		if f == "files" || f == "file" {
			if i > 0 {
				if v, err := strconv.Atoi(fields[i-1]); err == nil {
					files = v
				}
			}
		}
		if strings.HasPrefix(f, "insertion") || strings.HasPrefix(f, "insertions") {
			if i > 0 {
				if v, err := strconv.Atoi(fields[i-1]); err == nil {
					adds = v
				}
			}
		}
		if strings.HasPrefix(f, "deletion") || strings.HasPrefix(f, "deletions") {
			if i > 0 {
				if v, err := strconv.Atoi(fields[i-1]); err == nil {
					dels = v
				}
			}
		}
	}
	return files, adds, dels
}
