package auth

import (
	"net/http"
	"net/url"

	"github.com/go-oauth2/oauth2/v4/server"
)

func AuthorizeHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := GetUsernameFromSession(r)
		if err != nil {
			q := url.Values{}
			q.Set("redirect", r.URL.String())
			http.Redirect(w, r, "/login?"+q.Encode(), http.StatusFound)
			return
		}

		srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
			return username, nil
		})

		err = srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
