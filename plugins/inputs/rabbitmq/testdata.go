package rabbitmq

var (
	exchangeHandleData = `
[
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "message_stats": {
            "publish_in": 28,
            "publish_in_details": {
                "rate": 0.0
            },
            "publish_out": 28,
            "publish_out_details": {
                "rate": 0.0
            }
        },
        "name": "",
        "type": "direct",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "name": "amq.direct",
        "type": "direct",
        "vhost": "/"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "name": "amq.fanout",
        "type": "fanout",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "name": "amq.headers",
        "type": "headers",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "name": "amq.match",
        "type": "headers",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": true,
        "name": "amq.rabbitmq.log",
        "type": "topic",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": true,
        "name": "amq.rabbitmq.trace",
        "type": "topic",
        "vhost": "hjj"
    },
    {
        "arguments": {},
        "auto_delete": false,
        "durable": true,
        "internal": false,
        "message_stats": {
            "publish_in": 20,
            "publish_in_details": {
                "rate": 0.0
            }
        },
        "name": "amq.topic",
        "type": "topic",
        "vhost": "hjj"
    }
]`

	overviewHandleData = `
{
    "cluster_name": "rabbit@tan-ThinkPad-E450",
    "contexts": [
        {
            "description": "RabbitMQ Management",
            "node": "rabbit@tan-ThinkPad-E450",
            "path": "/",
            "port": "15672",
            "ssl_opts": []
        }
    ],
    "erlang_full_version": "Erlang/OTP 20 [erts-9.2] [source] [64-bit] [smp:4:4] [ds:4:4:10] [async-threads:64] [kernel-poll:true]",
    "erlang_version": "20.2.2",
    "exchange_types": [
        {
            "description": "AMQP fanout exchange, as per the AMQP specification",
            "enabled": true,
            "name": "fanout"
        },
        {
            "description": "AMQP direct exchange, as per the AMQP specification",
            "enabled": true,
            "name": "direct"
        },
        {
            "description": "AMQP headers exchange, as per the AMQP specification",
            "enabled": true,
            "name": "headers"
        },
        {
            "description": "AMQP topic exchange, as per the AMQP specification",
            "enabled": true,
            "name": "topic"
        }
    ],
    "listeners": [
        {
            "ip_address": "::",
            "node": "rabbit@tan-ThinkPad-E450",
            "port": 5672,
            "protocol": "amqp",
            "socket_opts": {
                "backlog": 128,
                "exit_on_close": false,
                "linger": [
                    true,
                    0
                ],
                "nodelay": true
            }
        },
        {
            "ip_address": "::",
            "node": "rabbit@tan-ThinkPad-E450",
            "port": 25672,
            "protocol": "clustering",
            "socket_opts": []
        },
        {
            "ip_address": "::",
            "node": "rabbit@tan-ThinkPad-E450",
            "port": 15672,
            "protocol": "http",
            "socket_opts": {
                "port": 15672
            }
        }
    ],
    "management_version": "3.6.10",
    "message_stats": {
        "ack": 0,
        "ack_details": {
            "rate": 0.0
        },
        "confirm": 48,
        "confirm_details": {
            "rate": 0.0
        },
        "deliver": 0,
        "deliver_details": {
            "rate": 0.0
        },
        "deliver_get": 7,
        "deliver_get_details": {
            "rate": 0.0
        },
        "deliver_no_ack": 0,
        "deliver_no_ack_details": {
            "rate": 0.0
        },
        "disk_reads": 0,
        "disk_reads_details": {
            "rate": 0.0
        },
        "disk_writes": 12,
        "disk_writes_details": {
            "rate": 0.0
        },
        "get": 7,
        "get_details": {
            "rate": 0.0
        },
        "get_no_ack": 0,
        "get_no_ack_details": {
            "rate": 0.0
        },
        "publish": 48,
        "publish_details": {
            "rate": 0.0
        },
        "redeliver": 4,
        "redeliver_details": {
            "rate": 0.0
        },
        "return_unroutable": 20,
        "return_unroutable_details": {
            "rate": 0.0
        }
    },
    "node": "rabbit@tan-ThinkPad-E450",
    "object_totals": {
        "channels": 0,
        "connections": 0,
        "consumers": 0,
        "exchanges": 8,
        "queues": 1
    },
    "queue_totals": {
        "messages": 28,
        "messages_details": {
            "rate": 0.0
        },
        "messages_ready": 28,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        }
    },
    "rabbitmq_version": "3.6.10",
    "rates_mode": "basic",
    "statistics_db_event_queue": 0
}`

	queueHandleData = `
[
    {
        "arguments": {},
        "auto_delete": false,
        "backing_queue_status": {
            "avg_ack_egress_rate": 0.0,
            "avg_ack_ingress_rate": 0.0,
            "avg_egress_rate": 0.0,
            "avg_ingress_rate": 0.0,
            "delta": [
                "delta",
                "undefined",
                0,
                0,
                "undefined"
            ],
            "len": 28,
            "mode": "default",
            "next_seq_id": 28,
            "q1": 0,
            "q2": 0,
            "q3": 0,
            "q4": 28,
            "target_ram_count": "infinity"
        },
        "consumer_utilisation": null,
        "consumers": 0,
        "durable": true,
        "exclusive": false,
        "exclusive_consumer_tag": null,
        "garbage_collection": {
            "fullsweep_after": 65535,
            "max_heap_size": 0,
            "min_bin_vheap_size": 46422,
            "min_heap_size": 233,
            "minor_gcs": 427
        },
        "head_message_timestamp": null,
        "memory": 143144,
        "message_bytes": 136,
        "message_bytes_paged_out": 0,
        "message_bytes_persistent": 72,
        "message_bytes_ram": 136,
        "message_bytes_ready": 136,
        "message_bytes_unacknowledged": 0,
        "message_stats": {
            "ack": 0,
            "ack_details": {
                "rate": 0.0
            },
            "deliver": 0,
            "deliver_details": {
                "rate": 0.0
            },
            "deliver_get": 7,
            "deliver_get_details": {
                "rate": 0.0
            },
            "deliver_no_ack": 0,
            "deliver_no_ack_details": {
                "rate": 0.0
            },
            "get": 7,
            "get_details": {
                "rate": 0.0
            },
            "get_no_ack": 0,
            "get_no_ack_details": {
                "rate": 0.0
            },
            "publish": 28,
            "publish_details": {
                "rate": 0.0
            },
            "redeliver": 4,
            "redeliver_details": {
                "rate": 0.0
            }
        },
        "messages": 28,
        "messages_details": {
            "rate": 0.0
        },
        "messages_paged_out": 0,
        "messages_persistent": 12,
        "messages_ram": 28,
        "messages_ready": 28,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_ready_ram": 28,
        "messages_unacknowledged": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        },
        "messages_unacknowledged_ram": 0,
        "name": "testhjj",
        "node": "rabbit@tan-ThinkPad-E450",
        "policy": null,
        "recoverable_slaves": null,
        "reductions": 1424483,
        "reductions_details": {
            "rate": 0.0
        },
        "state": "running",
        "vhost": "hjj"
    }
]
`

	nodeHandleData = `
[
    {
        "applications": [
            {
                "description": "RabbitMQ AMQP Client",
                "name": "amqp_client",
                "version": "3.6.10"
            },
            {
                "description": "The Erlang ASN1 compiler version 5.0.4",
                "name": "asn1",
                "version": "5.0.4"
            },
            {
                "description": "ERTS  CXC 138 10",
                "name": "compiler",
                "version": "7.1.4"
            },
            {
                "description": "Small, fast, modular HTTP server.",
                "name": "cowboy",
                "version": "1.0.4"
            },
            {
                "description": "Support library for manipulating Web protocols.",
                "name": "cowlib",
                "version": "1.0.2"
            },
            {
                "description": "CRYPTO",
                "name": "crypto",
                "version": "4.2"
            },
            {
                "description": "INETS  CXC 138 49",
                "name": "inets",
                "version": "6.4.5"
            },
            {
                "description": "ERTS  CXC 138 10",
                "name": "kernel",
                "version": "5.4.1"
            },
            {
                "description": "MNESIA  CXC 138 12",
                "name": "mnesia",
                "version": "4.15.3"
            },
            {
                "description": "CPO  CXC 138 46",
                "name": "os_mon",
                "version": "2.4.4"
            },
            {
                "description": "Public key infrastructure",
                "name": "public_key",
                "version": "1.5.2"
            },
            {
                "description": "RabbitMQ",
                "name": "rabbit",
                "version": "3.6.10"
            },
            {
                "description": "Modules shared by rabbitmq-server and rabbitmq-erlang-client",
                "name": "rabbit_common",
                "version": "3.6.10"
            },
            {
                "description": "RabbitMQ Management Console",
                "name": "rabbitmq_management",
                "version": "3.6.10"
            },
            {
                "description": "RabbitMQ Management Agent",
                "name": "rabbitmq_management_agent",
                "version": "3.6.10"
            },
            {
                "description": "RabbitMQ Web Dispatcher",
                "name": "rabbitmq_web_dispatch",
                "version": "3.6.10"
            },
            {
                "description": "Socket acceptor pool for TCP protocols.",
                "name": "ranch",
                "version": "1.3.0"
            },
            {
                "description": "SASL  CXC 138 11",
                "name": "sasl",
                "version": "3.1.1"
            },
            {
                "description": "Erlang/OTP SSL application",
                "name": "ssl",
                "version": "8.2.3"
            },
            {
                "description": "ERTS  CXC 138 10",
                "name": "stdlib",
                "version": "3.4.3"
            },
            {
                "description": "Syntax tools",
                "name": "syntax_tools",
                "version": "2.1.4"
            },
            {
                "description": "XML parser",
                "name": "xmerl",
                "version": "1.3.16"
            }
        ],
        "auth_mechanisms": [
            {
                "description": "SASL PLAIN authentication mechanism",
                "enabled": true,
                "name": "PLAIN"
            },
            {
                "description": "QPid AMQPLAIN mechanism",
                "enabled": true,
                "name": "AMQPLAIN"
            },
            {
                "description": "RabbitMQ Demo challenge-response authentication mechanism",
                "enabled": false,
                "name": "RABBIT-CR-DEMO"
            }
        ],
        "cluster_links": [],
        "config_files": [
            "/etc/rabbitmq/rabbitmq.config (not found)"
        ],
        "context_switches": 13660854,
        "context_switches_details": {
            "rate": 19.0
        },
        "contexts": [
            {
                "description": "RabbitMQ Management",
                "path": "/",
                "port": "15672"
            }
        ],
        "db_dir": "/var/lib/rabbitmq/mnesia/rabbit@tan-ThinkPad-E450",
        "disk_free": 570389422080,
        "disk_free_alarm": false,
        "disk_free_details": {
            "rate": 0.0
        },
        "disk_free_limit": 50000000,
        "enabled_plugins": [
            "rabbitmq_management"
        ],
        "exchange_types": [
            {
                "description": "AMQP fanout exchange, as per the AMQP specification",
                "enabled": true,
                "name": "fanout"
            },
            {
                "description": "AMQP direct exchange, as per the AMQP specification",
                "enabled": true,
                "name": "direct"
            },
            {
                "description": "AMQP headers exchange, as per the AMQP specification",
                "enabled": true,
                "name": "headers"
            },
            {
                "description": "AMQP topic exchange, as per the AMQP specification",
                "enabled": true,
                "name": "topic"
            }
        ],
        "fd_total": 65536,
        "fd_used": 26,
        "fd_used_details": {
            "rate": -0.4
        },
        "gc_bytes_reclaimed": 76880682016,
        "gc_bytes_reclaimed_details": {
            "rate": 128436.8
        },
        "gc_num": 2950191,
        "gc_num_details": {
            "rate": 4.6
        },
        "io_file_handle_open_attempt_avg_time": 0.05287272727272727,
        "io_file_handle_open_attempt_avg_time_details": {
            "rate": 0.0
        },
        "io_file_handle_open_attempt_count": 55,
        "io_file_handle_open_attempt_count_details": {
            "rate": 0.0
        },
        "io_read_avg_time": 19.055,
        "io_read_avg_time_details": {
            "rate": 0.0
        },
        "io_read_bytes": 1,
        "io_read_bytes_details": {
            "rate": 0.0
        },
        "io_read_count": 1,
        "io_read_count_details": {
            "rate": 0.0
        },
        "io_reopen_count": 0,
        "io_reopen_count_details": {
            "rate": 0.0
        },
        "io_seek_avg_time": 0.09627272727272726,
        "io_seek_avg_time_details": {
            "rate": 0.0
        },
        "io_seek_count": 11,
        "io_seek_count_details": {
            "rate": 0.0
        },
        "io_sync_avg_time": 50.0522,
        "io_sync_avg_time_details": {
            "rate": 0.0
        },
        "io_sync_count": 20,
        "io_sync_count_details": {
            "rate": 0.0
        },
        "io_write_avg_time": 1.0112,
        "io_write_avg_time_details": {
            "rate": 0.0
        },
        "io_write_bytes": 9078,
        "io_write_bytes_details": {
            "rate": 0.0
        },
        "io_write_count": 20,
        "io_write_count_details": {
            "rate": 0.0
        },
        "log_file": "/var/log/rabbitmq/rabbit@tan-ThinkPad-E450.log",
        "mem_alarm": false,
        "mem_limit": 3302432768,
        "mem_used": 66322072,
        "mem_used_details": {
            "rate": -39776.0
        },
        "metrics_gc_queue_length": {
            "channel_closed": 0,
            "channel_consumer_deleted": 0,
            "connection_closed": 0,
            "consumer_deleted": 0,
            "exchange_deleted": 0,
            "node_node_deleted": 0,
            "queue_deleted": 0,
            "vhost_deleted": 0
        },
        "mnesia_disk_tx_count": 10,
        "mnesia_disk_tx_count_details": {
            "rate": 0.0
        },
        "mnesia_ram_tx_count": 544,
        "mnesia_ram_tx_count_details": {
            "rate": 0.0
        },
        "msg_store_read_count": 0,
        "msg_store_read_count_details": {
            "rate": 0.0
        },
        "msg_store_write_count": 0,
        "msg_store_write_count_details": {
            "rate": 0.0
        },
        "name": "rabbit@tan-ThinkPad-E450",
        "net_ticktime": 60,
        "os_pid": "9160",
        "partitions": [],
        "proc_total": 1048576,
        "proc_used": 334,
        "proc_used_details": {
            "rate": -0.4
        },
        "processors": 4,
        "queue_index_journal_write_count": 15,
        "queue_index_journal_write_count_details": {
            "rate": 0.0
        },
        "queue_index_read_count": 0,
        "queue_index_read_count_details": {
            "rate": 0.0
        },
        "queue_index_write_count": 5,
        "queue_index_write_count_details": {
            "rate": 0.0
        },
        "rates_mode": "basic",
        "run_queue": 0,
        "running": true,
        "sasl_log_file": "/var/log/rabbitmq/rabbit@tan-ThinkPad-E450-sasl.log",
        "sockets_total": 58890,
        "sockets_used": 0,
        "sockets_used_details": {
            "rate": 0.0
        },
        "type": "disc",
        "uptime": 595874646
    }
]`

	bindingHandleData = `
[
    {
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
    }
]
`
)
