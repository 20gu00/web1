package v2

import (
	"net/http"
	"strings"
)

var ErrInvalidPattern = errors.New("invalid pattern error")

type HandlerBaseTree struct {
	root *node
}

func NewHandlerBaseTree() Handler {
	return &HandlerBaseTree{
		root: &node{},
	}
}

type node struct {
	path     string
	children []*node
	handler  handlerFunc
}

func newNode(path string) *node {
	return &node{
		path:     path,
		children: make([]*node, 0, 10),
	}
}
func (h *HandlerBaseTree) ServeHTTP(c *Context) {
	handler, found := h.searchRouter(c.R.URL.Path)
	if !found {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("not found!"))
		return
	}

	handler(c)
}

func (h *HandlerBaseTree) searchRouter(pattern string) (handlerFunc, bool) {
	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")
	currentRoot := h.root
	for _, path := range paths {
		child, ok := h.searchChild(currentRoot, path)
		if !ok {
			return nil, false
		}
		currentRoot = child
	}
	if currentRoot.handler == nil {
		return nil, false
	}
	return currentRoot.handler, true
}

func (h *HandlerBaseTree) HttpRoute(
	method string,
	pattern string,
	handlerFunc handlerFunc) {
	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")
	currentRoot := h.root
	for index, path := range paths {
		matchChild, ok := h.searchChild(currentRoot, path)
		if ok {
			currentRoot = matchChild
		} else {
			// 为当前节点根据
			h.buildSubTree(currentRoot, paths[index:], handlerFunc)
			return
		}
	}
	currentRoot.handler = handlerFunc
}

func (h *HandlerBaseTree) searchChild(root *node, path string) (*node, bool) {
	var wildcardNode *node
	for _, child := range root.children {
		if child.path == path && child.path != "*"{
			return child, true
		}

		if child.path == "*" {
			wildcardNode = child
		}
	}
	b:=wildcardNode!=nil
	return wildcardNode, b  //nil,err
}

func (h *HandlerBaseTree) buildSubTree(root *node, paths []string, handlerFn handlerFunc) {
	currentRoot := root
	for _, path := range paths {
		newN := newNode(path)
		currentRoot.children = append(currentRoot.children, newN)
		currentRoot = newN
	}
	currentRoot.handler = handlerFn
}

func (h *HandlerBaseTree) validatePattern(pattern string) error {
	//校验 *，如果存在，必须在最后一个，并且它前面必须是/
	//只接受 /* 的存在，abc*这种是非法 *abc */abc **都不允许,因为这样会要求处理路由的回溯匹配,太麻烦了

	p := strings.Index(pattern, "*")
	if p > 0 {
		if p != len(pattern) - 1 {
			return ErrInvalidPattern
		}

		if pattern[p-1] != '/' {
			return ErrInvalidPattern
		}

	}
	return nil
}
