package httperr

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/vasiliiperfilev/ddd/internal/common/logs"
)

func InternalError(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Internal server error", http.StatusInternalServerError)
}

func Unauthorised(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Unauthorised", http.StatusUnauthorized)
}

func BadRequest(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Bad request", http.StatusBadRequest)
}

func RateLimitExceeded(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Too Many Requests", http.StatusTooManyRequests)
}

func httpRespondWithError(err error, slug string, w http.ResponseWriter, r *http.Request, logMSg string, status int) {
	logs.GetLogEntry(r).WithError(err).WithField("error-slug", slug).Warn(logMSg)
	resp := ErrorResponse{slug, status}

	if err := render.Render(w, r, resp); err != nil {
		panic(err)
	}
}

type ErrorResponse struct {
	Slug       string `json:"slug"`
	httpStatus int
}

func (e ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(e.httpStatus)
	return nil
}
