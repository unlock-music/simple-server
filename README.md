# Simple Server for UnlockMusic
- 一个简易的，单文件部署的本地Http服务器
- 基于Golang 1.16的embed功能
- 方便普通使用者运行和部署

##使用方法
### 从Release下载
- [Release](https://github.com/unlock-music/simple-server/releases)
### 手动构建
- 安装golang 1.16以上的版本(Tips 2020/11/07: 当前未发布，需要从源码构建)
```shell script
go generate
go build
```