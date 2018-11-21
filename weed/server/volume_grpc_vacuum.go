package weed_server

import (
	"context"

	"github.com/draleyva/seaweedfs/weed/glog"
	"github.com/draleyva/seaweedfs/weed/pb/volume_server_pb"
	"github.com/draleyva/seaweedfs/weed/storage"
)

func (vs *VolumeServer) VacuumVolumeCheck(ctx context.Context, req *volume_server_pb.VacuumVolumeCheckRequest) (*volume_server_pb.VacuumVolumeCheckResponse, error) {

	resp := &volume_server_pb.VacuumVolumeCheckResponse{}

	garbageRatio, err := vs.store.CheckCompactVolume(storage.VolumeId(req.VolumdId))

	resp.GarbageRatio = garbageRatio

	if err != nil {
		glog.V(3).Infof("check volume %d: %v", req.VolumdId, err)
	}

	return resp, err

}

func (vs *VolumeServer) VacuumVolumeCompact(ctx context.Context, req *volume_server_pb.VacuumVolumeCompactRequest) (*volume_server_pb.VacuumVolumeCompactResponse, error) {

	resp := &volume_server_pb.VacuumVolumeCompactResponse{}

	err := vs.store.CompactVolume(storage.VolumeId(req.VolumdId), req.Preallocate)

	if err != nil {
		glog.Errorf("compact volume %d: %v", req.VolumdId, err)
	} else {
		glog.V(1).Infof("compact volume %d", req.VolumdId)
	}

	return resp, err

}

func (vs *VolumeServer) VacuumVolumeCommit(ctx context.Context, req *volume_server_pb.VacuumVolumeCommitRequest) (*volume_server_pb.VacuumVolumeCommitResponse, error) {

	resp := &volume_server_pb.VacuumVolumeCommitResponse{}

	err := vs.store.CommitCompactVolume(storage.VolumeId(req.VolumdId))

	if err != nil {
		glog.Errorf("commit volume %d: %v", req.VolumdId, err)
	} else {
		glog.V(1).Infof("commit volume %d", req.VolumdId)
	}

	return resp, err

}

func (vs *VolumeServer) VacuumVolumeCleanup(ctx context.Context, req *volume_server_pb.VacuumVolumeCleanupRequest) (*volume_server_pb.VacuumVolumeCleanupResponse, error) {

	resp := &volume_server_pb.VacuumVolumeCleanupResponse{}

	err := vs.store.CommitCleanupVolume(storage.VolumeId(req.VolumdId))

	if err != nil {
		glog.Errorf("cleanup volume %d: %v", req.VolumdId, err)
	} else {
		glog.V(1).Infof("cleanup volume %d", req.VolumdId)
	}

	return resp, err

}
