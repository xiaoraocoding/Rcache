package servers

import (
	"Rcache/caches"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

// HTTPServer 是提供 http 服务的服务器。
type HTTPServer struct {
	// cache 是内部存储用的缓存实例。
	cache *caches.Cache
}

// NewHTTPServer 返回一个 http 服务器。
func NewHTTPServer(cache *caches.Cache) *HTTPServer {
	return &HTTPServer{
		cache: cache,
	}
}

// Run 启动这个 http 服务器。
func (hs *HTTPServer) Run(address string) error {
	return http.ListenAndServe(address, hs.routerHandler())
}

// =======================================================================

// wrapUriWithVersion 会用 API 版本去包装 uri，比如 "v1" 版本的 API 包装 "/cache" 就会变成 "/v1/cache"。
func wrapUriWithVersion(uri string) string {
	return path.Join("/", APIVersion, uri)
}

// routerHandler 返回注册的路由处理器。
func (hs *HTTPServer) routerHandler() http.Handler {
	router := httprouter.New()
	router.GET(wrapUriWithVersion("/cache/:key"), hs.getHandler)
	router.PUT(wrapUriWithVersion("/cache/:key"), hs.setHandler)
	router.DELETE(wrapUriWithVersion("/cache/:key"), hs.deleteHandler)
	router.GET(wrapUriWithVersion("/status"), hs.statusHandler)
	return router
}

// getHandler 获取缓存中的数据并返回。
func (hs *HTTPServer) getHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := hs.cache.Get(key)
	if !ok {
		// 返回 404 错误码
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	writer.Write(value)
}

// setHandler 添加数据到缓存中。
func (hs *HTTPServer) setHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	key := params.ByName("key")
	value, err := ioutil.ReadAll(request.Body)
	if err != nil {
		// 返回 500 错误码
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 从请求中获取 ttl
	ttl, err := ttlOf(request)
	if err != nil {
		// 返回 500 错误码
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 添加数据，并设置为指定的 ttl
	err = hs.cache.SetWithTTL(key, value, ttl)
	if err != nil {
		// 如果返回了错误，说明触发了写满保护机制，返回 413 错误码，这个错误码表示请求体中的数据太大了
		// 同时返回错误信息，加上一个 "Error: " 的前缀，方便识别为错误码
		writer.WriteHeader(http.StatusRequestEntityTooLarge)
		writer.Write([]byte("Error: " + err.Error()))
		return
	}
	// 成功添加就返回 201 的状态码，其实 200 的状态码也可以，不过 201 的语义更符合，所以就选了这个状态码
	writer.WriteHeader(http.StatusCreated)
}

// ttlOf 从请求中解析 ttl 并返回，如果 error 不为空，说明 ttl 解析出错。
func ttlOf(request *http.Request) (int64, error) {

	// 从请求头中获取 ttl 头部，如果没有设置或者 ttl 为空均按不设置 ttl 处理，也就是不会过期
	ttls, ok := request.Header["Ttl"]
	if !ok || len(ttls) < 1 {
		return caches.NeverDie, nil
	}
	return strconv.ParseInt(ttls[0], 10, 64)
}

// deleteHandler 从缓存中删除指定数据。
func (hs *HTTPServer) deleteHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	hs.cache.Delete(key)
}

// statusHandler 返回缓存信息。
func (hs *HTTPServer) statusHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	status, err := json.Marshal(hs.cache.Status())
	if err != nil {
		// 返回 500 错误码
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(status)
}