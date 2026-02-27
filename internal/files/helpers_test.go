package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExistsFileAndDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.txt")

	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o600))
	require.True(t, Exists(filePath))
	require.False(t, Exists(tmpDir))
	require.False(t, Exists(filepath.Join(tmpDir, "missing.txt")))
}
