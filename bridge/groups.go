package bridge

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type groupType string
type roomClass string

const (
	lightGroup groupType = "LightGroup"
	roomGroup  groupType = "Room"

	livingRoom     roomClass = "Living room"
	kitchenRoom    roomClass = "Kitchen"
	diningRoom     roomClass = "Dining"
	bedRoom        roomClass = "Bedroom"
	kidsBedroom    roomClass = "Kids bedroom"
	bathRoom       roomClass = "Bathroom"
	nurseryRoom    roomClass = "Nursery"
	recreationRoom roomClass = "Recreation"
	officeRoom     roomClass = "Office"
	gymRoom        roomClass = "Gym"
	hallwayRoom    roomClass = "Hallway"
	toiletRoom     roomClass = "Toilet"
	frontDoorRoom  roomClass = "Front door"
	garageRoom     roomClass = "Garage"
	terraceRoom    roomClass = "Terrace"
	gardenRoom     roomClass = "Garden"
	drivewayRoom   roomClass = "Driveway"
	carportRoom    roomClass = "Carport"
	otherRoom      roomClass = "Other"
	homeRoom       roomClass = "Home"
	downstairsRoom roomClass = "Downstairs"
	upstairsRoom   roomClass = "Upstairs"
	topFloorRoom   roomClass = "Top floor"
	atticRoom      roomClass = "Attic"
	guestRoom      roomClass = "Guest room"
	staircaseRoom  roomClass = "Staircase"
	loungeRoom     roomClass = "Lounge"
	manCaveRoom    roomClass = "Man cave"
	computerRoom   roomClass = "Computer"
	studioRoom     roomClass = "Studio"
	musicRoom      roomClass = "Music"
	tvRoom         roomClass = "TV"
	readingRoom    roomClass = "Reading"
	closetRoom     roomClass = "Closet"
	storageRoom    roomClass = "Storage"
	laundryRoom    roomClass = "Laundry room"
	balconyRoom    roomClass = "Balcony"
	porchRoom      roomClass = "Porch"
	barbecueRoom   roomClass = "Barbecue"
	poolRoom       roomClass = "Pool"
)

var allRooms = []roomClass{
	livingRoom, kitchenRoom, diningRoom, bedRoom, kidsBedroom, bathRoom,
	nurseryRoom, recreationRoom, officeRoom, gymRoom, hallwayRoom, toiletRoom,
	frontDoorRoom, garageRoom, terraceRoom, gardenRoom, drivewayRoom, carportRoom,
	otherRoom, homeRoom, downstairsRoom, upstairsRoom, topFloorRoom, atticRoom,
	guestRoom, staircaseRoom, loungeRoom, manCaveRoom, computerRoom, studioRoom,
	musicRoom, tvRoom, readingRoom, closetRoom, storageRoom, laundryRoom, balconyRoom,
	porchRoom, barbecueRoom, poolRoom,
}

func findRoomForGroup(g string) roomClass {
	for _, r := range allRooms {
		if strings.Contains(strings.ToLower(g), strings.ToLower(string(r))) {
			return r
		}
	}
	return otherRoom
}

type groupState struct {
	AllOn bool `json:"all_on"`
	AnyOn bool `json:"any_on"`
}

type group struct {
	Name    string     `json:"name"`
	Type    groupType  `json:"type"`
	Class   roomClass  `json:"class"`
	Lights  []string   `json:"lights"`
	Sensors []string   `json:"sensors"`
	State   groupState `json:"state"`
	Action  lightState `json:"action"`
	Recycle bool       `json:"recycle"`
}

func (*group) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type groups map[string]*group

func (groups) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type groupRenameReq struct {
	Name   *string  `json:"name"`
	Lights []string `json:"lights"`
	Class  *string  `json:"class"`

	no []string
}

func (req *groupRenameReq) Bind(r *http.Request) error {
	if req.Name != nil {
		req.no = append(req.no, "name")
	}
	if len(req.Lights) != 0 {
		req.no = append(req.no, "lights")
	}
	if req.Class != nil {
		req.no = append(req.no, "class")
	}
	return nil
}

func (s *Server) groupRename(w http.ResponseWriter, r *http.Request) {
	data := &groupRenameReq{}
	if err := render.Bind(r, data); err != nil {
		renderListOK(w, r, errInvalidJSON())
		return
	}
	msg := []render.Renderer{}
	for _, param := range data.no {
		msg = append(msg, errParameterReadOnly(r, param))
	}
	renderListOK(w, r, msg...)
}

func (s *Server) createGroups() groups {
	grps := map[string]*group{}
	devs := s.getAllLightsFromMQTT()
	for name, dev := range devs {
		grps[name] = &group{
			Name:    dev.Name,
			Type:    roomGroup,
			Class:   findRoomForGroup(dev.Name),
			Lights:  []string{name},
			Sensors: []string{},
			State: groupState{
				AllOn: dev.State.On,
				AnyOn: dev.State.On,
			},
			Action: dev.State,
		}
	}
	return grps
}

func (s *Server) getGroups(w http.ResponseWriter, r *http.Request) {
	g := s.createGroups()
	s.config.Lock()
	s.config.groups = len(g)
	s.config.Unlock()
	renderOK(w, r, g)
}

func (s *Server) getGroup(id string) *group {
	groups := s.createGroups()
	for name, g := range groups {
		if name == id {
			return g
		}
	}
	return nil
}

func (s *Server) groupByID(w http.ResponseWriter, r *http.Request) {
	groupID := chi.RouteContext(r.Context()).URLParam("groupID")
	group := s.getGroup(groupID)
	if group == nil {
		renderOK(w, r, errInvalidResource(r))
		return
	}
	renderOK(w, r, group)
}

func (s *Server) groupUpdateState(w http.ResponseWriter, r *http.Request) {
	groupID := chi.RouteContext(r.Context()).URLParam("groupID")
	group := s.getGroup(groupID)
	if group == nil {
		renderOK(w, r, errInvalidResource(r))
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

	res := []render.Renderer{}

	for _, l := range group.Lights {
		res = append(res,
			s.renderLightStateUpdate(s.updateLightState(s.getLight(l), data), true, groupID)...)
	}
	renderListOK(w, r, res...)
}
