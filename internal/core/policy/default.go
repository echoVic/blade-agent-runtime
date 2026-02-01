package policy

func DefaultPolicy() *Policy {
	return &Policy{
		Version: 1,
		Rules: []Rule{
			{
				Name:    "no-rm-rf-root",
				Pattern: `rm\s+(-[a-zA-Z]*r[a-zA-Z]*f|(-[a-zA-Z]*f[a-zA-Z]*r))\s+/\s*$`,
				Action:  "block",
				Reason:  "Dangerous: recursive delete from root",
			},
			{
				Name:    "no-rm-rf-home",
				Pattern: `rm\s+(-[a-zA-Z]*r[a-zA-Z]*f|(-[a-zA-Z]*f[a-zA-Z]*r))\s+~/?\s*$`,
				Action:  "block",
				Reason:  "Dangerous: recursive delete home directory",
			},
			{
				Name:    "no-disk-write",
				Pattern: `>\s*/dev/sd[a-z]`,
				Action:  "block",
				Reason:  "Dangerous: write to disk device",
			},
			{
				Name:    "no-mkfs",
				Pattern: `mkfs\s+`,
				Action:  "block",
				Reason:  "Dangerous: format filesystem",
			},
			{
				Name:    "no-dd-disk",
				Pattern: `dd\s+.*of=/dev/sd[a-z]`,
				Action:  "block",
				Reason:  "Dangerous: write to disk device with dd",
			},
			{
				Name:    "warn-sudo",
				Pattern: `^sudo\s+`,
				Action:  "warn",
				Reason:  "Warning: running with elevated privileges",
			},
			{
				Name:    "warn-chmod-777",
				Pattern: `chmod\s+777`,
				Action:  "warn",
				Reason:  "Warning: setting world-writable permissions",
			},
			{
				Name:    "warn-curl-pipe-sh",
				Pattern: `curl\s+.*\|\s*(ba)?sh`,
				Action:  "warn",
				Reason:  "Warning: piping curl output to shell",
			},
			{
				Name:    "log-git-push",
				Pattern: `git\s+push`,
				Action:  "log",
				Reason:  "Logging: git push operation",
			},
			{
				Name:    "log-npm-publish",
				Pattern: `npm\s+publish`,
				Action:  "log",
				Reason:  "Logging: npm publish operation",
			},
		},
	}
}

func DefaultPolicyYAML() string {
	return `version: 1

rules:
  - name: no-rm-rf-root
    pattern: 'rm\s+(-[a-zA-Z]*r[a-zA-Z]*f|(-[a-zA-Z]*f[a-zA-Z]*r))\s+/\s*$'
    action: block
    reason: "Dangerous: recursive delete from root"

  - name: no-rm-rf-home
    pattern: 'rm\s+(-[a-zA-Z]*r[a-zA-Z]*f|(-[a-zA-Z]*f[a-zA-Z]*r))\s+~/?\s*$'
    action: block
    reason: "Dangerous: recursive delete home directory"

  - name: no-disk-write
    pattern: '>\s*/dev/sd[a-z]'
    action: block
    reason: "Dangerous: write to disk device"

  - name: no-mkfs
    pattern: 'mkfs\s+'
    action: block
    reason: "Dangerous: format filesystem"

  - name: no-dd-disk
    pattern: 'dd\s+.*of=/dev/sd[a-z]'
    action: block
    reason: "Dangerous: write to disk device with dd"

  - name: warn-sudo
    pattern: '^sudo\s+'
    action: warn
    reason: "Warning: running with elevated privileges"

  - name: warn-chmod-777
    pattern: 'chmod\s+777'
    action: warn
    reason: "Warning: setting world-writable permissions"

  - name: warn-curl-pipe-sh
    pattern: 'curl\s+.*\|\s*(ba)?sh'
    action: warn
    reason: "Warning: piping curl output to shell"

  - name: log-git-push
    pattern: 'git\s+push'
    action: log
    reason: "Logging: git push operation"

  - name: log-npm-publish
    pattern: 'npm\s+publish'
    action: log
    reason: "Logging: npm publish operation"
`
}
