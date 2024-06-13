**bpf_net_l4_log**:

tx_seq_pos: 10000
rx_seq_pos: 30000

|No. | txrx | seq | ack| payload | flag |
| -| -| - | -| - | - |
| 1 | *rx* | 1 |  901 |100 | psh,ack |
| 2 | *rx* | 101  | 901 | 100 | psh,ack|
| 3 | tx | 901 | 201 | 0 | ack |
| 4 | tx | 901 | 201 | 60 | psh,ack|
| 5 | *rx* | 201 | 961 |  0 | ack|

**bpf_net_l7_log**

direction = incoming
rx_seq = 30101
tx_seq = 10901

method=POST
path=/path
status_code=200

**公式**：

对于 No.2

l4_current_rx_seq = rx_seq_pos + seq(1)  // == 30101
l4_current_rx_seq_end= l4_current_rx_seq + payload(100)  // == 30201

l4_current_rx_seq **<=** rx_seq(l7) **<** l4_current_rx_seq_end

由于 incoming，这一条 http request : POST /path

对于 No. 4

l4_current_tx_seq = tx_seq_pos + seq(901)  // == 10901
l4_current_tx_seq_end= l4_current_tx_seq + payload(60)  // == 10961

l4_current_tx_seq **<=** tx_seq(l7) **<** l4_current_tx_seq_end

这一条填充 http response ： status 200
