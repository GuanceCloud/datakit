package config

// config.Cfg.Binlog = &config.BinlogConfig{}

// var datasource config.BinlogDatasource

// datasource.Addr = ""
// datasource.User = ""
// datasource.Pwd = ""
// datasource.Password = ""
// // if datasource.pwd != "" {
// // 	datasource.Password = XorEncode(datasource.pwd)
// // }

// datasource.ServerID = uint32(rand.New(rand.NewSource(time.Now().Unix())).Intn(1000)) + 1001

// tb1 := &config.BinlogTable{
// 	Table:       "table1",
// 	Measurement: "table1",
// 	Columns: map[string]string{
// 		"tag1":   "tag",
// 		"field1": "field",
// 	},
// 	ExcludeListenEvents: []string{"delete"},
// }

// tb2 := &config.BinlogTable{
// 	Table:       "table2",
// 	Measurement: "table2",
// 	//Tags:        []string{"tag1", "tag2"},
// 	//Fields:      []string{"field1", "field2"},
// 	Columns: map[string]string{
// 		"tag1":   "tag",
// 		"field1": "field",
// 	},
// }

// input := &config.BinlogInput{
// 	Database:      "test",
// 	Tables:        []*config.BinlogTable{tb1, tb2},
// 	ExcludeTables: []string{"table3"},
// }

// datasource.Inputs = []*config.BinlogInput{input}

// config.Cfg.Binlog.Datasources = append(config.Cfg.Binlog.Datasources, &datasource)

// outdt, _ := yaml.Marshal(&Cfg)
// ioutil.WriteFile("aa.cfg", outdt, 0664)
