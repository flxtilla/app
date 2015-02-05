package flotilla

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

const (
	static   nodeType = 0
	param    nodeType = 1
	catchAll nodeType = 2
)

type (
	Param struct {
		Key   string
		Value string
	}

	Params []Param

	nodeType uint8

	node struct {
		path      string
		wildChild bool
		nType     nodeType
		maxParams uint8
		indices   []byte
		children  []*node
		manage    Manage
		priority  uint32
	}

	engine struct {
		p                     sync.Pool
		trees                 map[string]*node
		app                   *App
		RedirectTrailingSlash bool
		RedirectFixedPath     bool
	}

	result struct {
		code   int
		manage Manage
		params Params
		tsr    bool
	}
)

func newEngine(app *App) *engine {
	e := &engine{app: app, RedirectTrailingSlash: true, RedirectFixedPath: true}
	e.p.New = func() interface{} { return NewCtx(app) }
	return e
}

func (e *engine) manage(method string, path string, m Manage) {
	if method != "STATUS" && path[0] != '/' {
		panic("path must begin with '/'")
	}

	if e.trees == nil {
		e.trees = make(map[string]*node)
	}

	root := e.trees[method]

	if root == nil {
		root = new(node)
		e.trees[method] = root
	}

	root.addRoute(path, m)
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	if n >= 255 {
		return 255
	}
	return uint8(n)
}

func (n *node) incrementChildPrio(i int) int {
	n.children[i].priority++
	prio := n.children[i].priority

	// adjust position (move to front)
	for j := i - 1; j >= 0 && n.children[j].priority < prio; j-- {
		// swap node positions
		tmpN := n.children[j]
		n.children[j] = n.children[i]
		n.children[i] = tmpN
		tmpI := n.indices[j]
		n.indices[j] = n.indices[i]
		n.indices[i] = tmpI

		i--
	}
	return i
}

func (n *node) addRoute(path string, manage Manage) {
	n.priority++
	numParams := countParams(path)

	// non-empty tree
	if len(n.path) > 0 || len(n.children) > 0 {
	WALK:
		for {
			// Update maxParams of the current node
			if numParams > n.maxParams {
				n.maxParams = numParams
			}

			// Find the longest common prefix.
			// This also implies that the commom prefix contains no ':' or '*'
			// since the existing key can't contain this chars.
			i := 0
			for max := min(len(path), len(n.path)); i < max && path[i] == n.path[i]; i++ {
			}

			// Split edge
			if i < len(n.path) {
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					manage:    n.manage,
					priority:  n.priority - 1,
				}

				// Update maxParams (max of all children)
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}

				n.children = []*node{&child}
				n.indices = []byte{n.path[i]}
				n.path = path[:i]
				n.manage = nil
				n.wildChild = false
			}

			// Make new node a child of this node
			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					n = n.children[0]
					n.priority++

					// Update maxParams of the child node
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					// Check if the wildcard matches
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						// check for longer wildcard, e.g. :name and :names
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue WALK
						}
					}

					panic("conflict with wildcard route")
				}

				c := path[0]

				// slash after param
				if n.nType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					continue WALK
				}

				// Check if a child with the next path byte exists
				for i, index := range n.indices {
					if c == index {
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue WALK
					}
				}

				// Otherwise insert it
				if c != ':' && c != '*' {
					n.indices = append(n.indices, c)
					child := &node{
						maxParams: numParams,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					n = child
				}
				n.insertChild(numParams, path, manage)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if n.manage != nil {
					panic("a Manage is already registered for this path")
				}
				n.manage = manage
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, path, manage)
	}
}

func (n *node) insertChild(numParams uint8, path string, manage Manage) {
	var offset int

	// find prefix until first wildcard (beginning with ':'' or '*'')
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		// Check if this Node existing children which would be
		// unreachable if we insert the wildcard here
		if len(n.children) > 0 {
			panic("wildcard route conflicts with existing children")
		}

		// find wildcard end (either '/' or path end)
		end := i + 1
		for end < max && path[end] != '/' {
			end++
		}

		if end-i < 2 {
			panic("wildcards must be named with a non-empty name")
		}

		if c == ':' { // param
			// split path at the beginning of the wildcard
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// if the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				n.children = []*node{child}
				n = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = []byte{path[i]}
			n = child
			n.priority++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				manage:    manage,
				priority:  1,
			}
			n.children = []*node{child}

			return
		}
	}

	// insert remaining path part and handle to the leaf
	n.path = path[offset:]
	n.manage = manage
}

func (n *node) getValue(path string) (manage Manage, p Params, tsr bool) {
walk: // Outer loop for walking the tree
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]
				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					c := path[0]
					for i, index := range n.indices {
						if c == index {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					tsr = (path == "/" && n.manage != nil)
					return

				}

				// handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					if p == nil {
						// lazy allocation
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.path[1:]
					p[i].Value = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						tsr = (len(path) == end+1)
						return
					}

					if manage = n.manage; manage != nil {
						return
					} else if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						tsr = (n.path == "/" && n.manage != nil)
					}

					return

				case catchAll:
					// save param value
					if p == nil {
						// lazy allocation
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.path[2:]
					p[i].Value = path

					manage = n.manage
					return

				default:
					panic("Unknown node type")
				}
			}
		} else if path == n.path {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if manage = n.manage; manage != nil {
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, index := range n.indices {
				if index == '/' {
					n = n.children[i]
					tsr = (n.path == "/" && n.manage != nil) ||
						(n.nType == catchAll && n.children[0].manage != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				path == n.path[:len(n.path)-1] && n.manage != nil)
		return
	}
}

func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	ciPath = make([]byte, 0, len(path)+1) // preallocate enough memory

	// Outer loop for walking the tree
	for len(path) >= len(n.path) && strings.ToLower(path[:len(n.path)]) == strings.ToLower(n.path) {
		path = path[len(n.path):]
		ciPath = append(ciPath, n.path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !n.wildChild {
				r := unicode.ToLower(rune(path[0]))
				for i, index := range n.indices {
					// must use recursive approach since both index and
					// ToLower(index) could exist. We must check both.
					if r == unicode.ToLower(rune(index)) {
						out, found := n.children[i].findCaseInsensitivePath(path, fixTrailingSlash)
						if found {
							return append(ciPath, out...), true
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				found = (fixTrailingSlash && path == "/" && n.manage != nil)
				return

			} else {
				n = n.children[0]

				switch n.nType {
				case param:
					// find param end (either '/' or path end)
					k := 0
					for k < len(path) && path[k] != '/' {
						k++
					}

					// add param value to case insensitive path
					ciPath = append(ciPath, path[:k]...)

					// we need to go deeper!
					if k < len(path) {
						if len(n.children) > 0 {
							path = path[k:]
							n = n.children[0]
							continue
						} else { // ... but we can't
							if fixTrailingSlash && len(path) == k+1 {
								return ciPath, true
							}
							return
						}
					}

					if n.manage != nil {
						return ciPath, true
					} else if fixTrailingSlash && len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists
						n = n.children[0]
						if n.path == "/" && n.manage != nil {
							return append(ciPath, '/'), true
						}
					}
					return

				case catchAll:
					return append(ciPath, path...), true

				default:
					panic("Unknown node type")
				}
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.manage != nil {
				return ciPath, true
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i, index := range n.indices {
					if index == '/' {
						n = n.children[i]
						if (n.path == "/" && n.manage != nil) ||
							(n.nType == catchAll && n.children[0].manage != nil) {
							return append(ciPath, '/'), true
						}
						return
					}
				}
			}
			return
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(path)+1 == len(n.path) && n.path[len(path)] == '/' &&
			strings.ToLower(path) == strings.ToLower(n.path[:len(path)]) &&
			n.manage != nil {
			return append(ciPath, n.path...), true
		}
	}
	return
}

func (e *engine) status(code int) *result {
	if root := e.trees["STATUS"]; root != nil {
		if manage, params, tsr := root.getValue(strconv.Itoa(code)); manage != nil {
			return &result{code, manage, params, tsr}
		}
	}
	return &result{code, statusManage(code), nil, false}
}

func (e *engine) statusfunc(c *Ctx, code int) error {
	rslt := e.status(code)
	rslt.manage(c)
	return nil
}

func (e *engine) lookup(method, path string) *result {
	if root := e.trees[method]; root != nil {
		if manage, params, tsr := root.getValue(path); manage != nil {
			return &result{200, manage, params, tsr}
		} else if method != "CONNECT" && path != "/" {
			code := 301
			if method != "GET" {
				code = 307
			}
			if tsr && e.RedirectTrailingSlash {
				var newpath string
				if path[len(path)-1] == '/' {
					newpath = path[:len(path)-1]
				} else {
					newpath = path + "/"
				}
				return &result{code: code,
					manage: func(c *Ctx) {
						c.Request.URL.Path = newpath
						http.Redirect(c.RW, c.Request, c.Request.URL.String(), code)
					}}
			}
			if e.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					e.RedirectTrailingSlash,
				)
				if found {
					return &result{code: code,
						manage: func(c *Ctx) {
							c.Request.URL.Path = string(fixedPath)
							http.Redirect(c.RW, c.Request, c.Request.URL.String(), code)
						}}
				}
			}

		}
	}
	for method := range e.trees {
		if method == method {
			continue
		}
		handle, _, _ := e.trees[method].getValue(path)
		if handle != nil {
			return e.status(405)
		}
	}
	return e.status(404)
}

func rcvr(c *Ctx) {
	if rcv := recover(); rcv != nil {
		p := newError(fmt.Sprintf("%s", rcv))
		c.errorTyped(p, ErrorTypePanic, stack(3))
		c.Call("status", 500)
	}
}

func (e *engine) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c, cancel := e.get(rw, req)
	handle(e, c)
	cancel(c)
}

func handle(e *engine, c *Ctx) {
	defer rcvr(c)
	rslt := e.lookup(c.Request.Method, c.Request.URL.Path)
	c.Result(rslt)
	rslt.manage(c)
}

func (e *engine) get(rw http.ResponseWriter, req *http.Request) (*Ctx, CancelFunc) {
	c := e.p.Get().(*Ctx)
	c.Reset(req, rw)
	//c.Request.ParseMultipartForm(e.app.Env.Store["UPLOAD_SIZE"].Int64())
	c.Start(e.app.SessionManager)
	cancel := func(c *Ctx) { e.put(c); c.Cancel() }
	return c, cancel
}

func (e *engine) put(c *Ctx) {
	e.p.Put(c)
}
