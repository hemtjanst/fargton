package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func registerTestingUser(t *testing.T, s *Server) string {
	t.Helper()
	register := &registrationReq{DeviceType: t.Name()}
	enc, err := json.Marshal(register)
	if err != nil {
		t.Fatalf(err.Error())
	}

	st, body := tReq(t, s, http.MethodPost, "/api", enc)
	if st != http.StatusOK {
		t.Fatalf("expected registration to return 200 OK, got %d", st)
	}

	dec := []registrationResp{}
	err = json.Unmarshal(body, &dec)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(dec) != 1 {
		t.Fatalf("expected registration response to contain 1 element, got %d", len(dec))
	}

	return dec[0].Success.Username
}

func TestRegistration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	t.Run("register", func(t *testing.T) {
		t.Run("complete garbage", func(t *testing.T) {
			st, body := tReq(t, b, http.MethodPost, "/api", []byte("/"))
			assert.Equal(t, http.StatusOK, st)

			dec := []errorResp{}
			err := json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
			assert.Equal(t, 901, dec[0].Error.Type)
		})
		t.Run("missing DeviceType", func(t *testing.T) {
			register := &registrationReq{DeviceType: ""}
			enc, err := json.Marshal(register)
			assert.NoError(t, err)

			st, body := tReq(t, b, http.MethodPost, "/api", enc)
			assert.Equal(t, http.StatusOK, st)

			dec := []errorResp{}
			err = json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
			assert.Equal(t, 5, dec[0].Error.Type)
		})
		t.Run("properly", func(t *testing.T) {
			register := &registrationReq{DeviceType: t.Name()}
			enc, err := json.Marshal(register)
			assert.NoError(t, err)

			st, body := tReq(t, b, http.MethodPost, "/api", enc)
			assert.Equal(t, http.StatusOK, st)

			dec := []registrationResp{}
			err = json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
		})
		t.Run("unsuccessfully due to bad whitelist", func(t *testing.T) {
			register := &registrationReq{DeviceType: t.Name()}
			enc, err := json.Marshal(register)
			assert.NoError(t, err)

			b.config.Lock()
			b.config.whitelistConfigPath = "/harhar/nope.derp"
			b.config.Unlock()

			defer func() {
				b.config.Lock()
				b.config.whitelistConfigPath = ""
				b.config.Unlock()
			}()

			st, body := tReq(t, b, http.MethodPost, "/api", enc)
			assert.Equal(t, http.StatusOK, st)

			dec := []errorResp{}
			err = json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
			assert.Equal(t, 901, dec[0].Error.Type)
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user1 := registerTestingUser(t, b)
			user2 := registerTestingUser(t, b)

			st, body := tReq(t, b, http.MethodDelete, fmt.Sprintf("/api/%s/config/whitelist/%s", user1, user2), nil)
			assert.Equal(t, http.StatusOK, st)

			dec := []deleteResp{}
			err := json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
			assert.Equal(t,
				fmt.Sprintf("/config/whitelist/%s deleted.", user2),
				dec[0].Success,
			)
		})
		t.Run("unsuccessfully due to bad whitelist", func(t *testing.T) {
			user1 := registerTestingUser(t, b)
			user2 := registerTestingUser(t, b)

			b.config.Lock()
			b.config.whitelistConfigPath = "/harhar/nope.derp"
			b.config.Unlock()

			defer func() {
				b.config.Lock()
				b.config.whitelistConfigPath = ""
				b.config.Unlock()
			}()

			st, body := tReq(t, b, http.MethodDelete, fmt.Sprintf("/api/%s/config/whitelist/%s", user1, user2), nil)
			assert.Equal(t, http.StatusOK, st)

			dec := []errorResp{}
			err := json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
			assert.Equal(t, 901, dec[0].Error.Type)
		})
	})
}
