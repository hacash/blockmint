package config

var (

	// consensus rule, prohibit change !
	MaximumBlockSize            = int64(1024 * 1024 * 2) // 区块最大尺寸值byte  =  2MB
	ChangeDifficultyBlockNumber = uint64(10)
	EachBlockTakesTime          = uint64(10) // 秒

)
