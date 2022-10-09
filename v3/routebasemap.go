package v3

import (
	"net/http"
	"sync"
)

var _ Handler = &HandlerBaseMap{} //判断HandlerBaseMap是否是Handler类型,这里其实就是HandlerBaseMap是否实现了Handler接口

type HandlerBaseMap struct {
	handlers sync.Map //线程安全的map
}

func NewHandlerBaseMap() *HandlerBaseMap {
	return &HandlerBaseMap{}
}

func (h *HandlerBaseMap) ServeHTTP(c *Context) {
	req := c.R
	key := h.getKey(req.Method, req.URL.Path)
	handler, ok := h.handlers.Load(key)
	if !ok {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("未匹配到相应的路由"))
		return
	}

	handler.(handlerFunc)(c)
}

func (h *HandlerBaseMap) HttpRoute(method string, pattern string,
	handlerFunc handlerFunc) error {
	key := h.getKey(method, pattern)
	h.handlers.Store(key, handlerFunc)
	return nil
}

func (h *HandlerBaseMap) getKey(method string,
	path string) string {
	return method + "@" + path
}
