package bridge

import (
	"net/http"
)

type capacity struct {
	Available int  `json:"available"`
	Total     int  `json:"total"`
	Channels  *int `json:"channels,omitempty"`
}

type sensorsCapacity struct {
	capacity
	Clip capacity `json:"clip"`
	ZLL  capacity `json:"zll"`
	ZGP  capacity `json:"zgp"`
}

type scenesCapacity struct {
	capacity
	LightStates capacity `json:"lightstates"`
}

type rulesCapacity struct {
	capacity
	Actions capacity `json:"actions"`
}

type capabilities struct {
	Lights        capacity            `json:"lights"`
	Sensors       sensorsCapacity     `json:"sensors"`
	Groups        capacity            `json:"groups"`
	Scenes        scenesCapacity      `json:"scenes"`
	Rules         capacity            `json:"rules"`
	Schedules     capacity            `json:"schedules"`
	ResourceLinks capacity            `json:"resourcelinks"`
	Whitelists    capacity            `json:"whitelists"`
	Timezones     map[string][]string `json:"timezones"`
	Streaming     capacity            `json:"streaming"`
}

func (capabilities) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) getCapabilities(w http.ResponseWriter, r *http.Request) {
	c := capabilities{
		Lights: capacity{
			Available: 0,
			Total:     s.config.lights,
		},
		Groups: capacity{
			Available: 0,
			Total:     s.config.groups,
		},
		Whitelists: capacity{
			Available: 1000,
			Total:     len(*s.config.Whitelist),
		},
		Timezones: map[string][]string{"values": []string{"UTC"}},
		Streaming: capacity{
			Available: 0,
			Total:     0,
			Channels:  IntPtr(0),
		},
	}
	renderOK(w, r, c)
}
