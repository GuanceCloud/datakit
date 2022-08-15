package cmds

import (
	"fmt"
	"os"
	"testing"
)

func TestScanJDKBinPath(t *testing.T) {
	if err := os.Chdir(os.TempDir()); err != nil {
		t.Fatal(err)
	}
	binDir, err := scanJDKBinPath("jdk-11.0.15+10")
	if err != nil {
		fmt.Println("scan bin path err:", err)
	}

	fmt.Println(binDir)
}

func TestScanProguardBinPath(t *testing.T) {
	path, err := scanProguardBinPath("/Users/zy/Downloads/proguard-7.2.2")
	if err != nil {
		fmt.Println("scan bin path err:", err)
	}

	fmt.Println(path)
}
