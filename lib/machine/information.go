package machine

import (
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/iesreza/foundation/lib"
	"github.com/iesreza/foundation/lib/network"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// DiskDrive disk drive info struct
type DiskDrive struct {
	Caption      string `json:"name"`
	DeviceID     string `json:"device_id"`
	Model        string `json:"model"`
	Partitions   uint   `json:"partitions"`
	Size         uint64 `json:"size"`
	SerialNumber string `json:"serial_number"`
	Active       bool   `json:"active"`
}

// Partition partition info struct
type Partition struct {
	Name               string  `json:"name" csv:"DeviceID"`
	FileSystem         string  `json:"file_system"`
	Size               uint64  `json:"size"`
	FreeSpace          uint64  `json:"free_space"`
	UsedSpace          uint64  `json:"used_space"`
	FreeSpacePercent   float64 `json:"free_space_percent"`
	VolumeSerialNumber string  `json:"volume_serial_number"`
	Active             bool    `json:"active"`
}

// Memory memory information
type Memory struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

// CPU cpu information
type CPU struct {
	Cores []float64
	Total float64
}

// UniqueHwID generates unique hardware id
func UniqueHwID() (string, error) {
	netconfig, _ := network.GetConfig()
	mac := netconfig.HardwareAddress.String()
	hardDiskId, _ := GetActiveHddSerial()
	biosId, _ := GetBiosId()
	if len(mac) == 17 {
		mac = strings.Replace(mac, ":", "", -1)[6:]
	}
	if len(biosId) > 6 {
		biosId = biosId[len(biosId)-6:]
	}
	if len(hardDiskId) > 2 && hardDiskId[0:2] == "0x" {
		hardDiskId = hardDiskId[2:]
	}
	a := 0
	b := 0
	c := 0

	hwid := ""
	for {

		if a >= 0 && a < len(mac) {
			hwid += string(mac[a])
			a++
		} else {
			a = -1
		}
		if b >= 0 && b < len(hardDiskId) {
			hwid += string(hardDiskId[b])
			b++
		} else {
			b = -1
		}
		if c >= 0 && c < len(biosId) {
			hwid += string(biosId[c])
			c++
		} else {
			c = -1
		}
		if a == -1 && b == -1 && c == -1 {
			break
		}

	}

	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	hwid = reg.ReplaceAllString(hwid, "")

	return hwid, nil
}

// GetBiosId get bios id
func GetBiosId() (string, error) {

	if runtime.GOOS == "windows" {
		res, err := exec.Command(`cmd`, "/C", `wmic csproduct get UUID`).CombinedOutput()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(res)), "\n")
			if len(lines) == 2 {
				return strings.TrimSpace(lines[1]), nil
			}
		}

	} else {
		res, err := exec.Command(`bash`, "-c", `dmidecode -s system-uuid`).CombinedOutput()

		if err == nil {
			return strings.TrimSpace(string(res)), nil
		}
	}
	return "", fmt.Errorf("unable to get bios id")

}

// GetActiveHddSerial get active hard disk serial
func GetActiveHddSerial() (string, error) {
	drive := GetActiveHardDisk()
	if drive.DeviceID == "" {
		return "", fmt.Errorf("active drive not found")

	}
	return drive.SerialNumber, nil
}

// GetHardDisks get list of hard disks
func GetHardDisks() ([]DiskDrive, error) {
	if runtime.GOOS == "windows" {
		res, err := exec.Command("cmd", "/C", `wmic DiskDrive get Caption,DeviceID,Model,Partitions,Size,SerialNumber /format:csv`).CombinedOutput()
		if err != nil {
			return []DiskDrive{}, err
		} else {

			dd := []DiskDrive{}

			clean := strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) || r == '\n' {
					return r
				}
				return -1
			}, string(res))

			err = gocsv.UnmarshalString(clean, &dd)

			if err == nil {
				res, err := exec.Command(`cmd`, "/C", `wmic partition where BootPartition=true get DeviceID`).CombinedOutput()
				if err == nil {
					re := regexp.MustCompile(`(?m)Disk\s+\#(\d+)`)
					match := re.FindAllStringSubmatch(string(res), 1)
					if len(match) > 0 && len(match[0]) == 2 {
						re := regexp.MustCompile(`.+(\d+)$`)
						for k, item := range dd {
							matches := re.FindAllStringSubmatch(item.DeviceID, 1)
							if len(matches) == 1 && len(matches[0]) == 2 && matches[0][1] == match[0][1] {
								dd[k].Active = true
								break
							}
						}

					}
				}
			}

			return dd, err
		}
	} else {
		res, err := exec.Command("bash", "-c", "lsblk -o name,serial,size,mountpoint,vendor,model,type,kname,fstype -b -J").CombinedOutput()
		if err != nil {
			return []DiskDrive{}, err
		} else {
			dd := struct {
				Blockdevices []struct {
					Name       string `json:"name"`
					KernelName string `json:"kname"`
					Serial     string `json:"serial"`
					Size       string `json:"size"`
					Mountpoint string `json:"mountpoint"`
					Vendor     string `json:"vendor"`
					Model      string `json:"model"`
					Type       string `json:"type"`
					Fstype     string `json:"fstype"`
					Children   []struct {
						Name       string      `json:"name"`
						Fstype     interface{} `json:"fstype"`
						Mountpoint string      `json:"mountpoint"`
					} `json:"children,omitempty"`
				} `json:"blockdevices"`
			}{}

			json.Unmarshal(res, &dd)
			response := []DiskDrive{}

			for _, item := range dd.Blockdevices {

				if item.Type != "disk" || item.Mountpoint == "[SWAP]" {
					continue
				}

				drive := DiskDrive{}
				drive.Caption = item.Vendor + " " + item.Model
				drive.DeviceID = item.KernelName
				drive.Model = item.Vendor + " " + item.Model
				drive.SerialNumber = item.Serial
				drive.Size, _ = strconv.ParseUint(item.Size, 0, 64)
				drive.Partitions = uint(len(item.Children))
				for _, child := range item.Children {
					if child.Mountpoint == "/" {
						drive.Active = true
						break
					}
				}
				if drive.Size < 20*lib.MB {
					continue
				}
				response = append(response, drive)

			}
			return response, nil
		}
	}
	return []DiskDrive{}, nil
}

// GetActiveHardDisk get active hard disk
func GetActiveHardDisk() DiskDrive {
	hdds, _ := GetHardDisks()

	for _, item := range hdds {
		if item.Active {
			return item
		}
	}

	return DiskDrive{}
}

// GetPartitions get list of partitions
func GetPartitions() ([]Partition, error) {
	partitions := []Partition{}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	if runtime.GOOS == "windows" {
		res, err := exec.Command("cmd", "/C", `wmic LogicalDisk Where DriveType=3 Get DeviceID,FileSystem,Size,FreeSpace,VolumeSerialNumber /format:csv`).CombinedOutput()

		clean := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) || r == '\n' {
				return r
			}
			return -1
		}, string(res))

		activeDrive := strings.Split(dir, ":")[0] + ":"
		if err == nil {

			err = gocsv.UnmarshalString(clean, &partitions)
			for k, item := range partitions {
				partitions[k].UsedSpace = item.Size - item.FreeSpace
				partitions[k].FreeSpacePercent = float64(item.FreeSpace) * 100 / float64(item.Size)
				partitions[k].Active = activeDrive == item.Name
			}
			return partitions, nil
		}
	} else {
		res, err := exec.Command("bash", "-c", "df -T").CombinedOutput()
		if err == nil {
			lines := strings.Split(string(res), "\n")
			foundActive := false
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) == 7 && strings.HasPrefix(fields[0], "/dev/") && !strings.HasPrefix(fields[0], "/dev/loop") {
					partition := Partition{}
					partition.Name = fields[6]
					partition.FileSystem = fields[1]
					partition.Size, _ = strconv.ParseUint(fields[2], 10, 64)
					partition.Size *= uint64(1024)
					partition.UsedSpace, _ = strconv.ParseUint(fields[3], 10, 64)
					partition.UsedSpace *= uint64(1024)
					partition.FreeSpace = partition.Size - partition.UsedSpace
					partition.FreeSpacePercent = float64(partition.FreeSpace) * 100 / float64(partition.Size)
					if partition.Name != "/" {
						if foundActive == false && strings.HasPrefix(dir, partition.Name) {
							foundActive = true
							partition.Active = true
						}
					}

					res, err := exec.Command("bash", "-c", "findmnt -fn -o UUID "+fields[0]).CombinedOutput()
					if err == nil {
						partition.VolumeSerialNumber = strings.TrimSpace(string(res))
					}
					partitions = append(partitions, partition)
				}
			}
			if foundActive == false {
				for k, partition := range partitions {
					if partition.Name == "/" {
						partitions[k].Active = true
					}
				}
			}

			return partitions, nil
		}

	}
	return partitions, fmt.Errorf("unable to get partitions")

}

// GetActivePartition get active partition
func GetActivePartition() (Partition, error) {
	partitions, err := GetPartitions()
	if err == nil {
		for _, item := range partitions {
			if item.Active {
				return item, nil
			}
		}
	}
	return Partition{}, fmt.Errorf("unable to get active partition")
}

// GetMemory get memory information
func GetMemory() (Memory, error) {
	var response Memory
	v, err := mem.VirtualMemory()
	if err != nil {
		return response, err
	}
	response.Total = v.Total
	response.Used = v.Used
	response.Free = v.Available
	response.UsedPercent = v.UsedPercent
	return response, nil
}

// GetCPUModel return cpu model
func GetCPUModel() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

// GetCPU return cpu usage
func GetCPU(interval time.Duration) (CPU, error) {
	response := CPU{}
	res, err := cpu.Percent(interval, true)
	if err != nil {
		return response, err
	}
	response.Cores = res
	for _, item := range res {
		response.Total += item
	}
	response.Total /= float64(len(res))
	return response, nil
}
