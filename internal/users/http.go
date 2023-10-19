package main

import (
	"net"
	"net/http"

	"github.com/go-chi/render"
	"github.com/vasiliiperfilev/ddd/internal/common/auth"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
	"github.com/vasiliiperfilev/ddd/internal/common/server/httperr"
	"github.com/vasiliiperfilev/ddd/internal/users/domain/user"
	"github.com/vasiliiperfilev/ddd/internal/users/oapi"
)

type HttpServer struct {
	userRepository user.Repository
	backgroundJobs *server.BackgroundJobs
}

func (h HttpServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.Unauthorised("no-user-found", err, w, r)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		err = h.userRepository.Update(r.Context(), authUser.UUID, func(u user.User) (user.User, error) {
			u.UpdateLastIP(host)
			return u, nil
		})
		if err != nil {
			httperr.InternalError("internal-server-error", err, w, r)
			return
		}
	}

	user, err := h.userRepository.Get(r.Context(), authUser.UUID)
	if err != nil {
		httperr.InternalError("cannot-get-user", err, w, r)
		return
	}

	userResponse := oapi.User{
		DisplayName: authUser.DisplayName,
		Balance:     user.Balance,
		Role:        authUser.Role,
	}

	render.Respond(w, r, userResponse)
}
