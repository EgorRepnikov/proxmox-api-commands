package constants

var (
	NODE_COMMANDS = map[string]struct{}{
		"shutdown": {},
		"reboot":   {},
	}
	VM_COMMANDS = map[string]struct{}{
		"start":    {},
		"stop":     {},
		"shutdown": {},
	}
)
