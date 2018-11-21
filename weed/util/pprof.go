package util

import (
	"os"
	"runtime/pprof"

	"github.com/draleyva/seaweedfs/weed/glog"
)

func SetupProfiling(cpuProfile, memProfile string) {
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			glog.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		OnInterrupt(func() {
			pprof.StopCPUProfile()
		})
	}
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			glog.Fatal(err)
		}
		OnInterrupt(func() {
			pprof.WriteHeapProfile(f)
			f.Close()
		})
	}

}
