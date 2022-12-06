package service

const (
	protocol = "Aiko-Server"
	TLS      = "tls"
	TCP      = "tcp"
	WS       = "ws"
	GRPC     = "grpc"
)

// Service is the interface of all the services running in the panel
type Service interface {
	Start() error
	Close() error
}

type CertConfig struct {
	CertFile string
	KeyFile  string
}

type FallBackConfig struct {
	Host string
	Port int
}
