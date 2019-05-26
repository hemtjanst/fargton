package bridge

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type registrationReq struct {
	DeviceType        string `json:"devicetype"`
	GenerateClientKey *bool  `json:"generateclientkey"`
}

func (rr *registrationReq) Bind(r *http.Request) error {
	if rr.DeviceType == "" {
		return errParamMissing
	}
	return nil
}

type registrationResp struct {
	Success struct {
		Username string `json:"username"`
	} `json:"success"`
}

func (rr *registrationResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// RegisterUser adds the user to the whitelist
func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	data := &registrationReq{}
	if err := render.Bind(r, data); err != nil {
		if err == errParamMissing {
			renderListOK(w, r, errMissingParameter(r))
			return
		}
		renderListOK(w, r, errInternalError(infoFromRequest(r).resource, "100"))
		return
	}

	u := uuid.New().String()
	entry := whitelist{
		Name:       data.DeviceType,
		ID:         u,
		CreatedAt:  DateTimeToISO8600(now().UTC()),
		LastUsedAt: DateTimeToISO8600(now().UTC()),
	}

	s.config.Lock()
	defer s.config.Unlock()
	wt := *s.config.Whitelist
	oldwt := map[string]whitelist{}
	copier.Copy(&oldwt, &wt)
	wt[u] = entry
	err := s.saveWhitelistToFile()
	if err != nil {
		s.logger.Error(err.Error())
		s.config.Whitelist = &oldwt
		renderListOK(w, r, errInternalError(infoFromRequest(r).resource, "100"))
		return
	}
	s.config.Whitelist = &wt

	regResp := &registrationResp{
		Success: struct {
			Username string `json:"username"`
		}{
			Username: u,
		},
	}
	renderListOK(w, r, regResp)
}

type deleteResp struct {
	Success string `json:"success"`
}

func (*deleteResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	s.config.Lock()
	defer s.config.Unlock()

	deleteID := chi.RouteContext(r.Context()).URLParam("deleteID")
	wt := *s.config.Whitelist
	oldwt := map[string]whitelist{}
	copier.Copy(&oldwt, &wt)
	delete(wt, deleteID)
	err := s.saveWhitelistToFile()
	if err != nil {
		s.logger.Error(err.Error())
		s.config.Whitelist = &oldwt
		renderListOK(w, r, errInternalError(infoFromRequest(r).resource, "100"))
		return
	}
	s.config.Whitelist = &wt
	renderListOK(w, r, &deleteResp{Success: fmt.Sprintf("/config/whitelist/%s deleted.", deleteID)})
}
