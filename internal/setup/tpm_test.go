package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallTPM_DetectsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tpmDir := filepath.Join(tmpDir, "plugins", "tpm")
	os.MkdirAll(tpmDir, 0755)

	result, err := InstallTPM(filepath.Join(tmpDir, "plugins"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AlreadyInstalled {
		t.Fatal("should detect existing tpm installation")
	}
}

func TestInstallTPM_FailsGracefullyWithoutGit(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, "plugins")

	// This test will actually try to clone if git is available
	// We just verify it doesn't panic
	_, _ = InstallTPM(pluginsDir)
}
