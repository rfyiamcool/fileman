# fileman

一个简单的文件中转服务器 !!!

**场景:**

存储临时文件, 在终端通过curl、wget就可对文件的上传下载，另外fileman对接oss存储, 屏蔽了oss sdk的使用.

**support storage:**

- local storage
- aliyun oss storage

**todo:**

- s3
- other oss

### API

#### upload

post

```
/upload          (select default storage)
/upload/local    (select local host storage)
/upload/oss      (select aliyun oss)
```

#### download

get

```
/file/{filename}              (...)
/downlaod/oss/{filename}      (...)
/downlaod/local/{filename}    (...)
```

### Run

#### config

fileman.yaml

```
# 监听地址
listen_address: ":8080"

# 文件存放目录
upload_dir: "/tmp"

basic_auth:
  enable: false
  username: "test-user"
  password: "hello"

# 阿里云OSS
oss:
  enable: true
  public: true
  endpoint: ""
  access_key: ""
  access_secret: ""
  bucket_name: ""

s3:
  enable: true
  use_ssl: false
  endpoint: ""
  access_key: ""
  access_secret: ""
  bucket_name: ""
```

#### dev run

```
make run
```

### Usage

#### build && install

```
make build
make install
```

默认安装到

```
/usr/local/bin/
```

#### 手动启动

```
fileman -c /tmp/fileman.yaml
```

如果不指定配置文件, 配置文件的搜索顺序为

- {current path}
- /etc/
- /etc/file_transfer

### Example

**no basic auth**

upload

```
$> curl -F file=@tt.png http://localhost:8080/upload/local

filename: 643616221782092862
download: http://localhost:8080/download/local/643616221782092862
```

download

```
$> curl -o {filename} http://localhost:8080/download/local/{filename}
```

**basic auth**

upload/downlaod (curl/wget)

```
$> curl --user user:password -F file=@tt.png http://localhost:8080/upload/local

$> curl --user user:password http://localhost:8080/download/local/{filename}
or
$> wget --http-user=user --http-password=password http://localhost:8080/download/local/{filename}
```
