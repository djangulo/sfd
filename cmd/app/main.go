package main

import (
	"crypto/sha256"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	_ "github.com/go-bindata/go-bindata"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/djangulo/sfd/accounts"
	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/crypto/session"
	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/filters"

	// _ "github.com/djangulo/sfd/db/memory"
	_ "github.com/djangulo/sfd/db/postgres"
	"github.com/djangulo/sfd/items"
	"github.com/djangulo/sfd/mail"
	_ "github.com/djangulo/sfd/mail/console"
	"github.com/djangulo/sfd/pagination"
	"github.com/djangulo/sfd/storage"
	_ "github.com/djangulo/sfd/storage/fs"
)

var (
	dbDriver       db.Driver
	mailDriver     mail.Mailer
	cfg            config.Configurer
	storageDriver  storage.Driver
	sessionManager session.Manager
	tokenManager   token.Manager
	r              chi.Router
	port           string
	tlsCert        string
	tlsKey         string
	cpuprofile     string
	configFile     string
)

func init() {
	const (
		portUsage         = "Port to serve the app at."
		portDefault       = "9000"
		tlsCertUsage      = "TLS certificate to sign the app with."
		tlsCertDefault    = ""
		tlsKeyUsage       = "TLS key to sign the app with."
		tlsKeyDefault     = ""
		cpuprofileUsage   = "write cpu profile to file"
		cpuprofileDefault = ""
		configFileUsage   = "write cpu profile to file"
		configFileDefault = "./config/config.conf"
	)
	flag.StringVar(&port, "p", portDefault, portUsage)
	flag.StringVar(&port, "port", portDefault, portUsage)
	flag.StringVar(&cpuprofile, "cpuprofile", cpuprofileDefault, cpuprofileUsage)
	flag.StringVar(&tlsCert, "tls-cert", tlsCertDefault, tlsCertUsage)
	flag.StringVar(&tlsKey, "tls-key", tlsKeyDefault, tlsKeyUsage)
	flag.StringVar(&configFile, "c", tlsCertDefault, configFileDefault+" (shorthand)")
	flag.StringVar(&configFile, "config-file", tlsKeyDefault, configFileDefault)
	flag.Parse()

}

func main() {
	cfg = config.Get()

	var err error
	storageDriver, err = storage.Open(cfg.StorageURL())
	if err != nil {
		panic(err)
	}

	dbDriver, err = db.Open(cfg.DatabaseURL())
	if err != nil {
		panic(err)
	}

	mailDriver, err = mail.Open(cfg.EmailURL())
	if err != nil {
		panic(err)
	}

	sessionManager, err = session.NewManager(dbDriver, cfg.AuthCookieName(), 86400, cfg)
	if err != nil {
		panic(err)
	}

	tokenManager, err = token.NewManager(dbDriver, sha256.New, cfg)
	if err != nil {
		panic(err)
	}

	r = chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.StripSlashes)
	// r.Use(realIPPrinter)
	r.Use(middleware.Compress(
		5,
		"text/html",
		"text/css",
		"text/plain",
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
		"application/json",
		"application/atom+xml",
		"application/rss+xml",
		"image/svg+xml",
	))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*", "http://localhost:3000"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// _, err = core.NewServer(sessionManager)
	// if err != nil {
	// 	panic(err)
	// }
	r.Get("/", IndexHandler)
	ServeSPA(r, "/*", AssetFile())
	FileServer(r, storageDriver.Path(), storageDriver.Dir())
	itemServer, err := items.NewServer(dbDriver, storageDriver)
	if err != nil {
		panic(err)
	}
	accountsServer, err := accounts.NewServer(
		dbDriver, mailDriver, cfg, storageDriver, tokenManager, sessionManager)
	if err != nil {
		panic(err)
	}
	// register these two outside, as they're unlikely to change and needed
	// for global tasks, i.e. sending emails, so the URL needs to remain
	// manageable
	r.Get(cfg.PasswordResetEndpoint(), accountsServer.CheckPassResetToken)
	r.Get(cfg.EmailVerificationEndpoint(), accountsServer.VerifyEmailToken)
	r.With(sessionManager.NoErrContext).Post(cfg.LoginEndpoint(), accountsServer.LoginUser)
	r.With(sessionManager.NoErrContext).Route(cfg.LogoutEndpoint(), func(r chi.Router) {
		r.Get("/", accountsServer.Logout)
		r.Post("/", accountsServer.Logout)
	})
	r.Route("/api/v0.1.0", func(r chi.Router) {
		r.Route("/accounts", func(r chi.Router) {
			r.With(
				sessionManager.NoErrContext,
				accountsServer.UserContext).Get(
				`/{userID:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
				accounts.GetUser)
			r.With(
				sessionManager.NoErrContext,
				accountsServer.UserContext).Get(
				`/{username:[a-z0-9-._]{1,255}}`,
				accounts.GetUser)
			r.Route("/register", func(r chi.Router) {
				r.Post("/", accountsServer.RegisterUser)
			})
			r.Route("/password", func(r chi.Router) {
				r.Post("/init", accountsServer.PasswordResetInit)
				r.With(tokenManager.CSRFContext).Post("/confirm", accountsServer.PasswordResetConfirm)
			})
			r.Get("/state", accountsServer.RestoreUserState)
		})
		r.Route("/items", func(r chi.Router) {
			r.With(
				sessionManager.NoErrContext,
				pagination.Context,
				filters.Context,
				itemServer.ItemListCtx).Get("/", items.ListItems)
			r.With(sessionManager.Context).Post("/", itemServer.CreateItem)
			r.Route("/{itemID}", func(r chi.Router) {
				r.Use(itemServer.ItemCtx)
				r.With(itemServer.ItemImagesCtx).Get("/", items.GetItem)
				r.Post("/watch", itemServer.WatchItem)
				r.Post("/unwatch", itemServer.UnWatchItem)
				r.Route("/images", func(r chi.Router) {
					r.With(pagination.Context, itemServer.ItemImagesCtx).Get("/", items.ItemImages)
					r.Post("/", itemServer.AddItemImages)
				})
				r.Route("/bids", func(r chi.Router) {
					r.With(pagination.Context, itemServer.ItemBidsCtx).Get("/", items.ItemBids)
					r.Post("/", itemServer.PlaceBid)
					r.With(pagination.Context, itemServer.ItemBidsCtx).Get("/bids", items.ItemBids)
					r.Get("/winning", itemServer.ItemWinningBid)
					r.Route("/{bidID}", func(r chi.Router) {
						r.Use(itemServer.BidCtx)
						r.Get("/", items.GetBid)
						r.Post("/invalidate", itemServer.InvalidateBid)
					})
				})
			})
		})
	})
	// r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	http.RedirectHandler("/index.html", 301).ServeHTTP(w, r)
	// }))

	// _, err = admin.NewServer(dbDriver, mailDriver, cfg, sm, tm)
	// if err != nil {
	// 	panic(err)
	// }
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	cancel := make(chan struct{})
	eChan := make(chan error)

	go sessionManager.GC(1*time.Hour, eChan, cancel)
	go tokenManager.GC(1*time.Minute, eChan, cancel)

	go func() {
		for {
			select {
			case e := <-eChan:
				log.Println(e)
			}
		}
	}()

	// cancel on panic
	defer func() {
		if rec := recover(); rec != nil {
			cancel <- struct{}{}
		}
		close(cancel)
		close(eChan)
	}()

	log.Println("Listening on :" + port)
	if tlsCert != "" && tlsKey != "" {
		log.Fatal(http.ListenAndServeTLS(":"+port, tlsCert, tlsKey, r))
	}
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// FileServer serve static files the chi way.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")

		w.Header().Set("Cache-Control", "max-age:2592000, public")
		uPath := r.URL.Path
		SetContentType(w, uPath)
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

// IndexHandler grabs the index.html from the go-bindata assets
// and serves its content.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	idx, err := AssetFile().Open("index.html")
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	finfo, err := idx.Stat()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.ServeContent(w, r, "index.html", finfo.ModTime(), idx)
}

// ServeSPA is meant to work as a catch-all for a single-page application.
// Delegates to client-side routing if the server cannot find the requested file.
// If the path leads to a file that exists in root (http.FileSystem), then
// said file is served, otherwise IndexHandler is called to delegate routing
// to the single-page application.
func ServeSPA(r chi.Router, path string, root http.FileSystem) {
	if !strings.Contains(path, "*") {
		panic("ServeSPA must be a catch-all route, but it does not contain a \"*\".")
	}

	if strings.ContainsAny(path, "{}") {
		panic("ServeSPA does not permit any URL parameters.")
	}

	if path != "/*" && path[(len(path)-2):(len(path)-1)] != "/*" {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/*"
	}

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")

		w.Header().Set("Cache-Control", "max-age:2592000, public")
		uPath := r.URL.Path
		file, err := root.Open(uPath)
		if err != nil {
			// No SPA file has been found, redirect the request to
			// the index.html
			IndexHandler(w, r)
			return
		}
		finfo, err := file.Stat()
		if err != nil {
			// file not found, error out
			http.Error(w, err.Error(), 404)
			return
		}
		SetContentType(w, finfo.Name())
		// http.ServeContent(w, r, finfo.Name(), finfo.ModTime(), file)
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

// SetContentType resolves the content-type based on the extension of path.
func SetContentType(w http.ResponseWriter, path string) {
	switch filepath.Ext(path) {
	case ".aac":
		w.Header().Set("Content-Type", "audio/aac")
	case ".abw":
		w.Header().Set("Content-Type", "application/x-abiword")
	case ".arc":
		w.Header().Set("Content-Type", "application/x-freearc")
	case ".avi":
		w.Header().Set("Content-Type", "video/x-msvideo")
	case ".azw":
		w.Header().Set("Content-Type", "application/vnd.amazon.ebook")
	case ".bin":
		w.Header().Set("Content-Type", "application/octet-stream")
	case ".bmp":
		w.Header().Set("Content-Type", "image/bmp")
	case ".bz":
		w.Header().Set("Content-Type", "application/x-bzip")
	case ".bz2":
		w.Header().Set("Content-Type", "application/x-bzip2")
	case ".csh":
		w.Header().Set("Content-Type", "application/x-csh")
	case ".csv":
		w.Header().Set("Content-Type", "text/csv")
	case ".doc":
		w.Header().Set("Content-Type", "application/msword")
	case ".docx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	case ".epub":
		w.Header().Set("Content-Type", "application/epub+zip")
	case ".gz":
		w.Header().Set("Content-Type", "application/gzip")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".htm":
		w.Header().Set("Content-Type", "text/html")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".ico":
		w.Header().Set("Content-Type", "image/vnd.microsoft.icon")
	case ".ics":
		w.Header().Set("Content-Type", "text/calendar")
	case ".jar":
		w.Header().Set("Content-Type", "application/java-archive")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".jsonld":
		w.Header().Set("Content-Type", "application/ld+json")
	case ".midi", ".mid":
		w.Header().Set("Content-Type", "audio/midi")
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".mpeg":
		w.Header().Set("Content-Type", "video/mpeg")
	case ".mpkg":
		w.Header().Set("Content-Type", "application/vnd.apple.installer+xml")
	case ".odp":
		w.Header().Set("Content-Type", "application/vnd.oasis.opendocument.presentation")
	case ".ods":
		w.Header().Set("Content-Type", "application/vnd.oasis.opendocument.spreadsheet")
	case ".odt":
		w.Header().Set("Content-Type", "application/vnd.oasis.opendocument.text")
	case ".oga":
		w.Header().Set("Content-Type", "audio/ogg")
	case ".ogv":
		w.Header().Set("Content-Type", "video/ogg")
	case ".ogx":
		w.Header().Set("Content-Type", "application/ogg")
	case ".opus":
		w.Header().Set("Content-Type", "audio/opus")
	case ".otf":
		w.Header().Set("Content-Type", "font/otf")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case ".php":
		w.Header().Set("Content-Type", "application/x-httpd-php")
	case ".ppt":
		w.Header().Set("Content-Type", "application/vnd.ms-powerpoint")
	case ".pptx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	case ".rar":
		w.Header().Set("Content-Type", "application/vnd.rar")
	case ".rtf":
		w.Header().Set("Content-Type", "application/rtf")
	case ".sh":
		w.Header().Set("Content-Type", "application/x-sh")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".swf":
		w.Header().Set("Content-Type", "application/x-shockwave-flash")
	case ".tar":
		w.Header().Set("Content-Type", "application/x-tar")
	case ".tiff", ".tif":
		w.Header().Set("Content-Type", "image/tiff")
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".vsd":
		w.Header().Set("Content-Type", "application/vnd.visio")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".weba":
		w.Header().Set("Content-Type", "audio/webm")
	case ".webm":
		w.Header().Set("Content-Type", "video/webm")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".xhtml":
		w.Header().Set("Content-Type", "application/xhtml+xml")
	case ".xls":
		w.Header().Set("Content-Type", "application/vnd.ms-excel")
	case ".xlsx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	case ".xml":
		w.Header().Set("Content-Type", "application/xml")
	case ".xul":
		w.Header().Set("Content-Type", "application/vnd.mozilla.xul+xml")
	case ".zip":
		w.Header().Set("Content-Type", "application/zip")
	case ".3gp":
		w.Header().Set("Content-Type", "video/3gpp")
	case ".3g2":
		w.Header().Set("Content-Type", "video/3gpp2")
	case ".7z":
		w.Header().Set("Content-Type", "application/x-7z-compressed")
	case ".js", ".mjs":
		w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	default:
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
}
