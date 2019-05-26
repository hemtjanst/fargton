package bridge

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/brutella/dnssd"
	"github.com/koron/go-ssdp"
	"go.uber.org/zap"
)

func (s *Server) descriptionXML(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
<specVersion>
<major>1</major>
<minor>0</minor>
</specVersion>
<URLBase>http://%s:%d/</URLBase>
<device>
<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
<friendlyName>Philips hue (%s)</friendlyName>
<manufacturer>Royal Philips Electronics</manufacturer>
<manufacturerURL>http://www.philips.com</manufacturerURL>
<modelDescription>Philips hue Personal Wireless Lighting</modelDescription>
<modelName>Philips hue bridge 2015</modelName>
<modelNumber>%s</modelNumber>
<modelURL>http://www.meethue.com</modelURL>
<serialNumber>%s</serialNumber>
<UDN>uuid:%s</UDN>
<presentationURL>index.html</presentationURL>
<iconList>
<icon>
<mimetype>image/png</mimetype>
<height>48</height>
<width>48</width>
<depth>24</depth>
<url>hue_logo_0.png</url>
</icon>
<icon>
<mimetype>image/png</mimetype>
<height>120</height>
<width>120</width>
<depth>24</depth>
<url>hue_logo_3.png</url>
</icon>
</iconList>
</device>
</root>`,
		s.config.advertiseIP, s.config.port,
		s.config.advertiseIP,
		s.config.ModelID,
		s.config.strippedMAC,
		s.config.uuid)
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func newMDNSService(c *Config) (dnssd.Service, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "fargton-bridge"
	}

	sdConfig := dnssd.Config{
		Name:   fmt.Sprintf("Philips-Hue-%s", c.BridgeID[10:]),
		Type:   "_hue._tcp",
		Host:   fmt.Sprintf("%s-%s", hostname, c.BridgeID[10:]),
		Domain: "local",
		IPs:    nil,
		Port:   int(c.port),
	}
	service, err := dnssd.NewService(sdConfig)
	if err != nil {
		return service, err
	}

	return service, nil
}

func newMDNSResponder(c *Config) (dnssd.Responder, error) {
	rp, err := dnssd.NewResponder()
	if err != nil {
		return rp, err
	}
	sv, err := newMDNSService(c)
	if err != nil {
		return rp, err
	}
	hdl, err := rp.Add(sv)
	if err != nil {
		return rp, err
	}
	hdl.UpdateText(map[string]string{
		"bridgeid": c.BridgeID,
		"modelid":  c.ModelID,
	}, rp)

	return rp, nil
}

func newSSDPResponder(c *Config, l *zap.Logger, quit chan bool) {
	ad1, err := ssdp.Advertise(
		"upnp:rootdevice",
		fmt.Sprintf("uuid:%s::upnp:rootdevice", c.uuid),
		fmt.Sprintf("http://%s:%d/description.xml", c.advertiseIP, c.port),
		fmt.Sprintf("FreeRTOS/7.4.2, UPnP/1.0, IpBridge/%s", c.APIVersion),
		100)
	if err != nil {
		l.Error(err.Error())
	}
	ad2, err := ssdp.Advertise(
		fmt.Sprintf("uuid:%s", c.uuid),
		fmt.Sprintf("uuid:%s", c.uuid),
		fmt.Sprintf("http://%s:%d/description.xml", c.advertiseIP, c.port),
		fmt.Sprintf("FreeRTOS/7.4.2, UPnP/1.0, IpBridge/%s", c.APIVersion),
		100)
	if err != nil {
		l.Error(err.Error())
	}
	ad3, err := ssdp.Advertise(
		"urn:schemas-upnp-org:device:basic:1",
		fmt.Sprintf("uuid:%s", c.uuid),
		fmt.Sprintf("http://%s:%d/description.xml", c.advertiseIP, c.port),
		fmt.Sprintf("FreeRTOS/7.4.2, UPnP/1.0, IpBridge/%s", c.APIVersion),
		100)
	if err != nil {
		l.Error(err.Error())
	}
	aliveTick := time.Tick(60 * time.Second)

	for {
		select {
		case <-aliveTick:
			// Yes, twice
			ad1.Alive()
			ad2.Alive()
			ad3.Alive()
			ad1.Alive()
			ad2.Alive()
			ad3.Alive()
		case <-quit:
			ad1.Bye()
			ad2.Bye()
			ad3.Bye()
			time.Sleep(5 * time.Millisecond)
			ad1.Close()
			ad2.Close()
			ad3.Close()
			l.Info("stopped SSDP/UPnP responder")
			return
		}
	}
}
