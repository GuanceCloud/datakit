// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sensitive

import "regexp"

var (
	datawayTokenRegexp  = regexp.MustCompile(`token=tkn_[A-Za-z0-9_]+`)
	passwordKVRegexp    = regexp.MustCompile(`(pass|password|bearer_token_string|sk|token)\s*=\s*(".*")`)
	passwordArgRegexp   = regexp.MustCompile(`('--password'\s*,\s*)'.*'\s*,`)
	userInfoURIRegexp   = regexp.MustCompile(`(["']?[A-Za-z0-9]+)\:\/\/([A-Za-z0-9_]+)\:(.+)\@`)
	querySecretRegexp   = regexp.MustCompile(`(?i)([?&](?:token|key|kv|secret|password|passwd|auth|credential)=)([^&\s]+)`)
	logSecretKVRegexp   = regexp.MustCompile(`(?i)((?:[A-Za-z0-9_]*)(?:token|key|kv|secret|password|passwd|auth|credential)(?:[A-Za-z0-9_]*)\s*[:=]\s*)(["']?)([^"',\s\]]+)`)
	sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|token|key_pw|secret|key)`)
)

func RedactBugReportString(str string, kinds []string) string {
	for _, kind := range kinds {
		switch kind {
		case "dataway":
			str = datawayTokenRegexp.ReplaceAllString(str, `token=******`)
		case "password":
			str = passwordKVRegexp.ReplaceAllString(str, `${1} = "******"`)
			str = passwordArgRegexp.ReplaceAllString(str, `${1}'******',`)
		case "uri":
			str = userInfoURIRegexp.ReplaceAllString(str, `${1}://${2}:******@`)
		default:
		}
	}

	return str
}

func IsSensitiveKey(key string) bool {
	return sensitiveKeyPattern.MatchString(key)
}

func RedactLogString(str string) string {
	str = RedactBugReportString(str, []string{"dataway", "password", "uri"})
	str = querySecretRegexp.ReplaceAllString(str, `${1}******`)
	str = logSecretKVRegexp.ReplaceAllString(str, `${1}${2}******`)
	return str
}
