## 基于自定义网络协议实现分布式kv存储

### 功能特征

- 加入一致性哈希，集群每个节点负责独立的数据

- 使用分段锁机制提高并发速度

- 提供 Get/Set/Delete/Status 几种调用接口

- 提供 HTTP / TCP 两种调用服务

- 使用httprouter提供HTTP的调用服务

- 使用自定义网络协议提供TCP的调用服务

- 支持获取缓存信息，比如 key 和 value 的占用空间

- 引入内存写满保护，使用 TTL 和 LRU 两种算法进行过期

- 引入 GC 机制，随机淘汰过期数据

- 基于内存快照实现持久化功能

- 使用基于Gossip协议的开源项目memberlist进行分布式通信

  

### 性能测试

`测试环境`  **阿里云ESC服务器2核心2GB**

`测试数据量`**10000数据**

| 服务类型 | 写入速度 | 读取速度  |
| -------- | -------- | --------- |
| `tcp`    | `1313ms` | `923ms`   |
| `http`   | `3952ms` | `10950ms` |

### 自定义协议

- 请求：版本 命令 参数个数 参数长度 参数内容
- 响应：版本 答复含义 数据长度 数据内容

```
请求：
version    command    argsLength    {argLength    arg}
 1byte      1byte       4byte          4byte    unknown

响应：
version    reply    bodyLength    {body}
 1byte     1byte      4byte      unknown
```



### 使用服务

`HTTP`

```go
go run main.go -serverType http -address 127.0.0.1
```

`HTTP集群中加入机器`

```go
go run main.go -serverType http -address 127.0.0.2 -cluster 127.0.0.1
```

`TCP`

```go
go run main.go -address 127.0.0.1
```

`TCP集群中加入机器`

```go
go run main.go -address 127.0.0.2 -cluster 127.0.0.1
```

