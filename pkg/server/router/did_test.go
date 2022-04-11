package router

import (
	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tbd54566975/ssi-service/pkg/service/did"
	"github.com/tbd54566975/ssi-service/pkg/service/framework"
	"github.com/tbd54566975/ssi-service/pkg/storage"
	"log"
	"os"
	"testing"
)

func TestDIDRouter(t *testing.T) {

	// remove the db file after the test
	t.Cleanup(func() {
		_ = os.Remove(storage.DBFile)
	})

	t.Run("Nil Service", func(tt *testing.T) {
		didRouter, err := NewDIDRouter(nil, nil)
		assert.Error(tt, err)
		assert.Empty(tt, didRouter)
		assert.Contains(tt, err.Error(), "service cannot be nil")
	})

	t.Run("Bad Service", func(tt *testing.T) {
		didRouter, err := NewDIDRouter(&testService{}, nil)
		assert.Error(tt, err)
		assert.Empty(tt, didRouter)
		assert.Contains(tt, err.Error(), "could not create DID router with service type: test")
	})

	t.Run("DID Service Test", func(tt *testing.T) {
		bolt, err := storage.NewBoltDB()
		assert.NoError(tt, err)
		assert.NotEmpty(tt, bolt)

		logger := log.New(os.Stdout, "ssi-test", log.LstdFlags)
		didService, err := did.NewDIDService(logger, []did.Method{did.KeyMethod}, bolt)
		assert.NoError(tt, err)
		assert.NotEmpty(tt, didService)

		// check type and status
		assert.Equal(tt, framework.DID, didService.Type())
		assert.Equal(tt, framework.StatusReady, didService.Status().Status)

		// get unknown handler
		_, err = didService.GetHandler("bad")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "could not get handler for DID method: bad")

		supported := didService.GetSupportedMethods()
		assert.NotEmpty(tt, supported)
		assert.Len(tt, supported, 1)
		assert.Equal(tt, did.KeyMethod, supported[0])

		// get known handler
		keyHandler, err := didService.GetHandler(did.KeyMethod)
		assert.NoError(tt, err)
		assert.NotEmpty(tt, keyHandler)

		// bad key type
		_, err = keyHandler.CreateDID(did.CreateDIDRequest{KeyType: "bad"})
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "could not create did:key")

		// good key type
		createDIDResponse, err := keyHandler.CreateDID(did.CreateDIDRequest{KeyType: crypto.Ed25519})
		assert.NoError(tt, err)
		assert.NotEmpty(tt, createDIDResponse)

		// check the DID is a did:key
		assert.Contains(tt, createDIDResponse.DID.ID, "did:key")

		// get it back
		getDIDResponse, err := keyHandler.GetDID(createDIDResponse.DID.ID)
		assert.NoError(tt, err)
		assert.NotEmpty(tt, getDIDResponse)

		// make sure it's the same value
		assert.Equal(tt, createDIDResponse.DID.ID, getDIDResponse.DID.ID)
	})
}

type testService struct{}

func (s *testService) Type() framework.Type {
	return "test"
}

func (s *testService) Status() framework.Status {
	return framework.Status{Status: "ready"}
}