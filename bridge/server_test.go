package bridge

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"lib.hemtjan.st/server"
	"lib.hemtjan.st/testutils"
	"lib.hemtjan.st/transport/mqtt"
)

const testingMAC = "52:41:67:64:ac:b7" // Locally Administered Unicast MAC Address
const testingTLSCert = "./testing_data/public.crt"
const testingTLSKey = "./testing_data/private.key"
const testingHost = "127.0.0.1"
const testingHostPort = "127.0.0.1:0"

func NewTestingTransport(t *testing.T,
	l *zap.Logger,
) (context.CancelFunc, mqtt.MQTT) {
	if l == nil {
		l = zap.NewNop()
	}

	ctx, clf := context.WithCancel(context.Background())

	name := fmt.Sprintf("%s-%s", t.Name(), TopicToStrInt(time.Now().String()))
	mqCfg := &mqtt.Config{
		ClientID:      name,
		Address:       []string{testutils.MQTTAddress(t)},
		AnnounceTopic: "announce",
		LeaveTopic:    "leave",
		DiscoverTopic: "discover",
	}
	cl, err := mqtt.New(ctx, mqCfg)
	if err != nil {
		clf()
		t.Fatalf(err.Error())
	}
	return clf, cl
}

// NewTestingBridge starts a bridge configured for testing
func NewTestingBridge(t *testing.T,
	l *zap.Logger,
) (*Server, func(context.Context)) {

	if l == nil {
		l = zap.NewNop()
	}

	clf, cl := NewTestingTransport(t, l)
	m := server.New(cl)

	opts := []ConfigOption{
		Name(t.Name()),
		MAC(testingMAC),
		AdvertiseIP(testingHost),
		Address(testingHostPort),
		Timezone("UTC"),
		TLSPublicKeyPath(testingTLSCert),
		TLSPrivateKeyPath(testingTLSKey),
		TLSAddress(testingHostPort),
	}

	c, err := NewConfig(opts...)
	if err != nil {
		clf()
		t.Fatalf(err.Error())
	}

	b := NewServer(c, m, l)
	cancel, err := b.Start(0)
	if err != nil {
		clf()
		t.Fatalf(err.Error())
	}

	return b, func(ctx context.Context) {
		clf()
		cancel(ctx)
	}
}

func tReq(t *testing.T,
	s *Server,
	method string,
	endpoint string,
	body []byte,
) (int, []byte) {
	t.Helper()
	hc := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	rr, err := http.NewRequest(
		method,
		fmt.Sprintf("https://%s:%d%s", testingHost, s.config.tlsPort, endpoint),
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	resp, err := hc.Do(rr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return resp.StatusCode, data
}

func TestNewServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	t.Run("config", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			st, body := tReq(t, b, http.MethodGet, "/api/nouser/config", nil)
			assert.Equal(t, http.StatusOK, st)

			dec := unauthenticatedConfig{}
			err := json.Unmarshal(body, &dec)
			assert.NoError(t, err)
		})
		t.Run("authenticated", func(t *testing.T) {
			username := registerTestingUser(t, b)
			st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/config", username), nil)
			assert.Equal(t, http.StatusOK, st)
			dec := unauthenticatedConfig{}
			err := json.Unmarshal(body, &dec)
			assert.NoError(t, err)
		})
	})

	t.Run("anonymous on authenticated endpoint", func(t *testing.T) {
		st, body := tReq(t, b, http.MethodGet, "/api/fake-turrible/lights/new", nil)
		assert.Equal(t, http.StatusOK, st)

		errDec := []*errorResp{}
		err := json.Unmarshal(body, &errDec)
		assert.NoError(t, err)
		assert.Equal(t, 1, errDec[0].Error.Type)
	})

	t.Run("method not allowed", func(t *testing.T) {
		st, body := tReq(t, b, http.MethodPut, "/api", nil)
		assert.Equal(t, http.StatusOK, st)

		errDec := []*errorResp{}
		err := json.Unmarshal(body, &errDec)
		assert.NoError(t, err)
		assert.Equal(t, 4, errDec[0].Error.Type)
	})
}

func TestWhitelist(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf(err.Error())
	}

	defer os.RemoveAll(dir) // clean up
	t.Run("save", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		b, shutdown := NewTestingBridge(t, nil)
		defer cancel()
		defer shutdown(ctx)
		b.config.Lock()
		b.config.Whitelist = &map[string]whitelist{
			"test": whitelist{
				ID:         "a",
				CreatedAt:  "a",
				LastUsedAt: "a",
				Name:       "zxcvf",
			},
		}
		b.config.whitelistConfigPath = filepath.Join(dir, "whitelist.json")
		b.config.Unlock()
		err := b.saveWhitelistToFile()
		assert.NoError(t, err)

		f, err := ioutil.ReadFile(filepath.Join(dir, "whitelist.json"))
		assert.NoError(t, err)
		assert.True(t, strings.Contains(string(f), "zxcvf"))
	})
	t.Run("load", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		b, shutdown := NewTestingBridge(t, nil)
		defer cancel()
		defer shutdown(ctx)
		b.config.Lock()
		b.config.whitelistConfigPath = filepath.Join(dir, "whitelist.json")
		b.config.Unlock()
		res, err := b.loadWhitelistFromFile()
		assert.NoError(t, err)
		assert.Len(t, *res, 1)
	})
}

func TestGetConfigAndData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	clf, m := NewTestingTransport(t, nil)
	defer clf()
	cleanup, err := testutils.DevicesFromJSON("./testing_data/light-dim.json", m)
	assert.NoError(t, err)
	defer cleanup()

	b.mqtt.WaitForDevice(ctx, "test/light1")

	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s", username), nil)
	assert.Equal(t, http.StatusOK, st)

	dec := configAndDataResp{}
	err = json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Len(t, *dec.Lights, 1)
	assert.Len(t, *dec.Groups, 1)
	assert.Len(t, *dec.Scenes, 0)
	assert.Len(t, *dec.Rules, 0)
	assert.Len(t, *dec.Schedules, 0)
	assert.Len(t, *dec.ResourceLinks, 0)
	assert.Len(t, *dec.Sensors, 1)
}
