package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/dart-scheduler/browser
var staticFiles embed.FS

// isMobileBrowser reports whether the User-Agent string belongs to a mobile
// browser. The "Mobi" substring is the de-facto standard indicator used by all
// major mobile browsers (Android Chrome, Safari on iPhone/iPod, Samsung
// Internet, etc.) and is recommended by MDN.
func isMobileBrowser(ua string) bool {
	return strings.Contains(ua, "Mobi")
}

// SPAHandler returns an HTTP handler that serves the embedded Angular SPA.
// All requests that don't match a static file are served index.html (for
// client-side routing).
func SPAHandler() http.Handler {
	sub, err := fs.Sub(staticFiles, "dist/dart-scheduler/browser")
	if err != nil {
		panic("web: embed sub-fs: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect mobile browsers landing on the desktop root to the mobile UI.
		if r.URL.Path == "/" && isMobileBrowser(r.Header.Get("User-Agent")) {
			http.Redirect(w, r, "/m", http.StatusFound)
			return
		}

		// Try to open the requested file.
		f, err := sub.Open(r.URL.Path[1:]) // strip leading "/"
		if err != nil {
			// Not found → serve index.html for SPA routing.
			r2 := *r
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, &r2)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
