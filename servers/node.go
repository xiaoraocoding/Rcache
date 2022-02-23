package servers

import (
	"Rcache/helpers"
	"github.com/hashicorp/memberlist"
	"io/ioutil"
	"stathat.com/c/consistent"
	"time"
)


// node 代表集群中的一个节点，会保存一些和集群相关的数据。
type node struct {

	// options 存储着一些服务器相关的选项。
	options *Options

	// address 记录的是当前节点的访问地址，包含 ip 或者主机、端口等信息。
	address string

	// circle 是一致性哈希的实例。
	circle *consistent.Consistent

	// nodeManager 是节点管理器，用于管理节点。
	nodeManager *memberlist.Memberlist
}

// newNode 创建一个节点实例，并使用 options 去初始化。
func newNode(options *Options) (*node, error) {

	// 如果没有需要加入的集群，则把当前节点当成新集群
	if options.Cluster == nil || len(options.Cluster) == 0 {
		options.Cluster = []string{options.Address}
	}

	// 创建节点管理器，后续所有和集群相关的操作都需要通过这个节点管理器
	nodeManager, err := createNodeManager(options)
	if err != nil {
		return nil, err
	}

	// 创建节点
	node := &node{
		options:     options,
		address:     helpers.JoinAddressAndPort(options.Address, options.Port),
		circle:      consistent.New(),
		nodeManager: nodeManager,
	}

	// 注意这里设置了一致性哈希的虚拟节点数，并开启了自动更新一致性哈希内的物理节点信息
	node.circle.NumberOfReplicas = options.VirtualNodeCount
	node.autoUpdateCircle()
	return node, nil
}

func createNodeManager(options *Options) (*memberlist.Memberlist, error) {

	// 在默认的 LAN 配置上进行设置
	config := memberlist.DefaultLANConfig()
	config.Name = helpers.JoinAddressAndPort(options.Address, options.Port)
	config.BindAddr = options.Address
	config.LogOutput = ioutil.Discard // 禁用日志输出

	// 创建 memberlist 实例
	nodeManager, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	// 加入到指定的集群
	_, err = nodeManager.Join(options.Cluster)
	return nodeManager, err
}

func (n *node) nodes() []string {
	members := n.nodeManager.Members()
	nodes := make([]string, len(members))
	for i, member := range members {
		nodes[i] = member.Name
	}
	return nodes
}

//  根据 name 选择出一个适合的 node。
func (n *node) selectNode(name string) (string, error) {
	return n.circle.Get(name)
}

//  判断 address 是否指当前节点。
func (n *node) isCurrentNode(address string) bool {
	return n.address == address
}

// 更新一致性哈希的信息。
// 一致性哈希的信息来源就是 memberlist 实例。
func (n *node) updateCircle() {
	n.circle.Set(n.nodes())
}

// autoUpdateCircle 开启一个定时任务去定期更新一致性哈希的信息。
func (n *node) autoUpdateCircle() {
	n.updateCircle()
	go func() {
		ticker := time.NewTicker(time.Duration(n.options.UpdateCircleDuration) * time.Second)
		for {
			select {
			case <-ticker.C:
				n.updateCircle()
			}
		}
	}()
}
