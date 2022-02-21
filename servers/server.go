package servers

const (
	// APIVersion 代表当前服务的版本。
	// 因为我们做的服务是提供给外部调用的，而版本的升级可能会带来 API 的改动。
	// 我们需要标记当前服务能提供 API 的版本，这样即使后面升级了 API 也不用担心，只要用户调用的版本是正确的，调用就不会出错
	APIVersion = "v1"
)

// Server 是服务器的抽象接口。
type Server interface {

	// Run 会将服务器启动指定的 address 上。
	Run(address string) error
}