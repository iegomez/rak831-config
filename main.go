package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
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

func main() {
	ethName := flag.String("eth", "eth0", "eth interface name")
	wlanName := flag.String("wlan", "wlan0", "wlan interface name")
	lcPath := flag.String("lc", "/opt/ttn-gateway/bin/local_conf.json", "local conf path")
	gpsPath := flag.String("gpsPath", "/dev/ttyAMA0", "gps tty path")
	hostName := flag.String("host", "", "set host name")
	serverName := flag.String("server", "localhost", "server to forward packets to")
	upPort := flag.Int("up", 1700, "udp up port")
	downPort := flag.Int("down", 1700, "udp down port")

	flag.Parse()

	server := Server{
		ServerAddress: *serverName,
		ServPortUp:    *upPort,
		ServPortDown:  *downPort,
		ServEnabled:   true,
	}

	gwConf := GWConf{
		GWC: GatewayConf{
			RefLatitude:  -33.433567,
			RefLongitude: -70.6217137,
			RefAltitude:  600,
			ContactEmail: "contacto@manglar.cl",
			Description:  "Manglar gateway",
			Servers: []Server{
				server,
			},
		},
	}

	if *gpsPath != "" {
		gwConf.GWC.FakeGPS = false
		gwConf.GWC.GPS = true
		gwConf.GWC.GPSttyPath = *gpsPath
	}

	fmt.Println(*ethName)
	fmt.Println(*wlanName)
	fmt.Println(*lcPath)

	addr, err := getMacAddr(*ethName)
	if err != nil {
		addr, err = getMacAddr(*wlanName)
		if err != nil {
			fmt.Println(err)
		} else {
			addr, err = formatAddr(addr)
			if err != nil {
				panic(err)
			}
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
	fmt.Println([]byte(*hostName))
	//unix.Sethostname([]byte(*hostName))
	jb, err := json.Marshal(gwConf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jb))

}

func getMacAddr(ifName string) (string, error) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", err
	}

	return iface.HardwareAddr.String(), nil
}

func formatAddr(addr string) (string, error) {
	nAddr := strings.Replace(addr, ":", "", -1)
	if len(nAddr) != 12 {
		return "", errors.New("address format not recognized")
	}
	nAddr = strings.ToUpper(fmt.Sprintf("%sFFFE%s", nAddr[:6], nAddr[6:]))
	return nAddr, nil
}
