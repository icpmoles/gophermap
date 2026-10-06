//go:build mage

//mage:multiline

// Set the general description you want to have displayed with mage -l here.
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/cpu"
	// mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

type target struct {
	goos   string
	goarch string
	goarm  int // only meaningful when goarch is "arm"
}

// targets mirrors the build matrix in .github/workflows/build.yaml
var targets = []target{
	{"windows", "amd64", 0},
	{"windows", "arm64", 0},
	{"darwin", "amd64", 0},
	{"darwin", "arm64", 0},
	{"linux", "amd64", 0},
	{"linux", "arm64", 0},
	{"linux", "arm", 7},
	{"linux", "arm", 6},
	{"linux", "arm", 5},
}

// Builds and packages gophermap for the host platform into dist/
func Build() error {
	version, err := buildVersion()
	if err != nil {
		return err
	}
	return build(target{runtime.GOOS, runtime.GOARCH, hostGoarm()}, version)
}

// Builds and packages gophermap for every release target into dist/
func BuildAll() error {
	version, err := buildVersion()
	if err != nil {
		return err
	}
	for _, t := range targets {
		if err := build(t, version); err != nil {
			return err
		}
	}
	return nil
}

// hostGoarm returns the GOARM level supported by the host CPU, or 0 when not
// running on arm. A GOARM already set in the environment takes precedence.
// GOARM=5 is soft-float, 6 requires VFP and 7 requires VFPv3.
func hostGoarm() int {
	if runtime.GOARCH != "arm" {
		return 0
	}
	if v, err := strconv.Atoi(os.Getenv("GOARM")); err == nil {
		return v
	}
	switch {
	case !cpu.ARM.HasVFP:
		return 5
	case !cpu.ARM.HasVFPv3:
		return 6
	default:
		return 7
	}
}

// buildVersion returns <tag>-<date>-<sha> when HEAD is tagged, <date>-<sha> otherwise.
func buildVersion() (string, error) {
	date := time.Now().UTC().Format("20060102")
	sha, err := git("rev-parse", "--short=7", "HEAD")
	if err != nil {
		return "", fmt.Errorf("reading commit: %w", err)
	}
	if tag, err := git("describe", "--tags", "--exact-match", "HEAD"); err == nil {
		return tag + "-" + date + "-" + sha, nil
	}
	return date + "-" + sha, nil
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func build(t target, version string) error {
	nameTriplet := "gophermap-" + t.goos + "-" + t.goarch
	if t.goarch == "arm" {
		nameTriplet += "v" + strconv.Itoa(t.goarm)
	}
	filename := "gophermap"
	if t.goos == "windows" {
		filename += ".exe"
	}
	fmt.Println("Building", nameTriplet, version)

	cmd := exec.Command("go", "build",
		"-ldflags=-s -w -X main.Version="+version,
		"-trimpath", "-v",
		"-o", filepath.Join("dist", nameTriplet, filename),
		"cmd/cli.go")
	cmd.Env = append(os.Environ(), "GOOS="+t.goos, "GOARCH="+t.goarch, "CGO_ENABLED=0")
	if t.goarch == "arm" {
		cmd.Env = append(cmd.Env, "GOARM="+strconv.Itoa(t.goarm))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building %s: %w", nameTriplet, err)
	}

	return tarGz(filepath.Join("dist", "release-"+nameTriplet+".tar.gz"), "dist", nameTriplet)
}

// tarGz archives dir (relative to root) into dst, like `tar -czf dst -C root dir`.
func tarGz(dst, root, dir string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	err = filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if d.IsDir() {
			hdr.Name += "/"
		} else if !strings.HasSuffix(hdr.Name, ".exe") {
			hdr.Mode = 0o755 // Windows hosts don't record the executable bit
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(tw, src)
		return err
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return f.Close()
}

// A custom install step if you need your bin someplace other than go/bin
// func Install() error {
// 	mg.Deps(Build)
// 	fmt.Println("Installing...")
// 	return os.Rename("./MyApp", "/usr/bin/MyApp")
// }

// Manage your deps, or running package managers.
// func InstallDeps() error {
// 	fmt.Println("Installing Deps...")
// 	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
// 	return cmd.Run()
// }

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("dist")
}
