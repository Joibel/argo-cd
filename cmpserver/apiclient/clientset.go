package apiclient

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpc_util "github.com/argoproj/argo-cd/v2/util/grpc"
	"github.com/argoproj/argo-cd/v2/util/io"
)

const (
	// MaxGRPCMessageSize contains max grpc message size
	MaxGRPCMessageSize = 100 * 1024 * 1024
)

// Clientset represents config management plugin server api clients
type Clientset interface {
	NewConfigManagementPluginClient() (io.Closer, ConfigManagementPluginServiceClient, error)
}

type ClientType int

const (
	Sidecar ClientType = iota
	Service
)

func (ct *ClientType) addrType() string {
	switch *ct {
	case Sidecar:
		return "unix"
	case Service:
		return "tcp"
	default:
		return ""
	}
}

func (ct *ClientType) String() string {
	switch *ct {
	case Sidecar:
		return "sidecar"
	case Service:
		return "service"
	default:
		return "unknown"
	}
}

type clientSet struct {
	address    string
	clientType ClientType
}

func (c *clientSet) addrType() string {
	return c.clientType.addrType()
}

func (c *clientSet) NewConfigManagementPluginClient() (io.Closer, ConfigManagementPluginServiceClient, error) {
	conn, err := c.newConnection()
	if err != nil {
		return nil, nil, err
	}
	return conn, NewConfigManagementPluginServiceClient(conn), nil
}

func (c *clientSet) newConnection() (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(1000 * time.Millisecond)),
	}
	unaryInterceptors := []grpc.UnaryClientInterceptor{grpc_retry.UnaryClientInterceptor(retryOpts...)}
	dialOpts := []grpc.DialOption{
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxGRPCMessageSize), grpc.MaxCallSendMsgSize(MaxGRPCMessageSize)),
		grpc.WithUnaryInterceptor(grpc_util.OTELUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpc_util.OTELStreamClientInterceptor()),
	}

	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc_util.BlockingDial(context.Background(), c.addrType(), c.address, nil, dialOpts...)
	if err != nil {
		log.Errorf("Unable to connect to config management plugin with address %s (type %s)", c.address, c.clientType.String())
		return nil, err
	}
	return conn, nil
}

// NewConfigManagementPluginClientSet creates new instance of config management plugin server Clientset
func NewConfigManagementPluginClientSet(address string, clientType ClientType) Clientset {
	return &clientSet{address: address, clientType: clientType}
}
