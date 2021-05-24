{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 tomcat 指标

## 前置条件

安装或下载 [Jolokia](http://repo1.maven.org/maven2/org/jolokia/jolokia-war/1.2.3/jolokia-war-1.2.3.war), 重命名为 jolokia.war, 并放置于 tomcat 的 webapps 目录下。
编辑 tomcat 的 conf 目录下的 tomcat-users.xml，增加 role 为 jolokia 的用户。

以 apache-tomcat-9.0.45 为例（示例中的 jolokia user 的 username 和 password 请务必修改！！！）:

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

前往 http://localhost:8080/jolokia 查看是否配置成功

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志采集

如需采集 Tomcat 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Tomcat 日志文件的绝对路径。比如：

``` toml
  [inputs.tomcat.log]
    files = []
```

开启日志采集以后，默认会产生日志来源（`source`）为 `tomcat` 的日志。

>注意：必须将 DataKit 安装在 NGINX 所在主机才能采集 Tomcat 日志
