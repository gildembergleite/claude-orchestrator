package configs

import _ "embed"

//go:embed zarc.tmux.conf
var TmuxConfig string

//go:embed claude-memory.md
var ClaudeMemorySection string
