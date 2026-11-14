package manage

import "net/http"

func (s *Service) coreInfo(w http.ResponseWriter, r *http.Request) {
	version, listen := s.opt.CoreInfo()
	respond(w, struct {
		Version       string `json:"version"`
		Listen        string `json:"listen"`
		AdminAPIMajor uint32 `json:"admin_api_major"`
	}{version, listen, AdminAPIMajor}, nil)
}
