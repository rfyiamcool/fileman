package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/beinan/fastid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	storageLocal     = "local"
	storageAliyunOss = "aliyun_oss"
	storageS3        = "s3"

	plain = "text/plain; charset=utf-8"
)

var (
	help       bool
	config     *Config
	configFile string

	sigch = make(chan os.Signal, 1)
)

func init() {
	flag.BoolVar(&help, "h", false, "help")
	flag.StringVar(&configFile, "c", "fileman.yaml", "-c {configFilePath}")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`fileman
Usage: file_transfer [-h] [-c configFilePath]

Options:
`)
	flag.PrintDefaults()
}

type server struct {
	cfg    *Config
	router *gin.Engine
	serv   *http.Server
}

func newServer(cfg *Config) *server {
	router := gin.Default()
	return &server{
		cfg:    cfg,
		router: router,
	}
}

func (s *server) handleUpload(c *gin.Context) {
	switch config.Base.DefaultStorage {
	case storageLocal:
		s.handleUploadLocal(c)
	case storageAliyunOss:
		s.handleUploadAliyunOss(c)
	case storageS3:
		// to do
	default:
		s.handleUploadLocal(c)
	}
}

func (s *server) handleGetFile(c *gin.Context) {
	switch config.Base.DefaultStorage {
	case storageLocal:
		s.handleGetLocal(c)
	case storageAliyunOss:
		s.handleGetAliyunOss(c)
	case storageS3:
		// to do
	default:
		s.handleUploadLocal(c)
	}
}

func (s *server) handleGetLocal(c *gin.Context) {
	fname := c.Param("file")
	if fname == "" {
		c.String(http.StatusBadRequest, "please input file name")
		return
	}

	file, err := os.Open(path.Join(config.Base.UploadDir, fname))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Data(http.StatusOK, plain, bs)
}

func (s *server) handleUploadLocal(c *gin.Context) {
	// only single file
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		logrus.Errorf("failed to upload local storage, err: %s", err.Error())
		return
	}

	logrus.Infof("upload local filename: %s, size: %v", file.Filename, file.Size)

	// Upload the file to specific dst.
	file.Filename = genid()
	dst := filepath.Join(config.Base.UploadDir, file.Filename)
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		logrus.Errorf("save uploaded file to local storage, err: %s", err.Error())
		return
	}

	dpath := s.spliceLocolPath(c, file.Filename)
	c.String(http.StatusOK, s.makeResponse(file.Filename, dpath))
}

func (s *server) makeResponse(fname, dpath string) string {
	tmpl := "filename: %s\ndownload: %s\n"
	return fmt.Sprintf(tmpl, fname, dpath)
}

func (s *server) spliceLocolPath(c *gin.Context, fname string) string {
	return fmt.Sprintf("http://%s/download/local/%s", c.Request.Host, fname)
}

func (s *server) handleGetAliyunOss(c *gin.Context) {
	file := c.Param("file")
	fio, err := defaultAlioss.getObject(file)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	content, err := ioutil.ReadAll(fio)
	c.Data(http.StatusOK, plain, content)
}

func (s *server) handleUploadAliyunOss(c *gin.Context) {
	if !config.Oss.Enable {
		c.String(http.StatusOK, "not enable oss config")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		logrus.Errorf("failed to upload local storage, err: %s", err.Error())
		return
	}

	file.Filename = genid()
	fio, err := file.Open()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		logrus.Errorf("failed to open file, err: %s", err.Error())
		return
	}

	err = defaultAlioss.putObject(file.Filename, fio)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		logrus.Errorf("failed to put object to aliyun oss, err: %s", err.Error())
		return
	}

	dpath := s.spliceAliossPath(c, file.Filename)
	c.String(http.StatusOK, s.makeResponse(file.Filename, dpath))
}

func (s *server) spliceAliossPath(c *gin.Context, fname string) string {
	if config.Oss.Public {
		return fmt.Sprintf("https://%s.%s/%s", config.Oss.BucketName, config.Oss.Endpoint, fname)
	}

	return fmt.Sprintf("http://%s/wget/oss/%s", c.Request.Host, fname)
}

func (s *server) registerWrapper() {
	if !config.BaseAuth.Enable {
		return
	}
	s.router.Use(gin.BasicAuth(gin.Accounts{
		config.BaseAuth.UserName: config.BaseAuth.Password,
	}))
}

func (s *server) configureRouter() {
	// upload
	s.router.POST("/upload", s.handleUpload)
	s.router.POST("/upload/local", s.handleUploadLocal)
	s.router.POST("/upload/oss", s.handleUploadAliyunOss)

	// download
	// s.router.StaticFS("/file/", http.Dir(config.Base.UploadDir))
	s.router.GET("/file/:file", s.handleGetFile)
	s.router.GET("/download/oss/:file", s.handleGetAliyunOss)
	s.router.GET("/download/local/:file", s.handleGetLocal)
}

func (s *server) run() {
	go s.start()
}

func (s *server) start() {
	s.serv = &http.Server{
		Addr:         config.Base.ListenAddress,
		Handler:      s.router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	err := s.serv.ListenAndServe()
	if err != nil {
		logrus.Fatalf("failed to listen http server, err: %s", err.Error())
	}
}

func (s *server) stop() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	s.serv.SetKeepAlivesEnabled(false)
	return s.serv.Shutdown(ctx)
}

// init
func initResource() {
	if config.Base.UploadDir == "" {
		os.MkdirAll(config.Base.UploadDir, os.ModePerm)
	}

	if config.Oss.Enable {
		osser := newOssHandler(config.Oss)
		err := osser.init()
		if err != nil {
			logrus.Fatalf("failed to init oss client, err: %s", err.Error())
		}
	}
}

func genid() string {
	id := fastid.CommonConfig.GenInt64ID()
	return fmt.Sprintf("%d", id)
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	// parse config
	config = parseConfig()
	config.validate()
	config.print()

	// init each resource
	initResource()

	// http router
	serv := newServer(config)
	serv.registerWrapper()
	serv.configureRouter()
	serv.run()

	// bind signal
	signal.Notify(sigch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// listen signal
	for sig := range sigch {
		logrus.Infof("recv signal: %s", sig.String())

		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			logrus.Info("http server will exit")
			close(sigch)

		case syscall.SIGHUP:
			// todo reload

		default:
			continue
		}
	}

	// grace shutdown
	err := serv.stop()
	if err != nil {
		logrus.Errorf("faield to shutdown server, err: %s", err.Error())
		os.Exit(1)
	}

	logrus.Info("http server exited")
	os.Exit(0)
}
