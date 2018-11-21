package s3api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/draleyva/seaweedfs/weed/glog"
	"github.com/draleyva/seaweedfs/weed/pb/filer_pb"
	"github.com/gorilla/mux"
)

var (
	OS_UID = uint32(os.Getuid())
	OS_GID = uint32(os.Getgid())
)

func (s3a *S3ApiServer) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {

	var response ListAllMyBucketsResponse

	entries, err := s3a.list(s3a.option.BucketsPath, "", "", false, 0)

	if err != nil {
		writeErrorResponse(w, ErrInternalError, r.URL)
		return
	}

	var buckets []ListAllMyBucketsEntry
	for _, entry := range entries {
		if entry.IsDirectory {
			buckets = append(buckets, ListAllMyBucketsEntry{
				Name:         entry.Name,
				CreationDate: time.Unix(entry.Attributes.Crtime, 0),
			})
		}
	}

	response = ListAllMyBucketsResponse{
		ListAllMyBucketsResponse: ListAllMyBucketsResult{
			Owner: CanonicalUser{
				ID:          "",
				DisplayName: "",
			},
			Buckets: ListAllMyBucketsList{
				Bucket: buckets,
			},
		},
	}

	writeSuccessResponseXML(w, encodeResponse(response))
}

func (s3a *S3ApiServer) PutBucketHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	// create the folder for bucket, but lazily create actual collection
	if err := s3a.mkdir(s3a.option.BucketsPath, bucket, nil); err != nil {
		writeErrorResponse(w, ErrInternalError, r.URL)
		return
	}

	writeSuccessResponseEmpty(w)
}

func (s3a *S3ApiServer) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	err := s3a.withFilerClient(func(client filer_pb.SeaweedFilerClient) error {

		ctx := context.Background()

		// delete collection
		deleteCollectionRequest := &filer_pb.DeleteCollectionRequest{
			Collection: bucket,
		}

		glog.V(1).Infof("delete collection: %v", deleteCollectionRequest)
		if _, err := client.DeleteCollection(ctx, deleteCollectionRequest); err != nil {
			return fmt.Errorf("delete collection %s: %v", bucket, err)
		}

		return nil
	})

	err = s3a.rm(s3a.option.BucketsPath, bucket, true, false, true)

	if err != nil {
		writeErrorResponse(w, ErrInternalError, r.URL)
		return
	}

	writeResponse(w, http.StatusNoContent, nil, mimeNone)
}

func (s3a *S3ApiServer) HeadBucketHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	err := s3a.withFilerClient(func(client filer_pb.SeaweedFilerClient) error {

		request := &filer_pb.LookupDirectoryEntryRequest{
			Directory: s3a.option.BucketsPath,
			Name:      bucket,
		}

		glog.V(1).Infof("lookup bucket: %v", request)
		if _, err := client.LookupDirectoryEntry(context.Background(), request); err != nil {
			return fmt.Errorf("lookup bucket %s/%s: %v", s3a.option.BucketsPath, bucket, err)
		}

		return nil
	})

	if err != nil {
		writeErrorResponse(w, ErrNoSuchBucket, r.URL)
		return
	}

	writeSuccessResponseEmpty(w)
}
