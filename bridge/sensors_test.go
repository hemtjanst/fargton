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

func TestNewDaylightSensor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)
	t.Run("default", func(t *testing.T) {
		d := b.newDaylightSensor()
		assert.Equal(t, "Daylight", d.Name)
		assert.Equal(t, "Daylight", d.Type)
		assert.Equal(t, "PHDL00", d.ModelID)
		assert.Equal(t, "Philips", d.ManufacturerName)
		assert.Equal(t, "1.0", *d.SWVersion)

		assert.False(t, *d.State.Daylight)
		assert.Equal(t, DateTimeToISO8600(now().UTC()), d.State.LastUpdated)
		assert.Nil(t, d.State.ButtonEvent)

		assert.True(t, d.Config.On)
		assert.Nil(t, d.Config.Reachable)
		assert.Nil(t, d.Config.Battery)
		assert.Equal(t, "0.0000E", *d.Config.Longitude)
		assert.Equal(t, "0.0000N", *d.Config.Latitude)
		assert.Equal(t, 0, *d.Config.SunriseOffset)
		assert.Equal(t, 0, *d.Config.SunsetOffset)
	})
	t.Run("south-western hemisphere", func(t *testing.T) {
		b.config.Lock()
		b.config.latitude = -1.0
		b.config.longitude = -1.0
		b.config.Unlock()
		defer func() {
			b.config.Lock()
			b.config.latitude = 0.0
			b.config.longitude = 0.0
			b.config.Unlock()
		}()
		d := b.newDaylightSensor()
		assert.Equal(t, "1.0000W", *d.Config.Longitude)
		assert.Equal(t, "1.0000S", *d.Config.Latitude)
	})
}
func TestGetAllSensors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/sensors", username), nil)
	assert.Equal(t, http.StatusOK, st)
	dec := sensors{}
	err := json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Len(t, dec, 1)
}

func TestSensorByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	username := registerTestingUser(t, b)

	t.Run("non-existant ID", func(t *testing.T) {
		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/sensors/2", username), nil)
		assert.Equal(t, http.StatusOK, st)
		errDec := []*errorResp{}
		err := json.Unmarshal(body, &errDec)
		assert.NoError(t, err)
		assert.Equal(t, 3, errDec[0].Error.Type)

	})
	t.Run("daylight sensor", func(t *testing.T) {
		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/sensors/1", username), nil)
		assert.Equal(t, http.StatusOK, st)
		dec := sensor{}
		err := json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Equal(t, "Daylight", dec.Type)
	})
}

func TestSearchSensors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodPost, fmt.Sprintf("/api/%s/sensors", username), nil)
	assert.Equal(t, http.StatusOK, st)
	dec := []successResp{}
	err := json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Len(t, dec, 1)
	assert.Len(t, dec[0].Success, 1)
}

func TestGetNewSensors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/sensors/new", username), nil)
	assert.Equal(t, http.StatusOK, st)
	dec := lastScanResp{}
	err := json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Equal(t, DateTimeToISO8600(now().UTC()), dec.LastScan)
}
