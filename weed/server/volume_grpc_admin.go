package weed_server

import (
	"context"

	"github.com/draleyva/seaweedfs/weed/glog"
	"github.com/draleyva/seaweedfs/weed/pb/volume_server_pb"
	"github.com/draleyva/seaweedfs/weed/storage"
)

func (vs *VolumeServer) DeleteCollection(ctx context.Context, req *volume_server_pb.DeleteCollectionRequest) (*volume_server_pb.DeleteCollectionResponse, error) {

	resp := &volume_server_pb.DeleteCollectionResponse{}

	err := vs.store.DeleteCollection(req.Collection)

	if err != nil {
		glog.Errorf("delete collection %s: %v", req.Collection, err)
	} else {
		glog.V(2).Infof("delete collection %v", req)
	}

	return resp, err

}

func (vs *VolumeServer) AssignVolume(ctx context.Context, req *volume_server_pb.AssignVolumeRequest) (*volume_server_pb.AssignVolumeResponse, error) {

	resp := &volume_server_pb.AssignVolumeResponse{}

	err := vs.store.AddVolume(
		storage.VolumeId(req.VolumdId),
		req.Collection,
		vs.needleMapKind,
		req.Replication,
		req.Ttl,
		req.Preallocate,
	)

	if err != nil {
		glog.Errorf("assign volume %v: %v", req, err)
	} else {
		glog.V(2).Infof("assign volume %v", req)
	}

	return resp, err

}

func (vs *VolumeServer) VolumeMount(ctx context.Context, req *volume_server_pb.VolumeMountRequest) (*volume_server_pb.VolumeMountResponse, error) {

	resp := &volume_server_pb.VolumeMountResponse{}

	err := vs.store.MountVolume(storage.VolumeId(req.VolumdId))

	if err != nil {
		glog.Errorf("volume mount %v: %v", req, err)
	} else {
		glog.V(2).Infof("volume mount %v", req)
	}

	return resp, err

}

func (vs *VolumeServer) VolumeUnmount(ctx context.Context, req *volume_server_pb.VolumeUnmountRequest) (*volume_server_pb.VolumeUnmountResponse, error) {

	resp := &volume_server_pb.VolumeUnmountResponse{}

	err := vs.store.UnmountVolume(storage.VolumeId(req.VolumdId))

	if err != nil {
		glog.Errorf("volume unmount %v: %v", req, err)
	} else {
		glog.V(2).Infof("volume unmount %v", req)
	}

	return resp, err

}
