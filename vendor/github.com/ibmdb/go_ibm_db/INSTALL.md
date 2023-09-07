# Installing go_ibm_db

*Copyright (c) 2014 IBM Corporation and/or its affiliates. All rights reserved.*

Permission is hereby granted, free of charge, to any person obtaining a copy of this
software and associated documentation files (the "Software"), to deal in the Software
without restriction, including without limitation the rights to use, copy, modify, 
merge, publish, distribute, sublicense, and/or sell copies of the Software, 
and to permit persons to whom the Software is furnished to do so, subject to the 
following conditions:

The above copyright notice and this permission notice shall be included in all copies 
or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, 
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR 
PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE 
FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR 
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER 
DEALINGS IN THE SOFTWARE.

## Contents

1. [Overview](#Installation)
2. [ibm_db Installation on Linux](#inslnx)
3. [ibm_db Installation on MacOS](#insmac)
4. [ibm_db Installation on Windows](#inswin)

## <a name="overview"></a> 1. Overview

The [*go_ibm_db*](https://github.com/ibmdb/go_ibm_db) is an asynchronous/synchronous interface for GoLang to IBM DB2.

Following are the steps to installation in your system.

This go_ibm_db driver has been tested on 64-bit/32-bit IBM Linux, MacOS and Windows.





## <a name="inslnx"></a> 2. Go_ibm_db Installation on Linux.

### 2.1 Install GoLang for Linux

Download the
[GoLang Linux binaries](https://golang.org/dl) or [Go Latest binaries](https://go.dev/dl) and
extract the file, for example into `/mygo`:

```
cd /mygo
wget -c https://golang.org/dl/go1.20.5.linux-amd64.tar.gz
tar -xzf go1.20.5.linux-amd64.tar.gz
```

Set PATH to include Go:

```
export PATH=/mygo/go/bin:$PATH
```

### 2.2 Install go_ibm_db

Following are the steps to install [*go_ibm_db*](https://github.com/ibmdb/go_ibm_db) from github.
using directory `/goapp` for example.

#### 2.2.1 Direct Installation.
```
1. mkdir goapp
2. cd goapp
3. go install github.com/ibmdb/go_ibm_db/installer@latest
   or
   go install github.com/ibmdb/go_ibm_db/installer@v0.4.3
```

It's Done.

#### 2.2.2 Manual Installation by using git clone.

```
1. mkdir goapp
2. cd goapp
3. git clone https://github.com/ibmdb/go_ibm_db/
```

### 2.3 Download clidriver

Download clidriver in your system, use below command:
go to installer folder where go_ibm_db is downloaded in your system 
(Example: /home/uname/go/src/github.com/ibmdb/go_ibm_db/installer or /home/uname/goapp/go_ibm_db/installer 
where uname is the username) and run setup.go file (go run setup.go)


### 2.4 Set environment variables to clidriver directory path

#### 2.4.1 Manual
```
export IBM_DB_HOME=/home/uname/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib 
export LD_LIBRARY_PATH=/home/uname/clidriver/lib
or
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$IBM_DB_HOME/lib
```

#### 2.4.2 Script file
```
cd .../go_ibm_db/installer
source setenv.sh
```
## <a name="insmac"></a> 3. Go_ibm_db Installation on MacOS.

### 3.1 Install GoLang for Mac

Download the
[GoLang MacOS binaries](https://golang.org/dl) or [GoLang Latest binaries](https://go.dev/dl) and
extract the file.

### 3.2 Install go_ibm_db

#### 3.2.1 Direct Installation.
```
1. mkdir goapp
2. cd goapp
3. go install github.com/ibmdb/go_ibm_db/installer@latest
   or
   go install github.com/ibmdb/go_ibm_db/installer@v0.4.3
```

It's Done.

#### 3.2.2 Manual Installation by using git clone.
```
1. mkdir goapp
2. cd goapp
3. git clone https://github.com/ibmdb/go_ibm_db/
```

### 3.3 Download clidriver

Download clidriver in your system, use below command:
go to installer folder where go_ibm_db is downloaded in your system 
(Example: /home/uname/go/src/github.com/ibmdb/go_ibm_db/installer or /home/uname/goapp/go_ibm_db/installer 
where uname is the username) and run setup.go file (go run setup.go)


### 3.4 Set environment variables to clidriver directory path

#### 3.4.1 Manual
```
export IBM_DB_HOME=/home/uname/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib

export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/home/uname/go/src/github.com/ibmdb/clidriver/lib
or
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:$IBM_DB_HOME/lib
```

#### 3.4.2 Script file
```
cd .../go_ibm_db/installer
source setenv.sh
```

## <a name="inswin"></a> 4. Go_ibm_db Installation on Windows.

### 4.1 Install GoLang for Windows

Download the [Go Windows binary/installer](https://golang.org/dl) or [Go Latest binaries](https://go.dev/dl/) and
install it.

### 4.2 Install go_ibm_db

Following are the steps to install [*go_ibm_db*](https://github.com/ibmdb/go_ibm_db) from github.
using directory `/goapp` for example.

#### 4.2.1 Direct Installation.
```
1. mkdir goapp
2. cd gopapp
3. go install github.com/ibmdb/go_ibm_db/installer@latest
   or
   go install github.com/ibmdb/go_ibm_db/installer@v0.4.3
```

#### 4.2.2 Manual Installation by using git clone.
```
1. mkdir goapp
2. cd goapp
3. git clone https://github.com/ibmdb/go_ibm_db/
```

### 4.3 Download clidriver

Download clidriver in your system, go to installer folder where go_ibm_db is downloaded in your system, use below command: 
(Example: C:\Users\uname\go\src\github.com\ibmdb\go_ibm_db\installer or C:\goapp\go_ibm_db\installer 
 where uname is the username ) and run setup.go file (go run setup.go).


### 4.4 Set environment variables to clidriver directory path

#### 4.4.1 Manual
```
set IBM_DB_HOME=C:\Users\uname\go\src\github.com\ibmdb\clidriver
set PATH=%PATH%;C:\Users\uname\go\src\github.com\ibmdb\clidriver\bin
or 
set PATH=%PATH%;%IBM_DB_HOME%\bin
```

### 4.4.2 Script file 
```
cd .../go_ibm_db/installer
Run setenvwin.bat
```
It's Done.

4. Download platform specific clidriver from https://public.dhe.ibm.com/ibmdl/export/pub/software/data/db2/drivers/odbc_cli/ , untar/unzip it and set `IBM_DB_HOME` environmental variable to full path of extracted 'clidriver' directory, for example if clidriver is extracted as: `/home/mysystem/clidriver`, then set system level environment variable `IBM_DB_HOME=/home/mysystem/clidriver`.



## <a name="m1chip"></a> 5. Steps to install ibm_db on MacOS M1/M2 Chip system (arm64 architecture)
**Warning:** If you use the ARM version of homebrew (as recommended for M1/M2 chip systems) you will get the following error message:
```
$ brew install gcc-12
Error: Cannot install in Homebrew on ARM processor in Intel default prefix (/usr/local)!
Please create a new installation in /opt/homebrew using one of the
"Alternative Installs" from:
  https://docs.brew.sh/Installation
You can migrate your previously installed formula list with:
  brew bundle dump
```
Install `gcc@12` using homebrew `(note: the x86_64 version of homebrew is needed for this, not the recommended ARM based homebrew)`. The clearest instructions on how to install and use the `x86_64` version of `homebrew` is by following below steps:
*	My arm64/M1 brew is installed here:
```
	$ which brew
	/opt/homebrew/bin/brew
```
*	Step 1. Install x86_64 brew under /usr/local/bin/brew
	`arch -x86_64 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"`
*	Step 2. Create an alias for the x86_64 brew
	I added this to my ~/.bashrc as below:
```
	# brew hack for x86_64
	alias brew64='arch -x86_64 /usr/local/bin/brew'
```
* Then install gcc@12 using the x86_64 homebrew:
```
	brew64 install gcc@12
```
* Now find location of `lib/gcc/12/libstdc++.6.dylib` file in your system. It should be inside `/usr/local/homebrew/lib/gcc/12` or `/usr/local/lib/gcc/12` or `/usr/local/homebrew/Cellar/gcc@12/12.2.0/lib/gcc/12` or something similar. You need to find the correct path.
Suppose path of gcc lib is `/usr/local/homebrew/lib/gcc/12`. Then update your .bashrc/.zshrc file with below line
```
export DYLD_LIBRARY_PATH=/usr/local/homebrew/lib/gcc/12:$DYLD_LIBRARY_PATH
```
