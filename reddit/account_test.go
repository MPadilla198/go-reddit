package reddit

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountService_GetIdentity(t *testing.T) {
	client, _ := setup(t)

	t.Log("Setup complete.")

	// blob, err := readFileContents(testDataPath + "/account/info.json")
	// require.NoError(t, err)
	var ctx = context.Background()
	t.Log("Background context acquired.")

	_, resp, err := client.Account.GetIdentity(ctx)
	t.Log("Received response.")
	require.NoError(t, err)
	// require.Equal(t, expectedIdentity, identity)
	require.NotNil(t, resp)

	t.Logf("%v\n", resp)
}
