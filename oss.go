package main

import (
	"io"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	defaultAlioss *alioss
)

type alioss struct {
	cfg    OssConfig
	client *oss.Client
	bucket *oss.Bucket
	sync.Mutex
}

func newOssHandler(cfg OssConfig) *alioss {
	defaultAlioss = &alioss{
		cfg: cfg,
	}
	return defaultAlioss
}

func (o *alioss) init() error {
	o.Lock()
	defer o.Unlock()

	// double check already init client
	if o.client != nil {
		return nil
	}

	var err error
	o.client, err = oss.New(config.Oss.Endpoint, config.Oss.AccessKeyID, config.Oss.AccessKeySecret)
	if err != nil {
		return err
	}
	o.bucket, err = o.client.Bucket(config.Oss.BucketName)
	if err != nil {
		return err
	}

	// dial timeout
	dialTimeout := time.Duration(5 * time.Second)
	o.client.Config.HTTPTimeout.ConnectTimeout = dialTimeout
	return nil
}

func (o *alioss) putObject(filename string, body io.Reader) error {
	return o.bucket.PutObject(filename, body)
}

func (o *alioss) getObject(key string) (io.ReadCloser, error) {
	return o.bucket.GetObject(key)
}
