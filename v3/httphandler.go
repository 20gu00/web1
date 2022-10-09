package v3

type handlerFunc func(c *Context)

type Handler interface {
	RouteDo
	ServeHTTP(c *Context)
}
