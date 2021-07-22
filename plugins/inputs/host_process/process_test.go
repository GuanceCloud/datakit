package host_process

// func TestParseField(t *testing.T) {
// 	var obj = Processes{}
// 	p, _ := pr.NewProcess(1)
// 	username, state, name, fields, message := obj.Parse(p)
// 	a := map[string]interface{}{
// 		"username": username,
// 		"state":    state,
// 		"name":     name,
// 		"fields":   fields,
// 		"message":  message,
// 	}
// 	b, _ := json.MarshalIndent(a, "", "  ")
// 	fmt.Println(string(b))
// }

// func TestProcessesRun(t *testing.T) {
// 	var obj = Processes{
// 		ProcessName:    []string{"zsh"},
// 		ObjectInterval: datakit.Duration{Duration: 5 * time.Minute},
// 		RunTime:        datakit.Duration{Duration: 10 * time.Minute},
// 		OpenMetric:     false,
// 		re:             ".*Google",
// 	}
// 	obj.WriteObject()
// }
