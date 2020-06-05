package network

import (
	"fmt"
	"github.com/getevo/evo/lib"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// PingResult ping result struct
type PingResult struct {
	Min, Max, Avg, Loss, Dev int
	Resolved                 bool
}

// String return ping result in string
func (result PingResult) String() string {
	if !result.Resolved {
		return "Timed out"
	}

	return fmt.Sprintf("Min:%d Max:%d Avg:%d Deviation:%d Loss:%d%%", result.Min, result.Max, result.Avg, result.Dev, result.Loss)
}

var windowsPingRegex = regexp.MustCompile(`(?s)(\d+)%\s*loss.*Minimum\s*=\s*(\d+).*Maximum\s*=\s*(\d+).*Average\s*=\s*(\d+).*`)
var linuxPingRegex = regexp.MustCompile(`(?s)(\d+)%\s*packet loss.*rtt.*mdev = (\d+).*?\/(\d+).*?\/(\d+)`)

// Ping ping destination return the info
func Ping(host string, pings int) (PingResult, error) {
	var result PingResult
	if runtime.GOOS == "windows" {
		output, err := exec.Command("ping", host, "-n", strconv.Itoa(pings)).Output()
		if err != nil {
			return result, err
		}
		if strings.Contains(string(output), "Ping statistics") {

			res := windowsPingRegex.FindStringSubmatch(string(output))
			result.Loss = lib.ParseSafeInt(res[1])
			result.Min = lib.ParseSafeInt(res[2])
			result.Max = lib.ParseSafeInt(res[3])
			result.Avg = lib.ParseSafeInt(res[4])
			result.Dev = result.Max - result.Min
			result.Resolved = true
		} else {
			result.Resolved = false
			result.Loss = 100
		}
	} else {
		output, err := exec.Command("ping", host, "-c", strconv.Itoa(pings)).Output()
		if err != nil {
			return result, err
		}
		if strings.Contains(string(output), "ping statistics") {
			res := linuxPingRegex.FindStringSubmatch(string(output))
			result.Loss = lib.ParseSafeInt(res[1])
			result.Min = lib.ParseSafeInt(res[2])
			result.Avg = lib.ParseSafeInt(res[3])
			result.Max = lib.ParseSafeInt(res[4])
			result.Dev = result.Max - result.Min
			result.Resolved = true
		} else {
			result.Resolved = false
			result.Loss = 100
		}
	}
	return result, nil
}
