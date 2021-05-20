package tomcat

var TomcatMetricType = map[string]string{
	"requestCount":       "int",
	"bytesReceived":      "int",
	"bytesSent":          "int",
	"processingTime":     "int",
	"errorCount":         "int",
	"jspCount":           "int",
	"jspReloadCount":     "int",
	"jspUnloadCount":     "int",
	"maxTHreads":         "int",
	"currentThreadCount": "int",
	"currentThreadsBusy": "int",
	"hitCount":           "int",
	"lookupCount":        "int",
}
