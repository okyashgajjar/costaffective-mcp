package targets

import "costaffective/internal/installer"

func init() {
	installer.RegisterTarget(&ClaudeTarget{})
	installer.RegisterTarget(&CursorTarget{})
	installer.RegisterTarget(&OpencodeTarget{})
	installer.RegisterTarget(&CodexTarget{})
	installer.RegisterTarget(&AntigravityTarget{})
}
