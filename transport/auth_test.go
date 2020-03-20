package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sync"
	"testing"
)

func TestSha256Length(t *testing.T) {
	hash := sha256.New()
	hash.Write([]byte("ahahahfdkalfdjaldfjasklfjlasjflkajdfasjkfafajlkdfjaklfjakljfakldfjasfjasjf"))
	sum := sha256.New().Sum(nil)
	if 32 != len(sum) {
		t.Error("sum not equal", len(sum))
		return
	}
	if 64 != len(hex.EncodeToString(sum)) {
		t.Error("str not equal")
		return
	}
}

func TestCheckAgentConn(t *testing.T) {
	client, server := net.Pipe()

	secret := "test"
	name := "cocopark"

	var wg sync.WaitGroup
	wg.Add(2)

	randKeyClient := []byte{}
	randKeyServer := []byte{}

	// client
	go func() {
		defer client.Close()
		defer wg.Done()
		randKey, err := RequireAuth(client, secret, name)
		randKeyClient = randKey
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// server
	go func() {
		defer server.Close()
		defer wg.Done()
		authName, randKey, err := CheckAgentConn(server, secret)
		randKeyServer = randKey
		if err != nil {
			t.Error(err)
			return
		}
		if authName != name {
			t.Error("name not match")
			return
		}
	}()

	wg.Wait()

	if !bytes.Equal(randKeyClient, randKeyServer) {
		t.Error("randKey not equal", randKeyClient, randKeyServer)
	}
}
