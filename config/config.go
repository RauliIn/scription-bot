package config

type Config struct {
	PriKey string //私钥
	//ethrpc请求配置
	EthRpcConf struct {
		Url          string // Url:请求rpc地址 修改来替换各链rpc地址
		IntervalTime int    //请求轮询时间
		PrefixNumber uint64 //查询当前区块之前多少块
	}
	//minted限制,用来过滤非热门铭文
	MintedLimit struct {
		AddrCount  int //minted该铭文的地址数 只有比当前数量大才去minted
		TotalCount int //minted该铭文的总交易数 只有比当前数量大才去minted
		MintCount  int //mint该铭文的张数
	}
}
