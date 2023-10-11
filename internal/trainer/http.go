package main

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/vasiliiperfilev/ddd/internal/common/auth"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
	"github.com/vasiliiperfilev/ddd/internal/common/server/httperr"
)

type HttpServer struct {
	db             db
	backgroundJobs *server.BackgroundJobs
}

func (h HttpServer) GetTrainerAvailableHours(w http.ResponseWriter, r *http.Request, queryParams GetTrainerAvailableHoursParams) {
	if queryParams.DateFrom.After(queryParams.DateTo) {
		httperr.BadRequest("date-from-after-date-to", nil, w, r)
		return
	}

	dates, err := h.db.GetDates(r.Context(), &queryParams)
	if err != nil {
		httperr.InternalError("unable-to-get-dates", err, w, r)
		return
	}

	render.Respond(w, r, dates)
}

func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.Unauthorised("no-user-found", err, w, r)
		return
	}

	if !auth.PermissionsMap.HasPermission(user.Role, auth.ChangeTrainerHoursPermission) {
		httperr.Unauthorised("invalid-role", nil, w, r)
		return
	}

	if err := h.db.UpdateAvailability(r, true); err != nil {
		httperr.InternalError("unable-to-update-availability", err, w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h HttpServer) MakeHourUnavailable(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.Unauthorised("no-user-found", err, w, r)
		return
	}
	if !auth.PermissionsMap.HasPermission(user.Role, auth.ChangeTrainerHoursPermission) {
		httperr.Unauthorised("invalid-role", nil, w, r)
		return
	}

	if err := h.db.UpdateAvailability(r, false); err != nil {
		httperr.InternalError("unable-to-update-availability", err, w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
