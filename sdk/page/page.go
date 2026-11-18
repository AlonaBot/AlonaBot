// Package page 提供插件页面的前端套件静态资源（kit.css 与 kit.js）服务。
//
// 插件通过 [Handler] 将套件静态资源挂载在自身的 HTTP 路由下；
// 插件页面引用后即可自动与 WebUI postMessage 桥对接，跟随设计令牌、
// 明暗主题与界面语言切换，并经桥请求插件自身路由。
// 契约见 spec/ADMIN-API.md §7 与 DECISIONS D-159。
package page

import (
	_ "embed"
	"fmt"
	"net/http"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

//go:embed kit.css
var kitCSS []byte

//go:embed kit.js
var kitJS []byte

const (
	coreModule        = "github.com/AlonaBot/AlonaBot"
	defaultSDKVersion = "(devel)"
	contentTypeCSS    = "text/css; charset=utf-8"
	contentTypeJS     = "text/javascript; charset=utf-8"
	headerContentType = "Content-Type"
	headerContentLen  = "Content-Length"
	headerETag        = "ETag"
	headerIfNoneMatch = "If-None-Match"
	filenameKitCSS    = "kit.css"
	filenameKitJS     = "kit.js"
)

var getSDKVersion = sync.OnceValue(detectSDKVersion)

func detectSDKVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return defaultSDKVersion
	}
	if info.Main.Path == coreModule && info.Main.Version != "" {
		return info.Main.Version
	}
	for _, dependency := range info.Deps {
		if dependency.Path == coreModule {
			if dependency.Replace != nil {
				dependency = dependency.Replace
			}
			if dependency.Version != "" {
				return dependency.Version
			}
			return defaultSDKVersion
		}
	}
	return defaultSDKVersion
}

// Handler 返回响应 kit.css 与 kit.js 静态资源的 HTTP 处理函数。
// 只响应路径末段为 kit.css 与 kit.js 的 GET 请求，其余均响应 404；
// 响应带有正确的 Content-Type 以及基于 SDK 版本的 ETag 头。
func Handler() http.Handler {
	etag := fmt.Sprintf("%q", getSDKVersion())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		filename := path.Base(r.URL.Path)
		var content []byte
		var contentType string
		switch filename {
		case filenameKitCSS:
			content = kitCSS
			contentType = contentTypeCSS
		case filenameKitJS:
			content = kitJS
			contentType = contentTypeJS
		default:
			http.NotFound(w, r)
			return
		}

		w.Header().Set(headerContentType, contentType)
		w.Header().Set(headerETag, etag)

		if match := r.Header.Get(headerIfNoneMatch); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set(headerContentLen, strconv.Itoa(len(content)))
		_, _ = w.Write(content)
	})
}
