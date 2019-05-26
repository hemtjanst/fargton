package bridge

import (
	"net/http"

	"github.com/go-chi/chi"
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
	LastUpdated string `json:lastupdated`
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
	sen := sensor{
		State: sensorState{
			Daylight:    BoolPtr(s.isDaylight()),
			LastUpdated: DateTimeToISO8600(now().UTC()),
		},
		Config: sensorConfig{
			On:            true,
			Longitude:     StrPtr("none"),
			Latitude:      StrPtr("none"),
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

	s.config.RLock()
	defer s.config.RUnlock()
	if h >= int(s.config.sunrise) && h < int(s.config.sunset) {
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
