package p2p

var (
	globalInstanceP2PServer *P2PServer = nil
)

type P2PServer struct {
}

func GetGlobalInstanceP2PServer() *P2PServer {
	if globalInstanceP2PServer == nil {
		globalInstanceP2PServer = NewP2PService()
	}
	return globalInstanceP2PServer
}

func NewP2PService() *P2PServer {
	return &P2PServer{}
}

func (this *P2PServer) Start() error {

	return nil
}
