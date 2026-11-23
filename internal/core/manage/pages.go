package manage

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"golang.org/x/text/language"
)

type page struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Purpose string `json:"purpose,omitempty"`
}

func (s *Service) pages(id string, locales []string) []page {
	pages := make([]page, 0)
	// 直接使用管理服务的调用方可以只装配持久状态，页面来自运行路由。
	if s.opt.Router == nil {
		return pages
	}
	plugin, ok := s.opt.Router.Plugin(id)
	if !ok {
		return pages
	}
	for _, route := range plugin.Routes() {
		if route.Kind != pb.RouteKind_ROUTE_KIND_HTTP || route.PageTitle == "" {
			continue
		}
		title := route.PageTitle
		path := route.GetHttp().GetPath()
		if catalog := plugin.Messages(); catalog != nil {
			title = catalog.PageTitle(locales, path, title)
		}
		pages = append(pages, page{Path: path, Title: title, Purpose: route.GetPagePurpose()})
	}
	return pages
}

func pageLanguages(header string) []string {
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return nil
	}
	locales := make([]string, 0, len(tags))
	for _, tag := range tags {
		locales = append(locales, tag.String())
	}
	return locales
}
