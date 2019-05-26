package bridge

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"lib.hemtjan.st/server"
)

const lightSWVersion = "1.46.13_r26312"
const manufacturer = "Philips"

type lightBulbModel string
type lightBulbType string
type lightBulbGamut string
type lightBulbProductName string

const (
	whiteModel       lightBulbModel = "LWB014"
	temperatureModel lightBulbModel = "LTW015"
	rgbModel         lightBulbModel = "LCT016"

	whiteType       lightBulbType = "Dimmable Light"
	temperatureType lightBulbType = "Color Temperature Light"
	rgbType         lightBulbType = "Extended Color Light"

	whiteGamut       lightBulbGamut = "-"
	temperatureGamut lightBulbGamut = "2200K-6500K"
	colorE26Gamut    lightBulbGamut = "B"

	whiteProductName       lightBulbProductName = "Hue White lamp"
	temperatureProductName lightBulbProductName = "Hue A19 White Ambiance"
	rgbProductName         lightBulbProductName = "Hue bulb A19"
)

type lightState struct {
	On             bool      `json:"on"`
	Brightness     int       `json:"bri"`
	XY             []float64 `json:"xy,omitempty"`
	Effect         string    `json:"effect"`
	MiredColorTemp int       `json:"ct,omitempty"`
	Alert          string    `json:"alert"`
	ColorMode      string    `json:"colormode"`
	Reachable      bool      `json:"reachable"`
	Mode           string    `json:"mode"`
}

type lightSWUpdate struct {
	State       string `json:"state"`
	LastInstall string `json:"lastinstall"`
}

type lightMiredColorTemperature struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type lightControl struct {
	MinDimLevel    int                         `json:"mindimlevel"`
	MaxLumen       int                         `json:"maxlumen"`
	ColorGamutType lightBulbGamut              `json:"colorgamuttype,omitempty"`
	ColorGamut     [][]float64                 `json:"colorgamut,omitempty"`
	MiredColorTemp *lightMiredColorTemperature `json:"ct,omitempty"`
}

type lightStreaming struct {
	Renderer bool `json:"renderer"`
	Proxy    bool `json:"proxy"`
}

type lightCapabilities struct {
	Certified bool            `json:"certified"`
	Control   *lightControl   `json:"control"`
	Streaming *lightStreaming `json:"streaming"`
}

type lightStartup struct {
	Mode       string `json:"mode"`
	Configured bool   `json:"configured"`
}

type lightConfig struct {
	Archetype string       `json:"archetype"`
	Function  string       `json:"function"`
	Direction string       `json:"direction"`
	Startup   lightStartup `json:"startup"`
}

type light struct {
	State            lightState           `json:"state"`
	SWUpdate         *lightSWUpdate       `json:"swupdate"`
	Type             lightBulbType        `json:"type"`
	Name             string               `json:"name"`
	Model            lightBulbModel       `json:"modelid"`
	ManufacturerName string               `json:"manufacturername"`
	ProductName      lightBulbProductName `json:"productname"`
	Capabilities     *lightCapabilities   `json:"capabilities"`
	Config           *lightConfig         `json:"config"`
	UUID             string               `json:"uniqueid"`
	SWVersion        string               `json:"swversion"`

	topic string
}

func (*light) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type lights map[string]*light

func (lights) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func newWhiteBulb(dev server.Device) (*light, error) {
	on, err := StringToBool(dev.Feature("on").Value())
	if err != nil {
		return nil, err
	}
	bri, err := StringToInt(dev.Feature("brightness").Value())
	if err != nil {
		return nil, err
	}

	l := &light{
		topic: dev.Info().Topic,
		State: lightState{
			On:         on,
			Brightness: ToPhilipsBrightness(bri),
			Reachable:  dev.IsReachable(),
			Mode:       "homeautomation",
			Effect:     "none",
			Alert:      "none",
		},
		SWUpdate: &lightSWUpdate{
			State:       "noupdates",
			LastInstall: DateTimeToISO8600(now().UTC()),
		},
		Type:             whiteType,
		Name:             dev.Name(),
		Model:            whiteModel,
		ManufacturerName: manufacturer,
		ProductName:      whiteProductName,
		Capabilities: &lightCapabilities{
			Certified: true,
			Control: &lightControl{
				MinDimLevel: 6000,
				MaxLumen:    900,
			},
			Streaming: &lightStreaming{
				Proxy:    true,
				Renderer: true,
			},
		},
		Config: &lightConfig{
			Archetype: "classicbulb",
			Function:  "functional",
			Direction: "omnidirectional",
			Startup: lightStartup{
				Mode:       "safety",
				Configured: true,
			},
		},
		SWVersion: lightSWVersion,
		UUID:      dev.Info().Topic,
	}
	return l, nil
}

func newColorTemperatureBulb(dev server.Device) (*light, error) {
	on, err := StringToBool(dev.Feature("on").Value())
	if err != nil {
		return nil, err
	}
	bri, err := StringToInt(dev.Feature("brightness").Value())
	if err != nil {
		return nil, err
	}
	ct, err := StringToInt(dev.Feature("colorTemperature").Value())
	if err != nil {
		return nil, err
	}

	l := &light{
		topic: dev.Info().Topic,
		State: lightState{
			On:             on,
			Brightness:     ToPhilipsBrightness(bri),
			MiredColorTemp: ct,
			ColorMode:      "ct",
			Mode:           "homeautomation",
			Reachable:      dev.IsReachable(),
			Effect:         "none",
			Alert:          "none",
		},
		SWUpdate: &lightSWUpdate{
			State:       "noupdates",
			LastInstall: DateTimeToISO8600(now().UTC()),
		},
		Type:             temperatureType,
		Name:             dev.Name(),
		Model:            temperatureModel,
		ManufacturerName: manufacturer,
		ProductName:      temperatureProductName,
		Capabilities: &lightCapabilities{
			Certified: true,
			Control: &lightControl{
				MinDimLevel: 6000,
				MaxLumen:    980,
				MiredColorTemp: &lightMiredColorTemperature{
					Min: 50,
					Max: 400,
				},
			},
			Streaming: &lightStreaming{
				Proxy:    true,
				Renderer: true,
			},
		},
		Config: &lightConfig{
			Archetype: "sultanbulb",
			Function:  "mixed",
			Direction: "omnidirectional",
			Startup: lightStartup{
				Mode:       "safety",
				Configured: true,
			},
		},
		SWVersion: lightSWVersion,
		UUID:      dev.Info().Topic,
	}
	return l, nil
}

func newRGBBulb(dev server.Device) (*light, error) {
	on, err := StringToBool(dev.Feature("on").Value())
	if err != nil {
		return nil, err
	}
	bri, err := StringToInt(dev.Feature("brightness").Value())
	if err != nil {
		return nil, err
	}
	hue, err := StringToInt(dev.Feature("hue").Value())
	if err != nil {
		return nil, err
	}
	sat, err := StringToInt(dev.Feature("saturation").Value())
	if err != nil {
		return nil, err
	}

	l := &light{
		topic: dev.Info().Topic,
		State: lightState{
			On:         on,
			Brightness: ToPhilipsBrightness(bri),
			ColorMode:  "xy",
			XY:         HemtjanstHStoCIExy(hue, sat),
			Mode:       "homeautomation",
			Reachable:  dev.IsReachable(),
			Effect:     "none",
			Alert:      "none",
		},
		SWUpdate: &lightSWUpdate{
			State:       "noupdates",
			LastInstall: DateTimeToISO8600(now().UTC()),
		},
		Type:             rgbType,
		Name:             dev.Name(),
		Model:            rgbModel,
		ManufacturerName: manufacturer,
		ProductName:      rgbProductName,
		Capabilities: &lightCapabilities{
			Certified: true,
			Control: &lightControl{
				MinDimLevel:    6000,
				MaxLumen:       600,
				ColorGamutType: "I",
				ColorGamut: [][]float64{
					{0.6812357, 0.318186},
					{0.3918985, 0.5250334},
					{0.1502415, 0.027116},
				},
			},
			Streaming: &lightStreaming{
				Proxy:    true,
				Renderer: true,
			},
		},
		Config: &lightConfig{
			Archetype: "sultanbulb",
			Function:  "mixed",
			Direction: "omnidirectional",
			Startup: lightStartup{
				Mode:       "safety",
				Configured: true,
			},
		},
		SWVersion: lightSWVersion,
		UUID:      dev.Info().Topic,
	}
	return l, nil
}

func (s *Server) getAllLightsFromMQTT() lights {
	devs := s.mqtt.DeviceByType("lightbulb")
	s.config.Lock()
	s.config.lights = len(devs)
	s.config.Unlock()
	bulbs := lights{}
	for _, l := range devs {
		if strings.HasPrefix(l.Info().Topic, "rpi") {
			continue
		}
		var b *light
		var err error
		if l.Feature("hue").Exists() || l.Feature("saturation").Exists() {
			b, err = newRGBBulb(l)
		} else if l.Feature("colorTemperature").Exists() {
			b, err = newColorTemperatureBulb(l)
		} else if l.Feature("brightness").Exists() {
			b, _ = newWhiteBulb(l)
		}
		if err != nil {
			s.logger.Error(err.Error(), zap.String("device", l.Info().Topic))
		}
		if b != nil {
			bulbs[TopicToStrInt(l.Info().Topic)] = b
		}
	}
	return bulbs
}

func (s *Server) getLight(id string) *light {
	devs := s.mqtt.DeviceByType("lightbulb")
	var l *light
	for _, d := range devs {
		if TopicToStrInt(d.Info().Topic) == id {
			var err error
			if d.Feature("hue").Exists() || d.Feature("saturation").Exists() {
				l, err = newRGBBulb(d)
			} else if d.Feature("colorTemperature").Exists() {
				l, err = newColorTemperatureBulb(d)
			} else if d.Feature("brightness").Exists() {
				l, err = newWhiteBulb(d)
			}
			if err != nil {
				s.logger.Error(err.Error(), zap.String("device", d.Info().Topic))
			}
			break
		}
	}
	return l
}

func (s *Server) getLights(w http.ResponseWriter, r *http.Request) {
	bulbs := s.getAllLightsFromMQTT()
	renderOK(w, r, bulbs)
}

func (s *Server) searchLights(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r,
		&successResp{Success: map[string]interface{}{"/lights": "Searching for new devices"}},
	)
}

func (s *Server) getNewLights(w http.ResponseWriter, r *http.Request) {
	t := now().UTC()
	renderOK(w, r, &lastScanResp{LastScan: DateTimeToISO8600(t)})
}

func (s *Server) lightByID(w http.ResponseWriter, r *http.Request) {
	lightID := chi.RouteContext(r.Context()).URLParam("lightID")
	l := s.getLight(lightID)
	if l == nil {
		renderListOK(w, r, errInvalidResource(r))
		return
	}
	renderOK(w, r, l)
}

func (s *Server) lightRename(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r, errParameterReadOnly(r, "name"))
}

type lightStateUpdate struct {
	On                  *bool      `json:"on"`
	Brightness          *int       `json:"bri"`
	Hue                 *int       `json:"hue"`
	Saturation          *int       `json:"sat"`
	XY                  *[]float64 `json:"xy"`
	ColorTemperature    *int       `json:"ct"`
	Alert               *string    `json:"alert"`
	Effect              *string    `json:"effect"`
	TransitionTime      *int       `json:"transitiontime"`
	BrightnessInc       *int       `json:"bri_inc"`
	HueInc              *int       `json:"hue_inc"`
	SaturationInc       *int       `json:"sat_inc"`
	ColorTemperatureInc *int       `json:"ct_inc"`
	XYInc               *[]float64 `json:"xy_inc"`

	no []string
}

func (upd *lightStateUpdate) Bind(r *http.Request) error {
	if upd.Hue != nil {
		upd.no = append(upd.no, "hue")
	}
	if upd.Saturation != nil {
		upd.no = append(upd.no, "sat")
	}
	if upd.Alert != nil {
		upd.no = append(upd.no, "alert")
	}
	if upd.Effect != nil {
		upd.no = append(upd.no, "effect")
	}
	if upd.TransitionTime != nil {
		upd.no = append(upd.no, "transitiontime")
	}
	if upd.HueInc != nil {
		upd.no = append(upd.no, "hue_inc")
	}
	if upd.SaturationInc != nil {
		upd.no = append(upd.no, "sat_inc")
	}
	if len(upd.no) > 0 {
		return fmt.Errorf("unsupported parameters %s", strings.Join(upd.no, ", "))
	}
	if upd.Brightness != nil && upd.BrightnessInc != nil {
		upd.BrightnessInc = nil
	}
	if upd.ColorTemperature != nil && upd.ColorTemperatureInc != nil {
		upd.ColorTemperatureInc = nil
	}
	if upd.XY != nil && upd.XYInc != nil {
		upd.XYInc = nil
	}
	return nil
}

func (s *Server) lightUpdateState(w http.ResponseWriter, r *http.Request) {
	lightID := chi.RouteContext(r.Context()).URLParam("lightID")
	l := s.getLight(lightID)
	if l == nil {
		renderListOK(w, r, renderAsList(errInvalidResource(r))...)
		return
	}
	data := &lightStateUpdate{}
	if err := render.Bind(r, data); err != nil {
		msg := []render.Renderer{}
		for _, param := range data.no {
			msg = append(msg, errParameterUnavailable(infoFromRequest(r).resource, param))
		}
		renderListOK(w, r, msg...)
		return
	}

	renderListOK(w, r,
		s.renderLightStateUpdate(s.updateLightState(l, data), false, lightID)...)
}

type lightUpdateStateResult struct {
	DeviceIsOff      []string
	InternalError    bool
	InvalidParameter []string
	Success          map[string]interface{}
}

func (s *Server) renderLightStateUpdate(l *lightUpdateStateResult, group bool, id string) []render.Renderer {
	list := []render.Renderer{}
	var resource string
	var action string
	if group {
		resource = fmt.Sprintf("/groups/%s", id)
		action = "action"
	} else {
		resource = fmt.Sprintf("/light/%s", id)
		action = "state"
	}
	if len(l.InvalidParameter) > 0 {
		for _, p := range l.InvalidParameter {
			list = append(list, errParameterUnavailable(resource, p))
		}
		return list
	}
	if len(l.DeviceIsOff) > 0 {
		for _, p := range l.DeviceIsOff {
			list = append(list, errDeviceIsOff(resource, p))
		}
		return list
	}
	if l.InternalError {
		return renderAsList(errInternalError(resource, "100"))
	}
	for param, value := range l.Success {
		list = append(list, &successResp{
			Success: map[string]interface{}{
				fmt.Sprintf("%s/%s/%s", resource, action, param): value,
			},
		})
	}
	return list
}

func (s *Server) updateLightState(light *light, state *lightStateUpdate) *lightUpdateStateResult {
	d := s.mqtt.Device(light.topic)
	lUpdate := &lightUpdateStateResult{
		Success: map[string]interface{}{},
	}

	switch light.Type {
	case whiteType:
		if state.ColorTemperature != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "ct")
		}
		if state.ColorTemperatureInc != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "ct_inc")
		}
		if state.XY != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "xy")
		}
		if state.XYInc != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "xy_inc")
		}
	case temperatureType:
		if state.XY != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "xy")
		}
		if state.XYInc != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "xy_inc")
		}
	case rgbType:
		if state.ColorTemperature != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "ct")
		}
		if state.ColorTemperatureInc != nil {
			lUpdate.InvalidParameter = append(lUpdate.InvalidParameter, "ct_inc")
		}
	}
	if len(lUpdate.InvalidParameter) > 0 {
		return lUpdate
	}

	on := false
	if d.Feature("on").Value() == "1" {
		on = true
	}
	params := []string{}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.Brightness != nil {
		params = append(params, "bri")
	}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.XY != nil {
		params = append(params, "xy")
	}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.ColorTemperature != nil {
		params = append(params, "ct")
	}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.BrightnessInc != nil {
		params = append(params, "bri_inc")
	}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.XYInc != nil {
		params = append(params, "xy_inc")
	}
	if !on && (state.On == nil || state.On != nil && !*state.On) && state.ColorTemperatureInc != nil {
		params = append(params, "ct_inc")
	}
	if len(params) > 0 {
		lUpdate.DeviceIsOff = params
		return lUpdate
	}

	/// Turn on first so any other characteristic updates will propagate
	if !on && state.On != nil && *state.On {
		d.Feature("on").Set("1")
		lUpdate.Success["on"] = true
	}
	if state.Brightness != nil {
		d.Feature("brightness").Set(IntToStr(ToHemtjanstBrightness(*state.Brightness)))
		lUpdate.Success["bri"] = *state.Brightness
	}
	if state.ColorTemperature != nil {
		d.Feature("colorTemperature").Set(IntToStr(*state.ColorTemperature))
		lUpdate.Success["ct"] = *state.ColorTemperature
	}
	if state.XY != nil {
		dt := *state.XY
		hue, sat := CIExyToHemtjanstHS(dt[0], dt[1])
		d.Feature("hue").Set(IntToStr(hue))
		d.Feature("saturation").Set(IntToStr(sat))
		lUpdate.Success["xy"] = *state.XY
	}
	if state.BrightnessInc != nil {
		v, err := StringToInt(d.Feature("brightness").Value())
		if err != nil {
			lUpdate.InternalError = true
			return lUpdate
		}
		d.Feature("brightness").Set(IntToStr(ToHemtjanstBrightness(ToPhilipsBrightness(v) + *state.BrightnessInc)))
		lUpdate.Success["bri_inc"] = *state.BrightnessInc
	}
	if state.ColorTemperatureInc != nil {
		v, err := StringToInt(d.Feature("colorTemperature").Value())
		if err != nil {
			lUpdate.InternalError = true
			return lUpdate
		}
		d.Feature("colorTemperature").Set(IntToStr(v + *state.ColorTemperatureInc))
		lUpdate.Success["ct_inc"] = *state.ColorTemperatureInc
	}
	if state.XYInc != nil {
		dt := *state.XYInc
		hue, err := StringToInt(d.Feature("hue").Value())
		if err != nil {
			lUpdate.InternalError = true
			return lUpdate
		}
		sat, err := StringToInt(d.Feature("saturation").Value())
		if err != nil {
			lUpdate.InternalError = true
			return lUpdate
		}
		xy := HemtjanstHStoCIExy(hue, sat)
		xN, yN := CIExyToHemtjanstHS(xy[0]+dt[0], xy[1]+dt[1])
		d.Feature("hue").Set(IntToStr(xN))
		d.Feature("saturation").Set(IntToStr(yN))
		lUpdate.Success["xy_inc"] = *state.XYInc
	}
	// Turn off last so we can update all the other characteristics first
	if on && state.On != nil && !*state.On {
		d.Feature("on").Set("0")
		lUpdate.Success["on"] = false
	}

	return lUpdate
}
