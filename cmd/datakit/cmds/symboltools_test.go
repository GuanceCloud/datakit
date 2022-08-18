package cmds

import (
	"fmt"
	"os"
	"testing"
)

func TestCheckJavaInstall(t *testing.T) {
	path, ok := checkToolInstalled("java")
	fmt.Println(path, ok)
}

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

func TestInstallProguard(t *testing.T) {
	if err := installProguard(); err != nil {
		fmt.Println(err)
	}
}

func TestInstallDwarf(t *testing.T) {
	if err := installDwarf(); err != nil {
		fmt.Println(err)
	}
}
