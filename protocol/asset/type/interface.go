package types


type AssetFormatChecker interface {
	check(*interface{}) error
}