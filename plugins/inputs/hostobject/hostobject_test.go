package hostobject

import (
	"log"
	"net"
	"testing"
)

func TestOSDetail(t *testing.T) {

	oi := getOSInfo()
	log.Printf("%s", oi.Release)

}

func TestGetIP(t *testing.T) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	log.Printf("%s", localAddr.IP)
}

func TestListInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
	}
	for _, i := range ifaces {
		log.Printf("mac: %s", i.HardwareAddr.String())

		addrs, err := i.Addrs()
		if err != nil {
			t.Error(err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				if ip.To4() != nil {
					log.Printf("%s", ip)
				}
			case *net.IPAddr:
				ip = v.IP
			}
		}
	}
}
