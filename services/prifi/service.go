package prifi

/*
* This is the internal part of the API. As probably the prifi-service will
* not have an external API, this will not have any API-functions.
 */

import (
	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/sda"
)

// ServiceName is the name to refer to the Template service from another
// package.
const ServiceName = "Prifi"

var serviceID sda.ServiceID

func init() {
	sda.RegisterNewService(ServiceName, newService)
	serviceID = sda.ServiceFactory.ServiceID(ServiceName)
}

// Service is our template-service
type Service struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
	*sda.ServiceProcessor
	path string
}

// StartTrustee has to take a configuration and start the necessary
// protocols to enable the trustee-mode.
func (s *Service) StartTrustee() error {
	log.Info("Service", s, "running in trustee mode")
	// Set up the configuration
	return nil
}

// StartRelay has to take a configuration and start the necessary
// protocols to enable the relay-mode.
func (s *Service) StartRelay() error {
	log.Info("Service", s, "running in relay mode")
	// Set up the configuration
	return nil
}

// StartClient has to take a configuration and start the necessary
// protocols to enable the client-mode.
func (s *Service) StartClient() error {
	log.Info("Service", s, "running in client mode")
	// Set up the configuration
	return nil
}

// NewProtocol is called on all nodes of a Tree (except the root, since it is
// the one starting the protocol) so it's the Service that will be called to
// generate the PI on all others node.
// If you use CreateProtocolSDA, this will not be called, as the SDA will
// instantiate the protocol on its own. If you need more control at the
// instantiation of the protocol, use CreateProtocolService, and you can
// give some extra-configuration to your protocol in here.
func (s *Service) NewProtocol(tn *sda.TreeNodeInstance, conf *sda.GenericConfig) (sda.ProtocolInstance, error) {
	log.Lvl3("Not templated yet")
	return nil, nil
}

// newTemplate receives the context and a path where it can write its
// configuration, if desired. As we don't know when the service will exit,
// we need to save the configuration on our own from time to time.
func newService(c *sda.Context, path string) sda.Service {
	s := &Service{
		ServiceProcessor: sda.NewServiceProcessor(c),
		path:             path,
	}
	return s
}
