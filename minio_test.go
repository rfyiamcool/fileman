package main

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestS3(t *testing.T) {
	cfg := S3Config{
		Endpoint:        "127.0.0.1:9000",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		AccessKeySecret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		BucketName:      "xiaorui",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mcli := newS3cli(ctx, cfg)
	err := mcli.connect()
	assert.Equal(t, err, nil)

	reader := strings.NewReader("haha")
	err = mcli.putObject("123123", reader, 4)
	assert.Equal(t, err, nil)

	bs, err := mcli.getObject("123123")
	assert.Equal(t, err, nil)
	t.Log(string(bs))
}
