# fsrv
simple file server

## 介绍
简单的文件上传，下载</br>
特点是简单，简单，简单</br>
缺点一大堆</br>

## 构建：
go build fsrv.go

## 启动
./fsrv</br>
或者</br>
./fsrv -port 8080 -delable true -store ./store</br>
或者</br>
nohup ./fsrv -port 8080 -delable true -store ./store >> nohup.out 2>&1 &</br>
