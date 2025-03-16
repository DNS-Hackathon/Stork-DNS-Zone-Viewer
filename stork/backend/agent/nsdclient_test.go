package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	input := `zone: arpa
	state: ok
	served-serial: "2025031500 since 2025-03-15T14:07:44"
	commit-serial: "2025031500 since 2025-03-15T15:19:08"
	wait: "1779 sec between attempts"
	`
	parsed, err := Parse(input)
	if err != nil {
		t.Fatalf("error parsing nsd-control zonestatus output: %s", err)
	}
	require.NoError(t, err)
	require.NotNil(t, parsed)

}
