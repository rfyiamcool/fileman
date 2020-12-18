package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Base     BaseConfig     `mapstructure:"base"`
	Oss      OssConfig      `mapstructure:"oss"`
	BaseAuth BaseAuthConfig `mapstructure:"basic_auth"`
}

type BaseConfig struct {
	ListenAddress  string `mapstructure:"listen_address"`
	UploadDir      string `mapstructure:"upload_dir"`
	URLPrefix      string `mapstructure:"url_prefix"`
	DefaultStorage string `mapstructure:"default_storage"`
}

type OssConfig struct {
	Enable          bool   `mapstructure:"enable"`
	Public          bool   `mapstructure:"public"`
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key"`
	AccessKeySecret string `mapstructure:"access_secret"`
	BucketName      string `mapstructure:"bucket_name"`
}

type BaseAuthConfig struct {
	Enable   bool   `mapstructure:"enable"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func parseConfig() {
	log.Println("config file: ", configFile)
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/etc/fileman")
		viper.SetConfigName("fileman")
	}

	// If a config file is found, read it in.
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("can't read config file")
		panic(err)
	}

	log.Println("read config from", viper.ConfigFileUsed())

	config = &Config{}
	err := viper.Unmarshal(config)
	if err != nil {
		log.Println("can't parse config file")
		panic(err)
	}

	log.Println("UploadDir", config.Base.UploadDir)
	log.Println("EnableBasicAuth", config.BaseAuth.Enable)
	log.Println("starting...")
}
