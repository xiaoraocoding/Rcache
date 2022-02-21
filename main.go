package main

import (
	"Rcache/caches"
	"Rcache/servers"
)

func main() {
	cache := caches.NewCache()
	err := servers.NewHTTPServer(cache).Run(":8888")
	if err != nil {
		panic(err)
	}
}
