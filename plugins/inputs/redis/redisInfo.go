package redis

type infoMeasurement struct {
	client *redis.Client
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	resData map[string]interface{}
}

// 生成行协议
func (m *infoMeasurement) LineProto() (io.Point, error) {
	return io.MakeMetric(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *infoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_info",
		Fields: map[string]*inputs.FieldInfo{
			"usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
			"info_latency_ms": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "active_defrag_running": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
			"redis_version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "active_defrag_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "active_defrag_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "active_defrag_key_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "active_defrag_key_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "aof_last_rewrite_time_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "aof_rewrite_in_progress": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "aof_current_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "aof_buffer_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "loading_total_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "loading_loaded_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "loading_loaded_perc": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "loading_eta_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "connected_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "connected_slaves": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "rejected_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "blocked_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "client_biggest_input_buf": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "client_longest_output_list": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "evicted_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "expired_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "latest_fork_usec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "bytes_received_per_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "bytes_sent_per_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "pubsub_channels": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "pubsub_patterns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "rdb_bgsave_in_progress": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "rdb_changes_since_last_save": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "rdb_last_bgsave_time_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "mem_fragmentation_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory_lua": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory_peak": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory_rss": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory_startup": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_memory_overhead": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "maxmemory": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "master_last_io_seconds_ago": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "master_sync_in_progress": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "master_sync_left_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "repl_backlog_histlen": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "master_repl_offset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "slave_repl_offset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_cpu_sys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_cpu_sys_children": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_cpu_user": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "used_cpu_user_children": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "keyspace_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "",
			},
	        "keyspace_misses": &inputs.FieldInfo{
				name: "keyspace_misses",
				plan: "result",
				parse: parseInt,
			},
		},
	}
}

func CollectInfoMeasurement(cli *redis.Client) *infoMeasurement {
	m := &infoMeasurement{
		client: cli,
	}

	m.getData()
	m.submit()

	return m
}

// 数据源获取数据
func (m *infoMeasurement) getData() error {
	start := time.Now()
	info, err := m.client.Info("ALL").Result()
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	latencyMs := Round(float64(elapsed)/float64(time.Millisecond), 2)

	m.resData["info_latency_ms"] = latencyMs
	m.parseInfoData(info)
}

// 解析返回结果
func (m *infoMeasurement) parseInfoData(info string) error {
	rdr := strings.NewReader(info)

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		val := strings.TrimSpace(parts[1])

		m.resData[key] = val
	}

	return nil
}

// 提交数据
func (m *infoMeasurement) submit() error {
	metricInfo := i.Info()
	for key, item := range metricInfo {
		if value, ok := i.resData[key]; ok {
			var val interface{}
			switch item.DataType {
				case inputs.Float:
					val = cast.ToFloat64(value)
				case inputs.Int:
					val = cast.ToInt64(value)
				case inputs.Bool:
					val = cast.ToBool(value)
			}
			m.fields[key] = val
		}
	}
}
