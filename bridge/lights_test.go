package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"lib.hemtjan.st/testutils"
)

func TestGetAllLights(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	t.Run("dimmable bulb", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-dim.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light1")

		lights := b.getAllLightsFromMQTT()
		assert.Len(t, lights, 1)
		l, ok := lights[TopicToStrInt("test/light1")]
		assert.True(t, ok)
		assert.Equal(t, whiteType, l.Type)
		assert.Equal(t, whiteModel, l.Model)
		assert.Equal(t, whiteProductName, l.ProductName)
		assert.False(t, l.State.On)
		assert.Equal(t, 12, l.State.Brightness)
	})
	t.Run("colour temperature bulb", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-ct.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light2")

		lights := b.getAllLightsFromMQTT()
		assert.Len(t, lights, 1)
		l, ok := lights[TopicToStrInt("test/light2")]
		assert.True(t, ok)
		assert.Equal(t, temperatureType, l.Type)
		assert.Equal(t, temperatureModel, l.Model)
		assert.Equal(t, temperatureProductName, l.ProductName)
		assert.False(t, l.State.On)
		assert.Equal(t, 2, l.State.Brightness)
		assert.Equal(t, 400, l.State.MiredColorTemp)
	})
	t.Run("RGB bulb", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-rgb.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light3")

		lights := b.getAllLightsFromMQTT()
		assert.Len(t, lights, 1)
		l, ok := lights[TopicToStrInt("test/light3")]
		assert.True(t, ok)
		assert.Equal(t, rgbType, l.Type)
		assert.Equal(t, rgbModel, l.Model)
		assert.Equal(t, rgbProductName, l.ProductName)
		assert.False(t, l.State.On)
		assert.Equal(t, 2, l.State.Brightness)
		assert.Equal(t, []float64{0.3000000083898227, 0.6000000167796454}, l.State.XY)
	})
	t.Run("through the API", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		c1, err := testutils.DevicesFromJSON("./testing_data/light-dim.json", m)
		assert.NoError(t, err)
		defer c1()
		c2, err := testutils.DevicesFromJSON("./testing_data/light-ct.json", m)
		assert.NoError(t, err)
		defer c2()
		c3, err := testutils.DevicesFromJSON("./testing_data/light-rgb.json", m)
		assert.NoError(t, err)
		defer c3()

		b.mqtt.WaitForDevice(ctx, "test/light1")
		b.mqtt.WaitForDevice(ctx, "test/light2")
		b.mqtt.WaitForDevice(ctx, "test/light3")

		username := registerTestingUser(t, b)

		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights", username), nil)
		assert.Equal(t, http.StatusOK, st)
		dec := lights{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Len(t, dec, 3)
	})
}

func TestGetLightByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)
	username := registerTestingUser(t, b)

	t.Run("non-existant ID", func(t *testing.T) {
		username := registerTestingUser(t, b)

		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights/1", username), nil)
		assert.Equal(t, http.StatusOK, st)
		errDec := []*errorResp{}
		err := json.Unmarshal(body, &errDec)
		assert.NoError(t, err)
		assert.Equal(t, 3, errDec[0].Error.Type)

	})
	t.Run("dimmable light ID", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-dim.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light1")

		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights/%s", username, TopicToStrInt("test/light1")), nil)
		assert.Equal(t, http.StatusOK, st)
		dec := light{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Equal(t, whiteType, dec.Type)
	})
	t.Run("color temperature light ID", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-ct.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light2")

		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights/%s", username, TopicToStrInt("test/light2")), nil)
		assert.Equal(t, http.StatusOK, st)
		dec := light{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Equal(t, temperatureType, dec.Type)
	})
	t.Run("rgb light ID", func(t *testing.T) {
		clf, m := NewTestingTransport(t, nil)
		defer clf()
		cleanup, err := testutils.DevicesFromJSON("./testing_data/light-rgb.json", m)
		assert.NoError(t, err)
		defer cleanup()

		b.mqtt.WaitForDevice(ctx, "test/light3")

		st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights/%s", username, TopicToStrInt("test/light3")), nil)
		assert.Equal(t, http.StatusOK, st)
		dec := light{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Equal(t, rgbType, dec.Type)
	})
}

func TestGetNewLights(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)
	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodGet, fmt.Sprintf("/api/%s/lights/new", username), nil)
	assert.Equal(t, http.StatusOK, st)
	dec := lastScanResp{}
	err := json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Equal(t, DateTimeToISO8600(now().UTC()), dec.LastScan)
}

func TestSearchLights(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)
	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodPost, fmt.Sprintf("/api/%s/lights", username), nil)
	assert.Equal(t, http.StatusOK, st)
	dec := []successResp{}
	err := json.Unmarshal(body, &dec)
	assert.NoError(t, err)
	assert.Len(t, dec, 1)
	assert.Len(t, dec[0].Success, 1)
}

func TestLightRename(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	username := registerTestingUser(t, b)

	st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s", username, TopicToStrInt("test/light1")), nil)
	assert.Equal(t, http.StatusOK, st)
	errDec := []*errorResp{}
	err := json.Unmarshal(body, &errDec)
	assert.NoError(t, err)
	assert.Equal(t, 8, errDec[0].Error.Type)
}

func TestLightStateUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)
	clf, m := NewTestingTransport(t, nil)
	defer clf()
	c1, err := testutils.DevicesFromJSON("./testing_data/light-dim.json", m)
	assert.NoError(t, err)
	defer c1()
	c2, err := testutils.DevicesFromJSON("./testing_data/light-ct.json", m)
	assert.NoError(t, err)
	defer c2()
	c3, err := testutils.DevicesFromJSON("./testing_data/light-rgb.json", m)
	assert.NoError(t, err)
	defer c3()
	c4, err := testutils.DevicesFromJSON("./testing_data/light-dim-on.json", m)
	assert.NoError(t, err)
	defer c4()

	b.mqtt.WaitForDevice(ctx, "test/light1")
	b.mqtt.WaitForDevice(ctx, "test/light2")
	b.mqtt.WaitForDevice(ctx, "test/light3")
	b.mqtt.WaitForDevice(ctx, "test/light4")

	username := registerTestingUser(t, b)

	t.Run("invalid light", func(t *testing.T) {
		q, err := json.Marshal(lightStateUpdate{On: BoolPtr(true)})
		assert.NoError(t, err)
		st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, "error"), q)
		assert.Equal(t, http.StatusOK, st)

		dec := []*errorResp{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Equal(t, 3, dec[0].Error.Type)
	})
	t.Run("errors", func(t *testing.T) {
		cases := map[string]lightStateUpdate{
			"hue":        lightStateUpdate{Hue: IntPtr(10)},
			"sat":        lightStateUpdate{Saturation: IntPtr(10)},
			"alert":      lightStateUpdate{Alert: StrPtr("test")},
			"effect":     lightStateUpdate{Effect: StrPtr("test")},
			"transition": lightStateUpdate{TransitionTime: IntPtr(10)},
			"hue_inc":    lightStateUpdate{HueInc: IntPtr(10)},
			"sat_inc":    lightStateUpdate{SaturationInc: IntPtr(10)},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				upd := c
				err := upd.Bind(&http.Request{})
				assert.Error(t, err)

				q, err := json.Marshal(c)
				assert.NoError(t, err)
				st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt("test/light1")), q)
				assert.Equal(t, http.StatusOK, st)

				dec := []*errorResp{}
				err = json.Unmarshal(body, &dec)
				assert.NoError(t, err)
				assert.Equal(t, 6, dec[0].Error.Type)
			})
		}
	})
	t.Run("override", func(t *testing.T) {
		t.Run("brightness", func(t *testing.T) {
			upd := lightStateUpdate{
				Brightness:    IntPtr(10),
				BrightnessInc: IntPtr(5),
			}
			err := upd.Bind(&http.Request{})
			assert.NoError(t, err)
			assert.Equal(t, 10, *upd.Brightness)
			assert.Nil(t, upd.BrightnessInc)
		})
		t.Run("colorTemperature", func(t *testing.T) {
			upd := lightStateUpdate{
				ColorTemperature:    IntPtr(10),
				ColorTemperatureInc: IntPtr(5),
			}
			err := upd.Bind(&http.Request{})
			assert.NoError(t, err)
			assert.Equal(t, 10, *upd.ColorTemperature)
			assert.Nil(t, upd.ColorTemperatureInc)
		})
		t.Run("xy", func(t *testing.T) {
			upd := lightStateUpdate{
				XY:    FloatPtr([]float64{1, 1}),
				XYInc: FloatPtr([]float64{10, 10}),
			}
			err := upd.Bind(&http.Request{})
			assert.NoError(t, err)
			assert.Equal(t, []float64{1, 1}, *upd.XY)
			assert.Nil(t, upd.XYInc)
		})
	})
	t.Run("set invalid param for device", func(t *testing.T) {
		t.Run("dimmable", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"colorTemp":     lightStateUpdate{ColorTemperature: IntPtr(10)},
				"colorTemp_inc": lightStateUpdate{ColorTemperatureInc: IntPtr(10)},
				"xy":            lightStateUpdate{XY: FloatPtr([]float64{1, 1})},
				"xy_inc":        lightStateUpdate{XYInc: FloatPtr([]float64{10, 10})},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light1"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.InvalidParameter, 1)
				})
			}
		})
		t.Run("colorTemperature", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"xy":     lightStateUpdate{XY: FloatPtr([]float64{1, 1})},
				"xy_inc": lightStateUpdate{XYInc: FloatPtr([]float64{10, 10})},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light2"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.InvalidParameter, 1)
				})
			}
		})
		t.Run("rgb", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"colorTemp":     lightStateUpdate{ColorTemperature: IntPtr(10)},
				"colorTemp_inc": lightStateUpdate{ColorTemperatureInc: IntPtr(10)},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light3"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.InvalidParameter, 1)
				})
			}
		})
	})
	t.Run("set param on device off", func(t *testing.T) {
		t.Run("dimmable", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"brightness":     lightStateUpdate{Brightness: IntPtr(10)},
				"brightness_inc": lightStateUpdate{BrightnessInc: IntPtr(10)},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light1"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.DeviceIsOff, 1)
				})
			}
		})
		t.Run("colorTemperature", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"colorTemp":     lightStateUpdate{ColorTemperature: IntPtr(10)},
				"colorTemp_inc": lightStateUpdate{ColorTemperatureInc: IntPtr(10)},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light2"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.DeviceIsOff, 1)
				})
			}
		})
		t.Run("rgb", func(t *testing.T) {
			cases := map[string]lightStateUpdate{
				"xy":     lightStateUpdate{XY: FloatPtr([]float64{1, 1})},
				"xy_inc": lightStateUpdate{XYInc: FloatPtr([]float64{10, 10})},
			}
			for name, c := range cases {
				t.Run(name, func(t *testing.T) {
					upd := c
					l := b.getLight(TopicToStrInt("test/light3"))
					assert.NotNil(t, l)
					res := b.updateLightState(l, &upd)
					assert.Len(t, res.DeviceIsOff, 1)
				})
			}
		})
	})
	t.Run("turn on", func(t *testing.T) {
		lights := []string{
			"test/light1",
			"test/light2",
			"test/light3",
		}
		on := lightStateUpdate{On: BoolPtr(true)}
		for _, l := range lights {
			lt := b.getLight(TopicToStrInt(l))
			assert.NotNil(t, l)
			res := b.updateLightState(lt, &on)
			assert.Len(t, res.DeviceIsOff, 0)
			assert.Len(t, res.InvalidParameter, 0)
			assert.False(t, res.InternalError)
			assert.Len(t, res.Success, 1)

			q, err := json.Marshal(on)
			assert.NoError(t, err)
			st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt(l)), q)
			assert.Equal(t, http.StatusOK, st)

			dec := []*successResp{}
			err = json.Unmarshal(body, &dec)
			assert.NoError(t, err)
			assert.Len(t, dec, 1)
		}
	})
	t.Run("change brightness", func(t *testing.T) {
		lights := []string{
			"test/light1",
			"test/light2",
			"test/light3",
		}
		cases := map[string]lightStateUpdate{
			"brightness":     lightStateUpdate{On: BoolPtr(true), Brightness: IntPtr(10)},
			"brightness_inc": lightStateUpdate{On: BoolPtr(true), BrightnessInc: IntPtr(10)},
		}
		for _, l := range lights {
			t.Run(l, func(t *testing.T) {
				for name, c := range cases {
					t.Run(name, func(t *testing.T) {
						upd := c
						lt := b.getLight(TopicToStrInt(l))
						assert.NotNil(t, lt)
						res := b.updateLightState(lt, &upd)
						assert.Len(t, res.DeviceIsOff, 0)
						assert.Len(t, res.InvalidParameter, 0)
						assert.False(t, res.InternalError)
						assert.Len(t, res.Success, 2)

						q, err := json.Marshal(c)
						assert.NoError(t, err)
						st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt(l)), q)
						assert.Equal(t, http.StatusOK, st)
						dec := []*successResp{}
						err = json.Unmarshal(body, &dec)
						assert.NoError(t, err)
						assert.Len(t, dec, 2)
					})
				}
			})
		}
	})
	t.Run("change ct", func(t *testing.T) {
		cases := map[string]lightStateUpdate{
			"colorTemp":     lightStateUpdate{On: BoolPtr(true), ColorTemperature: IntPtr(10)},
			"colorTemp_inc": lightStateUpdate{On: BoolPtr(true), ColorTemperatureInc: IntPtr(10)},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				upd := c
				lt := b.getLight(TopicToStrInt("test/light2"))
				assert.NotNil(t, lt)
				res := b.updateLightState(lt, &upd)
				assert.Len(t, res.DeviceIsOff, 0)
				assert.Len(t, res.InvalidParameter, 0)
				assert.False(t, res.InternalError)
				assert.Len(t, res.Success, 2)

				q, err := json.Marshal(c)
				assert.NoError(t, err)
				st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt("test/light2")), q)
				assert.Equal(t, http.StatusOK, st)
				dec := []*successResp{}
				err = json.Unmarshal(body, &dec)
				assert.NoError(t, err)
				assert.Len(t, dec, 2)
			})
		}
	})
	t.Run("change xy", func(t *testing.T) {
		cases := map[string]lightStateUpdate{
			"xy":     lightStateUpdate{On: BoolPtr(true), XY: FloatPtr([]float64{1, 1})},
			"xy_inc": lightStateUpdate{On: BoolPtr(true), XYInc: FloatPtr([]float64{10, 10})},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				upd := c
				lt := b.getLight(TopicToStrInt("test/light3"))
				assert.NotNil(t, lt)
				res := b.updateLightState(lt, &upd)
				assert.Len(t, res.DeviceIsOff, 0)
				assert.Len(t, res.InvalidParameter, 0)
				assert.False(t, res.InternalError)
				assert.Len(t, res.Success, 2)

				q, err := json.Marshal(c)
				assert.NoError(t, err)
				st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt("test/light3")), q)
				assert.Equal(t, http.StatusOK, st)
				dec := []*successResp{}
				err = json.Unmarshal(body, &dec)
				assert.NoError(t, err)
				assert.Len(t, dec, 2)
			})
		}
	})
	t.Run("turn off", func(t *testing.T) {
		off := lightStateUpdate{On: BoolPtr(false)}
		lt := b.getLight(TopicToStrInt("test/light4"))
		assert.NotNil(t, lt)
		res := b.updateLightState(lt, &off)
		assert.Len(t, res.DeviceIsOff, 0)
		assert.Len(t, res.InvalidParameter, 0)
		assert.False(t, res.InternalError)
		assert.Len(t, res.Success, 1)

		q, err := json.Marshal(off)
		assert.NoError(t, err)
		st, body := tReq(t, b, http.MethodPut, fmt.Sprintf("/api/%s/lights/%s/state", username, TopicToStrInt("test/light4")), q)
		assert.Equal(t, http.StatusOK, st)
		dec := []*successResp{}
		err = json.Unmarshal(body, &dec)
		assert.NoError(t, err)
		assert.Len(t, dec, 1)
	})
}

func TestRenderLightStateUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	b, shutdown := NewTestingBridge(t, nil)
	defer cancel()
	defer shutdown(ctx)

	t.Run("invalid parameter", func(t *testing.T) {
		upd := &lightUpdateStateResult{InvalidParameter: []string{"one"}}
		t.Run("group=false", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, false, "test1")
			assert.Len(t, res, 1)
		})
		t.Run("group=true", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, true, "test1")
			assert.Len(t, res, 1)
		})
	})
	t.Run("device is off", func(t *testing.T) {
		upd := &lightUpdateStateResult{DeviceIsOff: []string{"one"}}
		t.Run("group=false", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, false, "test1")
			assert.Len(t, res, 1)
		})
		t.Run("group=true", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, true, "test1")
			assert.Len(t, res, 1)
		})
	})
	t.Run("internal error", func(t *testing.T) {
		upd := &lightUpdateStateResult{InternalError: true}
		t.Run("group=false", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, false, "test1")
			assert.Len(t, res, 1)
		})
		t.Run("group=true", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, true, "test1")
			assert.Len(t, res, 1)
		})
	})
	t.Run("success", func(t *testing.T) {
		upd := &lightUpdateStateResult{Success: map[string]interface{}{
			"on": true,
		}}
		t.Run("group=false", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, false, "test1")
			assert.Len(t, res, 1)
		})
		t.Run("group=true", func(t *testing.T) {
			res := b.renderLightStateUpdate(upd, true, "test1")
			assert.Len(t, res, 1)
		})
	})
}
