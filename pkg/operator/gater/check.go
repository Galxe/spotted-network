package gater

import (
	"context"
	"fmt"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
)

// isBlocked checks if the peer is in the active operators map
func (g *ConnectionGater) isBlocked(peerID peer.ID) (bool, error) {
	blocked, err := g.blacklistRepo.IsBlocked(context.Background(), peerID.String())
	if err != nil {
		log.Printf("[Gater] Error checking if peer %s is blocked: %v", peerID, err)
		return false, err
	}
	if blocked == nil {
		log.Printf("[Gater] Peer %s not found in blacklist", peerID)
		return false, nil
	}
	return *blocked, nil
}

// isActiveOperator checks if the peer is an active operator
func (g *ConnectionGater) isActiveOperator(peerID peer.ID) (bool, error) {
	p2pKey, err := utils.PeerIDToP2PKey(peerID)
	if err != nil {
		log.Printf("[Gater] Error converting peerID to p2p key for %s: %v", peerID, err)
		return false, fmt.Errorf("failed to convert peerID to p2p key: %w", err)
	}

	operator, err := g.operatorRepo.GetOperatorByP2PKey(context.Background(), p2pKey)
	if err != nil {
		log.Printf("[Gater] Error getting operator by p2p key %s: %v", p2pKey, err)
		return false, err
	}

	if operator == nil {
		log.Printf("[Gater] No operator found for p2p key: %s", p2pKey)
		return false, nil
	}

	log.Printf("[Gater] Found operator: address=%s, p2p_key=%s, is_active=%v, active_epoch=%d, exit_epoch=%d",
		operator.Address,
		operator.P2pKey,
		operator.IsActive,
		operator.ActiveEpoch,
		operator.ExitEpoch,
	)

	return operator.IsActive, nil
}

// checkPeerPermission checks if a peer is allowed to connect
func (g *ConnectionGater) checkPeerPermission(peerID peer.ID) bool {
	log.Printf("[Gater] Checking permission for peer %s", peerID)

	// check if the peer is blocked
	blocked, err := g.isBlocked(peerID)
	if err != nil {
		log.Printf("[Gater] Error checking if peer %s is blocked: %v", peerID, err)
		return false
	}
	if blocked {
		log.Printf("[Gater] Peer %s is blocked, denying connection", peerID)
		return false
	}

	// check if the peer is active operator
	isActive, err := g.isActiveOperator(peerID)
	if err != nil {
		log.Printf("[Gater] Error checking if peer %s is active operator: %v", peerID, err)
		return false
	}
	if !isActive {
		log.Printf("[Gater] Peer %s is not an active operator, denying connection", peerID)
		return false
	}

	log.Printf("[Gater] Peer %s is allowed to connect", peerID)
	return true
}
