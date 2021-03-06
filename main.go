package main

import (
	"crypto/tls"
	"flag"
	"github.com/mariusor/littr.go/app/db"
	"github.com/mariusor/littr.go/app/oauth"
	"github.com/writeas/go-nodeinfo"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/mariusor/littr.go/internal/errors"

	"github.com/mariusor/littr.go/app"
	"github.com/mariusor/littr.go/app/api"
	"github.com/mariusor/littr.go/app/frontend"
	"github.com/mariusor/littr.go/internal/log"
)

var version = "HEAD"

const defaultPort = 3000
const defaultTimeout = time.Second * 15

func main() {
	var wait time.Duration
	var port int
	var host string
	var env string

	flag.DurationVar(&wait, "graceful-timeout", defaultTimeout, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.IntVar(&port, "port", defaultPort, "the port on which we should listen on")
	flag.StringVar(&host, "host", "", "the host on which we should listen on")
	flag.StringVar(&env, "env", "unknown", "the environment type")
	flag.Parse()

	e := app.EnvType(env)
	app.Instance = app.New(host, port, e, version)

	errors.IncludeBacktrace = app.Instance.Config.Env == app.DEV

	db.Init(&app.Instance)
	defer db.Config.DB.Close()

	os, err := oauth.NewOAuth(
		db.Config.DB,
		app.Instance.Logger.New(log.Ctx{"package": "oauth"}),
	)
	if err != nil {
		app.Instance.Logger.Warn(err.Error())
	}

	front, err := frontend.Init(frontend.Config{
		Env:         e,
		Logger:      app.Instance.Logger.New(log.Ctx{"package": "frontend"}),
		Secure:      app.Instance.Secure,
		BaseURL:     app.Instance.BaseURL,
		HostName:    app.Instance.HostName,
		OAuthServer: os,
	})
	if err != nil {
		app.Instance.Logger.Warn(err.Error())
	}

	a := api.Init(api.Config{
		Logger:      app.Instance.Logger.New(log.Ctx{"package": "api"}),
		BaseURL:     app.Instance.APIURL,
		OAuthServer: os,
	})
	//processing.InitQueues(&app.Instance)
	//processing.Logger = app.Instance.Logger.Dev(log.Ctx{"package": "processing"})

	app.Logger = app.Instance.Logger.New(log.Ctx{"package": "app"})
	db.Logger = app.Instance.Logger.New(log.Ctx{"package": "db"})

	// Routes
	r := chi.NewRouter()
	r.Use(middleware.RequestID)

	if app.Instance.Config.Env == app.PROD {
		r.Use(middleware.Recoverer)
	} else {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Frontend
	r.With(a.Repository).Route("/", front.Routes())

	// API
	r.With(db.Repository).Route("/api", a.Routes())

	cfg := api.NodeInfoConfig()
	// Web-Finger
	r.With(db.Repository).Route("/.well-known", func(r chi.Router) {
		r.Use(app.NeedsDBBackend(a.HandleError))

		ni := nodeinfo.NewService(cfg, api.NodeInfoResolver{})
		r.Get("/webfinger", a.HandleWebFinger)
		r.Get("/host-meta", api.HandleHostMeta)
		r.Get("/nodeinfo", ni.NodeInfoDiscover)
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			a.HandleError(w, r, errors.NotFoundf("%s", r.RequestURI))
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		front.HandleErrors(w, r, errors.NotFoundf("%s", r.RequestURI))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		front.HandleErrors(w, r, errors.MethodNotAllowedf("%s not allowed", r.Method))
	})

	app.Instance.Run(r, wait)
}
