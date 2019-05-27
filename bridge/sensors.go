package bridge

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/kelvins/sunrisesunset"
	"go.uber.org/zap"
)

type sensorConfig struct {
	On            bool    `json:"on"`
	Reachable     *bool   `json:"reachable,omitempty"`
	Battery       *uint8  `json:"battery,omitempty"`
	Longitude     *string `json:"long,omitempty"`
	Latitude      *string `json:"lat,omitempty"`
	SunriseOffset *int    `json:"sunriseoffset,omitempty"`
	SunsetOffset  *int    `json:"sunsetoffset,omitempty"`
}

type sensorState struct {
	Daylight    *bool  `json:"daylight,omitempty"`
	ButtonEvent *int   `json:"buttonevent,omitempty"`
	LastUpdated string `json:"lastupdated"`
}

type sensor struct {
	State            sensorState  `json:"state"`
	Config           sensorConfig `json:"config"`
	Name             string       `json:"name"`
	Type             string       `json:"type"`
	ModelID          string       `json:"modelid"`
	ManufacturerName string       `json:"manufacturername"`
	SWVersion        *string      `json:"swversion,omitempty"`
}

func (sensor) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type sensors map[string]sensor

func (sensors) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) newDaylightSensor() sensor {
	s.config.RLock()
	lat := s.config.latitude
	long := s.config.longitude
	s.config.RUnlock()

	var lats *string
	if math.Signbit(lat) {
		lats = StrPtr(fmt.Sprintf("%.4fS", lat*-1))
	} else {
		lats = StrPtr(fmt.Sprintf("%.4fN", lat))
	}

	var longs *string
	if math.Signbit(long) {
		longs = StrPtr(fmt.Sprintf("%.4fW", long*-1))
	} else {
		longs = StrPtr(fmt.Sprintf("%.4fE", long))
	}

	sen := sensor{
		State: sensorState{
			Daylight:    BoolPtr(s.isDaylight()),
			LastUpdated: DateTimeToISO8600(now().UTC()),
		},
		Config: sensorConfig{
			On:            true,
			Latitude:      lats,
			Longitude:     longs,
			SunriseOffset: IntPtr(0),
			SunsetOffset:  IntPtr(0),
		},
		Name:             "Daylight",
		Type:             "Daylight",
		ModelID:          "PHDL00",
		ManufacturerName: "Philips",
		SWVersion:        StrPtr("1.0"),
	}

	return sen
}

func (s *Server) isDaylight() bool {
	t := now().UTC()
	h := t.Hour()
	m := t.Minute()
	year, month, day := t.Date()

	s.config.RLock()
	p := sunrisesunset.Parameters{
		Latitude:  s.config.latitude,
		Longitude: s.config.longitude,
		UtcOffset: 0.0,
		Date:      time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
	s.config.RUnlock()

	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		s.logger.Error(err.Error())
		return false
	}

	sunriseT := time.Date(year, month, day, sunrise.Hour(), sunrise.Minute(), 0, 0, time.UTC)
	sunsetT := time.Date(year, month, day, sunset.Hour(), sunset.Minute(), 0, 0, time.UTC)

	s.logger.Debug("daylight",
		zap.String("sunrise", fmt.Sprintf("%02d:%02d", sunrise.Hour(), sunrise.Minute())),
		zap.String("sunset", fmt.Sprintf("%02d:%02d", sunset.Hour(), sunset.Minute())),
		zap.String("current", fmt.Sprintf("%02d:%02d", h, m)))

	if t.After(sunriseT) && t.Before(sunsetT) {
		return true
	}

	return false
}

func (s *Server) getNewSensors(w http.ResponseWriter, r *http.Request) {
	t := now().UTC()
	renderOK(w, r, &lastScanResp{LastScan: DateTimeToISO8600(t)})
}

func (s *Server) searchSensors(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r,
		&successResp{Success: map[string]interface{}{"/sensors": "Searching for new devices"}},
	)
}

func (s *Server) sensorRename(w http.ResponseWriter, r *http.Request) {
	renderListOK(w, r, errParameterReadOnly(r, "name"))
}

func (s *Server) sensorByID(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.RouteContext(r.Context()).URLParam("sensorID")
	if sensorID != "1" {
		renderListOK(w, r, errInvalidResource(r))
		return
	}
	renderOK(w, r, s.newDaylightSensor())
}

func (s *Server) getAllSensors(w http.ResponseWriter, r *http.Request) {
	sensors := sensors{
		"1": s.newDaylightSensor(),
	}
	renderOK(w, r, sensors)
}
