package manage

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
)

// ErrNeedsConfirm 列出需要管理员显式确认的受影响对象。
type ErrNeedsConfirm struct{ Affected []string }

// Error 列出本次操作需要确认的对象。
func (e *ErrNeedsConfirm) Error() string {
	return "confirmation required: " + strings.Join(e.Affected, ", ")
}

// Handler 将管理端点全部放在相同的 Bearer 或 Cookie/CSRF 认证之后。
func (s *Service) Handler(auth *adminauth.Auth) http.Handler {
	mux := http.NewServeMux()
	s.configRoutes(mux, auth)
	s.accessRoutes(mux)
	s.policyRoutes(mux)
	s.optionRoutes(mux)
	s.toolRoutes(mux)
	s.settingsRoutes(mux)
	s.instanceRoutes(mux)
	s.conversationRoutes(mux)
	s.logRoutes(mux)
	s.jobRoutes(mux)
	s.dataRoutes(mux)
	mux.HandleFunc("GET /api/v1/plugins/{id}/generations", s.generations)
	mux.HandleFunc("GET /api/v1/adapter/instances", s.listInstances)
	mux.HandleFunc("GET /api/v1/chats", s.chats)
	s.profileRoutes(mux)
	mux.HandleFunc("POST /api/v1/diagnostics", s.diagnostics)
	mux.HandleFunc("GET /api/v1/core", s.coreInfo)
	mux.HandleFunc("GET /api/v1/core/changelog", s.coreChangelog)
	mux.HandleFunc("GET /api/v1/packages/{id}/changelog", s.packageChangelog)
	mux.HandleFunc("GET /api/v1/packages/{id}/icon", s.icon)
	mux.HandleFunc("GET /api/v1/packages/{id}/examples/{path...}", s.example)
	mux.HandleFunc("POST /api/v1/packages/install", func(w http.ResponseWriter, r *http.Request) {
		if multipartUpload(r) {
			result, err := s.uploadPackage(r, false)
			respond(w, result, err)
			return
		}
		var request InstallRequest
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		result, err := s.Install(r.Context(), request)
		respond(w, result, err)
	})
	for _, action := range []string{"enable", "disable", "restart", "uninstall", "pin", "reload", "update", "rollback"} {
		mux.HandleFunc("POST /api/v1/packages/"+action, func(w http.ResponseWriter, r *http.Request) {
			if action == "update" && multipartUpload(r) {
				result, err := s.uploadPackage(r, true)
				respond(w, result, err)
				return
			}
			var request struct {
				ID         string         `json:"id"`
				Confirm    bool           `json:"confirm"`
				Purge      PurgeScope     `json:"purge"`
				Constraint string         `json:"constraint"`
				Generation string         `json:"generation"`
				Install    InstallRequest `json:"install"`
			}
			if err := decode(r, &request); err != nil {
				respond(w, nil, err)
				return
			}
			var err error
			switch action {
			case "enable":
				err = s.Enable(r.Context(), request.ID)
			case "disable":
				err = s.Disable(r.Context(), request.ID, request.Confirm)
			case "restart":
				err = s.Restart(r.Context(), request.ID)
			case "uninstall":
				err = s.Uninstall(r.Context(), request.ID, request.Purge, request.Confirm)
			case "pin":
				err = s.Pin(r.Context(), request.ID, request.Constraint)
			case "reload":
				err = s.Reload(r.Context(), request.ID)
			case "update":
				request.Install.Confirm = request.Confirm
				result, err := s.Update(r.Context(), request.ID, request.Install)
				respond(w, result, err)
				return
			case "rollback":
				err = s.Rollback(r.Context(), request.ID, request.Generation, request.Confirm)
			}
			respond(w, struct{}{}, err)
		})
	}
	mux.HandleFunc("GET /api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.List(r.Context())
		locales := pageLanguages(r.Header.Get("Accept-Language"))
		for i := range result {
			result[i].Pages = s.pages(result[i].ID, locales)
		}
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/admins", func(w http.ResponseWriter, r *http.Request) { respond(w, s.Admins(r.Context()), nil) })
	for _, action := range []string{"add", "remove"} {
		mux.HandleFunc("POST /api/v1/admins/"+action, func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				User string `json:"user"`
			}
			if err := decode(r, &request); err != nil {
				respond(w, nil, err)
				return
			}
			var err error
			if action == "add" {
				err = s.AddAdmin(r.Context(), request.User)
			} else {
				err = s.RemoveAdmin(r.Context(), request.User)
			}
			respond(w, struct{}{}, err)
		})
	}
	mux.HandleFunc("POST /api/v1/adapter/instances/create", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			PackageID  string          `json:"package_id"`
			InstanceID string          `json:"instance_id"`
			Config     json.RawMessage `json:"config"`
			Confirm    bool            `json:"confirm"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		respond(w, struct{}{}, s.CreateInstance(r.Context(), request.PackageID, request.InstanceID, request.Config, request.Confirm))
	})
	for _, action := range []string{"start", "stop", "delete"} {
		mux.HandleFunc("POST /api/v1/adapter/instances/"+action, func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				InstanceID string `json:"instance_id"`
			}
			if err := decode(r, &request); err != nil {
				respond(w, nil, err)
				return
			}
			var err error
			switch action {
			case "start":
				err = s.StartInstance(r.Context(), request.InstanceID)
			case "stop":
				err = s.StopInstance(r.Context(), request.InstanceID)
			case "delete":
				err = s.DeleteInstance(r.Context(), request.InstanceID)
			}
			respond(w, struct{}{}, err)
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "management endpoint not found"})
	})
	return auth.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.dataKVRequest(w, r) {
			return
		}
		if s.deleteConversation(w, r) {
			return
		}
		if s.setJobPaused(w, r) {
			return
		}
		// ServeMux 会清理路径并重定向；在此之前拒绝原始的越界或非规范路径。
		if !validStaticPath(strings.TrimPrefix(r.URL.Path, "/")) {
			respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "management endpoint not found"})
			return
		}
		mux.ServeHTTP(w, r)
	}))
}

func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return invalid("invalid request JSON: " + err.Error())
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return invalid("request must contain one JSON object")
	}
	return nil
}

func respond(w http.ResponseWriter, result any, err error) {
	status := http.StatusOK
	if err != nil {
		var confirmation *ErrNeedsConfirm
		if errors.As(err, &confirmation) {
			status = http.StatusConflict
			result = struct {
				Affected []string `json:"affected"`
			}{confirmation.Affected}
		} else {
			var detail *validate.Error
			errors.As(failure(err), &detail)
			status = http.StatusInternalServerError
			switch detail.Code {
			case pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT:
				status = http.StatusBadRequest
			case pb.ErrorCode_ERROR_CODE_NOT_FOUND:
				status = http.StatusNotFound
			case pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS:
				status = http.StatusConflict
			case pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED:
				status = http.StatusForbidden
			case pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION:
				status = http.StatusPreconditionFailed
			case pb.ErrorCode_ERROR_CODE_UNSUPPORTED:
				status = http.StatusNotImplemented
			case pb.ErrorCode_ERROR_CODE_UNAVAILABLE:
				status = http.StatusServiceUnavailable
			}
			var conflict *configConflict
			if errors.As(err, &conflict) {
				status = http.StatusConflict
			}
			result = struct {
				Code    string `json:"code"`
				Message string `json:"safe_message"`
			}{strings.TrimPrefix(detail.Code.String(), "ERROR_CODE_"), detail.Message}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("write management response failed", "error", err)
	}
}
