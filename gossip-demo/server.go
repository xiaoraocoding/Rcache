package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"github.com/hashicorp/memberlist"
)

// address 是当前服务器绑定的地址
// existedCluster 是一个现有的机器群
func RunServer(address string, existedCluster string) {

	// 创建一个配置，并将绑定的地址都配置上
	conf := memberlist.DefaultLANConfig()
	conf.Name = address
	conf.BindAddr = address

	// 创建一个成员实例
	cluster, err := memberlist.Create(conf)
	if err != nil {
		panic(err)
	}

	// 如果没有指定机器群，当前服务器就是为新的机器群
	if existedCluster == "" {
		existedCluster = address
	}

	// 加入指定的机器群
	_, err = cluster.Join([]string{existedCluster})
	if err != nil {
		panic(err)
	}

	// 提供一个 http 服务，用于查询当前服务器的机器群有哪些成员
	http.HandleFunc("/servers", func(writer http.ResponseWriter, request *http.Request) {

		// 查询成员，并取出地址存到一个切片中，响应 Json 编码的数据
		members := cluster.Members()
		hosts := make([]string, len(members))
		for i, node := range members {
			hosts[i] = node.Addr.String()
		}
		membersJson, _ := json.Marshal(hosts)
		writer.Write(membersJson)
	})
	http.ListenAndServe(address+":8080", nil)
}

func main() {

	// 指定当前服务器的地址，默认是 127.0.0.1
	ip := flag.String("ip", "127.0.0.1", "The ip of this server")

	// 指定要加入的机器群
	cluser := flag.String("cluster", "", "The existed server of a cluster")
	flag.Parse()

	// 启动服务器
	RunServer(*ip, *cluser)
}
