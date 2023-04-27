# Tomcat collector

Collected by jolokia

Install or download [Jolokia](http://repo1.maven.org/maven2/org/jolokia/jolokia-war/1.2.3/jolokia-war-1.2.3.war), rename it to jolokia.war, and Placed in the webapps directory of tomcat.
Edit tomcat-users.xml under the conf directory of tomcat, and add a user whose role is jolokia.

Take apache-tomcat-9.0.45 as an example (the username and password of jolokia user in the example must be modified!!!):

```ssh
$ cd apache-tomcat-9.0.45/

$ export tomcat_dir=`pwd`

$ wget http://repo1.maven.org/maven2/org/jolokia/jolokia-war/1.2.3/jolokia-war-1.2.3.war \
-O ./$tomcat_dir/webapps/jolokia.war

$ vim conf/tomcat-users.xml

 37 <!--
 38   <role rolename="tomcat"/>
 39   <role rolename="role1"/>
 40   <user username="tomcat" password="<must-be-changed>" roles="tomcat"/>
 41   <user username="both" password="<must-be-changed>" roles="tomcat,role1"/>
 42   <user username="role1" password="<must-be-changed>" roles="role1"/>
 43 -->
 44   <role rolename="jolokia"/>
 45   <user username="jolokia_user" password="secPassWd@123" roles="jolokia"/>
 46 
 47 </tomcat-users>


$ ./$tomcat_dir/bin/startup.sh 

 ...
 Tomcat started.
```

Go to http://localhost:8080/jolokia to check if the configuration is successful


Configuration example

```toml
[[inputs.tomcat]]
  ### Tomcat user(rolename="jolokia"). For example:
  # username = "jolokia_user"
  # password = "secPassWd@123"
  #

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

  ### Extra tags (optional)
  [inputs.tomcat.tags]
    # tag1 = "a"
```

## Tomcat metrics collection

* tomcat_global_request_processor

    |Metric|Description|Data Type|Unit|
    |:-- | - | - | -|
    |bytesReceived|Amount of data received, in bytes|int|count|
    |bytesSent|Amount of data sent, in bytes|int|count|
    |errorCount|Number of errors|int|count|
    |processingTime|Total time to process the requests|int|-|
    |requestCount|Number of requests processed|int|count|

* tomcat_jsp_monitor

    |Metric|Description|Data Type|Unit|
    |:-- | - | - | -|
    |jspCount|The number of JSPs that have been loaded into a webapp|int|count|
    |jspReloadCount|The number of JSPs that have been reloaded|int|count|
    |jspUnloadCount|The number of JSPs that have been unloaded|int|count|

* tomcat_thread_pool

    |Metric|Description|Data Type|Unit|
    |:-- | - | - | -|
    |currentThreadCount|currentThreadCount|int|count|
    |currentThreadsBusy|currentThreadsBusy|int|count|
    |maxThreads|maxThreads|int|count|

* tomcat_servlet

    |Metric|Description|Data Type|Unit|
    |:-- | - | - | -|
    |errorCount|Error|count|int|count|
    |processingTime|Totalexecutiontime of the servlet’s service method|int|-|
    requestCount|Number of requests processed by this wrapper|int|count|

* tomcat_cache

    |Metric|Description|Data Type|Unit|
    |:-- | - | - | -|
    |hitCount|The number of requests for resources that were served from the cache|int|count|
    |lookupCount|The number of requests for resources|int|count|
