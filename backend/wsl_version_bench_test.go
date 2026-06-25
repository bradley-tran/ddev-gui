package backend

import (
	"strings"
	"testing"
)

var wslOutputStr = `  NAME                   STATE           VERSION
* Ubuntu                 Running         2
  Ubuntu-20.04           Stopped         2
  docker-desktop         Running         2
  docker-desktop-data    Running         2
`

func BenchmarkParseWSLLinesOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines := strings.Split(wslOutputStr, "\n")
		ubuntuInstalled := false
		var wslLines []string
		for _, ln := range lines {
			l := strings.TrimSpace(ln)
			if l == "" || strings.HasPrefix(strings.ToUpper(l), "NAME") {
				continue
			}
			if strings.Contains(strings.ToLower(l), "ubuntu") || strings.Contains(l, "DDEV") {
				ubuntuInstalled = true
			}
			wslLines = append(wslLines, l)
		}
		if !ubuntuInstalled {
			wslLines = append(wslLines, "Warning: ddev requires a ubuntu-based wsl distro ⚠️")
		}
	}
}

func BenchmarkParseWSLLinesNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines := strings.Split(wslOutputStr, "\n")
		ubuntuInstalled := false
		var wslLines []string
		for _, ln := range lines {
			l := strings.TrimSpace(ln)
			if l == "" || (len(l) >= 4 && strings.EqualFold(l[:4], "NAME")) {
				continue
			}
			if !ubuntuInstalled {
				// quick check for DDEV or ubuntu case-insensitive
				if strings.Contains(l, "DDEV") || strings.Contains(strings.ToLower(l), "ubuntu") {
					ubuntuInstalled = true
				}
			}
			wslLines = append(wslLines, l)
		}
		if !ubuntuInstalled {
			wslLines = append(wslLines, "Warning: ddev requires a ubuntu-based wsl distro ⚠️")
		}
	}
}
