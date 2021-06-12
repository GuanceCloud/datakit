package tomcat

const (
	tomcatSampleCfg = `
[[inputs.tomcat]]
  ### Tomcat user(rolename="jolokia"). For example:
  # username = "jolokia_user"
  # password = "secPassWd@123"

  # response_timeout = "5s"

  urls = ["http://localhost:8080/jolokia"]

  ### Optional TLS config
  # tls_ca = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ### Monitor Interval
  # interval = "15s"

  [inputs.tomcat.log]
    # files = []
    ## grok pipeline script path
    # pipeline = "tomcat.p"

  [inputs.tomcat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...

  ### Tomcat metrics
  [[inputs.tomcat.metric]]
    name     = "tomcat_global_request_processor"
    mbean    = '''Catalina:name="*",type=GlobalRequestProcessor'''
    paths    = ["requestCount","bytesReceived","bytesSent","processingTime","errorCount"]
    tag_keys = ["name"]

  [[inputs.tomcat.metric]]
    name     = "tomcat_jsp_monitor"
    mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,name=jsp,type=JspMonitor"
    paths    = ["jspReloadCount","jspCount","jspUnloadCount"]
    tag_keys = ["J2EEApplication","J2EEServer","WebModule"]

  [[inputs.tomcat.metric]]
    name     = "tomcat_thread_pool"
    mbean    = "Catalina:name=\"*\",type=ThreadPool"
    paths    = ["maxThreads","currentThreadCount","currentThreadsBusy"]
    tag_keys = ["name"]

  [[inputs.tomcat.metric]]
    name     = "tomcat_servlet"
    mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,j2eeType=Servlet,name=*"
    paths    = ["processingTime","errorCount","requestCount"]
    tag_keys = ["name","J2EEApplication","J2EEServer","WebModule"]

  [[inputs.tomcat.metric]]
    name     = "tomcat_cache"
    mbean    = "Catalina:context=*,host=*,name=Cache,type=WebResourceRoot"
    paths    = ["hitCount","lookupCount"]
    tag_keys = ["context","host"]
    tag_prefix = "tomcat_"`

	pipelineCfg = `
# juli OneLineFormatter format
# cataline / host-manager / localhost / manager log 
add_pattern("olf_time", "%{MONTHDAY}-%{MONTH}-%{YEAR} %{TIME}")
grok(_, "%{olf_time:time} %{LOGLEVEL:status} \\[%{NOTSPACE:thread_name}\\] %{NOTSPACE:report_source} %{GREEDYDATA:msg}")
  
# localhost_access_log log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")
group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)

nullif(http_ident, "-")
nullif(http_auth, "-")

default_time(time)
`
)
