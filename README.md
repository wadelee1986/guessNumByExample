# guessNumByExample


## 需要查资料事情
1、sdk invoke 后没有拿到chaincode返回值；目前, blockchain sdk需要通过query做出进入下一居的判断。

2、前端页面需要通过js控制动画。

3、写入账簿失败的提示。

## 游戏规则

1、玩家通过浏览器访问服务器，会得到一个随机名字，（可能会有重名）；直接进去游戏；默认balance是0.

2、所有玩家开始压自己的数字，玩家互相之间可以看到压的次数；

3、当所有玩家压的数字加起来到达一局的最大值时，（最大值写在了程序里，目前是100）；开始结算；

4、从所有玩家中找出压的最小的数字，并且不重复的数字，那么这个玩家将得到这一居玩家压的所有数字；
   如果玩家压的都是重复的数字，那么所有数字都返回给对应的玩家。


## 程序安装

fabric -->  fabric-sdk-go server -->h5

1、安装 fabric-sdk-go的环境；

```bash

#### download code
go get github.com/hyperledger/fabric-sdk-go
#### In the Fabric SDK Go directory
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/



2、运行fabric环境

```bash
  ##### into fixture directory
  cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/
  ##### run fabric env
  source latest-env.sh && sudo docker-compose up --force-recreate

3、安装fabric-sdk-go server

```bash
  # download  code
  go  get  github.com/wadelee1986/guessNumByExample
  # In the server directory
  cd $GOPATH/src/github.com/wadelee1986/guessNumByExample/src/server
  # run ...
  go run *.go

4、访问 http://127.0.0.1:8080/

5、在send前的输入框，输入猜测的数字， 小于1没有意义； 大于100有可能全输掉。
