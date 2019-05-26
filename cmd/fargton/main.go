package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"hemtjan.st/fargton/bridge"
	"lib.hemtjan.st/server"
	"lib.hemtjan.st/transport/mqtt"
)

func main() {
	mqttCfg := mqtt.MustFlags(flag.String, flag.Bool)
	flgName := flag.String("bridge.name", "Philips hue", "Hue bridge name")
	flgAddress := flag.String("bridge.listen-address", "0.0.0.0:0", "address:port the bridge will listen on")
	flgMAC := flag.String("bridge.mac", "00:17:88:a1:b2:c3", "MAC address for this bridge (only used for config)")
	flgIP := flag.String("bridge.ip", "", "IP address to advertise the bridge on")

	flgTLSAddress := flag.String("bridge.tls-listen-address", "0.0.0.0:0", "address:port the bridge will listen on for TLS connections")
	flgTLSPrivKey := flag.String("bridge.tls-private-key", "./private.key", "path to TLS private key")
	flgTLSPubKey := flag.String("bridge.tls-public-key", "./public.crt", "path to TLS public key")

	flgWhitelist := flag.String("bridge.whitelist", "./whitelist.json", "path to where we will load and store whitelist entries")

	flgAuth := flag.Bool("bridge.auth-disable", false, "Disable checking requests against whitelist")

	flgLatitude := flag.Float64("location.lat", 0, "latitude of the bridge location")
	flgLongitude := flag.Float64("location.long", 0, "longitude of the bridge location")

	flag.Parse()

	l, _ := zap.NewDevelopment()
	defer l.Sync()

	m, err := mqtt.New(context.Background(), mqttCfg())
	if err != nil {
		l.Fatal(err.Error())
	}

	mqttManager := server.New(m)

	cfg, err := bridge.NewConfig(
		bridge.Name(*flgName),
		bridge.Address(*flgAddress),
		bridge.TLSAddress(*flgTLSAddress),
		bridge.MAC(*flgMAC),
		bridge.AdvertiseIP(*flgIP),
		bridge.TLSPublicKeyPath(*flgTLSPubKey),
		bridge.TLSPrivateKeyPath(*flgTLSPrivKey),
		bridge.DisableAuthentication(*flgAuth),
		bridge.WhitelistConfigPath(*flgWhitelist),
		bridge.Latitude(*flgLatitude),
		bridge.Longitude(*flgLongitude),
	)
	if err != nil {
		l.Fatal(err.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	s := bridge.NewServer(cfg, mqttManager, l)
	l.Info("initiating server startup")
	shutdown, err := s.Start(3 * time.Second)
	if err != nil {
		l.Fatal(err.Error())
	}
	l.Info("completed server startup")

	<-stop
	l.Info("initiating server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdown(ctx)

	l.Info("completed server shutdown")
}
