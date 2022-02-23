package servers

// Options 是服务器的选项配置。
type Options struct {

	// Address 是服务器监听使用的地址。
	Address string

	// Port 是服务器监听使用的端口。
	Port int

	// ServerType 是服务器的类型。
	ServerType string

	// VirtualNodeCount 是指一致性哈希的虚拟节点个数。
	VirtualNodeCount int

	// UpdateCircleDuration 是指更新一致性哈希信息的时间间隔。
	// 单位是秒。
	UpdateCircleDuration int

	// cluster 是指需要加入的集群，只需要集群中一个节点的地址即可。
	Cluster []string
}

func DefaultOptions() Options {
	return Options{
		Address:              "127.0.0.1",
		Port:                 5837,
		ServerType:           "tcp",
		VirtualNodeCount:     1024,
		UpdateCircleDuration: 3,  //这里的单位是秒
	}
}
