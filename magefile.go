//go:build mage

//mage:multiline

// Build the gophermap package
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

// Builds a debug gophermap for the host platform into dist/ (full debug info, no optimizations)
func Build() error {
	_, err := compile(hostTarget(), buildVersion(), false)
	return err
}

// Builds an optimized gophermap for the host platform and packages it into dist/
func BuildRelease() error {
	return release(hostTarget(), buildVersion())
}

// Builds optimized gophermap for every release target and packages them into dist/
func BuildAll() error {
	version := buildVersion()
	for _, t := range targets {
		if err := release(t, version); err != nil {
			return err
		}
	}
	return nil
}

func hostTarget() target {
	return target{runtime.GOOS, runtime.GOARCH, hostGoarm()}
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
// Without git (or outside a git checkout) it falls back to just <date>.
func buildVersion() string {
	date := time.Now().UTC().Format("20060102")
	sha, err := git("rev-parse", "--short=7", "HEAD")
	if err != nil {
		return date
	}
	if tag, err := git("describe", "--tags", "--exact-match", "HEAD"); err == nil {
		return tag + "-" + date + "-" + sha
	}
	return date + "-" + sha
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func release(t target, version string) error {
	nameTriplet, err := compile(t, version, true)
	if err != nil {
		return err
	}
	return tarGz(filepath.Join("dist", "release-"+nameTriplet+".tar.gz"), "dist", nameTriplet)
}

// compile builds gophermap for t into dist/<nameTriplet>/ and returns nameTriplet.
func compile(t target, version string, release bool) (string, error) {
	nameTriplet := "gophermap-" + t.goos + "-" + t.goarch
	if t.goarch == "arm" {
		nameTriplet += "v" + strconv.Itoa(t.goarm)
	}
	filename := "gophermap"
	if t.goos == "windows" {
		filename += ".exe"
	}

	args := []string{"build", "-o", filepath.Join("dist", nameTriplet, filename)}
	if release {
		fmt.Println("Building release", nameTriplet, version)
		// strip symbols and DWARF, drop local paths, apply cmd/default.pgo if present
		args = append(args, "-trimpath", "-pgo=auto", "-ldflags=-s -w -X main.Version="+version)
	} else {
		fmt.Println("Building debug", nameTriplet, version)
		// keep DWARF and disable optimizations and inlining so debuggers can step through
		args = append(args, "-gcflags=all=-N -l", "-ldflags=-X main.Version="+version+"-debug")
	}
	args = append(args, "./cmd")

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOOS="+t.goos, "GOARCH="+t.goarch, "CGO_ENABLED=0")
	if t.goarch == "arm" {
		cmd.Env = append(cmd.Env, "GOARM="+strconv.Itoa(t.goarm))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("building %s: %w", nameTriplet, err)
	}
	return nameTriplet, nil
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

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("dist")
}
