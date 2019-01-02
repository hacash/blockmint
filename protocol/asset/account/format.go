package account


import (
	"github.com/tribetrust/blockmint/sys/err"
)


type Address struct {
	address string
}



// check format
func (a Address) check() error {

	length := len(a.address)
	if ( length < 26 || length> 34 ) {
		return err.New("Address length must be between 26 and 34")
	}

	firstchar := a.address[0:1]
	if ( firstchar != "1" ) {
		return err.New("Address must start with char \"1\"")
	}


	return nil
}

