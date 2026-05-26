//go:build linux
// +build linux

package procwatch

import (
	"path/filepath"
	"strings"
)

var scriptExts = map[string]struct{}{
	".jar": {},
	".js":  {},
	".mjs": {},
	".py":  {},
	".pyc": {},
	".rb":  {},
	".pl":  {},
	".php": {},
	".sh":  {},
}

func detectServiceName(procName string, envKeys []string, env map[string]string, cmdline []string) string {
	for _, envKey := range envKeys {
		if value := strings.TrimSpace(env[envKey]); value != "" {
			return value
		}
	}

	if value := detectServiceNameFromCmdline(procName, cmdline); value != "" {
		return value
	}

	return procName
}

func detectServiceNameFromCmdline(procName string, cmdline []string) string {
	if len(cmdline) == 0 {
		return ""
	}

	mainProg := normalizeServiceToken(cmdline[0])
	switch mainProg {
	case "java":
		if value := findArgAfterFlag(cmdline[1:], "-jar"); value != "" {
			return value
		}
	case "python", "python2", "python3", "node", "nodejs", "ruby", "perl", "php", "bash", "sh":
		if value := findScriptArg(cmdline[1:]); value != "" {
			return value
		}
	}

	if procName == "" || procName == mainProg {
		return findScriptLikeArg(cmdline[1:])
	}

	return ""
}

func findArgAfterFlag(args []string, flag string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			return normalizeServiceToken(args[i+1])
		}
	}
	return ""
}

func findScriptArg(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(strings.Trim(args[i], `"'`))
		if arg == "" {
			continue
		}
		switch arg {
		case "-m":
			if i+1 < len(args) {
				return normalizeModuleToken(args[i+1])
			}
			return ""
		case "-c":
			return ""
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return normalizeServiceToken(arg)
	}
	return ""
}

func findScriptLikeArg(args []string) string {
	for _, raw := range args {
		arg := strings.TrimSpace(strings.Trim(raw, `"'`))
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(arg))
		if _, ok := scriptExts[ext]; ok {
			return normalizeServiceToken(arg)
		}
	}
	return ""
}

func normalizeModuleToken(module string) string {
	module = strings.TrimSpace(strings.Trim(module, `"'`))
	module = strings.TrimSuffix(module, ".py")
	return module
}

func normalizeServiceToken(path string) string {
	path = strings.TrimSpace(strings.Trim(path, `"'`))
	if path == "" {
		return ""
	}

	base := filepath.Base(path)
	if base == "." || base == string(filepath.Separator) {
		base = path
	}

	if ext := strings.ToLower(filepath.Ext(base)); ext != "" {
		if _, ok := scriptExts[ext]; ok {
			base = strings.TrimSuffix(base, ext)
		}
	}

	return base
}
