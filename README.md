# fsrv
## simple file server

## 简单的文件上传，下载
## 特点是简单，简单，简单
## 缺点一大堆

## 构建：
### go build fsrv.go

## 启动
### ./fsrv
### 或者
### ./fsrv -port 8080 -del true -store store
### 或者
### nohup ./fsrv -port 8080 -del true -store ./store > nohup.out 2>&1 &
