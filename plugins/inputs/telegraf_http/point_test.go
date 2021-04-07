package telegraf_http

import (
	"testing"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
)

func TestPointHandle(t *testing.T) {
	testcase := []string{
		`kubernetes_pod_container,container_name=deis-controller,namespace=deis,node_name=ip-10-0-0-0.ec2.internal,pod_name=deis-controller-3058870187-xazsr cpu_usage_core_nanoseconds=1072255403860672i,cpu_usage_nanocores=173497581i,logsfs_available_bytes=121128271872i,logsfs_capacity_bytes=153567944704i,logsfs_used_bytes=20787200i,memory_major_page_faults=0i,memory_page_faults=175i,memory_rss_bytes=0i,memory_usage_bytes=0i,memory_working_set_bytes=0i,rootfs_available_bytes=121128271872i,rootfs_capacity_bytes=153567944704i,rootfs_used_bytes=1110016i 1476477530000000000`,

		`kubernetes_node,host=ubuntu_test,node_name=cn-hangzhou.172.16.2.3 cpu_usage_core_nanoseconds=1861297536945863i,cpu_usage_nanocores=401704442i,fs_available_bytes=73116872704i,fs_capacity_bytes=126692048896i,fs_used_bytes=48189890560i,memory_available_bytes=1527259136i,memory_major_page_faults=7883i,memory_page_faults=100640248i,memory_rss_bytes=5999329280i,memory_usage_bytes=7411822592i,memory_working_set_bytes=6673932288i,network_rx_bytes=530490969458i,network_rx_errors=0i,network_tx_bytes=493173160592i,network_tx_errors=0i,runtime_image_fs_available_bytes=73116872704i,runtime_image_fs_capacity_bytes=126692048896i,runtime_image_fs_used_bytes=41885561550i 1614585811000000000`,

		`docker_container_mem,container_image=bitnami/etcd,container_name=etcd-server,container_status=running,container_version=latest,engine_host=ubuntu-server,host=ubuntu_test,maintainer=Bitnami\ <containers@bitnami.com>,server_version=20.10.2 total_active_file=5988352i,total_inactive_file=5627904i,usage_percent=0.10702535187524687,hierarchical_memory_limit=9223372036854771712i,pgfault=7566i,rss=6807552i,total_pgfault=7566i,total_pgmajfault=551i,unevictable=0i,inactive_anon=2220032i,mapped_file=8388608i,rss_huge=0i,total_pgpgout=14674i,total_inactive_anon=2220032i,total_rss=6807552i,total_active_anon=2134016i,max_usage=30453760i,active_anon=2134016i,pgmajfault=551i,total_cache=9162752i,total_rss_huge=0i,total_writeback=0i,writeback=0i,pgpgin=18573i,pgpgout=14674i,total_unevictable=0i,limit=8358453248i,usage=8945664i,container_id="ef22db1c35cd6b32fd139b07045335d4b40eb45f7a3889a2fafff931fc6b7b35",active_file=5988352i,cache=9162752i,inactive_file=5627904i,total_mapped_file=8388608i,total_pgpgin=18573i 1614250591000000000`,

		`docker_container_cpu,container_image=bitnami/etcd,container_name=etcd-server,container_status=running,container_version=latest,cpu=cpu-total,engine_host=ubuntu-server,host=ubuntu_test,maintainer=Bitnami\ <containers@bitnami.com>,server_version=20.10.2 container_id="ef22db1c35cd6b32fd139b07045335d4b40eb45f7a3889a2fafff931fc6b7b35",usage_percent=0.6716146633416459,throttling_periods=0,throttling_throttled_periods=0,throttling_throttled_time=0,usage_in_kernelmode=495640000000,usage_in_usermode=348640000000,usage_system=1465374460000000,usage_total=1301636763471 1614250591000000000`,

		// 非指定measurement
		`kubernetes_daemonset,daemonset_name=telegraf,selector_select1=s1,namespace=logging number_unavailable=0i,desired_number_scheduled=11i,number_available=11i,number_misscheduled=8i,number_ready=11i,updated_number_scheduled=11i,created=1527758699000000000i,generation=16i,current_number_scheduled=11i 1547597616000000000`,

		// 缺少必要的字段，取得对应的 nil 值，在类型断言时是否会 panic
		// cpu_usage_core_nanoseconds=1861297536945863i,
		`kubernetes_node,host=ubuntu_test,node_name=cn-hangzhou.172.16.2.3 cpu_usage_nanocores=401704442i,fs_available_bytes=73116872704i,fs_capacity_bytes=126692048896i,fs_used_bytes=48189890560i,memory_available_bytes=1527259136i,memory_major_page_faults=7883i,memory_page_faults=100640248i,memory_rss_bytes=5999329280i,memory_usage_bytes=7411822592i,memory_working_set_bytes=6673932288i,network_rx_bytes=530490969458i,network_rx_errors=0i,network_tx_bytes=493173160592i,network_tx_errors=0i,runtime_image_fs_available_bytes=73116872704i,runtime_image_fs_capacity_bytes=126692048896i,runtime_image_fs_used_bytes=41885561550i 1614585811000000000`,

		// memory_available_bytes=1527259136i,
		`kubernetes_node,host=ubuntu_test,node_name=cn-hangzhou.172.16.2.3 cpu_usage_core_nanoseconds=1861297536945863i,cpu_usage_nanocores=401704442i,fs_available_bytes=73116872704i,fs_capacity_bytes=126692048896i,fs_used_bytes=48189890560i,memory_major_page_faults=7883i,memory_page_faults=100640248i,memory_rss_bytes=5999329280i,memory_usage_bytes=7411822592i,memory_working_set_bytes=6673932288i,network_rx_bytes=530490969458i,network_rx_errors=0i,network_tx_bytes=493173160592i,network_tx_errors=0i,runtime_image_fs_available_bytes=73116872704i,runtime_image_fs_capacity_bytes=126692048896i,runtime_image_fs_used_bytes=41885561550i 1614585811000000000`,
	}

	for _, tc := range testcase {
		pts, err := influxm.ParsePointsWithPrecision([]byte(tc), time.Now().UTC(), "n")
		if err != nil {
			t.Fatal(err)
		}

		for _, pt := range pts {
			t.Logf("source: %s\n", pt.String())

			measurement := string(pt.Name())

			fn, ok := globalPointHandle[measurement]
			if !ok {
				t.Logf("not found measurement: %s\n\n", measurement)
				continue
			}

			data, err := fn(pt)
			if !ok {
				t.Error(err)
			}

			t.Logf("ending: %s\n\n", string(data))
		}
	}
}
