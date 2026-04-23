package configs

import _ "embed"

//go:embed claude-orchestrator.tmux.conf
var TmuxConfig string

//go:embed claude-memory.md
var ClaudeMemorySection string

//go:embed claude-sessions.md
var ClaudeSessionsSection string
