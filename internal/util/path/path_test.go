package path

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobalBarDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, ".bar")
	actual := GlobalBarDir()
	if actual != expected {
		t.Errorf("GlobalBarDir() = %q, want %q", actual, expected)
	}
}

func TestProjectID(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot string
	}{
		{
			name:     "simple path",
			repoRoot: "/Users/test/my-project",
		},
		{
			name:     "nested path",
			repoRoot: "/Users/test/work/repos/my-project",
		},
		{
			name:     "path with special chars",
			repoRoot: "/Users/test/my-project-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProjectID(tt.repoRoot)
			baseName := filepath.Base(tt.repoRoot)
			if !strings.HasPrefix(result, baseName+"-") {
				t.Errorf("ProjectID(%q) = %q, should start with %q", tt.repoRoot, result, baseName+"-")
			}
			parts := strings.Split(result, "-")
			hashPart := parts[len(parts)-1]
			if len(hashPart) != 4 {
				t.Errorf("ProjectID(%q) hash part = %q, want 4 hex chars", tt.repoRoot, hashPart)
			}
		})
	}
}

func TestProjectID_Uniqueness(t *testing.T) {
	path1 := "/Users/test/work/my-project"
	path2 := "/Users/test/personal/my-project"

	id1 := ProjectID(path1)
	id2 := ProjectID(path2)

	if id1 == id2 {
		t.Errorf("ProjectID should be unique for different paths: %q and %q both got %q", path1, path2, id1)
	}
}

func TestProjectID_Consistency(t *testing.T) {
	repoRoot := "/Users/test/my-project"
	id1 := ProjectID(repoRoot)
	id2 := ProjectID(repoRoot)

	if id1 != id2 {
		t.Errorf("ProjectID should be consistent: got %q and %q for same path", id1, id2)
	}
}

func TestBarDir(t *testing.T) {
	repoRoot := "/Users/test/my-project"
	result := BarDir(repoRoot)

	globalDir := GlobalBarDir()
	if !strings.HasPrefix(result, globalDir) {
		t.Errorf("BarDir(%q) = %q, should start with %q", repoRoot, result, globalDir)
	}
	if !strings.Contains(result, "projects") {
		t.Errorf("BarDir(%q) = %q, should contain 'projects'", repoRoot, result)
	}
	projectID := ProjectID(repoRoot)
	if !strings.HasSuffix(result, projectID) {
		t.Errorf("BarDir(%q) = %q, should end with %q", repoRoot, result, projectID)
	}
}

func TestFindRepoRoot(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	result, err := FindRepoRoot(subDir)
	if err != nil {
		t.Fatalf("FindRepoRoot(%q) error: %v", subDir, err)
	}
	if result != tmpDir {
		t.Errorf("FindRepoRoot(%q) = %q, want %q", subDir, result, tmpDir)
	}
}

func TestFindRepoRoot_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindRepoRoot(tmpDir)
	if err == nil {
		t.Error("FindRepoRoot should return error for non-git directory")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "a", "b", "c")

	if err := EnsureDir(newDir); err != nil {
		t.Fatalf("EnsureDir(%q) error: %v", newDir, err)
	}
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("os.Stat(%q) error: %v", newDir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q should be a directory", newDir)
	}
}
