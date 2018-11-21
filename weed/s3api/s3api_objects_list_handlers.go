package s3api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/draleyva/seaweedfs/weed/filer2"
	"github.com/draleyva/seaweedfs/weed/glog"
	"github.com/draleyva/seaweedfs/weed/pb/filer_pb"
	"github.com/gorilla/mux"
)

const (
	maxObjectListSizeLimit = 1000 // Limit number of objects in a listObjectsResponse.
)

func (s3a *S3ApiServer) ListObjectsV2Handler(w http.ResponseWriter, r *http.Request) {

	// https://docs.aws.amazon.com/AmazonS3/latest/API/v2-RESTBucketGET.html

	// collect parameters
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	glog.V(4).Infof("read v2: %v", vars)

	originalPrefix, marker, startAfter, delimiter, _, maxKeys := getListObjectsV2Args(r.URL.Query())

	if maxKeys < 0 {
		writeErrorResponse(w, ErrInvalidMaxKeys, r.URL)
		return
	}
	if delimiter != "" && delimiter != "/" {
		writeErrorResponse(w, ErrNotImplemented, r.URL)
		return
	}

	if marker == "" {
		marker = startAfter
	}

	response, err := s3a.listFilerEntries(bucket, originalPrefix, maxKeys, marker)

	if err != nil {
		writeErrorResponse(w, ErrInternalError, r.URL)
		return
	}

	writeSuccessResponseXML(w, encodeResponse(response))
}

func (s3a *S3ApiServer) ListObjectsV1Handler(w http.ResponseWriter, r *http.Request) {

	// https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketGET.html

	// collect parameters
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	originalPrefix, marker, delimiter, maxKeys := getListObjectsV1Args(r.URL.Query())

	if maxKeys < 0 {
		writeErrorResponse(w, ErrInvalidMaxKeys, r.URL)
		return
	}
	if delimiter != "" && delimiter != "/" {
		writeErrorResponse(w, ErrNotImplemented, r.URL)
		return
	}

	response, err := s3a.listFilerEntries(bucket, originalPrefix, maxKeys, marker)

	if err != nil {
		writeErrorResponse(w, ErrInternalError, r.URL)
		return
	}

	writeSuccessResponseXML(w, encodeResponse(response))
}

func (s3a *S3ApiServer) listFilerEntries(bucket, originalPrefix string, maxKeys int, marker string) (response *s3.ListObjectsOutput, err error) {

	// convert full path prefix into directory name and prefix for entry name
	dir, prefix := filepath.Split(originalPrefix)

	// check filer
	err = s3a.withFilerClient(func(client filer_pb.SeaweedFilerClient) error {

		request := &filer_pb.ListEntriesRequest{
			Directory:          fmt.Sprintf("%s/%s/%s", s3a.option.BucketsPath, bucket, dir),
			Prefix:             prefix,
			Limit:              uint32(maxKeys + 1),
			StartFromFileName:  marker,
			InclusiveStartFrom: false,
		}

		resp, err := client.ListEntries(context.Background(), request)
		if err != nil {
			return fmt.Errorf("list buckets: %v", err)
		}

		var contents []*s3.Object
		var commonPrefixes []*s3.CommonPrefix
		var counter int
		var lastEntryName string
		var isTruncated bool
		for _, entry := range resp.Entries {
			counter++
			if counter > maxKeys {
				isTruncated = true
				break
			}
			lastEntryName = entry.Name
			if entry.IsDirectory {
				commonPrefixes = append(commonPrefixes, &s3.CommonPrefix{
					Prefix: aws.String(fmt.Sprintf("%s%s/", dir, entry.Name)),
				})
			} else {
				contents = append(contents, &s3.Object{
					Key:          aws.String(fmt.Sprintf("%s%s", dir, entry.Name)),
					LastModified: aws.Time(time.Unix(entry.Attributes.Mtime, 0)),
					ETag:         aws.String("\"" + filer2.ETag(entry.Chunks) + "\""),
					Size:         aws.Int64(int64(filer2.TotalSize(entry.Chunks))),
					Owner: &s3.Owner{
						ID:          aws.String("bcaf161ca5fb16fd081034f"),
						DisplayName: aws.String("webfile"),
					},
					StorageClass: aws.String("STANDARD"),
				})
			}
		}

		response = &s3.ListObjectsOutput{
			Name:           aws.String(bucket),
			Prefix:         aws.String(originalPrefix),
			Marker:         aws.String(marker),
			NextMarker:     aws.String(lastEntryName),
			MaxKeys:        aws.Int64(int64(maxKeys)),
			Delimiter:      aws.String("/"),
			IsTruncated:    aws.Bool(isTruncated),
			Contents:       contents,
			CommonPrefixes: commonPrefixes,
		}

		glog.V(4).Infof("read directory: %v, found: %v", request, counter)

		return nil
	})

	return
}

func getListObjectsV2Args(values url.Values) (prefix, token, startAfter, delimiter string, fetchOwner bool, maxkeys int) {
	prefix = values.Get("prefix")
	token = values.Get("continuation-token")
	startAfter = values.Get("start-after")
	delimiter = values.Get("delimiter")
	if values.Get("max-keys") != "" {
		maxkeys, _ = strconv.Atoi(values.Get("max-keys"))
	} else {
		maxkeys = maxObjectListSizeLimit
	}
	fetchOwner = values.Get("fetch-owner") == "true"
	return
}

func getListObjectsV1Args(values url.Values) (prefix, marker, delimiter string, maxkeys int) {
	prefix = values.Get("prefix")
	marker = values.Get("marker")
	delimiter = values.Get("delimiter")
	if values.Get("max-keys") != "" {
		maxkeys, _ = strconv.Atoi(values.Get("max-keys"))
	} else {
		maxkeys = maxObjectListSizeLimit
	}
	return
}
