package network

import (
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// Network is the interface which store network configuration data
type Network struct {
	LocalIP                       net.IP
	DNS                           []string
	SubnetMask                    net.IP
	DefaultGateway                net.IP
	DefaultGatewayHardwareAddress net.HardwareAddr
	InterfaceName                 string
	HardwareAddress               net.HardwareAddr
	Suffix                        string
	Interface                     *net.Interface
}

var instance *Network

// RefreshConfig refetch network configuration
func RefreshConfig() (*Network, error) {
	if instance != nil {
		instance = nil
	}
	return GetConfig()
}

// GetConfig return  instance of network configuration.
func GetConfig() (*Network, error) {
	if instance != nil {
		return instance, nil
	}
	network := Network{}

	if runtime.GOOS == "windows" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		network.LocalIP = conn.LocalAddr().(*net.UDPAddr).IP

		interfaces, _ := net.Interfaces()
		for _, interf := range interfaces {

			if addrs, err := interf.Addrs(); err == nil {
				for _, addr := range addrs {
					if strings.Contains(addr.String(), network.LocalIP.String()) {
						network.InterfaceName = interf.Name
						network.HardwareAddress = interf.HardwareAddr
						network.Interface = &interf
					}
				}
			}
		}

		network.getWindows()
	} else {
		network.getLinux()
	}
	instance = &network
	return &network, nil
}

// getLinux read network data for linux
func (network *Network) getLinux() error {

	out, err := exec.Command("/bin/ip", "route", "get", "8.8.8.8").Output()

	if err != nil {
		return err
	}
	parts := strings.Split(string(out), " ")
	network.DefaultGateway = net.ParseIP(parts[2])
	network.InterfaceName = parts[4]
	network.LocalIP = net.ParseIP(parts[6])

	interf, err := net.InterfaceByName(network.InterfaceName)
	if err == nil {
		network.HardwareAddress = interf.HardwareAddr
		network.Interface = interf

	} else {
		return err
	}

	out, err = exec.Command("/sbin/ifconfig", network.InterfaceName).Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")

		if len(lines) > 1 {
			network.SubnetMask = net.ParseIP(strings.Split(strings.TrimSpace(lines[1]), " ")[4])
		}
	} else {
		return err
	}

	out, err = exec.Command("grep", "domain-name", "/var/lib/dhcp/dhclient."+network.InterfaceName+".leases").Output()

	if err == nil {
		dnslist := ""
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if strings.Contains(line, "domain-name-servers") {
				if len(line) > 26 {
					line = strings.TrimRight(strings.TrimSpace(line)[26:], ";")
					list := strings.Split(line, ",")
					for _, dnsitem := range list {
						if !strings.Contains(dnslist, dnsitem) {
							dnslist += dnsitem + ","
						}
					}
				}

			} else {
				if len(line) > 18 {
					network.Suffix = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(line)[18:], ";"))
				}
			}
			dnslist = strings.TrimRight(dnslist, ",")
		}

		network.DNS = strings.Split(dnslist, ",")
	} else {
		return err
	}
	out, err = exec.Command("arp", "-e", network.DefaultGateway.String()).Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")

		if len(lines) >= 2 {

			network.DefaultGatewayHardwareAddress, _ = net.ParseMAC(strings.Fields(lines[1])[2])
		}
	} else {
		return err
	}
	return nil
}

// String return network information as string
func (network *Network) String() string {

	res := "InterfaceName:" + network.InterfaceName + "\r\n"
	res += "HardwareAddress:" + network.HardwareAddress.String() + "\r\n"
	res += "LocalIP:" + network.LocalIP.String() + "\r\n"
	res += "DNS:" + strings.Join(network.DNS, ",") + "\r\n"
	res += "SubnetMask:" + network.SubnetMask.String() + "\r\n"
	res += "DefaultGateway:" + network.DefaultGateway.String() + "\r\n"
	res += "DefaultGatewayHardwareAddress:" + network.DefaultGatewayHardwareAddress.String() + "\r\n"
	res += "Suffix:" + network.Suffix + "\r\n"

	return res
}

// getWindows read network data in windows
func (network *Network) getWindows() error {
	out, err := exec.Command("ipconfig", "/all").Output()
	if err != nil {
		return err
	}
	items := strings.Split(string(out), "Ethernet adapter ")
	for _, item := range items {
		lines := strings.Split(item, "\r\n")
		if strings.HasPrefix(item, network.InterfaceName) {

			network.DNS = extractDotted(lines, "DNS Servers")
			if network.Suffix == "" {
				network.Suffix = extractDotted(lines, "Connection-specific DNS Suffix")[0]
			}
			network.SubnetMask = net.ParseIP(extractDotted(lines, "Subnet Mask")[0])

		}
		for _, line := range lines {
			if strings.Contains(line, "Default Gateway") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					ip := net.ParseIP(strings.TrimSpace(parts[1]))
					if ip != nil {
						network.DefaultGateway = ip
					}
				}
			}
		}
	}

	out, err = exec.Command("arp", "-a", network.DefaultGateway.String()).Output()
	if err != nil {
		return err
	}
	chunks := strings.Split(string(out), network.DefaultGateway.String())

	if len(chunks) == 3 {
		network.DefaultGatewayHardwareAddress, _ = net.ParseMAC(strings.Fields(chunks[2])[0])
	}
	return nil
}

// extractDotted extract data of ipconfig
func extractDotted(lines []string, key string) []string {
	result := ""
	found := false

	for _, line := range lines {
		if !found {
			if strings.HasPrefix(line, "   "+key) {
				result = line[39:] + ""
				found = true
			}
		} else {

			if len(line) > 39 && strings.TrimSpace(line[0:39]) == "" {
				result += "," + strings.TrimSpace(line[39:])
			} else {
				break
			}
		}

	}

	return strings.Split(strings.Trim(result, ","), ",")
}
