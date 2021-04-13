package mysql

const (
	configSample = `
[[inputs.mysql]]
    host = "localhost"
    user = "datakit"
    pass = "<PASS>"
    port = 3306
    # sock = "<SOCK>"
    # charset = "utf8"

    ## @param connect_timeout - number - optional - default: 10
    # connect_timeout = 10

   	## @param service - string - optional
    # service = "<SERVICE>"

    interval = "10s"

    # [[inputs.mysql.custom_queries]]
    #     sql = "SELECT foo, COUNT(*) FROM table.events GROUP BY foo"
    #     metric = "xxxx"
    #     tagKeys = ["column1", "column1"]
    #     fieldKeys = ["column3", "column1"]

    [inputs.mysql.tags]
    # tag1 = val1
    # tag2 = val2
`
)
