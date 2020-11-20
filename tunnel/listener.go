package tunnel

import "strings"

// Listen 注册cmd监听回调
func (t *tunnel) Listen(cmd string, fn OnRequestFunc) {
	cmd = strings.Trim(cmd, " ")
	if fn != nil && cmd != "" {
		if t.cmdFuncMap == nil {
			t.cmdFuncMap = make(map[string]OnRequestFunc)
		}
		t.cmdFuncMap[cmd] = fn
	}
}
