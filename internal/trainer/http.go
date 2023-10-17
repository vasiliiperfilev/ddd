package main

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/vasiliiperfilev/ddd/internal/common/auth"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
	"github.com/vasiliiperfilev/ddd/internal/common/server/httperr"
	"github.com/vasiliiperfilev/ddd/internal/trainer/domain/hour"
	"github.com/vasiliiperfilev/ddd/internal/trainer/openapi_types"
)

type HttpServer struct {
	db             db
	hourRepository hour.Repository
	backgroundJobs *server.BackgroundJobs
}

func (h HttpServer) GetTrainerAvailableHours(w http.ResponseWriter, r *http.Request, queryParams openapi_types.GetTrainerAvailableHoursParams) {
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

	hourUpdate := &openapi_types.HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		httperr.BadRequest("unable-to-update-availability", err, w, r)
	}

	for _, hourToUpdate := range hourUpdate.Hours {
		err := h.hourRepository.UpdateHour(r.Context(), hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeAvailable(); err != nil {
				return nil, err
			}
			return h, nil
		})

		if err != nil {
			httperr.InternalError("unable-to-update-availability", err, w, r)
			return
		}
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

	hourUpdate := &openapi_types.HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		httperr.BadRequest("unable-to-update-availability", err, w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}
