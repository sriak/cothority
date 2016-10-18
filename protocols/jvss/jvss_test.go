package jvss

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/sda"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestSID(t *testing.T) {
	sl := newSID(LTSS)
	ss := newSID(STSS)

	assert.True(t, sl.IsLTSS())
	assert.False(t, sl.IsSTSS())
	assert.NotEqual(t, sl, ss)

	assert.True(t, ss.IsSTSS())
	assert.False(t, ss.IsLTSS())
}

func TestJVSS(t *testing.T) {
	// Setup parameters
	var name string = "JVSS" // Protocol name
	var nodes uint32 = 5     // Number of nodes
	var rounds int = 1       // Number of rounds

	//msg, _ := hex.DecodeString("c20089ec2b85cdcb587c40bafb7ab03bb3773b18d9b263836f720206c2f7f6fc")
	//msg, _ := hex.DecodeString("324c4fab45576044b92d3dbdc6b19cde1bdd6dc612468eacd4d4eb7714b4d0e1")
	msg, _ := hex.DecodeString("0175b946a32825cf130741e790892361245b5b9e9c7f563d8ddc02a5ab9c317e")
	//msg := []byte("OpenPGP")
	fmt.Println("message:" + hex.EncodeToString(msg))

	local := sda.NewLocalTest()
	_, _, tree := local.GenTree(int(nodes), false, true, true)
	defer local.CloseAll()

	log.Lvl1("JVSS - starting")
	leader, err := local.CreateProtocol(name, tree)
	if err != nil {
		t.Fatal("Couldn't initialise protocol tree:", err)
	}
	jv := leader.(*JVSS)
	leader.Start()
	log.Lvl1("JVSS - setup done")

	for i := 0; i < rounds; i++ {
		log.Lvl1("JVSS - starting round", i)
		log.Lvl1("JVSS - requesting signature")
		sig, err := jv.Sign(msg)
		if err != nil {
			t.Fatal("Error signature failed", err)
		}
		b, _ := jv.keyPair.Public.MarshalBinary()
		fmt.Println("Public key :" + hex.EncodeToString(b))
		r, _ := sig.Random.SecretCommit().MarshalBinary()
		fmt.Println("Random commit R:" + hex.EncodeToString(r))
		s, _ := (*sig.Signature).MarshalBinary()
		fmt.Println("Signature (point) S :" + hex.EncodeToString(s))
		log.Lvl1("JVSS - signature received")
		err = jv.Verify(msg, sig)
		if err != nil {
			t.Fatal("Error signature verification failed", err)
		}
		log.Lvl1("JVSS - signature verification succeded")
	}

}
