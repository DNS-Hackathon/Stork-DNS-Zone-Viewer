package nsdconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNSDVersion(t *testing.T) {
	input := `version: 4.1.1-rc1`
	version, err := ParseNSDVersion(input)
	require.NoError(t, err)
	require.EqualValues(t, "4.1.1-rc1", version)
}
