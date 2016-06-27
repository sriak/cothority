package medco

import (
	"errors"
	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/network"
	"github.com/dedis/cothority/sda"
	. "github.com/dedis/cothority/services/medco/libmedco"
	"github.com/dedis/crypto/abstract"
)

const KEY_SWITCHING_PROTOCOL_NAME = "KeySwitching"

type KeySwitchedCipherMessage struct {
	Data                  map[TempID]CipherVector
	NewKey                abstract.Point
	OriginalEphemeralKeys map[TempID][]abstract.Point
	Proof                 map[TempID][]CompleteProof
}

type KeySwitchedCipherStruct struct {
	*sda.TreeNode
	KeySwitchedCipherMessage
}

func init() {
	network.RegisterMessageType(KeySwitchedCipherMessage{})
	sda.ProtocolRegisterName(KEY_SWITCHING_PROTOCOL_NAME, NewKeySwitchingProtocol)
}

type KeySwitchingProtocol struct {
	*sda.TreeNodeInstance

	// Protocol feedback channel
	FeedbackChannel chan map[TempID]CipherVector

	// Protocol communication channels
	PreviousNodeInPathChannel chan KeySwitchedCipherStruct

	// Protocol state data
	nextNodeInCircuit *sda.TreeNode
	TargetOfSwitch    *map[TempID]CipherVector
	TargetPublicKey   *abstract.Point
	originalEphemKeys map[TempID][]abstract.Point
}

func NewKeySwitchingProtocol(n *sda.TreeNodeInstance) (sda.ProtocolInstance, error) {
	keySwitchingProtocol := &KeySwitchingProtocol{
		TreeNodeInstance: n,
		FeedbackChannel:  make(chan map[TempID]CipherVector),
	}

	if err := keySwitchingProtocol.RegisterChannel(&keySwitchingProtocol.PreviousNodeInPathChannel); err != nil {
		return nil, errors.New("couldn't register data reference channel: " + err.Error())
	}

	var i int
	var node *sda.TreeNode
	var nodeList = n.Tree().List()
	for i, node = range nodeList {
		if n.TreeNode().Equal(node) {
			keySwitchingProtocol.nextNodeInCircuit = nodeList[(i+1)%len(nodeList)]
			break
		}
	}

	return keySwitchingProtocol, nil
}

// Starts the protocol
func (p *KeySwitchingProtocol) Start() error {

	if p.TargetOfSwitch == nil {
		return errors.New("No ciphertext given as key switching target.")
	}

	if p.TargetPublicKey == nil {
		return errors.New("No new public key to be switched on provided.")
	}

	log.Lvl1(p.ServerIdentity(), "started a Key Switching Protocol")

	initialMap := make(map[TempID]CipherVector, len(*p.TargetOfSwitch))
	p.originalEphemKeys = make(map[TempID][]abstract.Point, len(*p.TargetOfSwitch))
	for k,_ := range *p.TargetOfSwitch {
		initialCipherVector := *NewCipherVector(len((*p.TargetOfSwitch)[k]))
		p.originalEphemKeys[k] = make([]abstract.Point, len((*p.TargetOfSwitch)[k]))
		for i, c := range (*p.TargetOfSwitch)[k] {
			initialCipherVector[i].C = c.C
			p.originalEphemKeys[k][i] = c.K
		}
		initialMap[k] = initialCipherVector
	}

	p.sendToNext(&KeySwitchedCipherMessage{
		initialMap,
		*p.TargetPublicKey,
		p.originalEphemKeys,
		map[TempID][]CompleteProof{}})

	return nil
}

// Dispatch is an infinite loop to handle messages from channels
func (p *KeySwitchingProtocol) Dispatch() error {

	keySwitchingTarget := <-p.PreviousNodeInPathChannel

	for k, v := range keySwitchingTarget.Data {
		origEphemKeys := keySwitchingTarget.OriginalEphemeralKeys[k]
		v.KeySwitching(&v, &origEphemKeys, keySwitchingTarget.NewKey, p.Private())
		keySwitchingTarget.Data[k] = v
	}
	
	if p.IsRoot() {
		log.Lvl1(p.ServerIdentity(), "completed key switching.")
		p.FeedbackChannel <- keySwitchingTarget.Data
	} else {
		log.Lvl1(p.ServerIdentity(), "carried on key switching.")
		p.sendToNext(&keySwitchingTarget.KeySwitchedCipherMessage)
	}

	return nil
}

// Sends the message msg to the next node in the circuit based on the next TreeNode in Tree.List() If not visited yet.
// If the message already visited the next node, doesn't send and returns false. Otherwise, return true.
func (p *KeySwitchingProtocol) sendToNext(msg interface{}) {
	err := p.SendTo(p.nextNodeInCircuit, msg)
	if err != nil {
		log.Lvl1("Had an error sending a message: ", err)
	}
}
