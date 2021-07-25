package template

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasEnv(t *testing.T) {
	os.Setenv("GODUB_TEST_HAS_ENV_EXISTING", "value")
	assert.True(t, hasEnv("GODUB_TEST_HAS_ENV_EXISTING"))
	assert.False(t, hasEnv("GODUB_TEST_HAS_ENV_SOMETHING_ELSE"))
}