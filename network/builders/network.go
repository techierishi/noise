package builders

import (
	"errors"
	"reflect"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/perlin-network/noise/crypto"
	"github.com/perlin-network/noise/network"
)

// NetworkBuilder is a Address->processors struct
type NetworkBuilder struct {
	keys    *crypto.KeyPair
	address string
	upnpEnabled bool

	// map[string]MessageProcessor
	processors *network.StringMessageProcessorSyncMap
}

// SetKeys pair created from crypto.KeyPair
func (builder *NetworkBuilder) SetKeys(pair *crypto.KeyPair) {
	builder.keys = pair
}

func (builder *NetworkBuilder) SetAddress(address string) {
	builder.address = address
}

func (builder *NetworkBuilder) SetUpnpEnabled(enabled bool) {
	builder.upnpEnabled = enabled
}

// AddProcessor for a given message,
// Example: builder.AddProcessor((*protobuf.LookupNodeRequest)(nil), MessageProcessor{})
func (builder *NetworkBuilder) AddProcessor(message proto.Message, processor network.MessageProcessor) {
	// Initialize map if not exist.
	if builder.processors == nil {
		builder.processors = &network.StringMessageProcessorSyncMap{}
	}

	name := reflect.TypeOf(message).String()

	// Store pointers to message processor only.
	if value := reflect.ValueOf(message); value.Kind() == reflect.Ptr && value.Pointer() == 0 {
		builder.processors.Store(name, processor)
	} else {
		glog.Fatal("message must be nil")
	}
}

// BuildNetwork verifies all parameters of the network and returns either an error due to
// misconfiguration, or a noise.network.Network.
func (builder *NetworkBuilder) BuildNetwork() (*network.Network, error) {
	if builder.keys == nil {
		return nil, errors.New("cryptography keys not provided to Network; cannot create node ID")
	}

	if len(builder.address) == 0 {
		return nil, errors.New("Network requires public server IP for peers to connect to")
	}

	// Initialize map if not exist.
	if builder.processors == nil {
		builder.processors = &network.StringMessageProcessorSyncMap{}
	}

	unifiedAddr, err := network.ToUnifiedAddress(builder.address)
	if err != nil {
		return nil, err
	}

	net := &network.Network{
		Keys:    builder.keys,
		Address: unifiedAddr,
		UpnpEnabled: builder.upnpEnabled,

		Processors: builder.processors,

		Peers: &network.StringPeerClientSyncMap{},

		Listening: make(chan struct{}),
	}

	return net, nil
}
