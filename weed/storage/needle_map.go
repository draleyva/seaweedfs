package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/draleyva/seaweedfs/weed/storage/needle"
	. "github.com/draleyva/seaweedfs/weed/storage/types"
	"github.com/draleyva/seaweedfs/weed/util"
)

type NeedleMapType int

const (
	NeedleMapInMemory NeedleMapType = iota
	NeedleMapLevelDb
	NeedleMapBoltDb
	NeedleMapBtree
)

type NeedleMapper interface {
	Put(key NeedleId, offset Offset, size uint32) error
	Get(key NeedleId) (element *needle.NeedleValue, ok bool)
	Delete(key NeedleId, offset Offset) error
	Close()
	Destroy() error
	ContentSize() uint64
	DeletedSize() uint64
	FileCount() int
	DeletedCount() int
	MaxFileKey() NeedleId
	IndexFileSize() uint64
	IndexFileContent() ([]byte, error)
	IndexFileName() string
}

type baseNeedleMapper struct {
	indexFile           *os.File
	indexFileAccessLock sync.Mutex

	mapMetric
}

func (nm *baseNeedleMapper) IndexFileSize() uint64 {
	stat, err := nm.indexFile.Stat()
	if err == nil {
		return uint64(stat.Size())
	}
	return 0
}

func (nm *baseNeedleMapper) IndexFileName() string {
	return nm.indexFile.Name()
}

func IdxFileEntry(bytes []byte) (key NeedleId, offset Offset, size uint32) {
	key = BytesToNeedleId(bytes[:NeedleIdSize])
	offset = BytesToOffset(bytes[NeedleIdSize : NeedleIdSize+OffsetSize])
	size = util.BytesToUint32(bytes[NeedleIdSize+OffsetSize : NeedleIdSize+OffsetSize+SizeSize])
	return
}
func (nm *baseNeedleMapper) appendToIndexFile(key NeedleId, offset Offset, size uint32) error {
	bytes := make([]byte, NeedleIdSize+OffsetSize+SizeSize)
	NeedleIdToBytes(bytes[0:NeedleIdSize], key)
	OffsetToBytes(bytes[NeedleIdSize:NeedleIdSize+OffsetSize], offset)
	util.Uint32toBytes(bytes[NeedleIdSize+OffsetSize:NeedleIdSize+OffsetSize+SizeSize], size)

	nm.indexFileAccessLock.Lock()
	defer nm.indexFileAccessLock.Unlock()
	if _, err := nm.indexFile.Seek(0, 2); err != nil {
		return fmt.Errorf("cannot seek end of indexfile %s: %v",
			nm.indexFile.Name(), err)
	}
	_, err := nm.indexFile.Write(bytes)
	return err
}
func (nm *baseNeedleMapper) IndexFileContent() ([]byte, error) {
	nm.indexFileAccessLock.Lock()
	defer nm.indexFileAccessLock.Unlock()
	return ioutil.ReadFile(nm.indexFile.Name())
}
