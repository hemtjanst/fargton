package bridge

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const apiVersion = "1.31.0"
const bridgeModel = "BSB002"
const fallbackMAC = "01:23:45:67:89:AB"
const swVersion = "1931069120"
const datastoreVersion = "80"
const uuidPrefix = "2f402f80-da50-11e1-9b23"

var now = time.Now

// Config represent bridge configuration
type Config struct {
	Name             string                `json:"name"`
	APIVersion       string                `json:"apiversion"`
	SWVersion        string                `json:"swversion"`
	DatastoreVersion string                `json:"datastoreversion"`
	MACAddress       MACAddr               `json:"mac"`
	BridgeID         string                `json:"bridgeid"`
	ModelID          string                `json:"modelid"`
	Whitelist        *map[string]whitelist `json:"whitelist,omitempty"`

	// Extra stuff we need
	advertiseIP net.IP
	uuid        string
	strippedMAC string

	address             string
	authDisabled        bool
	port                uint16
	tlsAddress          string
	tlsPort             uint16
	tlsPubKey           string
	tlsPrivKey          string
	timezone            *time.Location
	whitelistConfigPath string

	lights int
	groups int

	sunrise uint8
	sunset  uint8

	sync.RWMutex
}

// MACAddr is a net.HardwareAddr with a JSON marshaller
type MACAddr net.HardwareAddr

// MarshalJSON marshals the MAC to its string representation
func (m MACAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.HardwareAddr(m).String())
}

// UnmarshalJSON turns the string back into a MACAddr
func (m *MACAddr) UnmarshalJSON(b []byte) error {
	s := new(string)
	err := json.Unmarshal(b, s)
	if err != nil {
		return err
	}
	mc, err := net.ParseMAC(*s)
	if err != nil {
		return err
	}
	*m = MACAddr(mc)
	return nil
}

// ConfigOption is a single option
type ConfigOption func(*Config) error

// Name sets the name in Config
func Name(n string) ConfigOption {
	return func(args *Config) error {
		args.Name = n
		return nil
	}
}

// DisableAuthentication allows any request through even when
// not on the whitelist
func DisableAuthentication(b bool) ConfigOption {
	return func(args *Config) error {
		args.authDisabled = b
		return nil
	}
}

// MAC configures the MAC address of the bridge
func MAC(m string) ConfigOption {
	return func(args *Config) error {
		pm, err := net.ParseMAC(m)
		if err != nil {
			return err
		}
		args.MACAddress = MACAddr(pm)
		args.strippedMAC = strings.ToUpper(strings.Replace(m, ":", "", -1))
		args.uuid = strings.ToLower(fmt.Sprintf("%s-%s", uuidPrefix, args.strippedMAC))
		return nil
	}
}

// AdvertiseIP sets the IP the bridge will advertise itself as listening on
func AdvertiseIP(ip string) ConfigOption {
	return func(args *Config) error {
		ipp := net.ParseIP(ip)
		if ipp == nil {
			return fmt.Errorf("could not parse %s as a valid IP", ip)
		}
		args.advertiseIP = ipp
		return nil
	}
}

// Address sets the host:port to bind on for plain tcp/http
func Address(a string) ConfigOption {
	return func(args *Config) error {
		args.address = a
		return nil
	}
}

// Timezone sets the local timezone
func Timezone(t string) ConfigOption {
	return func(args *Config) error {
		tz, err := time.LoadLocation(t)
		if err != nil {
			return err
		}
		args.timezone = tz
		return nil
	}
}

// TLSAddress sets the host:port to bind on for tls/https
func TLSAddress(a string) ConfigOption {
	return func(args *Config) error {
		args.tlsAddress = a
		return nil
	}
}

// TLSPublicKeyPath sets the path to the TLS public key
func TLSPublicKeyPath(a string) ConfigOption {
	return func(args *Config) error {
		args.tlsPubKey = a
		return nil
	}
}

// TLSPrivateKeyPath sets the path to the TLS private key
func TLSPrivateKeyPath(a string) ConfigOption {
	return func(args *Config) error {
		args.tlsPrivKey = a
		return nil
	}
}

// WhitelistConfigPath sets the path from where whitelist.json
// will be loaded and saved to
func WhitelistConfigPath(a string) ConfigOption {
	return func(args *Config) error {
		args.whitelistConfigPath = a
		return nil
	}
}

// Sunrise configures when the sun rises
func Sunrise(i uint) ConfigOption {
	return func(args *Config) error {
		args.sunrise = uint8(i)
		return nil
	}
}

// Sunset configures when the sun rises
func Sunset(i uint) ConfigOption {
	return func(args *Config) error {
		args.sunset = uint8(i)
		return nil
	}
}

// bridgeIDfromMAC generates a unique bridge ID based on the MAC
// address
func bridgeIDfromMAC(m MACAddr) string {
	return fmt.Sprintf("%XFFFE%X", []byte(m[:3]), []byte(m[3:]))
}

// NewConfig returns a new bridge configuration with
// the specified options
func NewConfig(setters ...ConfigOption) (*Config, error) {
	c := &Config{
		APIVersion:       apiVersion,
		ModelID:          bridgeModel,
		SWVersion:        swVersion,
		DatastoreVersion: datastoreVersion,
		Whitelist:        &map[string]whitelist{},
	}

	for _, setter := range setters {
		err := setter(c)
		if err != nil {
			return nil, err
		}
	}

	if c.Name == "" {
		return nil, fmt.Errorf("must give the bridge a name")
	}

	if strings.HasPrefix(c.address, ":") && !strings.HasPrefix(c.address, ":::") {
		_ = Address(fmt.Sprintf("0.0.0.0%s", c.address))(c)
	}

	if c.address == "" {
		_ = Address("0.0.0.0:0")(c)
	}

	if strings.HasPrefix(c.tlsAddress, ":") && !strings.HasPrefix(c.tlsAddress, ":::") {
		_ = TLSAddress(fmt.Sprintf("0.0.0.0%s", c.tlsAddress))(c)
	}

	if c.tlsAddress == "" {
		_ = TLSAddress("0.0.0.0:0")(c)
	}

	if c.advertiseIP == nil {
		_ = AdvertiseIP("127.0.0.1")(c)
	}

	if len(c.MACAddress) == 0 {
		_ = MAC(fallbackMAC)(c)
	}

	if c.timezone == nil {
		_ = Timezone("UTC")(c)
	}

	c.BridgeID = bridgeIDfromMAC(c.MACAddress)

	return c, nil
}

type whitelist struct {
	ID         string `json:"-"`
	CreatedAt  string `json:"create date"`
	LastUsedAt string `json:"last use date"`
	Name       string `json:"name"`
}

type unauthenticatedConfig struct {
	ReplacesBridgeID *string `json:"replacesbridgeid"`
	FactoryNew       bool    `json:"factorynew"`
	StarterKitID     string  `json:"starterkitid"`
	Config
}

func (*unauthenticatedConfig) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func createUnauthenticatedConfig(c *Config) *unauthenticatedConfig {
	resp := &unauthenticatedConfig{
		ReplacesBridgeID: nil,
		FactoryNew:       false,
		StarterKitID:     "",
	}
	resp.Name = c.Name
	resp.APIVersion = c.APIVersion
	resp.SWVersion = c.SWVersion
	resp.DatastoreVersion = c.DatastoreVersion
	resp.MACAddress = c.MACAddress
	resp.BridgeID = c.BridgeID
	resp.ModelID = c.ModelID
	resp.Whitelist = nil
	return resp
}

func (s *Server) getUnauthenticatedConfig(w http.ResponseWriter, r *http.Request) {
	renderOK(w, r, createUnauthenticatedConfig(s.config))
}

type bridge struct {
	State       string `json:"state"`
	LastInstall string `json:"lastinstall"`
}

type autoInstall struct {
	UpdateTime string `json:"updatetime"`
	Enabled    bool   `json:"on"`
}

type swupdateDeviceTypes struct {
	Bridge  bool     `json:"bridge"`
	Lights  []string `json:"lights"`
	Sensors []string `json:"sensors"`
}

type swupdate struct {
	UpdateState    int                 `json:"updatestate"`
	CheckForUpdate bool                `json:"checkforupdate"`
	DeviceTypes    swupdateDeviceTypes `json:"devicetypes"`
	URL            string              `json:"url"`
	Text           string              `json:"text"`
	Notify         bool                `json:"notify"`
}

type swupdate2 struct {
	CheckForUpdate bool        `json:"checkforupdate"`
	LastChange     string      `json:"lastchange"`
	Bridge         bridge      `json:"bridge"`
	State          string      `json:"state"`
	AutoInstall    autoInstall `json:"autoinstall"`
}

type internetServices struct {
	Internet     string `json:"internet"`
	RemoteAccess string `json:"remoteaccess"`
	Time         string `json:"time"`
	SWUpdate     string `json:"swupdate"`
}

type backup struct {
	Status    string `json:"status"`
	ErrorCode int    `json:"errorcode"`
}

type portalState struct {
	SignedOn      bool   `json:"signedon"`
	Incoming      bool   `json:"incoming"`
	Outgoing      bool   `json:"outgoing"`
	Communication string `json:"communication"`
}

type authenticatedConfig struct {
	unauthenticatedConfig

	Backup            backup           `json:"backup"`
	DHCP              bool             `json:"dhcp"`
	Gateway           string           `json:"gateway"`
	InternetServices  internetServices `json:"internetservices"`
	IPAddress         string           `json:"ipaddress"`
	LinkButtonPressed bool             `json:"linkbutton"`
	LocalTime         string           `json:"localtime"`
	Netmask           string           `json:"netmask"`
	PortalConnection  string           `json:"portalconnection"`
	PortalServices    bool             `json:"portalservices"`
	PortalState       portalState      `json:"portalstate"`
	ProxyAddress      string           `json:"proxyaddress"`
	ProxyPort         int              `json:"proxyport"`
	SWUpdate          swupdate         `json:"swupdate"`
	SWUpdate2         swupdate2        `json:"swupdate2"`
	Timezone          string           `json:"timezone"`
	UTC               string           `json:"UTC"`
	ZigbeeChannel     int              `json:"zigbeechannel"`
}

func (*authenticatedConfig) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func createAuthenticatedConfig(c *Config) *authenticatedConfig {
	t := now()
	resp := &authenticatedConfig{}

	resp.Name = c.Name
	resp.APIVersion = c.APIVersion
	resp.SWVersion = c.SWVersion
	resp.DatastoreVersion = c.DatastoreVersion
	resp.MACAddress = c.MACAddress
	resp.BridgeID = c.BridgeID
	resp.ModelID = c.ModelID

	resp.Backup = backup{
		Status:    "idle",
		ErrorCode: 0,
	}

	resp.DHCP = false
	resp.Gateway = "0.0.0.0"
	resp.InternetServices = internetServices{
		Internet:     "connected",
		RemoteAccess: "disconnected",
		Time:         "connected",
		SWUpdate:     "disconnected",
	}
	resp.IPAddress = c.advertiseIP.String()
	resp.LinkButtonPressed = true
	resp.LocalTime = DateTimeToISO8600(t.In(c.timezone))
	resp.Netmask = "255.255.255.0"
	resp.PortalConnection = "disconnected"
	resp.PortalServices = false
	resp.PortalState = portalState{
		SignedOn:      false,
		Incoming:      false,
		Outgoing:      false,
		Communication: "disconnected",
	}
	resp.ProxyAddress = "none"
	resp.ProxyPort = 0
	resp.SWUpdate = swupdate{
		UpdateState:    0,
		CheckForUpdate: false,
		DeviceTypes: swupdateDeviceTypes{
			Bridge:  false,
			Lights:  []string{},
			Sensors: []string{},
		},
		URL:    "",
		Text:   "",
		Notify: false,
	}
	resp.SWUpdate2 = swupdate2{
		CheckForUpdate: false,
		LastChange:     DateTimeToISO8600(t.UTC()),
		Bridge: bridge{
			State:       "noupdates",
			LastInstall: DateTimeToISO8600(t.UTC()),
		},
		State: "noupdates",
		AutoInstall: autoInstall{
			UpdateTime: "T14:00:00",
			Enabled:    false,
		},
	}
	resp.Timezone = c.timezone.String()
	resp.UTC = DateTimeToISO8600(t.UTC())
	resp.Whitelist = c.Whitelist
	resp.ZigbeeChannel = 15
	return resp
}

func (s *Server) getAuthenticatedConfig(w http.ResponseWriter, r *http.Request) {
	if !r.Context().Value(AuthenticatedCtxKey).(bool) {
		renderOK(w, r, createUnauthenticatedConfig(s.config))
		return
	}
	renderOK(w, r, createAuthenticatedConfig(s.config))
}
