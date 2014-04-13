package topology

import (
	"code.google.com/p/weed-fs/go/glog"
	"code.google.com/p/weed-fs/go/storage"
	"fmt"
	"math/rand"
	"sync"
)

/*
This package is created to resolve these replica placement issues:
1. growth factor for each replica level, e.g., add 10 volumes for 1 copy, 20 volumes for 2 copies, 30 volumes for 3 copies
2. in time of tight storage, how to reduce replica level
3. optimizing for hot data on faster disk, cold data on cheaper storage,
4. volume allocation for each bucket
*/

type VolumeGrowOption struct {
	Collection       string
	ReplicaPlacement *storage.ReplicaPlacement
	DataCenter       string
	Rack             string
	DataNode         string
}

type VolumeGrowth struct {
	accessLock sync.Mutex
}

func NewDefaultVolumeGrowth() *VolumeGrowth {
	return &VolumeGrowth{}
}

// one replication type may need rp.GetCopyCount() actual volumes
// given copyCount, how many logical volumes to create
func (vg *VolumeGrowth) findVolumeCount(copyCount int) (count int) {
	switch copyCount {
	case 1:
		count = 7
	case 2:
		count = 6
	case 3:
		count = 3
	default:
		count = 1
	}
	return
}

func (vg *VolumeGrowth) AutomaticGrowByType(option *VolumeGrowOption, topo *Topology) (count int, err error) {
	count, err = vg.GrowByCountAndType(vg.findVolumeCount(option.ReplicaPlacement.GetCopyCount()), option, topo)
	if count > 0 && count%option.ReplicaPlacement.GetCopyCount() == 0 {
		return count, nil
	}
	return count, err
}
func (vg *VolumeGrowth) GrowByCountAndType(targetCount int, option *VolumeGrowOption, topo *Topology) (counter int, err error) {
	vg.accessLock.Lock()
	defer vg.accessLock.Unlock()

	for i := 0; i < targetCount; i++ {
		if c, e := vg.findAndGrow(topo, option); e == nil {
			counter += c
		} else {
			return counter, e
		}
	}
	return
}

func (vg *VolumeGrowth) findAndGrow(topo *Topology, option *VolumeGrowOption) (int, error) {
	servers, e := vg.findEmptySlotsForOneVolume(topo, option)
	if e != nil {
		return 0, e
	}
	vid := topo.NextVolumeId()
	err := vg.grow(topo, vid, option, servers...)
	return len(servers), err
}

func (vg *VolumeGrowth) findEmptySlotsForOneVolume(topo *Topology, option *VolumeGrowOption) (servers []*DataNode, err error) {
	//find main datacenter and other data centers
	rp := option.ReplicaPlacement
	mainDataCenter, otherDataCenters, dc_err := topo.RandomlyPickNodes(rp.DiffDataCenterCount+1, func(node Node) error {
		if option.DataCenter != "" && node.IsDataCenter() && node.Id() != NodeId(option.DataCenter) {
			return fmt.Errorf("Not matching preferred data center:%s", option.DataCenter)
		}
		if node.FreeSpace() < rp.DiffRackCount+rp.SameRackCount+1 {
			return fmt.Errorf("Free:%d < Expected:%d", node.FreeSpace(), rp.DiffRackCount+rp.SameRackCount+1)
		}
		return nil
	})
	if dc_err != nil {
		return nil, dc_err
	}

	//find main rack and other racks
	mainRack, otherRacks, rack_err := mainDataCenter.(*DataCenter).RandomlyPickNodes(rp.DiffRackCount+1, func(node Node) error {
		if option.Rack != "" && node.IsRack() && node.Id() != NodeId(option.Rack) {
			return fmt.Errorf("Not matching preferred rack:%s", option.Rack)
		}
		if node.FreeSpace() < rp.SameRackCount+1 {
			return fmt.Errorf("Free:%d < Expected:%d", node.FreeSpace(), rp.SameRackCount+1)
		}
		return nil
	})
	if rack_err != nil {
		return nil, rack_err
	}

	//find main rack and other racks
	mainServer, otherServers, server_err := mainRack.(*Rack).RandomlyPickNodes(rp.SameRackCount+1, func(node Node) error {
		if option.DataNode != "" && node.IsDataNode() && node.Id() != NodeId(option.DataNode) {
			return fmt.Errorf("Not matching preferred data node:%s", option.DataNode)
		}
		if node.FreeSpace() < 1 {
			return fmt.Errorf("Free:%d < Expected:%d", node.FreeSpace(), 1)
		}
		return nil
	})
	if server_err != nil {
		return nil, server_err
	}

	servers = append(servers, mainServer.(*DataNode))
	for _, server := range otherServers {
		servers = append(servers, server.(*DataNode))
	}
	for _, rack := range otherRacks {
		r := rand.Intn(rack.FreeSpace())
		if server, e := rack.ReserveOneVolume(r); e == nil {
			servers = append(servers, server)
		} else {
			return servers, e
		}
	}
	for _, datacenter := range otherDataCenters {
		r := rand.Intn(datacenter.FreeSpace())
		if server, e := datacenter.ReserveOneVolume(r); e == nil {
			servers = append(servers, server)
		} else {
			return servers, e
		}
	}
	return
}

func (vg *VolumeGrowth) grow(topo *Topology, vid storage.VolumeId, option *VolumeGrowOption, servers ...*DataNode) error {
	for _, server := range servers {
		if err := AllocateVolume(server, vid, option.Collection, option.ReplicaPlacement); err == nil {
			vi := storage.VolumeInfo{Id: vid, Size: 0, Collection: option.Collection, ReplicaPlacement: option.ReplicaPlacement, Version: storage.CurrentVersion}
			server.AddOrUpdateVolume(vi)
			topo.RegisterVolumeLayout(vi, server)
			glog.V(0).Infoln("Created Volume", vid, "on", server)
		} else {
			glog.V(0).Infoln("Failed to assign", vid, "to", servers, "error", err)
			return fmt.Errorf("Failed to assign %s: %s", vid.String(), err.Error())
		}
	}
	return nil
}