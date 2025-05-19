package auth

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/go-oauth2/oauth2/v4/server"
	"golang.org/x/net/websocket"
)

func AuthorizeHandler(db *sql.DB, srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := GetUsernameFromSession(db, r)
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

func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "relay.tdccore.nl" {
			http.Error(w, "Invalid host", http.StatusForbidden)
			return
		}
		_, err := GetUsernameFromSession(db, r)
		if err != nil {
			redirect := "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, redirect, http.StatusFound)
			return
		}

		next(w, r)
	}
}

func RequireAuthWS(db *sql.DB, ws websocket.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := GetUsernameFromSession(db, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ws.ServeHTTP(w, r)
	})
}
