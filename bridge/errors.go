package bridge

import (
	"fmt"
	"net/http"
)

type innerErrResp struct {
	Type        int    `json:"type"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type errorResp struct {
	Error innerErrResp `json:"error"`
}

func (rr *errorResp) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func errUnauthorized(r *http.Request) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        1,
			Address:     infoFromRequest(r).resource,
			Description: "unauthorized user",
		},
	}
}

func errInvalidJSON() *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        2,
			Address:     "",
			Description: "body contains invalid json",
		},
	}
}

func errInvalidResource(r *http.Request) *errorResp {
	rs := infoFromRequest(r).resource
	return &errorResp{
		Error: innerErrResp{
			Type:        3,
			Address:     rs,
			Description: fmt.Sprintf("resource, %s, not available", rs),
		},
	}
}

func errMethod(r *http.Request) *errorResp {
	info := infoFromRequest(r)
	return &errorResp{
		Error: innerErrResp{
			Type:    4,
			Address: info.resource,
			Description: fmt.Sprintf(
				"method, %s, not allowed for resource, %s",
				info.method,
				info.resource),
		},
	}
}

func errMissingParameter(r *http.Request) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        5,
			Address:     infoFromRequest(r).resource,
			Description: "invalid/missing parameters in body",
		},
	}
}

func errParameterUnavailable(resource, param string) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        6,
			Address:     resource,
			Description: fmt.Sprintf("parameter, %s, not available", param),
		},
	}
}

func errInvalidValueforParam(r *http.Request, param, value string) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        7,
			Address:     infoFromRequest(r).resource,
			Description: fmt.Sprintf("invalid value, %s, for parameter, %s", value, param),
		},
	}
}

func errParameterReadOnly(r *http.Request, param string) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        8,
			Address:     infoFromRequest(r).resource,
			Description: fmt.Sprintf("parameter, %s, is not modifiable", param),
		},
	}
}

func errDeviceIsOff(resource string, param string) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        201,
			Address:     resource,
			Description: fmt.Sprintf("parameter, %s, is not modifiable. Device is set to off", param),
		},
	}
}

func errInternalError(resource, code string) *errorResp {
	return &errorResp{
		Error: innerErrResp{
			Type:        901,
			Address:     resource,
			Description: fmt.Sprintf("Internal error, %s", code),
		},
	}
}
