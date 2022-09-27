package rawdb

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/everFinance/turing/store/schema"
	"strings"
	"time"
)

type S3DB struct {
	downloader   s3manager.Downloader
	uploader     s3manager.Uploader
	s3Api        s3iface.S3API
	bucketPrefix string
	bkts         []string
}

func NewS3DB(cfg schema.Config) (*S3DB, error) {
	mySession := session.Must(session.NewSession())
	cred := credentials.NewStaticCredentials(cfg.AccKey, cfg.SecretKey, "")
	cfgs := aws.NewConfig().WithRegion(cfg.Region).WithCredentials(cred)
	if cfg.Use4EVER {
		cfgs.WithEndpoint("https://endpoint.4everland.co") // inject 4everland endpoint
	}
	s3Api := s3.New(mySession, cfgs)
	bkts, err := createS3Bucket(s3Api, cfg.BktPrefix, cfg.Bkt)
	if err != nil {
		return nil, err
	}

	log.Info("run with s3 success")
	return &S3DB{
		downloader: s3manager.Downloader{
			S3: s3Api,
		},
		uploader: s3manager.Uploader{
			S3: s3Api,
		},
		s3Api:        s3Api,
		bucketPrefix: cfg.BktPrefix,
		bkts:         bkts,
	}, nil
}

func (s *S3DB) Put(bucket, key string, value []byte) (err error) {
	bkt := getS3Bucket(s.bucketPrefix, bucket)
	uploadInfo := &s3manager.UploadInput{
		Bucket: aws.String(bkt),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
	}
	var retry uint64
	_, err = s.uploader.Upload(uploadInfo)
	for err != nil && retry < 5 {
		time.Sleep(time.Duration(retry) * time.Second)
		_, err = s.uploader.Upload(uploadInfo)
	}
	return
}

func (s *S3DB) Get(bucket, key string) (data []byte, err error) {
	bkt := getS3Bucket(s.bucketPrefix, bucket)
	downloadInfo := &s3.GetObjectInput{
		Bucket: aws.String(bkt),
		Key:    aws.String(key),
	}
	buf := aws.NewWriteAtBuffer([]byte{})
	n, err := s.downloader.Download(buf, downloadInfo)
	if n == 0 {
		return nil, schema.ErrNotExist
	}
	data = buf.Bytes()
	return
}

func (s *S3DB) GetAllKey(bucket string) (keys []string, err error) {
	bkt := getS3Bucket(s.bucketPrefix, bucket)
	resp, err := s.s3Api.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bkt)})
	if err != nil {
		return
	}
	keys = make([]string, 0)
	for _, item := range resp.Contents {
		keys = append(keys, *item.Key)
	}
	return
}

func (s *S3DB) Delete(bucket, key string) (err error) {
	bkt := getS3Bucket(s.bucketPrefix, bucket)
	_, err = s.s3Api.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bkt), Key: aws.String(key)})
	return
}

func (s *S3DB) Close() (err error) {
	return
}

func (s *S3DB) Clear() (err error) {
	// delete all keys for each bucket
	for _, bkt := range s.bkts {
		keys, err := s.GetAllKey(bkt)
		if err != nil {
			return err
		}
		for _, key := range keys {
			err = s.Delete(bkt, key)
			if err != nil {
				return err
			}
		}
	}
	// delete bucket
	for _, bkt := range s.bkts {
		s3bkt := getS3Bucket(s.bucketPrefix, bkt)
		_, err = s.s3Api.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(s3bkt)})
		if err != nil {
			return
		}
	}
	return
}

func createS3Bucket(svc s3iface.S3API, prefix string, bucketNames []string) ([]string, error) {
	if len(bucketNames) == 0 {
		bucketNames = append(bucketNames, schema.AllBkt...)
	}
	for _, bucketName := range bucketNames {
		s3Bkt := getS3Bucket(prefix, bucketName) // s3 bucket name only accept lower case
		_, err := svc.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(s3Bkt)})
		if err != nil && !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			return nil, err
		}
	}
	return bucketNames, nil
}

func getS3Bucket(prefix, bktName string) string {
	return strings.ToLower(prefix + "-" + bktName)
}
