package aliyunprice

type (
	// //DataDisk 数据盘
	// DataDisk struct {
	// 	Region   string
	// 	ImageOs  string //windows or linux
	// 	Category string //cloud, cloud_ssd...
	// }

	// //InstanceType 实例规格
	// InstanceType struct {
	// 	Region       string
	// 	ImageOs      string //windows or linux
	// 	InstanceType string //ecs.t1.small
	// 	IoOptimized  bool   //是否I/O 优化实例
	// }

	// //SystemDisk 系统盘
	// SystemDisk struct {
	// 	Region   string
	// 	ImageOs  string //windows or linux
	// 	Category string //cloud, cloud_ssd...
	// 	Size     int    //GB
	// }

	// InternetMaxBandwidthOut struct {
	// 	Region                  string
	// 	InternetMaxBandwidthOut int //M
	// }

	// InternetTrafficOut struct {
	// 	Region string
	// }

	Ecs struct {
		Region string

		//数据盘
		DataDiskCategory string //cloud, cloud_ssd...
		DataDiskSize     int    //GB

		//实例规格
		InstanceType string //ecs.t1.small
		IoOptimized  bool   //是否I/O 优化实例，默认false
		ImageOs      string //windows or linux

		//系统盘
		SystemDiskCategory string //cloud, cloud_ssd...
		SystemDiskSize     int    //GB

		//带宽/流量
		UseBandwidth            bool  //按固定带宽或使用流量，默认true
		InternetMaxBandwidthOut int64 //Kbps 使用固定带宽时的固定带宽值，默认1M
	}
)
