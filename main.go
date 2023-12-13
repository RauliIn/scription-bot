package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"scription-bot/config"
	"strings"
	"sync"
	"time"
)

var configFile = flag.String("f", "etc/deploy.yaml", "the config file")

func main() {
	//读取配置文件
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.Infof("===============read config success===============")

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var ok bool
	var realData string
	txData := &TxData{}

	//key fmt.Sprintf("%s:%s", inputData, transaction.To()) , 用来识别(xx)地址是否mint过(xx)tick
	addrMintTickHashMap := make(map[string]struct{}, 2<<10)

	//key：inputData value:地址数量, 用来统计inputData mint过的地址数量
	mintedAddrCountHashMap := make(map[string]int, 2<<4)

	//key：inputData value:总计minted数量, 用来统计总计minted数量
	mintedTotalCountHashMap := make(map[string]int, 2<<4)
	//key：inputData 用来识别已经minted过的inputData
	isMintedInputDataHashMap := make(map[string]struct{}, 2<<4)
	inputDataChannel := make(chan string, 1)
	ctx := context.Background()

	client, err := ethclient.Dial(c.EthRpcConf.Url)
	if err != nil {
		logx.Errorf(" ethclient.Dial:%s", err.Error())
		return
	}
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		logx.Errorf("client.BlockNumber:%s", err.Error())
		return
	}
	logx.Infof("blockNumber:%d", blockNumber)

	//接收mint数据,执行tx
	go func() {
		for data := range inputDataChannel {

			if _, ok = isMintedInputDataHashMap[data]; !ok {
				logx.Infof("===============minting===============")
				logx.Infof(data)

				//设置为已经minted
				isMintedInputDataHashMap[data] = struct{}{}
				for i := 0; i < c.MintedLimit.MintCount; i++ {
					sendTx(client, []byte(data), c.PriKey)
					time.Sleep(3 * time.Second)
				}

				logx.Infof("===============minted===============")
			}

		}
	}()

	//定时执行
	ticker := time.NewTicker(time.Duration(c.EthRpcConf.IntervalTime) * time.Second)
	defer ticker.Stop()
	defer wg.Done()
	defer ants.Release()

	for range ticker.C {
		logx.Infof("===============scan blockNumber:%d===============", blockNumber)
		//获取区块
		block, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNumber-c.EthRpcConf.PrefixNumber)))

		if err != nil {
			logx.Errorf("client.BlockByNumber:%s", err.Error())
			return
		}

		//获取区块交易
		transactions := block.Transactions()
		if transactions.Len() < 1 {
			continue
		}

		//遍历交易
		for _, transaction := range transactions {
			wg.Add(1)

			ants.Submit(func() {

				if transaction.Value().Int64() != 0 {
					return
				}

				//如果已经minted，直接放弃
				inputData := string(transaction.Data())

				//判断是否为有效数据,这里不做具体json解析,因为100以上矿工mint的数据格式应该没问题
				if realData, ok = strings.CutPrefix(inputData, "data:,"); !ok {
					return
				}

				err = json.Unmarshal([]byte(realData), &txData)
				if txData.Op != "mint" {
					return
				}

				if err != nil {
					logx.Errorf("json.Unmarshal:%s", inputData)
					return
				}

				//构建key,这里key不能用tick代替,p有可能不同,op也有可能有问题
				addrMintTickHashMapKey := fmt.Sprintf("%s:%s", inputData, transaction.To())

				mutex.Lock()
				//如果已经minted过的
				if _, ok = isMintedInputDataHashMap[inputData]; ok {
					logx.Infof("addrMintTickHashMap.len:%d", len(addrMintTickHashMap))
					//惰性删除minted的key
					delete(addrMintTickHashMap, addrMintTickHashMapKey)
					return
				}

				mintedTotalCountHashMap[inputData] += 1
				// 识别(xx)地址是否mint过(xx)tick 如果是新mint账号加入统计
				if _, ok = addrMintTickHashMap[addrMintTickHashMapKey]; !ok {

					addrMintTickHashMap[addrMintTickHashMapKey] = struct{}{}
					mintedAddrCountHashMap[inputData] += 1

				}
				//minted限制,用来过滤非热门铭文
				if mintedAddrCountHashMap[inputData] >= c.MintedLimit.AddrCount && mintedTotalCountHashMap[inputData] >= c.MintedLimit.TotalCount {
					//logx.Infof("inputData:%s,addrMintedCount:%d,totalMintedCount:%d", txData.Tick, mintedAddrCountHashMap[inputData], mintedTotalCountHashMap[inputData])

					inputDataChannel <- inputData

				}
				mutex.Unlock()

				wg.Done()
			})

		}

		blockNumber++

	}

	wg.Wait()

}

func sendTx(client *ethclient.Client, inputData []byte, priKey string) {
	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		logx.Errorf("crypto.HexToECDSA:%s", err.Error())
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logx.Errorf("ecdsa.PublicKey:%s", err.Error())

	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := fromAddress
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logx.Errorf("crypto.PubkeyToAddress:%s", err.Error())
	}
	value := big.NewInt(0)
	gasLimit := uint64(210000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("client.SuggestGasPrice:%s", err.Error())
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, inputData)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logx.Errorf(" client.NetworkID:%s", err.Error())
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		logx.Errorf(" client.NetworkID:%s", err.Error())
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logx.Errorf(" client.NetworkID:%s", err.Error())
	}
	logx.Infof("tx hash: %s,data:%s", signedTx.Hash().Hex(), inputData)
}

// TxData 截取 data:, 后的结构体
type TxData struct {
	P    string `json:"p"`
	Op   string `json:"op"`
	Tick string `json:"tick"`
	Amt  string `json:"amt"`
}
