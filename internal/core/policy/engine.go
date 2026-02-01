package policy

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Engine struct {
	Policy *Policy
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	p := &Policy{}
	if err := yaml.Unmarshal(data, p); err != nil {
		return err
	}
	e.Policy = p
	return nil
}

func (e *Engine) Check(cmd []string) (*Result, error) {
	if e.Policy == nil {
		return &Result{Allowed: true}, nil
	}
	full := strings.Join(cmd, " ")
	events := []Event{}
	allowed := true
	for _, rule := range e.Policy.Rules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, err
		}
		if re.MatchString(full) {
			action := strings.ToLower(rule.Action)
			events = append(events, Event{
				Rule:    rule.Name,
				Action:  action,
				Matched: rule.Pattern,
				Reason:  rule.Reason,
			})
			if action == "block" {
				allowed = false
			}
		}
	}
	return &Result{Allowed: allowed, Events: events}, nil
}
