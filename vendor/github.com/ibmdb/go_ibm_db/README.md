# go_ibm_db

Interface for GoLang to DB2 for z/OS, DB2 for LUW, DB2 for i.

## API Documentation

> For complete list of go_ibm_db APIs and examples please check [APIDocumentation.md](https://github.com/ibmdb/go_ibm_db/blob/master/API_DOCUMENTATION.md)

## Prerequisite

Golang should be installed(Golang version should be >=1.12.x and <= 1.20.X)

Git should be installed in your system.

For non-windows users, GCC and tar should be present in your system.

```
For Docker Linux Container(Ex: Amazon Linux2), use below commands:
yum install go git tar libpam
```

## Note:
Environment variable DB2HOME name is changed to IBM_DB_HOME.

## How to Install in Windows
```
You may install go_ibm_db using either of below commands
go get -d github.com/ibmdb/go_ibm_db
go install github.com/ibmdb/go_ibm_db/installer@latest
go install github.com/ibmdb/go_ibm_db/installer@v0.4.3

If you already have a clidriver available in your system, add the path of the same to your Path windows environment variable
Example: Path = C:\Program Files\IBM\IBM DATA SERVER DRIVER\bin


If you do not have a clidriver in your system, go to installer folder where go_ibm_db is downloaded in your system, use below command: 
(Example: C:\Users\uname\go\src\github.com\ibmdb\go_ibm_db\installer or C:\Users\uname\go\pkg\mod\github.com\ibmdb\go_ibm_db\installer 
 where uname is the username ) and run setup.go file (go run setup.go).


Set IBM_DB_HOME to clidriver downloaded path and
set this path to your PATH windows environment variable
(Example: Path=C:\Users\uname\go\src\github.com\ibmdb\clidriver)
set IBM_DB_HOME=C:\Users\uname\go\src\github.com\ibmdb\clidriver
set PATH=%PATH%;C:\Users\uname\go\src\github.com\ibmdb\clidriver\bin
or 
set PATH=%PATH%;%IBM_DB_HOME%\bin



Script file to set environment variable 
cd .../go_ibm_db/installer
setenvwin.bat

```

## How to Install in Linux/Mac
```
You may install go_ibm_db using either of below commands
go get -d github.com/ibmdb/go_ibm_db
go install github.com/ibmdb/go_ibm_db/installer@latest
go install github.com/ibmdb/go_ibm_db/installer@v0.4.2


If you already have a clidriver available in your system, set the below environment variables with the clidriver path

export IBM_DB_HOME=/home/uname/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib 
Linux:
export LD_LIBRARY_PATH=/home/uname/clidriver/lib
or
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$IBM_DB_HOME/lib
Mac:
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/Applications/clidriver/lib

If you do not have a clidriver available in your system, use below command:
go to installer folder where go_ibm_db is downloaded in your system 
(Example: /home/uname/go/src/github.com/ibmdb/go_ibm_db/installer or /home/uname/go/pkg/mod/github.com/ibmdb/go_ibm_db/installer 
where uname is the username) and run setup.go file (go run setup.go)

Set the below environment variables with the path of the clidriver downloaded

export IBM_DB_HOME=/home/uname/go/src/github.com/ibmdb/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib
Linux:
export LD_LIBRARY_PATH=/home/uname/go/src/github.com/ibmdb/clidriver/lib
or
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$IBM_DB_HOME/lib
Mac:
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/home/uname/go/src/github.com/ibmdb/clidriver/lib
or
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:$IBM_DB_HOME/lib


Script file to set environment variables in Linux/Mac 
cd .../go_ibm_db/installer
source setenv.sh

For Docker Linux Container, use below commands
yum install -y gcc git go wget tar xz make gcc-c++
cd /root
curl -OL https://golang.org/dl/go1.17.X.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.17.X.linux-amd64.tar.gz

rm /usr/bin/go
rm /usr/bin/gofmt
cp /usr/local/go/bin/go /usr/bin/
cp /usr/local/go/bin/gofmt /usr/bin/

go install github.com/ibmdb/go_ibm_db/installer@v0.4.2
or 
go install github.com/ibmdb/go_ibm_db/installer@latest

```

### <a name="Licenserequirements"></a> License requirements for connecting to databases

go_ibm_db driver can connect to DB2 on Linux Unix and Windows without any additional license/s, however, connecting to databases on DB2 for z/OS or DB2 for i(AS400) Servers require either client side or server side license/s. The client side license would need to be copied under `license` folder of your `clidriver` installation directory and for activating server side license, you would need to purchase DB2 Connect Unlimited Edition for System z® and DB2 Connect Unlimited Edition for System i®.

To know more about license and purchasing cost, please contact [IBM Customer Support](http://www-05.ibm.com/support/operations/zz/en/selectcountrylang.html).

To know more about server based licensing viz db2connectactivate, follow below links:
* [Activating the license certificate file for DB2 Connect Unlimited Edition](https://www.ibm.com/developerworks/community/blogs/96960515-2ea1-4391-8170-b0515d08e4da/entry/unlimited_licensing_in_non_java_drivers_using_db2connectactivate_utlility1?lang=en).
* [Unlimited licensing using db2connectactivate utility](https://www.ibm.com/developerworks/community/blogs/96960515-2ea1-4391-8170-b0515d08e4da/entry/unlimited_licensing_in_non_java_drivers_using_db2connectactivate_utlility1?lang=en.)

## How to run sample program

### example1.go:-

```go
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/ibmdb/go_ibm_db"
)

func main() {
	con := "HOSTNAME=host;DATABASE=name;PORT=number;UID=username;PWD=password"
	db, err := sql.Open("go_ibm_db", con)
	if err != nil {
		fmt.Println(err)
	}
	db.Close()
}
```
To run the sample:- go run example1.go

For complete list of connection parameters please check [this.](https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.swg.im.dbclient.config.doc/doc/c0054698.html)

### example2.go:-

```go
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/ibmdb/go_ibm_db"
)

func Create_Con(con string) *sql.DB {
	db, err := sql.Open("go_ibm_db", con)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

// Creating a table.

func create(db *sql.DB) error {
	_, err := db.Exec("DROP table SAMPLE")
	if err != nil {
		_, err := db.Exec("create table SAMPLE(ID varchar(20),NAME varchar(20),LOCATION varchar(20),POSITION varchar(20))")
		if err != nil {
			return err
		}
	} else {
		_, err := db.Exec("create table SAMPLE(ID varchar(20),NAME varchar(20),LOCATION varchar(20),POSITION varchar(20))")
		if err != nil {
			return err
		}
	}
	fmt.Println("TABLE CREATED")
	return nil
}

// Inserting row.

func insert(db *sql.DB) error {
	st, err := db.Prepare("Insert into SAMPLE(ID,NAME,LOCATION,POSITION) values('3242','mike','hyd','manager')")
	if err != nil {
		return err
	}
	st.Query()
	return nil
}

// This API selects the data from the table and prints it.

func display(db *sql.DB) error {
	st, err := db.Prepare("select * from SAMPLE")
	if err != nil {
		return err
	}
	err = execquery(st)
	if err != nil {
		return err
	}
	return nil
}

func execquery(st *sql.Stmt) error {
	rows, err := st.Query()
	if err != nil {
		return err
	}
	cols, _ := rows.Columns()
	fmt.Printf("%s    %s   %s    %s\n", cols[0], cols[1], cols[2], cols[3])
	fmt.Println("-------------------------------------")
	defer rows.Close()
	for rows.Next() {
		var t, x, m, n string
		err = rows.Scan(&t, &x, &m, &n)
		if err != nil {
			return err
		}
		fmt.Printf("%v  %v   %v         %v\n", t, x, m, n)
	}
	return nil
}

func main() {
	con := "HOSTNAME=host;DATABASE=name;PORT=number;UID=username;PWD=password"
	type Db *sql.DB
	var re Db
	re = Create_Con(con)
	err := create(re)
	if err != nil {
		fmt.Println(err)
	}
	err = insert(re)
	if err != nil {
		fmt.Println(err)
	}
	err = display(re)
	if err != nil {
		fmt.Println(err)
	}
}
```
To run the sample:- go run example2.go

### example3.go:-(POOLING)

```go
package main

import (
	_ "database/sql"
	"fmt"
	a "github.com/ibmdb/go_ibm_db"
)

func main() {
	con := "HOSTNAME=host;PORT=number;DATABASE=name;UID=username;PWD=password"
	pool := a.Pconnect("PoolSize=100")

	// SetConnMaxLifetime will take the value in SECONDS
	db := pool.Open(con, "SetConnMaxLifetime=30")
	st, err := db.Prepare("Insert into SAMPLE values('hi','hi','hi','hi')")
	if err != nil {
		fmt.Println(err)
	}
	st.Query()

	// Here the time out is default.
	db1 := pool.Open(con)
	st1, err := db1.Prepare("Insert into SAMPLE values('hi1','hi1','hi1','hi1')")
	if err != nil {
		fmt.Println(err)
	}
	st1.Query()

	db1.Close()
	db.Close()
	pool.Release()
	fmt.println("success")
}
```
To run the sample:- go run example3.go

### example4.go:-(POOLING- Limit on the number of connections)

```go
package main

import (
           "database/sql"
           "fmt"
           "time"
          a "github.com/ibmdb/go_ibm_db"
       )

func ExecQuery(st *sql.Stmt) error {
        res, err := st.Query()
        if err != nil {
             fmt.Println(err)
        }
        cols, _ := res.Columns()

        fmt.Printf("%s    %s   %s     %s\n", cols[0], cols[1], cols[2], cols[3])
        defer res.Close()
        for res.Next() {
                    var t, x, m, n string
                    err = res.Scan(&t, &x, &m, &n)
                    fmt.Printf("%v  %v   %v     %v\n", t, x, m, n)
        }
        return nil
}

func main() {
	con := "HOSTNAME=host;PORT=number;DATABASE=name;UID=username;PWD=password"
        pool := a.Pconnect("PoolSize=5")

        ret := pool.Init(5, con)
        if ret != true {
                fmt.Println("Pool initializtion failed")
        }

        for i:=0; i<20; i++ {
                db1 := pool.Open(con, "SetConnMaxLifetime=10")
                if db1 != nil {
                        st1, err1 := db1.Prepare("select * from VMSAMPLE")
                        if err1 != nil {
                                fmt.Println("err1 : ", err1)
                        }else{
                                go func() {
                                        execquery(st1)
                                        db1.Close()
                                }()
                        }
                }
        }
        time.Sleep(30*time.Second)
        pool.Release()
}
```
To run the sample:- go run example4.go

For Running the Tests:
======================

1) Put your connection string in the main.go file in testdata folder

2) Now run go test command (use go test -v command for details) 

3) To run a particular test case (use "go test sample_test.go main.go", example "go test Arraystring_test.go main.go")
