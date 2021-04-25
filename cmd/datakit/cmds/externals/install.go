package externals

import (
	"fmt"
	"strings"
)

func InstallExternal(service, url string) error {
	name := strings.ToLower(service)
	switch name {
	case "telegraf":
		return InstallTelegraf(url)
	default:
		return fmt.Errorf("Unsupport install %s", service)
	}

	return nil
}
