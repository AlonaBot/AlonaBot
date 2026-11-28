package manage

import (
	"net/http"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/telemetry"
)

func (s *Service) coreTelemetry(w http.ResponseWriter, r *http.Request) {
	if s.opt.Telemetry == nil {
		respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "core telemetry is not enabled"})
		return
	}
	window, err := telemetry.ParseWindow(r.URL.Query().Get("window"))
	if err != nil {
		respond(w, nil, invalid(err.Error()))
		return
	}
	report, err := s.opt.Telemetry.Snapshot(r.Context(), window)
	respond(w, report, err)
}
