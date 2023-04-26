
# DataKit Development Manual
---

## How to Add a Collector {#add-input}

Assuming that a new collector `zhangsan` is added, the following steps are generally followed:

- Add the module `zhangsan` under `plugins/inputs` and create an `input.go`
- Create a new structure in `input.go`

```golang
// Uniformly named Input
type Input struct {
	// Some configurable fields
	...

	// Generally, each collector can add a user-defined tag
	Tags   map[string]string
}
```

- The structure implements the following interfaces, for example, see `demo` collector:

```Golang
Catalog() string                  // Collector classifications, such as MySQL collectors, belong to the `db` classification
Run()                             // Collector entry function, which usually collects data here and sends the data to the `io` module
SampleConfig() string             // Sample collector configuration file
SampleMeasurement() []Measurement // Auxiliary structure of collector document generation
AvailableArchs() []string         // Operating system applicable to collector
```

> As some collector features are constantly being added, ==new collectors should implement all interfaces in plugins/inputs/inputs.go as much as possible==

- In `input.go`, add the following module initialization entry:

```Golang
func init() {
	inputs.Add("zhangsan", func() inputs.Input {
		return &Input{
			// Here you can initialize a bunch of default configuration parameters for this collector
		}
	})
}
```

- Add `import` in `plugins/inputs/all/all.go`

```Golang
import (
	... // Other existing collectors
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zhangsan"
)
```

- Add collectors to the top-level directory `checked.go`:

```Golang
allInputs = map[string]bool{
	"zhangsan":       false, // Note that it is initially set to false, and then changed to true when the collector is released
	...
}
```

- Perform compilation and replace the existing DataKit with the compiled binary. Take the Mac platform as an example:

```shell
$ make
$ tree dist/
dist/
└── datakit-darwin-amd64
    └── datakit          # Replace this dakakit with the existing datakit binary, typically /usr/local/datakit/datakit

sudo datakit service -T                                         # stop existing datakit
sudo truncate -s 0 /var/log/datakit/log                         # Empty the log
sudo cp -r dist/datakit-darwin-amd64/datakit /usr/local/datakit # Overlay binary
sudo datakit service -S                                         # restart datakit
```

- At this point, you typically have a `zhangsan.conf.sample` in the `/usr/local/datakit/conf.d/<Catalog>/` directory. Note that the `<Catalog>` here is the return value of the interface `Catalog() string` above.
- Open the `zhangsan` collector, make a copy of `zhangsan.conf` from `zhangsan.conf.sample`, modify the corresponding configuration (such as user name, directory configuration, etc.), and restart DataKit
- Check the collector condition by executing the following command:

```shell
sudo datakit check --config # Check whether the collector configuration file is normal
datakit -M --vvv            # Check the operation of all collectors
```

- If the collector function is complete, add `man/manuals/zhangsan.md` document, this can refer to `demo.md`, install the template inside to write

- For measurements in the document, the default is to list all the measurements that can be collected and their respective metrics in the document. Some special measurements or metrics, if there are preconditions, need to be explained in the document.
  - If a metric set needs to meet certain conditions, it should be described in `MeasurementInfo.Desc` of measurement
  - If there is a specific precondition for a metric in the measurement, it should be described on `FieldInfo.Desc`.

## Compile Environment Build {#setup-compile-env}

=== "Linux"

    #### Install Golang
    
    the current Go version [1.18.3](https://golang.org/dl/go1.18.3.linux-amd64.tar.gz)
    
    #### CI Settings
    
    > Assume go is installed in the /root/golang directory
    
    - Setting the directory
    
    ```
    # Create Go project path
    mkdir /root/go
    ```
    
    - Set the following environment variables
    
    ```
    export GO111MODULE=on
    # Set the GOPROXY environment variable
    export GOPRIVATE=gitlab.jiagouyun.com/*
    
    export GOPROXY=https://goproxy.io
    
    # Assume that golang is installed in the /root directory
    export GOROOT=/root/golang-1.18.3
    # Clone go code into GOPATH
    export GOPATH=/root/go
    export PATH=$GOROOT/bin:~/go/bin:$PATH
    ```
    
    Create a set of environment variables under `~/.ossenv` and fill in OSS Access Key and Secret Key for release:
    
    ```shell
    export RELEASE_OSS_ACCESS_KEY='LT**********************'
    export RELEASE_OSS_SECRET_KEY='Cz****************************'
    export RELEASE_OSS_BUCKET='zhuyun-static-files-production'
    export RELEASE_OSS_PATH=''
    export RELEASE_OSS_HOST='oss-cn-hangzhou-internal.aliyuncs.com'
    ```
    
    #### Install packr2
    
    Install [packr2](https://github.com/gobuffalo/packr/tree/master/v2){:target="_blank"}
    
    `go install github.com/gobuffalo/packr/v2/packr2@v2.8.3`
    
    #### Install common tools
    
    - tree
    - make
    - [goyacc](https://gist.github.com/tlightsky/9a163e59b6f3b05dbac8fc6b459a43c0): `go install golang.org/x/tools/cmd/goyacc@master`
    - [golangci-lint](https://golangci-lint.run/usage/install/#local-installation): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2`
    - gofumpt: `go install mvdan.cc/gofumpt@v0.1.1`
    - wget
    - docker
    - curl
    - [llvm](https://apt.llvm.org/): version >= 10.0
    - clang: version >= 10.0
    - linux kernel（>= 5.4.0-99-generic）header file: `apt-get install -y linux-headers-$(uname -r)` 
    
    #### Installing third-party libraries
    
    - `gcc-multilib`
    
    ```shell
    # Debian/Ubuntu
    sudo apt-get install -y gcc-multilib
    sudo apt-get install -y linux-headers-$(uname -r)
    # Centos: TODO
    ```

=== "Mac"

    not supported

=== "Windows"

    not supported

## Install, Upgrade and Test {#install-upgrade-testing}

After DataKit released new features, we had better do a full set of testing, including installation, upgrade and other processes. All existing DataKit installation files are stored on OSS. Let's use another isolated OSS bucket to do installation and upgrade tests.

Try this *default OSS path*：`oss://df-storage-dev/` (East China region). The following AK/SK can be obtained if necessary:

> Available for download [OSS Browser](https://help.aliyun.com/document_detail/209974.htm?spm=a2c4g.11186623.2.4.2f643d3bbtPfN8#task-2065478){:target="_blank"} client tool to view files in OSS.

- AK: `LTAIxxxxxxxxxxxxxxxxxxxx`
- SK: `nRr1xxxxxxxxxxxxxxxxxxxxxxxxxx`

In this OSS bucket, we specify that each developer has a subdirectory for storing their DataKit test files. The specific script is in the source code `scripts/build.sh`. Copy it to datakit source root directory, and slightly modify, can be used for local compilation and publishing.

### Custom Directory Running DataKit {#customize-workdir}

DataKit runs in the specified directory (/usr/local/DataKit under Linux) as ==service== by default, but you can customize the DataKit working directory to run in a non-service mode and read configuration and data from the specified directory in an additional way, which is mainly used to debug the functions of DataKit during development.

1. Update the latest code (dev branch)
2. Compile
3. Create the expected datakit working directory, such as `mkdir -p ~/datakit/conf.d`
4. Generate the default datakit.conf configuration file. Take Linux as an example, execute

```shell
./dist/datakit-linux-amd64/datakit tool --default-main-conf > ~/datakit/conf.d/datakit.conf
```

1. Modify the datakit.conf generated above:

	- Fill in `default_enabled_inputs` and add the list of collectors you want to open, typically `cpu,disk,mem` and so on
	- `http_api.listen` change the address
	- Change the token in `dataway.urls`
	- Change the logging directory/level if necessary
	- No more

2. Start the datakit, taking Linux as an example: `DK_DEBUG_WORKDIR=~/datakit ./dist/datakit-linux-amd64/datakit`
3. You can add a new alias to your local bash so that you can just run `ddk` each time you compile the DataKit (that is, Debugging-DataKit)

```shell
echo 'alias ddk="DK_DEBUG_WORKDIR=~/datakit ./dist/datakit-linux-amd64/datakit"' >> ~/.bashrc
source ~/.bashrc
```

In this way, the DataKit does not run as a service, and you can end the DataKit directly by ctrl+c

```shell
$ ddk
2021-08-26T14:12:54.647+0800    DEBUG   config  config/load.go:55       apply main configure...
2021-08-26T14:12:54.647+0800    INFO    config  config/cfg.go:361       set root logger to /tmp/datakit/log ok
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
  - using code:  gin.SetMode(gin.ReleaseMode)

	[GIN-debug] GET    /stats                    --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func1 (4 handlers)
	[GIN-debug] GET    /monitor                  --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func2 (4 handlers)
	[GIN-debug] GET    /man                      --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func3 (4 handlers)
	[GIN-debug] GET    /man/:name                --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func4 (4 handlers)
	[GIN-debug] GET    /restart                  --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func5 (4 handlers)
	...
```

You can also execute some command-line tools directly with ddk:

```shell
# Install IPDB
ddk install --ipdb iploc

# Query IP information
ddk debug --ipinfo 1.2.3.4
	    city: Brisbane
	province: Queensland
	 country: AU
	     isp: unknown
	      ip: 1.2.3.4
```

## Testing {#testing}

There are 2 types of testing in Datakit，one is integration testing, another is unit testing. There is no essential difference between them, but for integration testing, we have to set more environments.

Most of the time, we just run `make ut` for all testing, and we have to setup a Docker(remote or local) to help these integration testings. Here we show a example to do these:

- Configure a remote Docker and enable it's [remote function](https://medium.com/@ssmak/how-to-enable-docker-remote-api-on-docker-host-7b73bd3278c6){:target="_blank"}. For local Docker, nothing required to configure.

- Make a shell alias, start `make ut` within it:

```shell
alias ut='REMOTE_HOST=<YOUR-DOCKER-REMOTE-HOST> make ut'
```

Sometimes we need to configure more for integration testing:

- If we need to exclude some testing on package, we can add `UT_EXCLUDE` in the alias: `UT_EXCLUDE="gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp"`

- We can post the testing result to Guance Cloud, add a dataway and the token: `DATAWAY_URL="https://openway.guance.com/v1/write/logging?token=<YOUR-TOKEN>"`

The complete example:

```shell
alias ut='REMOTE_HOST=<YOUR-DOCKER-REMOTE-HOST> make ut UT_EXCLUDE="<package-name>" DATAWAY_URL="https://openway.guance.com/v1/write/logging?token=<YOUR-TOKEN>"'
```

## Release {#release}

The DataKit release consists of two parts:

- DataKit version release
- Document release

### DataKit Release {#release-dk}

The current release of DataKit is implemented in GitLab, which triggers the release of a specific branch of code once it is pushed to GitLab, as shown in _.gitlab-ci.yml_.

In versions prior to 1.2. 6 inclusive, the DataKit release relied on the output of the command `git describe --tags`. Since 1.2. 7, DataKit versions no longer rely on this mechanism, but by manually specifying the version number, the steps are as follows:

> Note: The current reliance on `git describe --tags` in script/build.sh is just a version acquisition policy issue and does not affect the main process.

- Edit *.gitlab-ci.yml* to modify the `VERSION`  variable inside, such as:

```yaml
    - make production GIT_BRANCH=$CI_COMMIT_BRANCH VERSION=1.2.8
```

Each release, you need to manually edit *.gitlab-ci.yml* to specify the version number.

- Add a tag to the code after the release is complete

```shell
git tag -f <same-as-the-new-version>
git push -f --tags
```

> Note: At present, the release of Mac version can only be released on Mac based on amd64 architecture. Because CGO is turned on, the Mac version of DataKit cannot be released on GitLab. It is implemented as follows:

```shell
make production_mac VERSION=<the-new-version>
make pub_production_mac VERSION=<the-new-version>
```

### DataKit Version Number Mechanism {#version-naming}

- Stable version: Its version number is `x.y.z`, where `y` must be an even number
- Non-stable version: its version number is `x.y.z`, where `y` must be cardinality

### Document Publishing {#release-docs}

Documentation can only be published on the development machine by installing [mkdocs](https://www.mkdocs.org/){:target="_blank"}. The process is as follows:

- Execute mkdocs.sh

```
./mkdocs.sh <the-new-version>
```

If no version is specified, the latest tag name is used as the version number.

> Note that if it is an online code release, it is best to ensure that it is consistent with **the current stable version number of online DataKit**, otherwise it will cause user trouble.

## About the Code Specification {#coding-rules}

We don't emphasize the specific code specification here. Existing tools can help us standardize our own code habits. At present, golint tools are introduced to check the existing code separately:

```golang
make lint
```

You can see various modification suggestions in check.err. For false positives, we can use `//nolint` to explicitly turn off:

```golang
// Obviously, 16 is the largest single-byte hexadecimal number, but gomnd in lint will report an error:
// mnd: Magic number: 16, in <return> detected (gomnd)
// But a suffix can be added here to mask this check
func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}

	// larger than any legal digit val
	return 16 //nolint:gomnd
}
```

> When to use `nolint`, see [here](https://golangci-lint.run/usage/false-positives/){:target="_blank"}

However, we do not recommend frequently adding `//nolint:xxx,yyy` to cover. Lint can be used in the following situations:

- Some well-known magic numbers in-such as 1024 for 1K and 16 for the maximum single-byte value.
- Security alerts that are really irrelevant, such as running a command in your code, but the command parameters are passed in from outside, but since the lint tool mentions them, it is necessary to consider whether there are possible security issues.

```golang
// cmd/datakit/cmds/monitor.go
cmd := exec.Command("/bin/bash", "-c", string(body)) //nolint:gosec
```
- Other places that may really need to be closed for inspection should be treated with caution.

## Troubleshoot DATA RACE {#data-race}

There are many DATA RACE problems in DataKit. These problems can be solved by adding a specific option when compiling DataKit, so that the compiled binary can automatically detect the code with DATA RACE during runtime.

To compile a DataKit with automatic detection of DATA RACE, the following conditions must be met:

- CGO must be turned on, so you can only make local (make by default).
- You must pass in the Makefile variable: `make RACE_DETECTION=on`

The compiled binary will increase a little, but it doesn't matter. We just need to test it locally. DATA RACE automatic detection has a feature, which can only be detected when the code runs to a specific code. Therefore, it is recommended that you automatically compile with `RACE_DETECTION=on` when testing your own functions daily, so as to find all the codes that cause DATA RACE as soon as possible.

### DATA RACE Doesn't Really Cause Data Disorder {#data-race-mess}

When a binary runtime with DATA RACE detection function encounters goroutines >=2 accessing the same data, and one of the goroutines executes write logic, it will print out code similar to the following on the terminal:

```shell hl_lines="8 9 10 11"
==================
WARNING: DATA RACE
Read at 0x00c000d40160 by goroutine 33:
  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/GuanceCloud/cliutils/dialtesting.(*HTTPTask).GetResults()
	  /Users/tanbiao/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/GuanceCloud/cliutils/dialtesting/http.go:208 +0x103c
	...

Previous write at 0x00c000d40160 by goroutine 74:
  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/GuanceCloud/cliutils/dialtesting.(*HTTPTask).Run.func2()
	  /Users/tanbiao/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/GuanceCloud/cliutils/dialtesting/http.go:306 +0x8c
	...
```

From these two pieces of information, we can know that the two codes work together on a data object, and at least one of them is a Write operation. However, it should be noted that only WARNING information is printed here, which means that this code does not necessarily lead to data problems, and the final problems need to be identified manually. For example, the following codes will not have data problems:

```golang

a = setupObject()

go func() {
	for {
		updateObject(a)
	}
}()
```

## Troubleshooting DataKit Memory Leaks {#mem-leak}

Edit DataKit.conf and add the following configuration fields at the top to turn on the DataKit remote pprof function:

```toml
enable_pprof = true
```

> If you install datakit for DaemonSet, you can inject environment variables:

```yaml
        - name: ENV_ENABLE_PPROF
          value: true
```

Restart DataKit to take effect.

### Get Pprof File {#get-pprof}

```shell
# Download the current DataKit active memory pprof file
wget http://<datakit-ip>:6060/debug/pprof/heap

# 下Download the current DataKit Total Allocated Memory pprof file (including memory that has been freed)
wget http://<datakit-ip>:6060/debug/pprof/allocs
```

> Port 6060 here is fixed and cannot be modified for the time being

Also accessed via the web `http://<datakit-ip>:6060/debug/pprof/heap?=debug=1`. You can also see some memory allocation information.

### View Pprof File {#use-pprof}

After downloading to the local, run the following command. After entering the interactive command, you can enter top to view the top10 hotspots of memory consumption:

```shell
$ go tool pprof heap 
File: datakit
Type: inuse_space
Time: Feb 23, 2022 at 9:06pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top                            <------ View top 10 memory hotspots
Showing nodes accounting for 7719.52kB, 88.28% of 8743.99kB total
Showing top 10 nodes out of 108
flat  flat%   sum%        cum   cum%
2048.45kB 23.43% 23.43%  2048.45kB 23.43%  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/alecthomas/chroma.NewLexer
1031.96kB 11.80% 35.23%  1031.96kB 11.80%  regexp/syntax.(*compiler).inst
902.59kB 10.32% 45.55%   902.59kB 10.32%  compress/flate.NewWriter
591.75kB  6.77% 52.32%   591.75kB  6.77%  bytes.makeSlice
561.50kB  6.42% 58.74%   561.50kB  6.42%  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/golang.org/x/net/html.init
528.17kB  6.04% 64.78%   528.17kB  6.04%  regexp.(*bitState).reset
516.01kB  5.90% 70.68%   516.01kB  5.90%  io.glob..func1
513.50kB  5.87% 76.55%   513.50kB  5.87%  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/github.com/gdamore/tcell/v2/terminfo/v/vt220.init.0
513.31kB  5.87% 82.43%   513.31kB  5.87%  gitlab.jiagouyun.com/cloudcare-tools/datakit/vendor/k8s.io/apimachinery/pkg/conversion.ConversionFuncs.AddUntyped
512.28kB  5.86% 88.28%   512.28kB  5.86%  encoding/pem.Decode
(pprof) 
(pprof) pdf                            <------ Output as pdf, that is, profile001. pdf will be generated in the current directory
Generating report in profile001.pdf
(pprof) 
(pprof) web                            <------ View it directly in the browser, and the effect is the same as PDF
```

> You can see the allocation of objects by `go tool pprof -sample_index=inuse_objects heap`, and consult `go tool pprof -help` for details.

In the same way, you can view the total allocated memory pprof file allocs. The effect of PDF is roughly as follows:

<figure markdown>
  ![](https://static.guance.com/images/datakit/datakit-pprof-pdf.png){ width="800" }
</figure>

For more ways to use pprof, see [here](https://www.freecodecamp.org/news/how-i-investigated-memory-leaks-in-go-using-pprof-on-a-large-codebase-4bec4325e192/){:target="_blank"}.

## DataKit Accessibility {#assist}

In addition to some of the accessibility features listed in the [official document](datakit-tools-how-to.md), DataKit supports other features that are used primarily during development.

### Check Sample Config is Correct {#check-sample-config}

```shell
datakit check --sample
------------------------
checked 52 sample, 0 ignored, 51 passed, 0 failed, 0 unknown, cost 10.938125ms
```

### Export Document {#export-docs}

Exports the existing DataKit document to the specified directory, specifies the document version, replaces the document marked `TODO` with `-` and ignores the collector `demo`.

```shell
man_version=`git tag -l | sort -nr | head -n 1` # Get the most recently released tag version
datakit doc --export-docs /path/to/doc --version $man_version --TODO "-" --ignore demo
```

## More Readings {#more-readings}

- [DataKit Monitor observer](datakit-monitor.md)
- [Introduction on DataKit to the overall architecture](datakit-arch.md)
