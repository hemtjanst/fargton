package bridge // import "hemtjan.st/fargton/bridge"

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"lib.hemtjan.st/server"

	"go.uber.org/zap"
)

var errParamMissing = errors.New("missing required parameter")

// Server represents the HTTP API
type Server struct {
	config      *Config
	logger      *zap.Logger
	httpRouter  *chi.Mux
	httpsRouter *chi.Mux
	mqtt        *server.Manager
}

// NewServer returns a new Server
func NewServer(c *Config, m *server.Manager, l *zap.Logger) *Server {
	r1 := chi.NewRouter()
	r2 := chi.NewRouter()
	s := &Server{
		config:      c,
		logger:      l,
		httpRouter:  r1,
		httpsRouter: r2,
		mqtt:        m,
	}

	if s.config.authDisabled {
		s.logger.Info("authentication has been disabled")
	}

	for _, r := range []*chi.Mux{r1, r2} {
		r.MethodNotAllowed(s.methodNotAllowed)
		r.NotFound(s.resourceNotFound)
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(Logger(l))
		r.Use(middleware.Recoverer)
		r.Use(middleware.SetHeader("Pragma", "no-cache"))
		r.Use(middleware.SetHeader("Expires", "Mon, 1 Aug 2011 09:00:00 GMT"))
		r.Use(middleware.SetHeader("Access-Control-Max-Age", "3600"))
		r.Use(middleware.SetHeader("Cache-Control",
			"no-store, no-cache, must-revalidate, post-check=0, pre-check=0"))
		r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
		r.Use(middleware.SetHeader("Access-Control-Allow-Credentials", "true"))
		r.Use(middleware.SetHeader("Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, DELETE, HEAD"))
		r.Use(middleware.SetHeader("Access-Control-Allow-Headers", "Content-Type"))
	}

	r1.Route("/description.xml", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeXML))
		r.Get("/", s.descriptionXML)
	})

	r1.With(render.SetContentType(render.ContentTypeJSON)).
		Get("/api/nouser/config", s.getUnauthenticatedConfig)

	r2.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Post("/", s.registerUser)
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(s.Authenticate)
			r.Get("/", s.getConfigAndData)
			r.Get("/config", s.getAuthenticatedConfig)
			r.Delete("/config/whitelist/{deleteID}", s.deleteUser)
			r.Get("/lights/new", s.getNewLights)
			r.Get("/lights/{lightID}", s.lightByID)
			r.Put("/lights/{lightID}", s.lightRename)
			r.Put("/lights/{lightID}/state", s.lightUpdateState)
			r.Get("/lights", s.getLights)
			r.Post("/lights", s.searchLights)
			r.Get("/groups", s.getGroups)
			r.Get("/groups/{groupID}", s.groupByID)
			r.Put("/groups/{groupID}", s.groupRename)
			r.Put("/groups/{groupID}/action", s.groupUpdateState)
			r.Get("/schedules", s.getDummies)
			r.Get("/scenes", s.getDummies)
			r.Get("/sensors/new", s.getNewSensors)
			r.Put("/sensors/{sensorID}", s.sensorRename)
			r.Get("/sensors/{sensorID}", s.sensorByID)
			r.Get("/sensors", s.getAllSensors)
			r.Post("/sensors", s.searchSensors)
			r.Get("/rules", s.getDummies)
			r.Get("/resourcelinks", s.getDummies)
			r.Get("/capabilities", s.getCapabilities)
		})
	})

	return s
}

func createListener(c *Config, l *zap.Logger, withTLS bool) (net.Listener, error) {
	var listener net.Listener
	var err error
	var port int

	switch withTLS {
	case false:
		listener, err = net.Listen("tcp", c.address)
	case true:
		cert, err := tls.LoadX509KeyPair(c.tlsPubKey, c.tlsPrivKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS key/cert: %v", err)
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		listener, err = tls.Listen("tcp", c.tlsAddress, tlsCfg)
	}

	if err != nil {
		return nil, err
	}

	ap := strings.Split(listener.Addr().String(), ":")
	port, _ = strconv.Atoi(ap[len(ap)-1])

	switch withTLS {
	case false:
		c.port = uint16(port)
	case true:
		c.tlsPort = uint16(port)
	}

	l.Info(fmt.Sprintf(
		"server listening on: %s (tls: %t)", listener.Addr().String(), withTLS))
	return listener, nil
}

// Start the HTTP listener and return a shutdown function we can call
// to shut everything down again
func (s *Server) Start(sleep time.Duration) (func(context.Context), error) {
	wt, err := s.loadWhitelistFromFile()
	if err != nil {
		return nil, err
	}
	s.config.Lock()
	s.config.Whitelist = wt
	s.config.Unlock()

	listener, err := createListener(s.config, s.logger, false)
	if err != nil {
		return nil, err
	}

	listenerTLS, err := createListener(s.config, s.logger, true)
	if err != nil {
		return nil, err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	s.logger.Info("starting MQTT")
	go s.mqtt.Start(ctx)
	time.Sleep(sleep) // Sleep a bit so the discover cycle can complete
	s.logger.Info("started MQT")
	s.logger.Info("fetching initial device data from MQTT, this may take a bit...")
	devs := s.mqtt.DeviceByType("lightbulb")
	s.config.lights = len(devs)
	s.config.groups = len(devs)
	wgDev := sync.WaitGroup{}
	for _, dev := range devs {
		fts := []string{"on", "colorTemperature", "brightness", "hue", "saturation"}
		for _, ft := range fts {
			wgDev.Add(1)
			go func(ft string) {
				defer wgDev.Done()
				if dev.Feature(ft).Exists() {
					_ = dev.Feature(ft).Value()
				}
			}(ft)
		}
	}
	wgDev.Wait()
	s.logger.Info("done fetching device data")

	h1 := &http.Server{
		Handler: s.httpRouter,
	}
	s.logger.Info("initialising Hue HTTP REST API")
	go func() {
		if err := h1.Serve(listener); err != http.ErrServerClosed {
			s.logger.Fatal(err.Error())
		}
	}()
	s.logger.Info(fmt.Sprintf(
		"started Hue HTTP REST API server on http://%s", listener.Addr().String()))

	h2 := &http.Server{
		Handler: s.httpsRouter,
	}
	s.logger.Info("initialising Hue HTTPS REST API")
	go func() {
		if err := h2.Serve(listenerTLS); err != http.ErrServerClosed {
			s.logger.Fatal(err.Error())
		}
	}()
	s.logger.Info(fmt.Sprintf(
		"started Hue HTTPS REST API server on https://%s", listenerTLS.Addr().String()))

	s.logger.Info("initialising mDNS responder for Hue bridge discovery")
	rp, err := newMDNSResponder(s.config)
	if err != nil {
		s.logger.Fatal(err.Error())
	}
	go func() {
		err := rp.Respond(ctx)
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()
	s.logger.Info("started mDNS responder")

	s.logger.Info("initialising SSDP/UPnP responder")
	wg := sync.WaitGroup{}
	wg.Add(1)
	quitSSDP := make(chan bool)
	go func() {
		defer wg.Done()
		newSSDPResponder(s.config, s.logger, quitSSDP)
	}()
	s.logger.Info("started SSDP/UPnP responder")

	return func(ctx context.Context) {
		quitSSDP <- true
		wg.Wait()
		ctxCancel()
		s.logger.Info("stopped mDNS responder")
		s.logger.Info("stopped MQTT")
		h1.Shutdown(ctx)
		s.logger.Info("stopped Hue HTTP REST API server")
		h2.Shutdown(ctx)
		s.logger.Info("stopped Hue HTTPS REST API server")
	}, nil
}

type successResp struct {
	Success map[string]interface{} `json:"success"`
}

func (*successResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type lastScanResp struct {
	LastScan string `json:"lastscan"`
}

func (*lastScanResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r, errMethod(r))
}

func (s *Server) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r, errInvalidResource(r))
}

type configAndDataResp struct {
	Config        *authenticatedConfig `json:"config"`
	Lights        *lights              `json:"lights"`
	Groups        *groups              `json:"groups"`
	Scenes        *dummies             `json:"scenes"`
	Rules         *dummies             `json:"rules"`
	Schedules     *dummies             `json:"schedules"`
	ResourceLinks *dummies             `json:"resourcelinks"`
	Sensors       *sensors             `json:"sensors"`
}

func (*configAndDataResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) getConfigAndData(w http.ResponseWriter, r *http.Request) {
	config := createAuthenticatedConfig(s.config)
	devs := s.getAllLightsFromMQTT()
	groups := s.createGroups()
	sensors := sensors{
		"1": s.newDaylightSensor(),
	}
	d := s.createDummies()

	resp := &configAndDataResp{
		Config:        config,
		Lights:        &devs,
		Groups:        &groups,
		Scenes:        &d,
		Rules:         &d,
		Schedules:     &d,
		ResourceLinks: &d,
		Sensors:       &sensors,
	}

	renderOK(w, r, resp)
}

type rInfo struct {
	authenticated bool
	httpVersion   string
	method        string
	resource      string
	sourceIP      string
	tls           bool
	uid           string
}

func infoFromRequest(r *http.Request) *rInfo {
	info := &rInfo{}
	info.method = r.Method
	info.httpVersion = r.Proto
	info.sourceIP = r.RemoteAddr

	info.uid = chi.RouteContext(r.Context()).URLParam("userID")
	info.resource = strings.TrimPrefix(
		r.URL.Path,
		fmt.Sprintf("/api/%s", info.uid))
	if info.resource == "" {
		info.resource = "/"
	}

	if r.TLS != nil {
		info.tls = true
	}

	if auth := r.Context().Value(AuthenticatedCtxKey); auth != nil {
		info.authenticated = auth.(bool)
	}

	return info
}

func (s *Server) loadWhitelistFromFile() (*map[string]whitelist, error) {
	path := s.config.whitelistConfigPath
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		s.logger.Info(fmt.Sprintf("whitelist does not exist at %s", path))
		return &map[string]whitelist{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", path, err)
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read contents of %s: %v", path, err)
	}
	var wt map[string]whitelist
	err = json.Unmarshal(data, &wt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s as JSON: %v", path, err)
	}
	s.logger.Info(fmt.Sprintf("whitelist loaded from: %s", path))
	return &wt, nil
}

func (s *Server) saveWhitelistToFile() error {
	if s.config.whitelistConfigPath == "" {
		s.logger.Debug("no whitelist config path specified, not persisting to disk")
		return nil
	}
	data, err := json.Marshal(s.config.Whitelist)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %v", err)
	}
	path := s.config.whitelistConfigPath
	s.logger.Debug("writing file")
	err = ioutil.WriteFile(path, data, 0600)
	s.logger.Debug("wrote file")
	if err != nil {
		return fmt.Errorf("failed to write whitelist to %s: %v", path, err)
	}
	return nil
}

func renderAsList(renderers ...render.Renderer) []render.Renderer {
	list := []render.Renderer{}
	for _, ren := range renderers {
		list = append(list, ren)
	}
	return list
}

func renderOK(w http.ResponseWriter, r *http.Request, ren render.Renderer) {
	render.Status(r, http.StatusOK)
	render.Render(w, r, ren)
}

func renderListOK(w http.ResponseWriter, r *http.Request, rens ...render.Renderer) {
	rs := renderAsList(rens...)
	render.Status(r, http.StatusOK)
	render.RenderList(w, r, rs)
}
