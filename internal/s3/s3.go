package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Config struct {
	AccessToken string 
	SecretKey   string
	Region      string
	Bucket      string
}

type Instance interface {
	UploadFile(context.Context, *s3manager.UploadInput) error
	DownloadFile(context.Context, io.WriterAt, *s3.GetObjectInput) error
	ListBuckets(context.Context) (*s3.ListBucketsOutput, error)
	CopyFile(context.Context, *s3.CopyObjectInput) error
	SetACL(context.Context, *s3.PutObjectAclInput) error
}

type s3Storage struct {
	session *session.Session
	downloader *s3manager.Downloader
	uploader *s3manager.Uploader

	s3 *s3.S3
	bucket string
}

func New(c Config) (Instance, error) {
	s, err := session.NewSession(&aws.Config{
		Credentials:    	credentials.NewStaticCredentials(c.AccessToken, c.SecretKey, ""),
		Region:				aws.String(c.Region),
		S3ForcePathStyle: 	aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	return &s3Storage{
		session: s,
		downloader: s3manager.NewDownloader(s),
		uploader: s3manager.NewUploader(s),
		s3: s3.New(s),
		bucket: c.Bucket,
	}, nil
}

func (s *s3Storage) UploadFile(ctx context.Context, input *s3manager.UploadInput) error {
	input.Bucket = &s.bucket
	
	_, err := s.uploader.UploadWithContext(ctx, input)
	return err
}

func (s *s3Storage) DownloadFile(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput) error {
	input.Bucket = &s.bucket
	
	_, err := s.downloader.DownloadWithContext(ctx, w, input)
	return err
}

func (s *s3Storage) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
	return s.s3.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
}

func (s *s3Storage) CopyFile(ctx context.Context, input *s3.CopyObjectInput) error {
	input.Bucket = &s.bucket
	
	_, err := s.s3.CopyObjectWithContext(ctx, input)
	return err
}

func (s *s3Storage) SetACL(ctx context.Context, input *s3.PutObjectAclInput) error {
	input.Bucket = &s.bucket
	
	_, err := s.s3.PutObjectAclWithContext(ctx, input)
	return err
}
