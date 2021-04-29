package rabbitmq

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"context"
	"github.com/gorilla/mux"
)

func TestGetMetric(t *testing.T) {
	r := mux.NewRouter()
	srv := &http.Server{Addr: ":8888", Handler: r}

	r.HandleFunc("/api/nodes", nodeHandle)
	r.HandleFunc("/api/exchanges", exchangeHandle)
	r.HandleFunc("/api/overview", overviewHandle)
	r.HandleFunc("/api/queues", queueHandle)
	r.HandleFunc("/api/queues/{vhost}/{name}/bindings", bindingHandle)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	time.Sleep(time.Second)
	n := &Input{
		Url: "http://0.0.0.0:8888",
	}
	cli, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err)
	}
	n.client = cli

	n.getMetric()
	for _, v := range collectCache {
		fmt.Println(v.LineProto())
	}
	srv.Shutdown(context.Background())
}

func nodeHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `[{"partitions":[],"os_pid":"9160","fd_total":65536,"sockets_total":58890,"mem_limit":3302432768,"mem_alarm":false,"disk_free_limit":50000000,"disk_free_alarm":false,"proc_total":1048576,"rates_mode":"basic","uptime":595874646,"run_queue":0,"processors":4,"exchange_types":[{"name":"fanout","description":"AMQP fanout exchange, as per the AMQP specification","enabled":true},{"name":"direct","description":"AMQP direct exchange, as per the AMQP specification","enabled":true},{"name":"headers","description":"AMQP headers exchange, as per the AMQP specification","enabled":true},{"name":"topic","description":"AMQP topic exchange, as per the AMQP specification","enabled":true}],"auth_mechanisms":[{"name":"PLAIN","description":"SASL PLAIN authentication mechanism","enabled":true},{"name":"AMQPLAIN","description":"QPid AMQPLAIN mechanism","enabled":true},{"name":"RABBIT-CR-DEMO","description":"RabbitMQ Demo challenge-response authentication mechanism","enabled":false}],"applications":[{"name":"amqp_client","description":"RabbitMQ AMQP Client","version":"3.6.10"},{"name":"asn1","description":"The Erlang ASN1 compiler version 5.0.4","version":"5.0.4"},{"name":"compiler","description":"ERTS  CXC 138 10","version":"7.1.4"},{"name":"cowboy","description":"Small, fast, modular HTTP server.","version":"1.0.4"},{"name":"cowlib","description":"Support library for manipulating Web protocols.","version":"1.0.2"},{"name":"crypto","description":"CRYPTO","version":"4.2"},{"name":"inets","description":"INETS  CXC 138 49","version":"6.4.5"},{"name":"kernel","description":"ERTS  CXC 138 10","version":"5.4.1"},{"name":"mnesia","description":"MNESIA  CXC 138 12","version":"4.15.3"},{"name":"os_mon","description":"CPO  CXC 138 46","version":"2.4.4"},{"name":"public_key","description":"Public key infrastructure","version":"1.5.2"},{"name":"rabbit","description":"RabbitMQ","version":"3.6.10"},{"name":"rabbit_common","description":"Modules shared by rabbitmq-server and rabbitmq-erlang-client","version":"3.6.10"},{"name":"rabbitmq_management","description":"RabbitMQ Management Console","version":"3.6.10"},{"name":"rabbitmq_management_agent","description":"RabbitMQ Management Agent","version":"3.6.10"},{"name":"rabbitmq_web_dispatch","description":"RabbitMQ Web Dispatcher","version":"3.6.10"},{"name":"ranch","description":"Socket acceptor pool for TCP protocols.","version":"1.3.0"},{"name":"sasl","description":"SASL  CXC 138 11","version":"3.1.1"},{"name":"ssl","description":"Erlang/OTP SSL application","version":"8.2.3"},{"name":"stdlib","description":"ERTS  CXC 138 10","version":"3.4.3"},{"name":"syntax_tools","description":"Syntax tools","version":"2.1.4"},{"name":"xmerl","description":"XML parser","version":"1.3.16"}],"contexts":[{"description":"RabbitMQ Management","path":"/","port":"15672"}],"log_file":"/var/log/rabbitmq/rabbit@tan-ThinkPad-E450.log","sasl_log_file":"/var/log/rabbitmq/rabbit@tan-ThinkPad-E450-sasl.log","db_dir":"/var/lib/rabbitmq/mnesia/rabbit@tan-ThinkPad-E450","config_files":["/etc/rabbitmq/rabbitmq.config (not found)"],"net_ticktime":60,"enabled_plugins":["rabbitmq_management"],"name":"rabbit@tan-ThinkPad-E450","type":"disc","running":true,"mem_used":66322072,"mem_used_details":{"rate":-39776.0},"fd_used":26,"fd_used_details":{"rate":-0.4},"sockets_used":0,"sockets_used_details":{"rate":0.0},"proc_used":334,"proc_used_details":{"rate":-0.4},"disk_free":570389422080,"disk_free_details":{"rate":0.0},"gc_num":2950191,"gc_num_details":{"rate":4.6},"gc_bytes_reclaimed":76880682016,"gc_bytes_reclaimed_details":{"rate":128436.8},"context_switches":13660854,"context_switches_details":{"rate":19.0},"io_read_count":1,"io_read_count_details":{"rate":0.0},"io_read_bytes":1,"io_read_bytes_details":{"rate":0.0},"io_read_avg_time":19.055,"io_read_avg_time_details":{"rate":0.0},"io_write_count":20,"io_write_count_details":{"rate":0.0},"io_write_bytes":9078,"io_write_bytes_details":{"rate":0.0},"io_write_avg_time":1.0112,"io_write_avg_time_details":{"rate":0.0},"io_sync_count":20,"io_sync_count_details":{"rate":0.0},"io_sync_avg_time":50.0522,"io_sync_avg_time_details":{"rate":0.0},"io_seek_count":11,"io_seek_count_details":{"rate":0.0},"io_seek_avg_time":0.09627272727272726,"io_seek_avg_time_details":{"rate":0.0},"io_reopen_count":0,"io_reopen_count_details":{"rate":0.0},"mnesia_ram_tx_count":544,"mnesia_ram_tx_count_details":{"rate":0.0},"mnesia_disk_tx_count":10,"mnesia_disk_tx_count_details":{"rate":0.0},"msg_store_read_count":0,"msg_store_read_count_details":{"rate":0.0},"msg_store_write_count":0,"msg_store_write_count_details":{"rate":0.0},"queue_index_journal_write_count":15,"queue_index_journal_write_count_details":{"rate":0.0},"queue_index_write_count":5,"queue_index_write_count_details":{"rate":0.0},"queue_index_read_count":0,"queue_index_read_count_details":{"rate":0.0},"io_file_handle_open_attempt_count":55,"io_file_handle_open_attempt_count_details":{"rate":0.0},"io_file_handle_open_attempt_avg_time":0.05287272727272727,"io_file_handle_open_attempt_avg_time_details":{"rate":0.0},"cluster_links":[],"metrics_gc_queue_length":{"connection_closed":0,"channel_closed":0,"consumer_deleted":0,"exchange_deleted":0,"queue_deleted":0,"vhost_deleted":0,"node_node_deleted":0,"channel_consumer_deleted":0}}]`
	w.Write([]byte(resp))
}

func queueHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `[{"messages_details":{"rate":0.0},"messages":28,"messages_unacknowledged_details":{"rate":0.0},"messages_unacknowledged":0,"messages_ready_details":{"rate":0.0},"messages_ready":28,"reductions_details":{"rate":0.0},"reductions":1424483,"message_stats":{"deliver_get_details":{"rate":0.0},"deliver_get":7,"ack_details":{"rate":0.0},"ack":0,"redeliver_details":{"rate":0.0},"redeliver":4,"deliver_no_ack_details":{"rate":0.0},"deliver_no_ack":0,"deliver_details":{"rate":0.0},"deliver":0,"get_no_ack_details":{"rate":0.0},"get_no_ack":0,"get_details":{"rate":0.0},"get":7,"publish_details":{"rate":0.0},"publish":28},"node":"rabbit@tan-ThinkPad-E450","arguments":{},"exclusive":false,"auto_delete":false,"durable":true,"vhost":"hjj","name":"testhjj","message_bytes_paged_out":0,"messages_paged_out":0,"backing_queue_status":{"mode":"default","q1":0,"q2":0,"delta":["delta","undefined",0,0,"undefined"],"q3":0,"q4":28,"len":28,"target_ram_count":"infinity","next_seq_id":28,"avg_ingress_rate":0.0,"avg_egress_rate":0.0,"avg_ack_ingress_rate":0.0,"avg_ack_egress_rate":0.0},"head_message_timestamp":null,"message_bytes_persistent":72,"message_bytes_ram":136,"message_bytes_unacknowledged":0,"message_bytes_ready":136,"message_bytes":136,"messages_persistent":12,"messages_unacknowledged_ram":0,"messages_ready_ram":28,"messages_ram":28,"garbage_collection":{"minor_gcs":427,"fullsweep_after":65535,"min_heap_size":233,"min_bin_vheap_size":46422,"max_heap_size":0},"state":"running","recoverable_slaves":null,"memory":143144,"consumer_utilisation":null,"consumers":0,"exclusive_consumer_tag":null,"policy":null}]`
	w.Write([]byte(resp))
}

func overviewHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `{"management_version":"3.6.10","rates_mode":"basic","exchange_types":[{"name":"fanout","description":"AMQP fanout exchange, as per the AMQP specification","enabled":true},{"name":"direct","description":"AMQP direct exchange, as per the AMQP specification","enabled":true},{"name":"headers","description":"AMQP headers exchange, as per the AMQP specification","enabled":true},{"name":"topic","description":"AMQP topic exchange, as per the AMQP specification","enabled":true}],"rabbitmq_version":"3.6.10","cluster_name":"rabbit@tan-ThinkPad-E450","erlang_version":"20.2.2","erlang_full_version":"Erlang/OTP 20 [erts-9.2] [source] [64-bit] [smp:4:4] [ds:4:4:10] [async-threads:64] [kernel-poll:true]","message_stats":{"publish":48,"publish_details":{"rate":0.0},"confirm":48,"confirm_details":{"rate":0.0},"return_unroutable":20,"return_unroutable_details":{"rate":0.0},"disk_reads":0,"disk_reads_details":{"rate":0.0},"disk_writes":12,"disk_writes_details":{"rate":0.0},"get":7,"get_details":{"rate":0.0},"get_no_ack":0,"get_no_ack_details":{"rate":0.0},"deliver":0,"deliver_details":{"rate":0.0},"deliver_no_ack":0,"deliver_no_ack_details":{"rate":0.0},"redeliver":4,"redeliver_details":{"rate":0.0},"ack":0,"ack_details":{"rate":0.0},"deliver_get":7,"deliver_get_details":{"rate":0.0}},"queue_totals":{"messages_ready":28,"messages_ready_details":{"rate":0.0},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0},"messages":28,"messages_details":{"rate":0.0}},"object_totals":{"consumers":0,"queues":1,"exchanges":8,"connections":0,"channels":0},"statistics_db_event_queue":0,"node":"rabbit@tan-ThinkPad-E450","listeners":[{"node":"rabbit@tan-ThinkPad-E450","protocol":"amqp","ip_address":"::","port":5672,"socket_opts":{"backlog":128,"nodelay":true,"linger":[true,0],"exit_on_close":false}},{"node":"rabbit@tan-ThinkPad-E450","protocol":"clustering","ip_address":"::","port":25672,"socket_opts":[]},{"node":"rabbit@tan-ThinkPad-E450","protocol":"http","ip_address":"::","port":15672,"socket_opts":{"port":15672}}],"contexts":[{"ssl_opts":[],"node":"rabbit@tan-ThinkPad-E450","description":"RabbitMQ Management","path":"/","port":"15672"}]}`
	w.Write([]byte(resp))
}

func exchangeHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `[{"name":"","vhost":"hjj","type":"direct","durable":true,"auto_delete":false,"internal":false,"arguments":{},"message_stats":{"publish_out":28,"publish_out_details":{"rate":0.0},"publish_in":28,"publish_in_details":{"rate":0.0}}},{"name":"amq.direct","vhost":"/","type":"direct","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.fanout","vhost":"hjj","type":"fanout","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.headers","vhost":"hjj","type":"headers","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.match","vhost":"hjj","type":"headers","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.rabbitmq.log","vhost":"hjj","type":"topic","durable":true,"auto_delete":false,"internal":true,"arguments":{}},{"name":"amq.rabbitmq.trace","vhost":"hjj","type":"topic","durable":true,"auto_delete":false,"internal":true,"arguments":{}},{"name":"amq.topic","vhost":"hjj","type":"topic","durable":true,"auto_delete":false,"internal":false,"arguments":{},"message_stats":{"publish_in":20,"publish_in_details":{"rate":0.0}}}]`
	w.Write([]byte(resp))
}

func bindingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `[{
	"source": "",
	"vhost": "/",
	"destination": "testhjj",
	"destination_type": "queue",
	"routing_key": "testhjj",
	"arguments": {},
	"properties_key": "testhjj"
	},
	{
	"source": "hjj",
	"vhost": "/",
	"destination": "testhjj",
	"destination_type": "queue",
	"routing_key": "fsaf",
	"arguments": {},
	"properties_key": "fsaf"
	}]`
	w.Write([]byte(resp))
}
