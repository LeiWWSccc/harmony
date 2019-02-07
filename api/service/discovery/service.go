package discovery

import (
	"github.com/ethereum/go-ethereum/log"
	proto_discovery "github.com/harmony-one/harmony/api/proto/discovery"
	"github.com/harmony-one/harmony/p2p"
	"github.com/harmony-one/harmony/p2p/host"
)

// Constants for discovery service.
const (
	numIncoming = 128
	numOutgoing = 16
)

// Service is the struct for discovery service.
type Service struct {
	Host        p2p.Host
	Rendezvous  string
	peerChan    chan p2p.Peer
	stakingChan chan p2p.Peer
	stopChan    chan struct{}
}

// New returns discovery service.
// h is the p2p host
// r is the rendezvous string, we use shardID to start (TODO: leo, build two overlays of network)
func New(h p2p.Host, r string, peerChan chan p2p.Peer, stakingChan chan p2p.Peer) *Service {
	return &Service{
		Host:        h,
		Rendezvous:  r,
		peerChan:    peerChan,
		stakingChan: stakingChan,
		stopChan:    make(chan struct{}),
	}
}

// StartService starts discovery service.
func (s *Service) StartService() {
	log.Info("Starting discovery service.")
	s.Init()
	s.Run()
}

// StopService shutdowns discovery service.
func (s *Service) StopService() {
	log.Info("Shutting down discovery service.")
	s.stopChan <- struct{}{}
	log.Info("discovery service stopped.")
}

// Run is the main function of the service
func (s *Service) Run() {
	go s.contactP2pPeers()
}

func (s *Service) contactP2pPeers() {
	for {
		select {
		case peer, ok := <-s.peerChan:
			if !ok {
				log.Debug("end of info", "peer", peer.PeerID)
				return
			}
			log.Debug("[DISCOVERY]", "peer", peer)
			s.Host.AddPeer(&peer)
			// TODO: stop ping if pinged before
			// TODO: call staking servcie here if it is a new node
			s.pingPeer(peer)
		case <-s.stopChan:
			return
		}
	}
}

// Init is to initialize for discoveryService.
func (s *Service) Init() {
	log.Info("Init discovery service")
}

func (s *Service) pingPeer(peer p2p.Peer) {
	ping := proto_discovery.NewPingMessage(s.Host.GetSelfPeer())
	buffer := ping.ConstructPingMessage()
	log.Debug("Sending Ping Message to", "peer", peer)
	content := host.ConstructP2pMessage(byte(0), buffer)
	s.Host.SendMessage(peer, content)
	log.Debug("Sent Ping Message to", "peer", peer)
	s.stakingChan <- peer
}