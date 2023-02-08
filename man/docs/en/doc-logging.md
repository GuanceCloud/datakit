# Design and Implementation of Datakit Log Collection System

## Preface {#head}

Log collection is an important item of Guance Cloud Datakit, which processes the actively collected or passively received log data and finally uploads it to the Guance Cloud center. Log collection can be divided into "network stream data" and "local disk file" according to data sources.

### Network Flow Data {#network}

Basically, they passively receive the data sent by the log generator by subscribing to the network interface.

The most common example is to look at the Docker log. When the command `docker logs -f CONTAIENR_NAME` is executed, the Docker starts a separate process and connects to the main process, receiving the data sent by the main process and outputting it to the terminal. Although the Docker logging process and the main process are on the same host, their interaction is through the local loopback network.

More complex weblog scenarios, such as Kubenetes clusters, have logs spread across different Nodes and need to be forwarded in api-server, which is twice as complex as Docker's single access link.

However, most logs obtained through the network have a problem-it is impossible to specify the log location. The log receiver can only choose to receive logs from the head, and may receive hundreds of thousands of logs at a time; Or start from the tail, something like `tail -f` only receives the latest data that is currently generated. If the process at the log receiver restarts, the log during this period will be lost.

Datakit's container log collection was initially received by the network, which was troubled by the above problems for a long time, and then gradually adjusted to the "local disk file" collection method mentioned below.

### Local Disk File {#disk-file}

Collecting local log files is the most common and efficient way, which saves the complicated transmission steps in the middle and accesses the disk files directly, which has higher maneuverability, but the implementation is more complicated and will encounter a series of detailed problems, such as:

- How to read data more efficiently on disk?
- What should we do if the file is deleted or rotated?
- How to locate the last location for "continuing reading" when reopening a file?

These problems are equivalent to spreading Docker log service, and all kinds of details and execution are handled by themselves, only the last network transmission part is omitted, and the complexity of implementation is much more troublesome than simply using network to receive.

This paper, focusing on "local disk file", will be divided into three aspects from bottom to top: "file discovery", "data collection and processing", "sending and synchronizing", and introduce the design and implementation details of Datakit log collection system in turn.

Additionally, the Datakit log collection execution flow is as follows, covering and subdividing the above "three aspects":

```
    glob Discover Files       Docker API Discover Files      Discover Files
         |                       |                            |
         ------------------------------------------------------
                                 |
                   Add to the log scheduler and assign to the specified lines
                                 |
         ---------------------------------------------------
         |                |                |               |
       line1            line2            line3          line4
                          |
                          |              |- Collect data, branch
                          |              |
                          |              |- Data transcoding
               |----->    |              |
               |          |              |- Special character processing
               |          |-  Document A      |
               |          | One collection cycle   |- Multi-line processing
               |          |              |
               |          |              |- Pipeline processing
               |          |              |
               |          |              |- Send
               |          |              |
               |          |              |- Synchronize file collection location
               |          |              |
               | Pipeline    |              |- File status detection
               | Cycling      |
               |          |
               |          |-  Document B |-
               |          |
               |          |
               |          |-  Document C |-
               |          |
               |----------|
```

## Discover and Locate Log File {#discovery}

Since you want to read and collect log files, you must first locate the file location on disk. There are mainly three kinds of file logs in Datakit, including two kinds of container logs and one kind of ordinary logs. Their collection methods are similar. This paper also mainly introduces these three kinds, which are:

- Normal log files
- Docker Stdout/Stderr, which is managed and dropped by the Docker service itself (Datakit currently only supports parsing the `json-file` driver)
- Containerd Stdout/Stderr. Containerd has no strategy for outputting logs. At present, Containerd Stdout/Stderr is managed by kubelet component of Kubenetes, which will be collectively referred to as `Containerd（CRI）`

### Found Common Log File {#discovery-log}

Ordinary log files are the most common, where processes write readable record data directly to disk files, such as the famous "log4j" framework or execute the `echo "this is log" >> /tmp/log` command to generate log files.

The file path of this log is mostly fixed. For example, the log path of mysql on Linux platform is `/var/log/mysql/mysql.log`. If you run Datakit MySQL collector, you will find a path to find the log file by default. However, log storage paths are configurable, and Datakit can't take care of all situations, so it must support manually specifying file paths.

In Datakit, the glob mode is used to configure the file path, which uses wildcard characters to locate the file name (of course, you can not use wildcard characters).

For example, there are now the following files:

```
$ tree /tmp
/tmp
├── datakit
│   ├── datakit-01.log
│   ├── datakit-02.log
│   └── datakit-03.log
└── mysql.d
    └── mysql
        └── mysql.log

3 directories, 4 files
```

In the Datakit logging collector, you can specify the log files to collect by configuring the `logfiles` parameter entry, such as:

- Collect all files in the `datakit` directory with glob `/tmp/datakit/*`
- Collect all files with the name `datakit` and the corresponding glob is `/tmp/datakit/datakit-*log`
- Collect `mysql.log`, but with `mysql.d` and `mysql` directories in between, there are several ways to navigate to the `mysql.log` file:
   - Specify directly: `/tmp/mysql.d/mysql/mysql.log`
   - A single asterisk specifies: `/tmp/*/*/mysql.log`, which is rarely used
   - Double star `double star`): `/tmp/**/mysql.log`, a simpler and more common way to use the double star `**` instead of the intermediate multi-tier directory structure

After using glob to specify the file path in the configuration file, Datakit periodically searches disk for files that meet the rules, and if it finds that they are not in the collection list, it adds them and collects them.

### Locate the container Stdout/Stderr Log File {#discovery-container-log}

There are two ways to output logs in a container:

- First, write directly to the mounted disk directory, which is the same as the above-mentioned "ordinary log files" in the host's view, and they are all files at a fixed location on the disk.
- Another way is to output to Stdout/Stderr, where the runtime of the container collects and manages the drop, which is also a more common way. This drop path can be obtained by accessing the runtime API.

Datakit gets the `LogPath` of the specified container by connecting to the sock file of Docker or Containerd and accessing their API, similar to executing `docker inspect --format='{{`{{.LogPath}}`}}' $INSTANCE_ID`:

```
$ docker inspect --format='{{`{{.LogPath}}`}}' cf681e
/var/lib/docker/containers/cf681eXXXX/cf681eXXXX-json.log
```

Once the container `LogPath` is obtained, the log collection is created using this path and the corresponding configuration.

## Collect Log Data and Process {#log-process}

###  Log Collection Scheduler {#scheduler}

After getting a log file path, because Datakit is written in Golang, it usually chooses to open one or more goroutines to collect files independently. The model is simple and easy to implement, which Datakit did before.

However, if the number of files is too large, the number of goroutines to be opened will also increase, which is very unfavorable to the management of goroutine. So Datakit implements a scheduler for log collection.

Like most scheduler models, Datakit implements multiple pipelines (lines) under the scheduler. When a new log collection is registered with the scheduler, Datakit allocates it according to the weights of each pipeline.

Each pipeline is executed cyclically, that is, A files are collected once (or continuously collected for N seconds, depending on the situation), then B files are collected, and then C files are collected, which can effectively control the number of goroutines and avoid the underlying scheduling and resource competition of a large number of goroutines.

If an error is found in this log collection, it will be removed from the pipeline and will not be collected again in the next cycle.

### Read Data and Split it into Rows {#read}

When it comes to reading log data, most of the time you think of a method function like `Readline()` that returns a full row of logs each time you call it. But this is not implemented in Datakit.

To ensure finer manipulation and higher performance, Datakit uses only the most basic `Read()` method, reading 4KiB of data at a time (buff size is 4KiB, and actually reading may be less), and manually dividing this 4KiB data into N parts by the newline character `\n`. This will lead to two situations:

- The last character of this 4KiB data is just a newline character, which can be divided into N parts, and there is no surplus.
- The last character of this 4KiB data is not a newline character. Compared with the above, there are only N-1 copies of this division, and there is a remaining part, which will be added to the head of the next 4KiB data, and so on.

In Datakit's code, here `update CursorPosition`, `copy` and `truncate` for the same buff to maximize memory reuse. 

After processing, the read data has become line by line, which can be moved to the next level of the execution stream, namely transcoding and special character processing.

### Transcoding and Special Character Processing {#decode}

Transcoding and special character processing should be carried out after data shaping, otherwise it will be truncated from the middle of characters and a truncated data will be processed. For example, a UTF-8 Chinese character occupies 3 bytes, and transcoding is done when the first byte is collected, which belongs to Undefined behavior.

Data transcoding is a very common behavior, which requires specifying the encoding type and the big end and small end (if any). This article focuses on "special character processing".

"Special characters" here refer to the color characters in the data. For example, the following command will output a red `rea`  word at the command line terminal:

```
$ RED='\033[0;31m' && NC='\033[0m' && print "${RED}red${NC}"
```

If you do not process and delete color characters, the final log data will also have `\033[0;31m`, not only lack of aesthetics, take up storage, and may have a negative impact on subsequent data processing. Therefore, special color characters should be screened out here.

There are many cases in the open source community, most of which are implemented using regular expressions, and the performance is not very good.

However, for a mature log output framework, there must be a way to turn off color characters, which Datakit recommends, and the log generator avoids printing color characters.

### Parse Row Data {#parse}

"Parsing row data" is primarily for container Stdout/Stderr logs. When the container runtime manages and drops the log, it adds some additional information fields, such as when it was generated, whether the source is `stdout` or `stderr`, whether this log is truncated, and so on. Datakit needs to parse this data and extract the corresponding fields.

- The Docker json-file log is in the following single format, JSON format, and the body is in the `log` field. If the end of the `log` content is `\n` , this row of data is complete and not truncated; If it is not `\n`, it means that the data is too long for more than 16KB and is truncated, and the rest of it is in the next JSON.
    ```
    {"log":"2022/09/14 15:11:11 Bash For Loop Examples. Hello, world! Testing output.\n","stream":"stdout","time":"2022-09-14T15:11:11.125641305Z"}
    ```
- The Containerd (CRI) single log format is as follows, with fields separated by spaces. Like Docker, Containerd (CRI) has a log truncation flag, the third field `P`, in addition to `F`. `P` means `Partial`, that is, incomplete and truncated; `F` means `Full`.
    ```
    2016-10-06T00:17:09.669794202Z stdout P log content 1
    2016-10-06T00:17:09.669794202Z stdout F log content 2
    ```
    The spliced log data is `log content 1 log content 2`.

By parsing the row data, we can get the log body, stdout/sterr and other information. According to the mark, we can determine whether it is an incomplete truncated log and splice the log. There is no truncation in ordinary log files, and a single line of data in a file can theoretically be infinitely long.

In addition, a single log line is truncated, and after splicing, it also belongs to a one-line log, instead of a multi-line log mentioned below, which are two different concepts.

### Multiline {#multiline}

Multi-line processing is a very important part of log collection, which makes some data that do not conform to the characteristics conform to the characteristics without losing the data. For example, the log file has the following data, which is a common Python stack print:

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

If there is no multi-line processing, the final data is the above 7 lines, which is exactly the same as the original text. This is not conducive to subsequent Pipeline cuts, because neither `Traceback (most recent call last):` in line 3 nor `File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app` in line 4 are fixed formats.

After effective multi-line processing, these 7 rows of data will become 3 rows, and the result is as follows:

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]\nTraceback (most recent call last):\n  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app\n    response = self.full_dispatch_request()
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

As you can see, each line of log data now begins with a character string such as `2020-10-23`, and lines 3, 4, and 5 that do not match the character in the original text are appended to the end of line 2. This looks much more beautiful and is beneficial to the subsequent Pipeline field cutting.

This function is not complicated, just specify the regular expression of the feature string.

The Datakit logging collector is configured with the `multiline_match` entry, which in the example above should be configured to `^\d{4}-\d{2}-\d{2}` to match a line header of the form `2020-10-23`.

The concrete implementation is similar to the stack structure in a data structure. If it conforms to the characteristics, it will stack the previous one and then put itself into the stack. If it does not conform to the characteristics, it will only add its own stack to the end of the previous one, so that the stack data received from the outside is consistent with the characteristics.

In addition, Datakit supports automatic multi-lines, in the configuration items of the logging collector `auto_multiline_detection` and `auto_multiline_extra_patterns`, and its logic is very simple: it provides a set of `multiline_match` that matches all the rules according to the original traversal, increasing its weight if the match succeeds so that it can be selected first next time.

Automatic multiline is a way to simplify configuration. In addition to user configuration, it also provides a "default automatic multiline rule list". See the end of the article for details.

### Pipeline Cut and Log Status {#pipeline}

Pipeline is a simple scripting language that provides various functions and syntax to write execution rules for a piece of text data. It is mainly used to cut unstructured text data, such as cutting a line of string text into multiple meaningful fields, or to extract some information from structured text (such as JSON).

The implementation of Pipeline is more complex, it consists of abstract syntax tree (AST) and a series of internal state machines and functional pure functions, which would not be described here.

Just look at the usage scenario, give a simple example, the original text is as follows:

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
```

pipeline script:

```python
grok(_, "%{date:time} %{NOTSPACE:status} %{GREEDYDATA:msg}")
default_time(time)
```

Final result:

```python
{
    "message": "2020-10-23 06:41:56,688 INFO demo.py 1.0",
    "msg": "demo.py 1.0",
    "status": "info",
    "time": 1603435316688000000
}
```

*Note: Pipeline's cut `status` field is `INFO`, but Datakit is mapped, so it appears as lowercase `info` for the sake of rigor*

Pipeline is the last step in log data processing, and Datakit uses the results of the Pipeline to build a row protocol, serialize the object, and prepare to package it and send it to Dataway.

## Send Data and Synchronize {#send}

Data sending is a very common behavior, and there are basically three steps in Datakit-"packaging", "transcoding" and "sending".

However, the operations after sending are essential and crucial, namely, "synchronizing the reading position of the current file" and "detecting the file status".

### Syncronization {#sync}

In the first section of the article, when introducing "network flow data", it was mentioned that in order to be able to continue reading log files at a fixed point, instead of only supporting "reading from the beginning of the file" or `tail -f` mode, Datakit introduced an important operation-recording the reading position of the current file.

Every time log collection reads data from a disk file, it will record the location of this data in the file. Only when a series of processing and sending are completed will this location information be connected with other data and synchronized to a single disk file.

The advantage of this is that Datakit starts log collection every time. If this log file has been collected before, it can locate the last location to continue collection this time, which will not cause repeated data collection and will not lose a certain piece of data in the middle.

Functional implementation is not complicated:

When Datakit starts log collection, it uses `File Absolute Path + File inode + N Bytes of File Header` to piece together an exclusive key value, and uses this key to find position in the file with the specified path.

- If the position can be found, it means that the file was collected last time and will be read again from the current position.
- If position is not found, it means that this is a new file, and it will be read from the beginning or the end of the file according to the situation.

### Detect File Status {#checking}

The state of a disk file is not static, and the file may be deleted, renamed, or left unchanged for a long time. Datakit has to deal with these situations.

- The file has not been modified for a long time:
    - Datakit will periodically get the modification time of the file (`file Modification Date`). If it finds that the distance from the current time exceeds a certain limit value, it will consider the file "inactive" and close it no longer to collect.
    - This logic also exists when searching for log files using glob rules. If a file that conforms to glob rules is found, but it has not been modified for a long time (or "updated"), it will not be logged.

- The file is rotated:
    - File rotate is a very complicated logic. Generally speaking, the file name remains unchanged, but the specific file pointed to by the file name changes, such as its inode. A typical example is Docker log drop.
    
Datakit will regularly check whether rotate has occurred in the file currently being collected. The logic of checking is to open a new file handle with this file name, call a function like `SameFile()`, and judge whether the two handles point to the same. If they are inconsistent, it means that rotate has occurred in the current file name.

Once it detects that the file has rotated, Datakit collects the remaining data of the current file (until EOF), reopens the file, which is already a new file, and then operates as usual.

## Summary {#end}

Log collection is a very complex system, which involves a lot of detail processing and optimization logic. The purpose of this paper is to introduce the design and implementation of Datakit log collection system, without Benchmark report and performance comparison with similar projects, the follow-up can be completed depending on the situation.

Supplementary links:

- [Introduction to the glob schema](https://en.wikipedia.org/wiki/Glob_(programming))
- [Datakit automatic multiline configuration](https://docs.guance.com/integrations/logging/#auto-multiline)
- [Datakit pipeline processing](https://docs.guance.com/datakit/pipeline/)
- [Docker truncates discussions over 16KiB logs](https://github.com/moby/moby/issues/34855)
- [Docker truncates more than 16KiB of source code](https://github.com/nalind/docker/blob/master/daemon/logger/copier.go#L13)
- [Docker logging driver](https://docs.docker.com/config/containers/logging/local/)
