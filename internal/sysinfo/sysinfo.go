package sysinfo

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

type DiskInfo struct {
	Mountpoint string  `json:"mountpoint"`
	Device     string  `json:"device"`
	Total      uint64  `json:"total"`
	Free       uint64  `json:"free"`
	UsedPct    float64 `json:"usedPct"`
	Fstype     string  `json:"fstype"`
}

type NetInterface struct {
	Name     string `json:"name"`
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
}

type CPUTemp struct {
	Label       string  `json:"label"`
	Temperature float64 `json:"temperature"`
}

type SystemInfo struct {
	Disks         []DiskInfo     `json:"disks"`
	MemTotal      uint64         `json:"memTotal"`
	MemUsed       uint64         `json:"memUsed"`
	MemUsedPct    float64        `json:"memUsedPct"`
	SwapTotal     uint64         `json:"swapTotal"`
	SwapUsed      uint64         `json:"swapUsed"`
	SwapUsedPct   float64        `json:"swapUsedPct"`
	UptimeSeconds uint64         `json:"uptimeSeconds"`
	CPUUsagePct   float64        `json:"cpuUsagePct"`
	CPUModel      string         `json:"cpuModel"`
	CPUCores      int            `json:"cpuCores"`
	CPUTemps      []CPUTemp      `json:"cpuTemps"`
	NetInterfaces []NetInterface `json:"netInterfaces"`
	Hostname      string         `json:"hostname"`
	OS            string         `json:"os"`
	KernelVersion string         `json:"kernelVersion"`
	LoadAvg1      float64        `json:"loadAvg1"`
	LoadAvg5      float64        `json:"loadAvg5"`
	LoadAvg15     float64        `json:"loadAvg15"`
	NumProcesses  uint64         `json:"numProcesses"`
}

// isVirtualFS returns true for pseudo/virtual filesystem types that should
// not appear as disk partitions in the system info.
func isVirtualFS(fstype string) bool {
	switch fstype {
	case "sysfs", "proc", "devtmpfs", "devpts", "tmpfs", "securityfs",
		"cgroup", "cgroup2", "pstore", "debugfs", "hugetlbfs", "mqueue",
		"configfs", "fusectl", "binfmt_misc", "autofs", "efivarfs",
		"bpf", "tracefs", "overlay", "nsfs", "ramfs", "rpc_pipefs",
		"nfsd", "fuse.gvfsd-fuse", "fuse.portal":
		return true
	}
	return false
}

func getDisks() []DiskInfo {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil
	}
	var disks []DiskInfo
	for _, p := range partitions {
		if isVirtualFS(p.Fstype) {
			continue
		}
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, DiskInfo{
			Mountpoint: p.Mountpoint,
			Device:     p.Device,
			Total:      usage.Total,
			Free:       usage.Free,
			UsedPct:    usage.UsedPercent,
			Fstype:     p.Fstype,
		})
	}
	return disks
}

func Get() (SystemInfo, error) {
	var info SystemInfo

	info.Disks = getDisks()

	// Memory
	v, err := mem.VirtualMemory()
	if err != nil {
		return info, err
	}
	info.MemTotal = v.Total
	info.MemUsed = v.Used
	info.MemUsedPct = v.UsedPercent

	// Swap
	sw, err := mem.SwapMemory()
	if err == nil {
		info.SwapTotal = sw.Total
		info.SwapUsed = sw.Used
		info.SwapUsedPct = sw.UsedPercent
	}

	// Uptime
	uptime, err := host.Uptime()
	if err != nil {
		return info, err
	}
	info.UptimeSeconds = uptime

	// CPU usage (200ms sample)
	pcts, err := cpu.Percent(200*time.Millisecond, false)
	if err == nil && len(pcts) > 0 {
		info.CPUUsagePct = pcts[0]
	}

	// CPU model + cores
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.CPUModel = cpuInfos[0].ModelName
	}
	info.CPUCores = runtime.NumCPU()

	// CPU temperatures (best-effort, may not be available)
	temps, err := sensors.TemperaturesWithContext(context.Background())
	if err == nil {
		for _, t := range temps {
			if t.Temperature > 0 {
				info.CPUTemps = append(info.CPUTemps, CPUTemp{
					Label:       t.SensorKey,
					Temperature: t.Temperature,
				})
			}
		}
	}

	// Network interfaces
	counters, err := net.IOCounters(true)
	if err == nil {
		for _, c := range counters {
			if c.BytesSent == 0 && c.BytesRecv == 0 {
				continue
			}
			info.NetInterfaces = append(info.NetInterfaces, NetInterface{
				Name:      c.Name,
				BytesSent: c.BytesSent,
				BytesRecv: c.BytesRecv,
			})
		}
	}

	// Host info
	hostInfo, err := host.Info()
	if err == nil {
		info.Hostname = hostInfo.Hostname
		info.OS = hostInfo.Platform + " " + hostInfo.PlatformVersion
		info.KernelVersion = hostInfo.KernelVersion
		info.NumProcesses = hostInfo.Procs
	}

	// Load average (Linux/macOS)
	loadAvg, err := load.Avg()
	if err == nil {
		info.LoadAvg1 = loadAvg.Load1
		info.LoadAvg5 = loadAvg.Load5
		info.LoadAvg15 = loadAvg.Load15
	}

	return info, nil
}
