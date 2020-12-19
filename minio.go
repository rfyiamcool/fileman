package main

import (
	"context"
	"io"
	"io/ioutil"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	defaultS3Client *s3cli
)

type s3cli struct {
	ctx    context.Context
	cfg    S3Config
	client *minio.Client
	sync.Mutex
}

func newS3cli(ctx context.Context, cfg S3Config) *s3cli {
	cli := s3cli{
		ctx: ctx,
		cfg: cfg,
	}
	return &cli
}

func (m *s3cli) connect() error {
	var err error
	m.client, err = minio.New(m.cfg.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			m.cfg.AccessKeyID, m.cfg.AccessKeySecret, "",
		),
		Secure: m.cfg.UseSSl,
	})
	if err != nil {
		return err
	}

	err = m.client.MakeBucket(m.ctx, m.cfg.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, err := m.client.BucketExists(m.ctx, m.cfg.BucketName)
		if err != nil && !exists {
			return err
		}
	}

	return nil
}

func (m *s3cli) putObject(fname string, fio io.Reader, size int64) error {
	_, err := m.client.PutObject(m.ctx, m.cfg.BucketName, fname, fio, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

func (m *s3cli) getObject(fname string) ([]byte, error) {
	reader, err := m.client.GetObject(m.ctx, m.cfg.BucketName, fname, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, err
}
