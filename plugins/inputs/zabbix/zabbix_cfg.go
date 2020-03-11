package zabbix

const (
	zabbixConfigSample = `
###
### Zabbix DB
### Select one provider by commenting [zabbix.postgres] or [zabbix.mysql]
###
#[zabbix]

  ##[zabbix.postgres]
  ## PostgreSQL:
  ##   postgres://[user[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ##   or a simple string:
  ##        host=localhost user=... password=... sslmode=... dbname=zabbix application_name=influxdb-zabbix
  ##
  ##address="postgres://zabbix:zabbix123***@192.168.1.101/zabbix?sslmode=disable"
  
#  [zabbix.mysql]
  ## MariaDB/MySQL
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/dbname[?tls=[true|false|skip-verify]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    db_user:passwd@tcp(127.0.0.1:3306)/zabbix'
#  address="zabbix:zabbix@tcp(127.0.0.1:3306)/zabbix"
  
###
### Zabbix tables 
### At least one table has to be active
###
### Controls tables for extracting data
###   name (string - mandatory) is the relation name in postgres.
###   active (boolean - mandatory) is to activate or not the data extraction.
###   startdate (string) is the starting date in yyyy-MM-ddTHH:mm:ss format. 
###       -- startdate is needed for the first load. After this, value stored in registry file prevails on this one.
###       -- example: 2016-10-01T00:00:00
###   daysperbatch (int) is the number of days to extractfrom Zabbix backend
###   hoursperbatch (int - default 360) is the number of hours to be loaded to InfluxDB 
###   interval in seconds (int - default 15) is time before each extraction poll.
###
#[tables]
#  [tables.history]
#  name="history"
#  active=false
#  startdate="2020-01-01T00:00:00"
#  hoursperbatch=720
#  interval=15
    
#  [tables.history_uint]
#  name="history_uint"
#  active=false
#  startdate="2020-01-01T00:00:00"
#  hoursperbatch=720
#  interval=15
  
#  [tables.trends]
#  name="trends"
#  active=false
#  startdate="2020-01-01T00:00:00"
#  hoursperbatch=720
#  interval=15
  
#  [tables.trends_uint]
#  name="trends_uint"
#  active=false
#  startdate="2020-01-01T00:00:00"
#  hoursperbatch=720
#  interval=15
   
###
### Registry file
### Name of the registry file. Per default, it is put in the current working directory. 
###
#[registry]
# File name 
#filename="influxdb-zabbix.json"

### End
`
)
