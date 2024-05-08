package cmd

// Options modifies the execution of git and git-lfs commands.
type Options struct {
	GitOptions
	*LFSOptions
}

// GitOptions modify the execution of git commands.
type GitOptions struct {
	Force      bool
	AltGitExec string
}

// LFSOptions modify the execution of git-lfs commands.
type LFSOptions struct {
	WithLFS    bool // will be overridden if git-lfs command is not found.
	AltLFSExec string
	ServerURL  string
}
