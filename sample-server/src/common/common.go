package common

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	TimeFmt = "2006-02-01 15:04:05.000"
)

type Request struct {
	Id   int
	Time string
	Uid  string
}

type SystemInfo struct {
	HostName      string
	KernelVersion string

	LogicalCpuCores  int32
	PhysicalCpuCores int32
	CpuMHZ           float64
	Memory           string
}

func ObtainClientInfo(r *http.Request) string {
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("client info:%s:%s", ip, port)
}
func ObtainSystemInfo(r *http.Request) *SystemInfo {
	systemInfo := &SystemInfo{}
	systemInfo.KernelVersion, _ = host.KernelVersion()
	hostInfo, _ := host.Info()
	systemInfo.HostName = hostInfo.Hostname

	cpuInfoStats, _ := cpu.Info()

	for _, info := range cpuInfoStats {
		if systemInfo.CpuMHZ == float64(0) {
			systemInfo.CpuMHZ = info.Mhz
		}
		systemInfo.LogicalCpuCores = systemInfo.LogicalCpuCores + info.Cores
		systemInfo.PhysicalCpuCores = systemInfo.PhysicalCpuCores + 1
	}

	vm, _ := mem.VirtualMemory()
	systemInfo.Memory = fmt.Sprintf("%d%s", strconv.FormatUint(vm.Total/1024/1024, 10), "mb")
	return systemInfo
}
func RequestToString(id int) ([]byte, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	req := Request{
		Id:   id,
		Time: time.Now().Format(TimeFmt),
		Uid:  uid.String(),
	}
	b, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	return b, nil
}
