package tunnel

import "strings"

// On 注册cmd监听回调
func (s *Server) On(cmd string, fn OnRequestFunc) {
	cmd = strings.Trim(cmd, " ")
	if fn != nil && cmd != "" {
		if s.cmdFuncMap == nil {
			s.cmdFuncMap = make(map[string]OnRequestFunc)
		}
		s.cmdFuncMap[cmd] = fn
	}
}
