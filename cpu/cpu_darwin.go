// +build darwin

package cpu

/*
#include <stdlib.h>
#include <sys/sysctl.h>
#include <sys/mount.h>
#include <mach/mach_init.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <libproc.h>
#include <mach/processor_info.h>
#include <mach/vm_map.h>
*/
import "C"

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

// sys/resource.h
const (
	CPUser    = 0
	CPNice    = 1
	CPSys     = 2
	CPIntr    = 3
	CPIdle    = 4
	CPUStates = 5
)

// time.h
const (
	ClocksPerSec = 128
)

func CPUTimes(percpu bool) ([]CPUTimesStat, error) {

	var ret []CPUTimesStat
	var count C.mach_msg_type_number_t = C.HOST_CPU_LOAD_INFO_COUNT
	var cpuload C.host_cpu_load_info_data_t

	status := C.host_statistics(C.host_t(C.mach_host_self()),
		C.HOST_CPU_LOAD_INFO,
		C.host_info_t(unsafe.Pointer(&cpuload)),
		&count)

	if status != C.KERN_SUCCESS {
		return nil, fmt.Errorf("host_statistics error=%d", status)
	}

	cpu := CPUTimesStat{
		User:   float32(cpuload.cpu_ticks[C.CPU_STATE_USER]),
		System: float32(cpuload.cpu_ticks[C.CPU_STATE_SYSTEM]),
		Idle:   float32(cpuload.cpu_ticks[C.CPU_STATE_IDLE]),
		Nice:   float32(cpuload.cpu_ticks[C.CPU_STATE_NICE]),
	}

	ret = append(ret, cpu)

	return ret, nil
}

// Returns only one CPUInfoStat on FreeBSD
func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat

	out, err := exec.Command("/usr/sbin/sysctl", "machdep.cpu").Output()
	if err != nil {
		return ret, err
	}

	c := CPUInfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		if len(values) < 1 {
			continue
		}

		t, err := strconv.ParseInt(values[1], 10, 64)
		// err is not checked here because some value is string.
		if strings.HasPrefix(line, "machdep.cpu.brand_string") {
			c.ModelName = strings.Join(values[1:], " ")
		} else if strings.HasPrefix(line, "machdep.cpu.family") {
			c.Family = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.model") {
			c.Model = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.stepping") {
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.leaf7_features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.extfeatures") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.core_count") {
			if err != nil {
				return ret, err
			}
			c.Cores = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}

		// TODO:
		// c.Mhz = mustParseFloat64(values[1])
	}

	return append(ret, c), nil
}
