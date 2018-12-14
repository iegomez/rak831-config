# rak831-config

This is a simple Rak831 + Raspberry Pi 3 configurator that allows to quickly replace `global_conf.json` with the desired band defaults (taken from [TTN's configurations](https://github.com/TheThingsNetwork/gateway-conf)), and `local_conf.json` with typical arguments. It's meant to be used after installing [ic880a-gateway](https://github.com/ttn-zh/ic880a-gateway.git) and for reconfiguring a gateway.  

The SD card that comes with the Rak831 tends to have issues, so I prefer to download a **fresh Raspbian image** and follow this configuration guide.

### Build

You need to have Go installed to build this program. There are no dependencies apart from the stdlib other than `golang.org/x/sys/unix` to allow setting the hostname. If you don't have it, get it with:

```
go get golang.org/x/sys/unix
```

You may cross compile for the Raspberry Pi 3 B+ with `make rpi`, or compile directly with `make`. Just change `GOARM` at the Makefile for other versions of the RPi.

### Usage

Run `./rak831 -h` to check the flags:

```
Usage of ./rak831:
  -alt int
    	ref altitude (default 600)
  -band string
    	band for global_conf.json: AS1, AS2, AU, CN, EU, IN, KR, RU OR US (default "AU")
  -down int
    	udp down port (default 1700)
  -eth string
    	eth interface name (default "eth0")
  -gc string
    	global conf path (default "/opt/ttn-gateway/bin/global_conf.json")
  -gpso string
    	options are gps, fake and none (default "gps")
  -gpsp string
    	gps tty path when using a gps
  -host string
    	name to set machine's hostname
  -lat float
    	ref latitude (default -33.433567)
  -lc string
    	local conf path (default "/opt/ttn-gateway/bin/local_conf.json")
  -lng float
    	ref longitude (default -70.6217137)
  -server string
    	server to forward packets to (default "localhost")
  -up int
    	udp up port (default 1700)
  -wlan string
    	wlan interface name (default "wlan0")
```
If no `host` argument is given, you'll be prompted to enter your hostname.  
The program will print your Gateway's ID (e.g., `Your Gateway ID is A0D6BFFFFEB5CEE2`), replace `global_conf.json` and **rewrite** `local_conf.json` (make sure it exists first, which should be true after running the `ic880a-gateway/install.sh`).

### Gateway conf

These are a series of typical configuration steps for a Rak831 based gateway.  
  
First, modify your SD `boot` partition to get wifi and ssh access by creating an empty file `boot/ssh` and a `wpa_supplicant` configuration file `boot/wpa_supplicant.conf` with your wifi configuration:

```
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=CL

network={
        ssid="your-ssid"
        psk="your-password"
}
```
  
Now boot your gateway to configure it.  
Start by running `sudo raspi-config` and changing your password and configuring localisation, enabling SPI at the interfaces and expanding your root filesystem.  
If you are using a GPS, then also disable serial login shell and enable hardware serial at the interfaces options.

Install necessary packages:
```
apt-get install git
apt-get install dirmngr
```
Get the ic880a-gateway code:
```
git clone -b spi https://github.com/ttn-zh/ic880a-gateway.git
cd ic880a-gateway
```
Open `start.sh` and change reset pin from 25 to 17, then install with `sudo ./install.sh`.  
This will install default configurations at `/opt/ttn-gateway/bin`. You may override them by hand or using this program by running `sudo ./rak831rpi`.

Finally, install and configure an optional gateway bridge (use https://www.loraserver.io/lora-gateway-bridge/overview/downloads/ for loraserver), and restart your services:

```
sudo service ttn-gateway restart
sudos ervice lora-gateway-bridge restart
```