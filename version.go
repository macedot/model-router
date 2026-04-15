package main

import "runtime/debug"

// FullVersion can be set via -ldflags "-X main.FullVersion=..." at build time.
// If unset, GetVersion() falls back to VCS metadata from debug.ReadBuildInfo.
var FullVersion = ""

func GetVersion() string {
	if FullVersion != "" {
		return FullVersion
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var rev, t string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) > 12 {
				rev = s.Value[:12]
			} else {
				rev = s.Value
			}
		case "vcs.time":
			if len(s.Value) >= 10 {
				t = s.Value[:10]
			}
		}
	}
	if rev == "" {
		return "dev"
	}
	return rev + " (" + t + ")"
}
