package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type GatewayConf struct {
	GatewayID    string   `json:"gateway_ID"`
	Servers      []Server `json:"servers"`
	RefLatitude  float64  `json:"ref_latitude"`
	RefLongitude float64  `json:"ref_longitude"`
	RefAltitude  int      `json:"ref_altitude"`
	ContactEmail string   `json:"contact_email"`
	Description  string   `json:"description"`
	FakeGPS      bool     `json:"fake_gps"`
	GPS          bool     `json:"gps"`
	GPSttyPath   string   `json:"gps_tty_path"`
}

type Server struct {
	ServerAddress string `json:"server_address"`
	ServPortUp    int    `json:"serv_port_up"`
	ServPortDown  int    `json:"serv_port_down"`
	ServEnabled   bool   `json:"serv_enabled"`
}

type GWConf struct {
	GWC GatewayConf `json:"gateway_conf"`
}

var bands = map[string]string{
	"AS1": "AS1-global_conf.json",
	"AS2": "AS2-global_conf.json",
	"AU":  "AU-global_conf.json",
	"CN":  "CN-global_conf.json",
	"EU":  "EU-global_conf.json",
	"IN":  "IN-global_conf.json",
	"KR":  "KR-global_conf.json",
	"RU":  "RU-global_conf.json",
	"US":  "US-global_conf.json",
}

func main() {
	ethName := flag.String("eth", "eth0", "eth interface name")
	wlanName := flag.String("wlan", "wlan0", "wlan interface name")
	gcPath := flag.String("gc", "/opt/ttn-gateway/bin/global_conf.json", "global conf path")
	lcPath := flag.String("lc", "/opt/ttn-gateway/bin/local_conf.json", "local conf path")
	gpsOption := flag.String("gpso", "gps", "options are gps, fake and none")
	gpsPath := flag.String("gpsp", "/dev/ttyS0", "gps tty path when using a gps")
	hostName := flag.String("host", "", "set host name")
	serverName := flag.String("server", "localhost", "server to forward packets to")
	upPort := flag.Int("up", 1700, "udp up port")
	downPort := flag.Int("down", 1700, "udp down port")
	band := flag.String("band", "AU", "band for global_conf.json: AS1, AS2, AU, CN, EU, IN, KR, RU OR US")
	lat := flag.Float64("lat", -33.433567, "ref latitude")
	lng := flag.Float64("lng", -70.6217137, "ref longitude")
	alt := flag.Int("alt", 600, "ref altitude")

	flag.Parse()

	server := Server{
		ServerAddress: *serverName,
		ServPortUp:    *upPort,
		ServPortDown:  *downPort,
		ServEnabled:   true,
	}

	gwConf := GWConf{
		GWC: GatewayConf{
			RefLatitude:  *lat,
			RefLongitude: *lng,
			RefAltitude:  *alt,
			ContactEmail: "contacto@manglar.cl",
			Description:  "Manglar GW",
			Servers: []Server{
				server,
			},
		},
	}

	if *gpsOption == "gps" && *gpsPath != "" {
		gwConf.GWC.GPS = true
		gwConf.GWC.FakeGPS = false
		gwConf.GWC.GPSttyPath = *gpsPath
	} else if *gpsOption == "fake" {
		gwConf.GWC.GPS = false
		gwConf.GWC.FakeGPS = true
	}

	addr, err := getMacAddr(*ethName)
	if err != nil {
		addr, err = getMacAddr(*wlanName)
		if err != nil {
			panic(err)
		}
	}

	gwConf.GWC.GatewayID = addr

	if *hostName == "" {
		hostName = &gwConf.GWC.GatewayID
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter hostname: ")
		scanner.Scan()
		text := scanner.Text()
		if text != "" {
			hostName = &text
		}
	}
	unix.Sethostname([]byte(*hostName))
	jb, err := json.Marshal(gwConf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Your Gateway ID is %s\n", addr)

	err = ioutil.WriteFile(*lcPath, jb, 0644)
	if err != nil {
		panic(err)
	}

	err = setGlobalConf(*band, *gcPath)
	if err != nil {
		panic(err)
	}

}

func getMacAddr(ifName string) (string, error) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", err
	}

	addr, err := formatAddr(iface.HardwareAddr.String())
	if err != nil {
		return "", err
	}

	return addr, nil
}

func formatAddr(addr string) (string, error) {
	nAddr := strings.Replace(addr, ":", "", -1)
	if len(nAddr) != 12 {
		return "", errors.New("address format not recognized")
	}
	nAddr = strings.ToUpper(fmt.Sprintf("%sFFFE%s", nAddr[:6], nAddr[6:]))
	return nAddr, nil
}

func setGlobalConf(gn, path string) error {
	input, err := ioutil.ReadFile(fmt.Sprintf("bands/%s", bands[gn]))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, input, 0644)
	if err != nil {
		return err
	}
	return nil
}
