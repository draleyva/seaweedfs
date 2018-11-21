package filer2

import (
	"os"

	"github.com/draleyva/seaweedfs/weed/glog"
	"github.com/spf13/viper"
)

var (
	Stores []FilerStore
)

func (f *Filer) LoadConfiguration(config *viper.Viper) {

	for _, store := range Stores {
		if config.GetBool(store.GetName() + ".enabled") {
			viperSub := config.Sub(store.GetName())
			if err := store.Initialize(viperSub); err != nil {
				glog.Fatalf("Failed to initialize store for %s: %+v",
					store.GetName(), err)
			}
			f.SetStore(store)
			glog.V(0).Infof("Configure filer for %s", store.GetName())
			return
		}
	}

	println()
	println("Supported filer stores are:")
	for _, store := range Stores {
		println("    " + store.GetName())
	}

	os.Exit(-1)
}
