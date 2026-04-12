package appcore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/utils"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	shortenedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortened_total",
			Help: "Total number of short links created",
		},
	)
	redirectCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_redirect_total",
			Help: "Total number of redirects",
		},
	)
	metricsOnce sync.Once
)

type Server struct {
	db *sql.DB
}

func NewServer() (*Server, error) {
	rand.Seed(time.Now().UnixNano())
	metricsOnce.Do(func() {
		prometheus.MustRegister(shortenedCounter, redirectCounter)
	})

	db, err := repository.NewDB()
	if err != nil {
		return nil, err
	}

	if err := ensureSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Server{db: db}, nil
}

func (s *Server) DynamicHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, code, err := resolveRoute(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch route {
		case "health":
			s.healthHandler(w, r)
		case "shorten":
			s.shortenHandler(w, r)
		case "metrics":
			promhttp.Handler().ServeHTTP(w, r)
		case "analytics":
			s.analyticsHandler(w, r, code)
		case "all":
			s.getAllLinksHandler(w, r)
		case "update":
			s.updateHandler(w, r, code)
		case "delete":
			s.deleteHandler(w, r, code)
		case "redirect":
			s.redirectHandler(w, r, code)
		default:
			http.NotFound(w, r)
		}
	})
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
		id TEXT PRIMARY KEY,
		original TEXT NOT NULL,
		short_code TEXT UNIQUE NOT NULL,
		custom_code TEXT,
		domain TEXT,
		visibility TEXT,
		created_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS visits (
		id SERIAL PRIMARY KEY,
		url_id TEXT REFERENCES urls(id) ON DELETE CASCADE,
		timestamp TEXT,
		ip_address TEXT,
		country TEXT,
		browser TEXT,
		device TEXT
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

func resolveRoute(r *http.Request) (route string, code string, err error) {
	if explicit := r.URL.Query().Get("route"); explicit != "" {
		return explicit, r.URL.Query().Get("code"), nil
	}

	switch {
	case r.URL.Path == "/health":
		return "health", "", nil
	case r.URL.Path == "/shorten":
		return "shorten", "", nil
	case r.URL.Path == "/metrics":
		return "metrics", "", nil
	case r.URL.Path == "/all":
		return "all", "", nil
	case strings.HasPrefix(r.URL.Path, "/analytics/"):
		return "analytics", strings.TrimPrefix(r.URL.Path, "/analytics/"), nil
	case strings.HasPrefix(r.URL.Path, "/update/"):
		return "update", strings.TrimPrefix(r.URL.Path, "/update/"), nil
	case strings.HasPrefix(r.URL.Path, "/delete/"):
		return "delete", strings.TrimPrefix(r.URL.Path, "/delete/"), nil
	case strings.HasPrefix(r.URL.Path, "/r/"):
		return "redirect", strings.TrimPrefix(r.URL.Path, "/r/"), nil
	default:
		return "", "", errors.New("route not found")
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func baseURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return scheme + "://" + host
}

func (s *Server) analyticsHandler(w http.ResponseWriter, r *http.Request, code string) {
	if code == "" {
		http.Error(w, "Analytics code required", http.StatusBadRequest)
		return
	}

	url, err := repository.GetURLByCode(s.db, code)
	if err != nil {
		http.Error(w, "Short code not found", http.StatusNotFound)
		return
	}

	if url.Visibility == "private" {
		http.Error(w, "Analytics not available for private URLs", http.StatusForbidden)
		return
	}

	visits, err := repository.GetVisitsByURLID(s.db, url.ID)
	if err != nil {
		http.Error(w, "Could not fetch visits", http.StatusInternalServerError)
		return
	}

	if len(visits) == 0 {
		fmt.Fprintln(w, "No visits yet.")
		return
	}

	for i, v := range visits {
		fmt.Fprintf(w, "Visit %d:\n", i+1)
		fmt.Fprintf(w, "  IP        : %s\n", v.IPAddress)
		fmt.Fprintf(w, "  Country   : %s\n", v.Country)
		fmt.Fprintf(w, "  Timestamp : %s\n", v.Timestamp)
		fmt.Fprintf(w, "  Browser   : %s\n", v.Browser)
		fmt.Fprintf(w, "  Device    : %s\n\n", v.Device)
	}
}

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	original := r.PostFormValue("url")
	if original == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	requestedCode := r.FormValue("code")
	shortCode := requestedCode
	if shortCode != "" {
		if _, err := repository.GetURLByCode(s.db, shortCode); err == nil {
			http.Error(w, "Custom short code already in use", http.StatusBadRequest)
			return
		}
	} else {
		shortCode = generateUniqueShortCode(s.db, 6)
	}

	visibility := r.PostFormValue("visibility")
	if visibility != "private" {
		visibility = "public"
	}

	url := model.URL{
		ID:         uuid.New().String(),
		Original:   original,
		ShortCode:  shortCode,
		Visibility: visibility,
		CreatedAt:  model.Now(),
	}

	if err := repository.SaveURL(s.db, url); err != nil {
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}
	shortenedCounter.Inc()

	root := baseURL(r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"short_url":     root + "/r/" + shortCode,
		"analytics_url": root + "/analytics/" + shortCode,
	})
}

func (s *Server) redirectHandler(w http.ResponseWriter, r *http.Request, code string) {
	if code == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	url, err := repository.GetURLByCode(s.db, code)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	ip := getClientIP(r)
	if ip == "::1" || ip == "127.0.0.1" {
		ip = "103.48.198.141"
	}

	ua := r.UserAgent()
	browser, osName, device := utils.ParseUserAgent(ua)
	location, _ := utils.GetLocation(ip)

	visit := model.Visit{
		URLID:     url.ID,
		Timestamp: time.Now(),
		IPAddress: ip,
		Country:   location.Country,
		Browser:   browser,
		OS:        osName,
		Device:    device,
	}

	_ = repository.SaveVisit(s.db, visit)
	redirectCounter.Inc()
	http.Redirect(w, r, url.Original, http.StatusFound)
}

func generateShortCode(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func generateUniqueShortCode(db *sql.DB, length int) string {
	for {
		code := generateShortCode(length)
		exists, err := repository.ShortCodeExists(db, code)
		if err != nil {
			continue
		}
		if !exists {
			return code
		}
	}
}

func (s *Server) getAllLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	links, err := repository.GetAllLinks(s.db)
	if err != nil {
		http.Error(w, "Failed to fetch links", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if code == "" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	var req struct {
		LongURL    string `json:"long_url"`
		Visibility string `json:"visibility"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	url, err := repository.GetURLByCode(s.db, code)
	if err != nil {
		http.Error(w, "Short code not found", http.StatusNotFound)
		return
	}

	if req.LongURL != "" {
		url.Original = req.LongURL
	}
	if req.Visibility != "" {
		url.Visibility = req.Visibility
	}

	if err := repository.UpdateURL(s.db, url); err != nil {
		http.Error(w, "Failed to update URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Updated short link '%s'", code)
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if code == "" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	if _, err := repository.GetURLByCode(s.db, code); err != nil {
		http.Error(w, "Short code not found", http.StatusNotFound)
		return
	}

	if err := repository.DeleteURLByCode(s.db, code); err != nil {
		http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Deleted short link '%s'", code)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"db unreachable"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
