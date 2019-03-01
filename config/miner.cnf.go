package config

var (

	// dir
	//DirBase               = "~/.hacash/"
	DirBase               = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/database"
	DirDataBlock          = "blocks/"
	DirDataChainState     = "chainstate/"
	DirDataMinerState     = "minerstate/"
	DirDataTemporaryState = "tempstate/" // 临时状态，最好是内存文件系统
	DirDataNodes          = "nodes/"

	// mining
	MinimalFeePurity = int64(10000 * 10000 * 100) // 0.000001枚 // 最低手续费比值 每byte收取多少烁代币 （枚铢烁埃渺）
	// 矿工手续费奖励地址，越多越好
	MinerRewardAddress = []string{
		"129877upotaKLajSRhmBwczeehmjXCPndG",
		"127717zvZWFjEghjEpyyRSnitEEbnMuuLn",
		"148633HaMRNMoBfZeHnZZkNKWqCFFYBhcQ",
		"145725fYGyXNcGRtFkzvNYdCMZmzSQvQsr",
		/*"154961JoyHLNusTARBcBPyYptJGKUharMR",
		"135361CpCMxbfLEEPdVrmudJKQnQpPKZJv",
		"149156xrQZfWXvYUfwMePTAKufxscyhKSo",
		"183515ApbcLzCWuTmKitodcvnCGHNusXJK",
		"176967SYYSPMyZdFcXabCoPmSGSUCeqrgK",
		"147151rAKfdKjwGyAitEkYpoWXjwARKppR",
		"142159YdbweQdhqnBdXpqbyUDWEbGkhKBq",
		"164831dMnhpdcUmwjnJsANrzPKUjucLWXH",
		"16359495TTMnsrsGapMQNAomkzeCJbVFoC",
		"1398746UFsxhwwzzWMiWAirdhqmRAWFLzH",
		"115425MdcHYuJRUSEnrThaRMQgvmGaSWnt",
		"164373iUBHformFbtMLdebhCrRbsXGQDBe",
		"123895JqdBxFdDJPAPxHutRZoiyAURFSAv",
		"194917kEkMgPfjBpSxuXQyXSsGDKzxtMrN",
		"129957RnnUGUSkbRQLtutGBfYSpRcTDCus",
		"124347qcRPrJJtosaQYdHWCUbKfrtfUjTJ",
		"179772ppQTHSLHxAvAKABGCzTfsaoSGdYS",
		"113372WDaowsRdcsYGNMPKZEgrrjknhiBE",
		"156919pzFQgCVgeZQcmWLthngCpdyYUskX",
		"195569fTbHrmYpPZLmWANWjFwLqwybHTTF",*/
	}

	///////////////////////////////////////////////

	// consensus rule, prohibit change !
	MaximumBlockSize = int64(1024 * 1024 * 2) // 区块最大尺寸值byte  =  2MB

)
