package go_ibm_db

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/ibmdb/go_ibm_db/api"
)

// CreateDb function will take the db name and user details as parameters
// and create the database.
func CreateDb(dbname string, connStr string, options ...string) (bool, error) {
	if dbname == "" {
		return false, fmt.Errorf("Database name cannot be empty")
	}
	var codeset, mode string
	count := len(options)
	if count > 0 {
		for i := 0; i < count; i++ {
			opt := strings.Split(options[i], "=")
			if opt[0] == "codeset" {
				codeset = opt[1]
			} else if opt[0] == "mode" {
				mode = opt[1]
			} else {
				return false, fmt.Errorf("not a valid parameter")
			}
		}
	}
	connStr = connStr + ";" + "ATTACH=true"
	return createDatabase(dbname, connStr, codeset, mode)
}

func createDatabase(dbname string, connStr string, codeset string, mode string) (bool, error) {
	var out api.SQLHANDLE
	in := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	bufDBN := api.StringToUTF16(dbname)
	bufCS := api.StringToUTF16(connStr)
	bufC := api.StringToUTF16(codeset)
	bufM := api.StringToUTF16(mode)

	ret := api.SQLAllocHandle(api.SQL_HANDLE_ENV, in, &out)
	if IsError(ret) {
		return false, NewError("SQLAllocHandle", api.SQLHENV(in))
	}
	drvH := api.SQLHENV(out)
	ret = api.SQLAllocHandle(api.SQL_HANDLE_DBC, api.SQLHANDLE(drvH), &out)
	if IsError(ret) {
		defer releaseHandle(drvH)
		return false, NewError("SQLAllocHandle", drvH)
	}
	hdbc := api.SQLHDBC(out)
	ret = api.SQLDriverConnect(hdbc, 0,
		(*api.SQLWCHAR)(unsafe.Pointer(&bufCS[0])), api.SQLSMALLINT(len(bufCS)),
		nil, 0, nil, api.SQL_DRIVER_NOPROMPT)
	if IsError(ret) {
		defer releaseHandle(hdbc)
		return false, NewError("SQLDriverConnect", hdbc)
	}
	if codeset == "" && mode == "" {
		ret = api.SQLCreateDb(hdbc, (*api.SQLWCHAR)(unsafe.Pointer(&bufDBN[0])), api.SQLINTEGER(len(bufDBN)), nil, 0, nil, 0)
	} else if codeset == "" {
		ret = api.SQLCreateDb(hdbc, (*api.SQLWCHAR)(unsafe.Pointer(&bufDBN[0])), api.SQLINTEGER(len(bufDBN)), nil, 0, (*api.SQLWCHAR)(unsafe.Pointer(&bufM[0])), api.SQLINTEGER(len(bufM)))
	} else if mode == "" {
		ret = api.SQLCreateDb(hdbc, (*api.SQLWCHAR)(unsafe.Pointer(&bufDBN[0])), api.SQLINTEGER(len(bufDBN)), (*api.SQLWCHAR)(unsafe.Pointer(&bufC[0])), api.SQLINTEGER(len(bufC)), nil, 0)
	} else {
		ret = api.SQLCreateDb(hdbc, (*api.SQLWCHAR)(unsafe.Pointer(&bufDBN[0])), api.SQLINTEGER(len(bufDBN)), (*api.SQLWCHAR)(unsafe.Pointer(&bufC[0])), api.SQLINTEGER(len(bufC)), (*api.SQLWCHAR)(unsafe.Pointer(&bufM[0])), api.SQLINTEGER(len(bufM)))
	}
	if IsError(ret) {
		defer releaseHandle(hdbc)
		return false, NewError("SQLCreateDb", hdbc)
	}
	defer releaseHandle(hdbc)
	return true, nil
}

// DropDb function will take the db name and user details as parameters
// and drop the database.
func DropDb(dbname string, connStr string) (bool, error) {
	if dbname == "" {
		return false, fmt.Errorf("Database name cannot be empty")
	}
	connStr = connStr + ";" + "ATTACH=true"
	return dropDatabase(dbname, connStr)
}

func dropDatabase(dbname string, connStr string) (bool, error) {
	var out api.SQLHANDLE
	in := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	bufDBN := api.StringToUTF16(dbname)
	bufCS := api.StringToUTF16(connStr)

	ret := api.SQLAllocHandle(api.SQL_HANDLE_ENV, in, &out)
	if IsError(ret) {
		return false, NewError("SQLAllocHandle", api.SQLHENV(in))
	}
	drvH := api.SQLHENV(out)
	ret = api.SQLAllocHandle(api.SQL_HANDLE_DBC, api.SQLHANDLE(drvH), &out)
	if IsError(ret) {
		defer releaseHandle(drvH)
		return false, NewError("SQLAllocHandle", drvH)
	}
	hdbc := api.SQLHDBC(out)
	ret = api.SQLDriverConnect(hdbc, 0,
		(*api.SQLWCHAR)(unsafe.Pointer(&bufCS[0])), api.SQLSMALLINT(len(bufCS)),
		nil, 0, nil, api.SQL_DRIVER_NOPROMPT)
	if IsError(ret) {
		defer releaseHandle(hdbc)
		return false, NewError("SQLDriverConnect", hdbc)
	}
	ret = api.SQLDropDb(hdbc, (*api.SQLWCHAR)(unsafe.Pointer(&bufDBN[0])), api.SQLINTEGER(len(bufDBN)))
	if IsError(ret) {
		defer releaseHandle(hdbc)
		return false, NewError("SQLDropDb", hdbc)
	}
	defer releaseHandle(hdbc)
	return true, nil
}
