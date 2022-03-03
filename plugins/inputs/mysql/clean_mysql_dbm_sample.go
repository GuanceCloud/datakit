package mysql

import "strings"

func getCleanMysqlVersion(r rows) *mysqlVersion {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	for r.Next() {
		var versionStr string
		var version mysqlVersion

		if err := r.Scan(&versionStr); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		parts := strings.Split(versionStr, "-")
		version.version = parts[0]

		for _, item := range parts {
			if item == strMariaDB {
				version.flavor = strMariaDB
			} else if len(version.flavor) == 0 {
				version.flavor = "MySQL"
			}

			if strings.Contains("log,standard,debug,valgrind,embedded,", item) { //nolint:gocritic
				version.build = item
			}
		}

		if len(version.build) == 0 {
			version.build = "unspecified"
		}

		return &version
	}

	return nil
}

func getCleanEnabledPerformanceSchemaConsumers(r rows) []string {
	if r == nil {
		return nil
	}

	var consumers []string

	defer closeRows(r)

	for r.Next() {
		var name string

		if err := r.Scan(&name); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		consumers = append(consumers, name)
	}

	return consumers
}
