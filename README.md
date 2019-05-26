# 🎨 Färgton 💡

Färgton (Swedish for hue/tint/tone) is a Philips Hue bridge exposing Hemtjänst
lights, light groups and sensors. It emulates a second generation Hue bridge
(the square one).

Please note that many endpoints will be read-only, despite the fact that on a
real bridge they might allow for modification. Only things that can normally
be controlled through Hemtjänst, like on/off state, colour etc. will be
controllable through this bridge.

It supports:

* [ ] Philips Hue REST API
    * [x] Discovery
        * [x] mDNS (DNS Service Discovery / zeroconf / Bonjour)
        * [x] SSDP (UPnP)
        * [x] `/description.xml`
    * [x] Authentication / Registration
    * [x] Configuration
        * Is read-only except for adding/deleting an entry from the whitelist
    * [x] Lights
    * [x] Groups
        * One group, a room, is created per light
        * Will be controllable once Hemtjanst gains a groups concept
    * [x] ~~Schedules~~
        * Returns empty
        * Use [Node-RED][nodered] for this
    * [x] ~~Scenes~~
        * Returns empty
        * Use [Node-RED][nodered] for this
    * [ ] Sensors
        * Only returns the hardcoded daylight sensor for now
    * [x] ~~Rules~~
        * Returns empty
        * Use [Node-RED][nodered] for this
    * [x] ~~Resource links~~
        * Returns empty since there is nothing right now to group multiple
          resources together
    * [x] Capabilities
* [ ] Philips Hue Entertainment API

[nodered]: https://nodered.org/

## Supported applications

Due to the implementation of SSDP and mDNS any application that follows the
[Hue Bridge Discovery][hbd] will be able to find this device. Some
implementations assume the API runs on port 80 and won't work if you run
Färgton on a different port.

### Official Hue app

This has only been tested with the Hue 3.18+ Android and iOS apps. It's
known to crash the 3.10

For the official Hue app to work with Färgton you'll need to bind the bridge
on port 80 for plain text traffic and port 443 for TLS encrypted traffic.

When the Hue app first connects it will fetch `http://ip:80/api/nouser/conf`
and switch to HTTPS after that. If you don't enable HTTPS the official Hue
app will not work.

#### 🔐 Generating TLS certificates

You'll need to generate a public/private keypair for the bridge. It's
important to note that the MAC address of the bridge becomes part of the
TLS certificate so if you change `-bridge.mac-address` you'll need to
generate a new certificate.

First create an `openssl.conf`:

```text
[ req ]
default_bits            = 1024
default_md              = sha256
default_keyfile         = privkey.pem
distinguished_name      = req_distinguished_name
attributes              = req_attributes
req_extensions  = v3_req
x509_extensions = usr_cert

[ usr_cert ]
basicConstraints=critical,CA:FALSE
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid,issuer
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[ v3_req ]
extendedKeyUsage = serverAuth, clientAuth, codeSigning, emailProtection
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment

[ req_distinguished_name ]

[ req_attributes ]
```

Now generate the certs:

```bash
$ mac="mac address of bridge"
$ serial="${mac:0:2}${mac:3:2}${mac:6:2}fffe${mac:9:2}${mac:12:2}${mac:15:2}"
$ dec_serial=`python3 -c "print(int(\"$serial\", 16))"`
$ openssl req -new -config openssl.conf  -nodes -x509 -newkey  ec -pkeyopt ec_paramgen_curve:P-256 -pkeyopt ec_param_enc:named_curve   -subj "/C=NL/O=Philips Hue/CN=$serial" -keyout private.key -out public.crt -set_serial $dec_serial -days 3650
...
```

Pass the path to the public and private keys to `-bridge.tls-public-key` and
`-bridge.tls-private-key` respectively.

[hbd]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/
