package sink

import (
	"github.com/draleyva/seaweedfs/weed/pb/filer_pb"
	"github.com/draleyva/seaweedfs/weed/replication/source"
	"github.com/draleyva/seaweedfs/weed/util"
)

type ReplicationSink interface {
	GetName() string
	Initialize(configuration util.Configuration) error
	DeleteEntry(key string, isDirectory, deleteIncludeChunks bool) error
	CreateEntry(key string, entry *filer_pb.Entry) error
	UpdateEntry(key string, oldEntry, newEntry *filer_pb.Entry, deleteIncludeChunks bool) (foundExistingEntry bool, err error)
	GetSinkToDirectory() string
	SetSourceFiler(s *source.FilerSource)
}

var (
	Sinks []ReplicationSink
)
