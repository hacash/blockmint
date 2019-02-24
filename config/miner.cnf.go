package config

var (

	// dir
	DirBase           = "~/.hacash/"
	DirDataBlock      = "blocks/"
	DirDataChainState = "chainstate/"
	DirDataMinerState = "minerstate/"

	// mining
	MinimalFees = int64(10000 * 10000 * 100) // 0.000001枚 // 最低手续费比值 每byte收取多少烁代币 （枚铢烁埃渺）

	///////////////////////////////////////////////

	// consensus rule, prohibit change !
	MaximumBlockSize = int64(1024 * 1024 * 2) // 区块最大尺寸值byte  =  2MB

)
