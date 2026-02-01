package policy

type Policy struct {
	Version int    `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

type Rule struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
	Action  string `yaml:"action"`
	Reason  string `yaml:"reason"`
}

type Result struct {
	Allowed bool
	Events  []Event
}

type Event struct {
	Rule    string
	Action  string
	Matched string
	Reason  string
}
