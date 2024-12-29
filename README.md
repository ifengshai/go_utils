
# 常用扩展包

## 配置文件处理库
github.com/spf13/viper
## 命令行工具包
github.com/spf13/cobra
## .env文件读取变量
github.com/joho/godotenv
## carbon时间处理包
github.com/golang-module/carbon
## 定时任务
github.com/robfig/cron/v3
## json处理
github.com/tidwall/gjson
## 下载工具
github.com/cavaliercoder/grab
## 参数验证包
github.com/go-playground/validator
## 加密解密验证包
golang.org/x/crypto
## redis操作
github.com/redis/go-redis
## ES操作
github.com/elastic/go-elasticsearch
## 组装数据库sql语句
go get github.com/Masterminds/squirrel
## 爬虫 
github.com/chromedp/chromedp
## html内容解析
github.com/PuerkitoBio/goquery
## GRPC
# 安装protoc
## 下载protoc命令到gopath/bin目录中
https://github.com/protocolbuffers/protobuf/releases/

## windows编译protoc：
windows下安装make命令
https://www.mingw-w64.org/downloads/
https://sourceforge.net/projects/mingw/
安装编译工具cmake：
https://cmake.org/download/
安装编译库visualstudio：
https://learn.microsoft.com/zh-cn/visualstudio/releases/2022/release-notes
下载grpc（可能需要梯子）:
git clone https://github.com/grpc/grpc
下载依赖包（必须需要梯子）：
git submodule update --init
配置编译库：
cmake .. -G "Visual Studio 17 2022"
开始编译：
cmake --build . --config Release

~~~shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

#生成go语言代码
protoc --proto_path=./proto/ --go_out=./protobuf/go --go-grpc_out=./protobuf/go ./proto/*.proto

#生成php语言代码
protoc --proto_path=./proto/ --php_out=./protobuf/php --plugin=protoc-gen-grpc=/fengsha/grpc/build-cmake/grpc_php_plugin ./proto/*.proto

protoc --proto_path=. --proto_path=./third_party --proto_path=./third_party/validate/ --proto_path=./third_party/google/api/ --proto_path=./marketing/interface/proto/marketing/v1/ --php_out=./src --grpc_out=./src --plugin=protoc-gen-grpc=D:\workspace\gopath\bin\grpc_php_plugin.exe ./marketing/interface/proto/marketing/v1/marketing.proto

~~~