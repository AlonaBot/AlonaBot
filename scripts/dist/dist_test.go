package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/release"
)

func TestDistribution(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	public, key, keyFile := distributionKey(t)
	target := runtime.GOOS + "/" + runtime.GOARCH
	first, second := t.TempDir(), t.TempDir()
	for _, out := range []string{first, second} {
		runDistribution(t, keyFile, out, target)
		verifyDistribution(t, out, target, public, key)
	}
	name := distributionName(target)
	a, err := os.ReadFile(filepath.Join(first, name))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(second, name))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("distribution archives differ between identical builds")
	}
}

func distributionKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey, string) {
	t.Helper()
	public, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyFile := filepath.Join(t.TempDir(), "signing.key")
	if err := os.WriteFile(keyFile, []byte(base64.StdEncoding.EncodeToString(key)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return public, key, keyFile
}

func runDistribution(t *testing.T, keyFile, out, platforms string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "run", "./scripts/dist", "--version", "0.1.0",
		"--key", keyFile, "--out", out, "--platforms", platforms)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dist: %v\n%s", err, output)
	}
}

func distributionName(target string) string {
	name := "alonabot-0.1.0-" + strings.ReplaceAll(target, "/", "-")
	if strings.HasPrefix(target, "windows/") {
		return name + ".zip"
	}
	return name + ".tar.gz"
}

func verifyDistribution(t *testing.T, out, target string, public ed25519.PublicKey, key ed25519.PrivateKey) {
	t.Helper()
	installer, err := os.ReadFile(filepath.Join(out, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("deploy", "install.sh"))
	if err != nil || !bytes.Equal(installer, source) {
		t.Fatalf("distribution install script differs from source: %v", err)
	}
	name := distributionName(target)
	unpacked := t.TempDir()
	unpackDistribution(t, filepath.Join(out, name), unpacked)
	root := filepath.Join(unpacked, strings.TrimSuffix(strings.TrimSuffix(name, ".tar.gz"), ".zip"))
	goos, goarch, _ := strings.Cut(target, "/")
	core := filepath.Join(root, executableName(goos))
	if _, err := os.Stat(core); err != nil {
		t.Fatal(err)
	}
	if target == runtime.GOOS+"/"+runtime.GOARCH {
		encoded := linkedPublicKey(t, core)
		got, err := release.DecodePublicKey(encoded)
		if err != nil || !bytes.Equal(got, public) {
			t.Fatalf("linked official key = %q, %v; want %x", encoded, err, public)
		}
	}
	for _, component := range components {
		dir := filepath.Join(root, "packages", component.id)
		read := func(name string) []byte {
			t.Helper()
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(raw, []byte(base64.StdEncoding.EncodeToString(key))) {
				t.Fatalf("private key leaked into %s", name)
			}
			return raw
		}
		raw, sig := read("release.json"), read("release.json.sig")
		pub, err := release.DecodePublicKey(read("release.pub"))
		if err != nil || !bytes.Equal(pub, public) {
			t.Fatalf("release public key = %x, %v", pub, err)
		}
		publication, err := release.Verify(raw, sig, pub)
		if err != nil {
			t.Fatal(err)
		}
		if publication.PluginID != component.id {
			t.Fatalf("plugin ID = %q, want %q", publication.PluginID, component.id)
		}
		asset, ok := publication.Asset(goos, goarch)
		if !ok {
			asset, ok = publication.Asset("any", "any")
		}
		if !ok {
			t.Fatalf("%s has no asset for %s", component.id, target)
		}
		archive := read(filepath.Base(asset.URL))
		digest := sha256.Sum256(archive)
		if int64(len(archive)) != asset.Size || hex.EncodeToString(digest[:]) != asset.SHA256 {
			t.Fatalf("%s asset size or digest differs", component.id)
		}
		if err := checkPrivateKey(filepath.Join(dir, filepath.Base(asset.URL)), key); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 4 {
			t.Fatalf("%s package entries = %d, %v; want archive and three metadata files", component.id, len(entries), err)
		}
		topLevel, err := os.ReadFile(filepath.Join(out, "packages", component.id, "release.json"))
		if err != nil || !bytes.Equal(topLevel, raw) {
			t.Fatalf("%s external release differs: %v", component.id, err)
		}
	}
	digest, err := archiveDigest(filepath.Join(out, name))
	if err != nil {
		t.Fatal(err)
	}
	sums, err := os.ReadFile(filepath.Join(out, "SHA256SUMS"))
	if err != nil || !strings.Contains(string(sums), fmt.Sprintf("%x  %s\n", digest, name)) {
		t.Fatalf("missing archive checksum: %v", err)
	}
}

func unpackDistribution(t *testing.T, path, destination string) {
	t.Helper()
	write := func(name string, mode os.FileMode, reader io.Reader) {
		t.Helper()
		path := filepath.Join(destination, filepath.FromSlash(name))
		if mode.IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			return
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, mode.Perm())
		if err != nil {
			t.Fatal(err)
		}
		_, copyErr := io.Copy(file, reader)
		closeErr := file.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("extract %s: copy %v; close %v", name, copyErr, closeErr)
		}
	}
	if strings.HasSuffix(path, ".zip") {
		archive, err := zip.OpenReader(path)
		if err != nil {
			t.Fatal(err)
		}
		defer archive.Close()
		for _, entry := range archive.File {
			reader, err := entry.Open()
			if err != nil {
				t.Fatal(err)
			}
			write(entry.Name, entry.Mode(), reader)
			if err := reader.Close(); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		write(header.Name, header.FileInfo().Mode(), archive)
	}
}

// linkedPublicKey 只读检查实际 Core 的链接字符串，避免给生产 CLI 增加测试开关。
func linkedPublicKey(t *testing.T, path string) []byte {
	t.Helper()
	output, err := exec.CommandContext(t.Context(), "go", "tool", "nm", "-size", path).CombinedOutput()
	if err != nil {
		t.Fatalf("read Core symbols: %v\n%s", err, output)
	}
	var address uint64
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[3] == officialKeySymbol {
			address, err = strconv.ParseUint(fields[0], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if address == 0 {
		t.Fatal("linked official public key symbol is missing")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type section struct{ address, size, offset uint64 }
	var sections []section
	var order binary.ByteOrder = binary.LittleEndian
	switch runtime.GOOS {
	case "darwin":
		binary, err := macho.NewFile(file)
		if err != nil {
			t.Fatal(err)
		}
		order = binary.ByteOrder
		for _, s := range binary.Sections {
			sections = append(sections, section{s.Addr, s.Size, uint64(s.Offset)})
		}
	case "linux":
		binary, err := elf.NewFile(file)
		if err != nil {
			t.Fatal(err)
		}
		order = binary.ByteOrder
		for _, s := range binary.Sections {
			sections = append(sections, section{s.Addr, s.Size, s.Offset})
		}
	case "windows":
		binary, err := pe.NewFile(file)
		if err != nil {
			t.Fatal(err)
		}
		base := binary.OptionalHeader.(*pe.OptionalHeader64).ImageBase
		for _, s := range binary.Sections {
			sections = append(sections, section{base + uint64(s.VirtualAddress), uint64(s.Size), uint64(s.Offset)})
		}
	}
	read := func(address, size uint64) []byte {
		for _, s := range sections {
			if address >= s.address && address+size <= s.address+s.size {
				raw := make([]byte, size)
				if _, err := file.ReadAt(raw, int64(s.offset+address-s.address)); err != nil {
					t.Fatal(err)
				}
				return raw
			}
		}
		t.Fatal("linked official public key section is missing")
		return nil
	}
	// 五个发行目标均为 64 位，Go string 由指针与长度两个机器字组成。
	header := read(address, 16)
	return read(order.Uint64(header[:8]), order.Uint64(header[8:]))
}
