






rabbitmq 采集器调研

采集指标如下：


- rabbitmq_overview （通过 API `/api/overview` 采集 ）

  | fields        | type   | unit     |description | 
  | :----:        | :----: | :----:   |  :----: | 
  |   object_totals_channels      |   int  |   count  | Total number of channels | 
  |   object_totals_connections      |   int  |   count  | Total number of connections | 
  |   object_totals_consumers      |   int  |   count  | Total number of consumers | 
  |   object_totals_queues      |   int  |   count  | Total number of queues | 
  |   message_ack_count      |   int  |   count  | Number of messages delivered to clients and acknowledged | 
  |   message_ack_rate      |   float  |   rate  | Rate of messages delivered to clients and acknowledged per second | 
  |   message_deliver_get_count     |   int  |   count  | Sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | 
  |   message_deliver_get_rate     |   float  |   rate  | Rate per second of the sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | 
  |   message_publish_count     |   int  |   count  | Count of messages published | 
  |   message_publish_rate     |   float  |   rate | Rate of messages published per second | 
  |   message_publish_in_rate     |   float  |   rate | Rate of messages published from channels into this overview per sec | 
  |   message_publish_in_count     |   int  |   count | Count of messages published from channels into this overview | 
  |   message_publish_out_count     |   int  |   count | Count of messages published from this overview into queues| 
  |   message_publish_out_rate     |   float  |   rate | Rate of messages published from this overview into queues per second | 
  |   message_redeliver_count     |   int  |   count | Count of subset of messages in deliver_get which had the redelivered flag set | 
  |   message_redeliver_rate     |   float  |   rate | Rate of subset of messages in deliver_get which had the redelivered flag set per second | 
  |   message_return_unroutable_count_rate     |   float  |   rate | Rate of messages returned to publisher as unroutable per second | 
  |   message_return_unroutable_count    |   int  |   count | Count of messages returned to publisher as unroutable | 
  |   queue_totals_messages_count    |   int  |   count | Total number of messages (ready plus unacknowledged) | 
  |   queue_totals_messages_rate    |   float  |   rate | Total rate of messages (ready plus unacknowledged) | 
  |   queue_totals_messages_ready    |   int  |   count | Number of messages ready for delivery | 
  |   queue_totals_messages_ready_rate    |   int  |   count | Rate of number of messages ready for delivery| 
  |   queue_totals_messages_unacknowledged   |   int  |   count | Number of unacknowledged messages | 
  |   queue_totals_messages_unacknowledged_rate   |   int  |   count | Rate of number of unacknowledged messages | 

- rabbitmq_exchange （通过 API `/api/exchanges` 采集 ）

  | fields        | type   | unit     |description | 
  | :----:        | :----: | :----:   |  :----: | 
  |   messages_ack_count      |   int  |   count  | Number of messages in exchanges delivered to clients and acknowledged | 
  |   messages_ack_rate      |   float  |   rate  | Rate of messages in exchanges delivered to clients and acknowledged per second | 
  |   message_deliver_get_count     |   int  |   count  | Sum of messages in exchanges delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | 
  |   message_deliver_get_rate     |   float  |   rate  | Rate per second of the sum of exchange messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get |
  |   message_publish_count     |   int  |   count  | Count of messages in exchanges published | 
  |   message_publish_rate     |   float  |   rate | Rate of messages in exchanges published per second | 
  |   message_publish_in_rate     |   float  |   rate | Rate of messages published from channels into this exchange per sec | 
  |   message_publish_in_count     |   int  |   count | Count of messages published from channels into this exchange|
  |   message_publish_out_count     |   int  |   count | Count of messages published from this exchange into queues｜
  |   message_publish_out_rate     |   float  |   rate | Rate of messages published from this exchange into queues per second | 
  |   message_redeliver_count     |   int  |   count | Count of subset of messages in exchanges in deliver_get which had the redelivered flag set | 
  |   message_redeliver_rate     |   float  |   rate | Rate of subset of messages in exchanges in deliver_get which had the redelivered flag set per second| 
  |   message_return_unroutable_count_rate     |   float  |   rate | Rate of messages in exchanges returned to publisher as unroutable per second | 
  |   message_return_unroutable_count    |   int  |   count | Count of messages in exchanges returned to publisher as unroutable| 
  
- rabbitmq_node （通过 API `/api/nodes` 采集 ）

  | fields        | type   | unit     |description | 
  | :----:        | :----: | :----:   |  :----: | 
  |   disk_free_alarm      |   bool  |   -  | Does the node have disk alarm｜
  |   disk_free      |   int  |   bytes  | Current free disk space｜
  |   fd_used      |   int  |   -  | Used file descriptors｜
  |   mem_alarm      |   bool  |   -  | Does the host has memory alarm｜
  |   mem_limit      |   int  |   bytes  | Memory usage high watermark in bytes｜
  |   mem_used       |   int  |   bytes  | Memory used in bytes｜
  |   run_queue       |   int  |   count  | Average number of Erlang processes waiting to run｜
  |   running       |   bool  |   -  | Is the node running or not｜
  |   sockets_used       |   int  |   count  | Number of file descriptors used as sockets｜
  
- rabbitmq_queue （通过 API `/api/queues` 采集 ）

  | fields        | type   | unit     |description | 
  | :----:        | :----: | :----:   |  :----: | 
  |   consumer_utilisation     |   float  |   rate  | The ratio of time that a queue's consumers can take new messages｜
  |   consumers     |   float  |   rate  | Number of consumers｜
  |   memory     |   int  |   bytes  | Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures｜
    |   messages_ack_count      |   int  |   count  | Number of messages in queues delivered to clients and acknowledged | 
    |   messages_ack_rate      |   float  |   rate  | Number per second of messages delivered to clients and acknowledged| 
    |   message_deliver_get_count     |   int  |   count  | Count of messages delivered in acknowledgement mode to consumers | 
    |   message_deliver_get_rate     |   float  |   rate  | Rate of messages delivered in acknowledgement mode to consumers |
    |   message_publish_count     |   int  |   count  | Count of messages in queues published| 
    |   message_publish_rate     |   float  |   rate | Rate per second of messages published | 
    |   message_redeliver_count     |   int  |   count | Count of subset of messages in queues in deliver_get which had the redelivered flag set| 
    |   message_redeliver_rate     |   float  |   rate | Rate per second of subset of messages in deliver_get which had the redelivered flag set| 
    