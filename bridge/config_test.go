package bridge

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	now = func() time.Time { return time.Unix(1554594787, 100) }
}

func TestNewConfig(t *testing.T) {
	t.Run("no name", func(t *testing.T) {
		c, err := NewConfig()
		assert.NotNil(t, err)
		assert.Nil(t, c)
	})
	t.Run("only name", func(t *testing.T) {
		c, err := NewConfig(Name(t.Name()))
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		if !assert.NotNil(t, c) {
			t.FailNow()
		}
		assert.Equal(t, t.Name(), c.Name)
		assert.Equal(t, apiVersion, c.APIVersion)
		assert.Equal(t, swVersion, c.SWVersion)
		assert.Equal(t, datastoreVersion, c.DatastoreVersion)
		assert.Equal(t, bridgeModel, c.ModelID)
		assert.Equal(t, MACAddr{0x1, 0x23, 0x45, 0x67, 0x89, 0xab}, c.MACAddress)
		assert.Equal(t, "012345FFFE6789AB", c.BridgeID)
		assert.Equal(t, "0123456789AB", c.strippedMAC)
		assert.Equal(t, fmt.Sprintf("%s-%s", uuidPrefix, "0123456789ab"), c.uuid)
		assert.Equal(t, net.ParseIP("127.0.0.1"), c.advertiseIP)
		assert.Equal(t, "0.0.0.0:0", c.address)
		assert.Equal(t, uint16(0), c.port)
		assert.Equal(t, "0.0.0.0:0", c.tlsAddress)
		assert.Equal(t, uint16(0), c.tlsPort)
		assert.Equal(t, "", c.tlsPrivKey)
		assert.Equal(t, "", c.tlsPubKey)
		assert.Equal(t, "UTC", c.timezone.String())
		assert.False(t, c.authDisabled)
		assert.Equal(t, "", c.whitelistConfigPath)
		assert.Equal(t, 0.0, c.latitude)
		assert.Equal(t, 0.0, c.longitude)
	})
	t.Run("MAC", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), MAC("aa:bb:cc:dd:ee:ff"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, MACAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}, c.MACAddress)
			assert.Equal(t, "AABBCCFFFEDDEEFF", c.BridgeID)
			assert.Equal(t, "AABBCCDDEEFF", c.strippedMAC)
			assert.Equal(t, fmt.Sprintf("%s-%s", uuidPrefix, "aabbccddeeff"), c.uuid)
		})
		t.Run("invalid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), MAC("aa:"))
			assert.NotNil(t, err)
			assert.Nil(t, c)
		})
	})
	t.Run("Address", func(t *testing.T) {
		t.Run("host:port", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Address("127.0.0.1:8080"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "127.0.0.1:8080", c.address)
		})
		t.Run(":port", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Address(":8080"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "0.0.0.0:8080", c.address)
		})
	})
	t.Run("Timezone", func(t *testing.T) {
		t.Run("Europe/Amsterdam", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Timezone("Europe/Amsterdam"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "Europe/Amsterdam", c.timezone.String())
		})
		t.Run("invalid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Timezone("nosuch/timezone"))
			assert.NotNil(t, err)
			assert.Nil(t, c)
		})
	})
	t.Run("TLSAddress", func(t *testing.T) {
		t.Run("host:port", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), TLSAddress("127.0.0.1:8080"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "127.0.0.1:8080", c.tlsAddress)
		})
		t.Run(":port", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), TLSAddress(":8080"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "0.0.0.0:8080", c.tlsAddress)
		})
	})
	t.Run("AdvertiseIP", func(t *testing.T) {
		t.Run("IPv4", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), AdvertiseIP("192.168.0.1"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xc0, 0xa8, 0x0, 0x1}, c.advertiseIP)
		})
		t.Run("IPv6", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), AdvertiseIP("fe80::ec92:ecff:fefc:6f25"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, net.IP{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xec, 0x92, 0xec, 0xff, 0xfe, 0xfc, 0x6f, 0x25}, c.advertiseIP)
		})
		t.Run("invalid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), AdvertiseIP("correct horse battery staple"))
			assert.NotNil(t, err)
			assert.Nil(t, c)
		})
	})
	t.Run("TLS keys", func(t *testing.T) {
		t.Run("public", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), TLSPublicKeyPath("public.crt"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "public.crt", c.tlsPubKey)
		})
		t.Run("private", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), TLSPrivateKeyPath("private.key"))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, "private.key", c.tlsPrivKey)
		})
	})
	t.Run("Disable authentication", func(t *testing.T) {
		t.Run("true", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), DisableAuthentication(true))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.True(t, c.authDisabled)
		})
		t.Run("false", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), DisableAuthentication(false))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.False(t, c.authDisabled)
		})
	})
	t.Run("Configure whitelist path", func(t *testing.T) {
		t.Run("/tmp/derp.json", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), WhitelistConfigPath(t.Name()))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, t.Name(), c.whitelistConfigPath)
		})
	})
	t.Run("latitude", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Latitude(50.85045))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, 50.85045, c.latitude)
		})
		t.Run("invalid", func(t *testing.T) {
			t.Run("over", func(t *testing.T) {
				c, err := NewConfig(Name(t.Name()), Latitude(95.0))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
			t.Run("under", func(t *testing.T) {
				c, err := NewConfig(Name(t.Name()), Latitude(-95.0))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
	})
	t.Run("longitude", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			c, err := NewConfig(Name(t.Name()), Longitude(4.34878))
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			if !assert.NotNil(t, c) {
				t.FailNow()
			}

			assert.Equal(t, 4.34878, c.longitude)
		})
		t.Run("invalid", func(t *testing.T) {
			t.Run("over", func(t *testing.T) {
				c, err := NewConfig(Name(t.Name()), Longitude(182.0))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
			t.Run("under", func(t *testing.T) {
				c, err := NewConfig(Name(t.Name()), Longitude(-182.0))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
	})
}

func TestMACMarshall(t *testing.T) {
	mac := MACAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	j, err := mac.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "\"aa:bb:cc:dd:ee:ff\"", string(j))
}

func TestUnauthenticateConfig(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		c, err := NewConfig(Name(t.Name()))
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		if !assert.NotNil(t, c) {
			t.FailNow()
		}

		unauth := createUnauthenticatedConfig(c)
		assert.Equal(t, c.Name, unauth.Name)
		assert.Equal(t, c.APIVersion, unauth.APIVersion)
		assert.Equal(t, c.SWVersion, unauth.SWVersion)
		assert.Equal(t, c.DatastoreVersion, unauth.DatastoreVersion)
		assert.Equal(t, c.MACAddress, unauth.MACAddress)
		assert.Equal(t, c.BridgeID, unauth.BridgeID)
		assert.Equal(t, c.ModelID, unauth.ModelID)
		assert.Nil(t, unauth.Whitelist)
	})
	t.Run("render", func(t *testing.T) {
		t.SkipNow()
	})
}

func TestAuthenticatedConfig(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		assert := assert.New(t)
		c, err := NewConfig(Name(t.Name()))
		if !assert.Nil(err) {
			t.FailNow()
		}
		if !assert.NotNil(c) {
			t.FailNow()
		}

		authed := createAuthenticatedConfig(c)

		assert.Equal(c.Name, authed.Name)
		assert.Equal(c.APIVersion, authed.APIVersion)
		assert.Equal(c.SWVersion, authed.SWVersion)
		assert.Equal(c.DatastoreVersion, authed.DatastoreVersion)
		assert.Equal(c.MACAddress, authed.MACAddress)
		assert.Equal(c.BridgeID, authed.BridgeID)
		assert.Equal(c.ModelID, authed.ModelID)

		assert.Equal(backup{
			Status:    "idle",
			ErrorCode: 0,
		}, authed.Backup)

		assert.False(authed.DHCP)
		assert.Equal("0.0.0.0", authed.Gateway)
		assert.Equal(internetServices{
			Internet:     "connected",
			RemoteAccess: "disconnected",
			Time:         "connected",
			SWUpdate:     "disconnected",
		}, authed.InternetServices)
		assert.Equal(c.advertiseIP.String(), authed.IPAddress)
		assert.True(authed.LinkButtonPressed)
		assert.Equal(DateTimeToISO8600(now().UTC()), authed.LocalTime)
		assert.Equal("255.255.255.0", authed.Netmask)
		assert.Equal("disconnected", authed.PortalConnection)
		assert.False(authed.PortalServices)
		assert.Equal(portalState{
			SignedOn:      false,
			Incoming:      false,
			Outgoing:      false,
			Communication: "disconnected",
		}, authed.PortalState)
		assert.Equal("none", authed.ProxyAddress)
		assert.Equal(0, authed.ProxyPort)
		assert.Equal(swupdate2{
			CheckForUpdate: false,
			LastChange:     DateTimeToISO8600(now().UTC()),
			Bridge: bridge{
				State:       "noupdates",
				LastInstall: DateTimeToISO8600(now().UTC()),
			},
			State: "noupdates",
			AutoInstall: autoInstall{
				UpdateTime: "T14:00:00",
				Enabled:    false,
			},
		}, authed.SWUpdate2)
		assert.Equal(c.timezone.String(), authed.Timezone)
		assert.Equal(DateTimeToISO8600(now().UTC()), authed.UTC)
		assert.Equal(c.Whitelist, authed.Whitelist)
		assert.Equal(15, authed.ZigbeeChannel)
	})
	t.Run("render", func(t *testing.T) {
		t.SkipNow()
	})
}
