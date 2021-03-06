package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/cloudfoundry"
)

// Check if the user is logged in, otherwise forward to login page.
func rootHandler(res http.ResponseWriter, req *http.Request) {
	s, _ := gothic.Store.Get(req, "uaa-proxy-session")
	if s.Values["logged"] != true {
		http.Redirect(res, req, "/auth/cloudfoundry", http.StatusTemporaryRedirect)
		return
	}

	newProxy(s.Values["user_email"].(string)).ServeHTTP(res, req)
}

// Handle callbacks from oauth.
func callbackHandler(res http.ResponseWriter, req *http.Request) {

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Fprintln(res, err)
		return
	}

	s, err := gothic.Store.Get(req, "uaa-proxy-session")
	s.Values["user_email"] = user.Email
	s.Values["logged"] = true
	gothic.Store.Save(req, res, s)

	http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
}

func newProxy(remote_user string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			forwardedURL := req.Header.Get(CF_FORWARDED_URL)
			url, err := url.Parse(forwardedURL)
			if err != nil {
				log.Fatalln(err.Error())
			}
			req.URL = url
			req.Host = url.Host
			req.Header.Add("X-Auth-User", remote_user)

			fmt.Println(req.Header)
		},
	}
	return proxy
}

func setProviders(callbackURL string) {
	goth.UseProviders(
		cloudfoundry.New(c.UAAUrl, c.ClientKey, c.ClientSecret, callbackURL, "openid"),
	)
}
