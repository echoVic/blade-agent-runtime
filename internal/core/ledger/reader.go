package ledger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

type Reader struct {
	tasksDir string
}

func NewReader(tasksDir string) *Reader {
	return &Reader{tasksDir: tasksDir}
}

func (r *Reader) ReadAll(taskID string) ([]Step, error) {
	ledgerPath := filepath.Join(r.tasksDir, taskID, "ledger.jsonl")
	file, err := os.Open(ledgerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Step{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []Step
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Step
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

func (r *Reader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
