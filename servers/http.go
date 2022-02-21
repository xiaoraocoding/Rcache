package servers

import (
	"Rcache/caches"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)


type HTTPServer struct {
	cache *caches.Cache
}


func NewHTTPServer(cache *caches.Cache) *HTTPServer {
	return &HTTPServer{
		cache: cache,
	}
}

func (hs *HTTPServer) Run(address string) error {
	return http.ListenAndServe(address, hs.routerHandler())
}


func (hs *HTTPServer) routerHandler() http.Handler {
	router := httprouter.New()
	router.GET("/cache/:key", hs.getHandler)
	router.PUT("/cache/:key", hs.setHandler)
	router.DELETE("/cache/:key", hs.deleteHandler)
	router.GET("/status", hs.statusHandler)
	return router
}


func (hs *HTTPServer) getHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := hs.cache.Get(key)
	if !ok {
		// 如果缓存中找不到数据，就返回 404 的状态码
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	writer.Write(value)
}


func (hs *HTTPServer) setHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	// value 从请求体中获取，整个请求体都被当作 value
	value, err := ioutil.ReadAll(request.Body)
	if err != nil {
		// 如果读取请求体失败，就返回 500 的状态码
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	hs.cache.Set(key, value)
}


func (hs *HTTPServer) deleteHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	hs.cache.Delete(key)
}


func (hs *HTTPServer) statusHandler(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// 将个数编码成 JSON 字符串
	status, err := json.Marshal(map[string]interface{}{
		"count": hs.cache.Count(),
	})
	if err != nil {
		// 如果编码失败，就返回 500 的状态码
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(status)
}
