# scription-bot
evm系链自动筛选热度铭文及进行跟单跟单

# 执行命令
下载golang 安装包,本项目基于golang 1.20版本开发，建议不低于该版本
下载地址：https://go.dev/ 根据系统下相应的版本

go env -w GO111MODULE=on 
go env -w GOPROXY=https://goproxy.io,direct

go mod tidy 

go build -o eth-scription-bot //建议输出格式 xx-scription-bot xx替换成请求的链 比如 eth链为 eth-scription-bot avax链为 avax-scription-bot

# 备注
## 项目相关

该项目写的比较匆忙,没有做细节优化,也没有做直接多链跑
1. 时间紧 
2. 链太多了,单真正活跃的就那么两三个,如果真有跑多链的需求可以参照下面开启多个链脚本方法
3. 处于安全考虑如果写多链代码逻辑方面不提,还需要大部分链都得放gas去实测,本人懒狗见谅

### 项目待优化部分
1. 项目没做map内存释放的优化,如果跑着跑着崩了,建议重开,后续有时间会完善
2. 存储也没做优化,比如minted过的铭文存到本地,重启还会去mint
3. 没有做黑名单功能,比如在测试avax链的时候,发现还是有很多地址在打avas,这个明明已经minted完了,很明显打了也是会打费,没做的原因有两点1是做黑名单功能需要扫块加判断该铭文的上限等等
如果谁有做了相关功能的api或者sdk可以让我接入下可以,自己有点繁琐,一个字还是懒,目前可选方案是加大/etc/deploy.yaml AddrCount和TotalCount,因为这里参数足够大能过滤大部分,
因为毕竟没有多少人去一直mint已经过上限的铭文
4. gas费用是读取rpc接口的建议费用,打之前建议去链浏览器看下gas费再做决定
5. 同样的需要打之前确认下该铭文是否已经打完,代码没有做铭文上限的检索和判断


## golang相关
安装golang后,建议配置GOPROXY,默认地址可能拉取依赖很慢

go env -w GOPROXY=https://goproxy.io,direct
go env -w GO111MODULE=on 该选项新版本默认都是开启的,根据go env 查看是否为on,不会查看的话直接执行就好了不会有影响

## 开启多个链脚本

建议输出格式 xx-scription-bot xx替换成请求的链 比如 eth链为 eth-scription-bot avax链为 avax-scription-bot

因为当前版本只能执行一个链，如果需要执行多个链建议把当前项目复制几份改成对应的链项目,然后打包 
例如在avax链执行另一个,可以把当前项目复制一份叫avax-scription-bot项目,修改对应的配置文件 /etc/deploy.yaml为avax链的rpc配置 执行 go build -o avax-scription-bot
这样免得混乱
## /etc/deploy.yaml 修改

主要修改私钥和rpc地址,其他参数可以使用默认配置,修改的可以查看配置文件注释修改

## 题外话
1. 大佬的话可以打赏支持下,不强求奥,以后也会是开源
打赏地址 0x04001842338fe79743680d8F3749eA53d16a41D9

2. 另外在avax链mint了个avi的铭文,有兴趣的小伙伴可以mint下
https://avascriptions.com/token/detail?tick=avi

3. twitter地址： https://twitter.com/guq43432217 微信 18001177545
4. 欢迎有问题的小伙伴github上提Issue 也可以twitter,微信聊

最后满意的小伙伴不介意可以右上角点个 star