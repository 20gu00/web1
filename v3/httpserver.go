package v3

import "net/http"

type HttpServer interface {
	RouteDo
	ServerStart(addr string) error
}

type RouteDo interface {
	HttpRoute(method string, pattern string, handlerFunc handlerFunc) error
}
type factServer struct {
	Name    string //标记下server,日志输出的时候方便识别,大写
	root    Filter
	handler Handler
}

func (f *factServer) ServerStart(addr string) error {
	//根
	//http.HandleFunc("/", func(writer http.ResponseWriter,
	//	request *http.Request) {
	//	c := NewContext(writer, request)
	//	f.root(c)
	//})
	return http.ListenAndServe(addr, nil)
}

func (f *factServer) HttpRoute(method string, pattern string,
	handlerFunc handlerFunc) error {
	return f.handler.HttpRoute(method, pattern, handlerFunc)
}

func NewFactServer(name string, builders ...FilterBuilder) HttpServer {
	handler := NewHandlerBaseTree()
	//handler := NewHandlerBasedOnMap()
	var root Filter = handler.ServeHTTP
	for i := len(builders) - 1; i >= 0; i-- {
		builder := builders[i]
		root = builder(root)
	}
	res := &factServer{
		Name:    name,
		handler: handler,
		root:    root,
	}
	return res
}
