package v3

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

var sustainMethods = [4]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
var ErrInvalidMethod = errors.New("invalid method error")
var ErrInvalidPattern = errors.New("invalid pattern error")

type HandlerBaseTree struct {
	//root *node
	trees map[string]*node
}

func NewHandlerBaseTree() Handler {
	trees := make(map[string]*node, len(sustainMethods))
	for _, method := range sustainMethods {
		trees[method] = newRootNode(method)
	}
	return &HandlerBaseTree{
		trees: trees,
	}
}

func (h *HandlerBaseTree) ServeHTTP(c *Context) {
	handler, found := h.searchRouter(c, c.R.Method, c.R.URL.Path)
	if !found {
		c.W.WriteHeader(http.StatusNotFound)
		_, _ = c.W.Write([]byte("not found!"))
		return
	}

	handler(c)
}

func (h *HandlerBaseTree) searchRouter(c *Context, method string, pattern string) (handlerFunc, bool) {
	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")
	currentRoot, ok := h.trees[method]
	if !ok {
		return nil, false
	}

	for _, path := range paths {
		// 从子节点里边找一个匹配到了当前 p 的节点
		child, ok := h.searchChild(currentRoot, path, c)
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

func (h *HandlerBaseTree) HttpRoute(method string, pattern string, handlerFunc handlerFunc) error {
	err := h.validatePattern(pattern)
	if err != nil {
		return err
	}

	pattern = strings.Trim(pattern, "/")
	paths := strings.Split(pattern, "/")
	currentRoot, ok := h.trees[method]
	if !ok {
		return ErrInvalidMethod
	}
	for index, path := range paths {
		child, ok := h.searchChild(currentRoot, path, nil) //context,nil
		// != nodeTypeAny 是考虑到 /order/* 和 /order/:id 这种注册顺序
		if ok && child.nodeType != nodeTypeAny {
			currentRoot = child
		} else {
			h.buildSubTree(currentRoot, paths[index:], handlerFunc)
			return nil
		}
	}
	currentRoot.handler = handlerFunc
	return nil
}

func (h *HandlerBaseTree) searchChild(root *node, path string, c *Context) (*node, bool) {
	candidates := make([]*node, 0, 2)
	for _, child := range root.children {
		if child.matchFunc(path, c) {
			candidates = append(candidates, child)
		}
	}
	if len(candidates) == 0 {
		return nil, false
	}

	// type 也决定了它们的优先级
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].nodeType < candidates[j].nodeType
	})
	return candidates[len(candidates)-1], true
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
		if p != len(pattern)-1 {
			return ErrInvalidPattern
		}

		if pattern[p-1] != '/' {
			return ErrInvalidPattern
		}

	}
	return nil
}
