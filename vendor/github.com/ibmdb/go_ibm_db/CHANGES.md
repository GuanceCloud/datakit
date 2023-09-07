2022-10-25, Version 0.4.2
=========================

 * Change IBM_DB2_HOME to IBM_DB_HOME (Bimal Kumar Jha)

 * Document corrections (Vikas Mathur)

 * Fix handling of os.LookupEnv return code in installer (Eric Newsom)

 * Changed environment variable name DB2HOME to IBM_DB_HOME (Bimal Kumar Jha)

 * Updated review comment (Vikas Mathur)

 * Changed env variable name DB2HOME to IBM_DB_HOME (Vikas Mathur)

 * Connection Pooling Enhancement to create connection in advance (Bimal Kumar Jha)

 * Added pooling with limit on the number of connetions example (Vikas Mathur)

 * Update API_DOCUMENTATION.md (VIKAS MATHUR)

 * Documented the new APIs (Vikas Mathur)

 * Addressed code review commments (Vikas Mathur)

 * Connection Pooling - Address review comment (Vikas Mathur)

 * Updated Connection Pooling (Vikas Mathur)

 * Added Context methods (Vikas Mathur)

 * Updated Golang install version (Vikas Mathur)

 * Addressed review comments (Vikas Mathur)

 * Updated for Docker Linux Container (Vikas Mathur)

 * Script file to set environment variables (Vikas Mathur)

 * Script file to set enviornment variables (Vikas Mathur)

 * Support for version 1.17 (Vikas Mathur)

 * Updated concurrent map write on the pool (Vikas Mathur)

 * Updated Library Path and fix decimal column type error (Vikas Mathur)

 * Update ExecDirect_test.go (yyt)

 * #126 fix (Ravuri Sai Ram Akhil)

 * error msg changed for when using db.Query (Ravuri Sai Ram Akhil)


2021-04-08, Version 0.4.1
=========================

 * GOARCh changed from 390 to s390x (Ravuri Sai Ram Akhil)


2021-03-17, Version 0.4.0
=========================

 * Readme updated to 1.16 (Ravuri Sai Ram Akhil)

 * Readme changed for 1.16 (Ravuri Sai Ram Akhil)

 * Clidriver Dowload location updated to support go1.16 (Ravuri Sai Ram Akhil)

 * println is removed (Ravuri Sai Ram Akhil)

 * Precision and Scale changed Decimal and Numeric (Ravuri Sai Ram Akhil)

 * Decimal value fix #116 (Ravuri Sai Ram Akhil)

 * Fix an additional typo. (David Mooney)

 * Run gofmt on API docs. (David Mooney)

 * Run gofmt on README code samples. (David Mooney)

 * Fix Misspell clidriver (Apipol Sukgler)

 * setup.go support AIX. (yunbozh)

 * Improve comment (alejandro.labad)

 * Improved solution as it is done by sql.go (alejandro.labad)

 * Propose solution for parametrized queries (alejandro.labad)

 * Fix issue # 100 where store procedure output parameters was being truncated. (Rajesh Jayakumar)

 * Support for aix (rahul-shinge)


2020-06-02, Version 0.3.0
=========================

 * ExecDirect test case added (Ravuri Sai Ram Akhil)

 * queryRow testcase updated and Direct Execute api added (Ravuri Sai Ram Akhil)

 * Allow null insertion to some binary data types (Santiago De la Cruz)

 * License requiements added (Ravuri Sai Ram Akhil)

 * Readme updated with <= in versions (Ravuri Sai Ram Akhil)

 * Readme updated with new line after versions (Ravuri Sai Ram Akhil)

 * Readme updated with versions (Ravuri Sai Ram Akhil)

 * Set 1 for default year, month and day SQL_C_TYPE_TIME (Santiago De la Cruz)

 * c_decfloat is removed (Ravuri Sai Ram Akhil)

 * Don't create a variable for len of data binding []byte (Santiago De la Cruz)

 * Don't get pointer of inexistent index when data is empty binding []byte (Santiago De la Cruz)

 * XML datatype support (Santiago De la Cruz)

 * DCFLOAT SUPPORT FOR WINDOWS AND NON WINDOWS (Ravuri Sai Ram Akhil)

 * Send size of LONG VARCHAR (Santiago De la Cruz)

 * added fmt package (Santiago De la Cruz)

 * Avoid register the driver when panic (Santiago De la Cruz)

 * Update bug_report.md (Santiago De la Cruz)

 * Update issue templates (Akhil Ravuri)

 * Connection attribute link added (Ravuri Sai Ram Akhil)

 * Connection attribute link added and create db options are defined. (Ravuri Sai Ram Akhil)

 * comment on how to use options in createDb api (Ravuri Sai Ram Akhil)

 * Documentation for admin operations (Ravuri Sai Ram Akhil)

 * Create database and Drop database api support (Ravuri Sai Ram Akhil)


2019-10-17, Version 0.2.0
=========================

 * changed query length from SQLINTEGER to SQLNTS (Ravuri Sai Ram Akhil)

 * sql_int added in prepare (Ravuri Sai Ram Akhil)

 * sqlnts added (Ravuri Sai Ram Akhil)

 * added sql_nts in prepare (Ravuri Sai Ram Akhil)

 * Boolean support for latest server (Ravuri Sai Ram Akhil)

 * input and output parametr is added for linux (Ravuri Sai Ram Akhil)

 * main  connection string changed (Ravuri Sai Ram Akhil)

 * sp InOut support and array insert support (Ravuri Sai Ram Akhil)


2019-09-09, Version 0.1.1
=========================

 * readme, null terminator modified and 2GB spported (Ravuri Sai Ram Akhil)

 * readme updated (Ravuri Sai Ram Akhil)

 * Removed extra bytes stored for char in sql.Out (Ravuri Sai Ram Akhil)

 * readme updated for linux2 (Ravuri Sai Ram Akhil)

 * Support for OUT type in Stored Procedure (Ravuri Sai Ram Akhil)

 * Now Clob supports 2GB (Ravuri Sai Ram Akhil)

 * null terminator has been added (Ravuri Sai Ram Akhil)

 * null terminator has been removed (Ravuri Sai Ram Akhil)

 * null termination added for utf16 to fix binvalue error for an empty variable (Ravuri Sai Ram Akhil)

 * pooling code updated (Ravuri Sai Ram Akhil)


2019-05-22, Version 0.1.0
=========================

 * SQLColAttributes,HasNextResultSet has been supported (Ravuri Sai Ram Akhil)

 * CLOB datatype is supported (Ravuri Sai Ram Akhil)

 * String termination modified (Ravuri Sai Ram Akhil)

 * fix for ';' issue (Ravuri Sai Ram Akhil)


2019-05-10, Version 0.0.1
=========================



2019-05-10, Version 0.0.0
=========================

 * First release!
