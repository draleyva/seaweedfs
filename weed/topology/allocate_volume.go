package topology

import (
	"context"

	"github.com/draleyva/seaweedfs/weed/operation"
	"github.com/draleyva/seaweedfs/weed/pb/volume_server_pb"
	"github.com/draleyva/seaweedfs/weed/storage"
)

type AllocateVolumeResult struct {
	Error string
}

func AllocateVolume(dn *DataNode, vid storage.VolumeId, option *VolumeGrowOption) error {

	return operation.WithVolumeServerClient(dn.Url(), func(client volume_server_pb.VolumeServerClient) error {
		_, deleteErr := client.AssignVolume(context.Background(), &volume_server_pb.AssignVolumeRequest{
			VolumdId:    uint32(vid),
			Collection:  option.Collection,
			Replication: option.ReplicaPlacement.String(),
			Ttl:         option.Ttl.String(),
			Preallocate: option.Prealloacte,
		})
		return deleteErr
	})

}
