package main

import (
	"bytes"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const (
	officialKeySymbol  = "github.com/AlonaBot/AlonaBot/internal/core/trust.officialKey"
	versionSymbol      = "github.com/AlonaBot/AlonaBot/internal/app.releaseVersion"
	goStringHeaderSize = 16 // 五个发行目标均为 64 位，string 为指针和长度。
)

func checkBuildInfo(executable, goos, goarch string) (*debug.BuildInfo, error) {
	info, err := buildinfo.ReadFile(executable)
	if err != nil {
		return nil, err
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["GOOS"] != goos || settings["GOARCH"] != goarch || settings["CGO_ENABLED"] != "0" {
		return nil, fmt.Errorf("binary target = %s/%s, CGO_ENABLED=%q; want %s/%s, CGO_ENABLED=0", settings["GOOS"], settings["GOARCH"], settings["CGO_ENABLED"], goos, goarch)
	}
	return info, nil
}

type binarySection struct {
	address, size, offset uint64
}

// linkedStrings 从实际加载节与符号读取 -X 注入值，检查器无需安装 Go 工具链。
func linkedStrings(filename, goos string, names ...string) (map[string][]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var sections []binarySection
	addresses := map[string]uint64{}
	var order binary.ByteOrder = binary.LittleEndian
	switch goos {
	case "linux":
		image, err := elf.NewFile(file)
		if err != nil {
			return nil, err
		}
		order = image.ByteOrder
		for _, section := range image.Sections {
			if section.Flags&elf.SHF_ALLOC != 0 && section.Type != elf.SHT_NOBITS {
				sections = append(sections, binarySection{section.Addr, section.Size, section.Offset})
			}
		}
		symbols, err := image.Symbols()
		if err != nil {
			return nil, err
		}
		for _, symbol := range symbols {
			addresses[symbol.Name] = symbol.Value
		}
	case "darwin":
		image, err := macho.NewFile(file)
		if err != nil {
			return nil, err
		}
		order = image.ByteOrder
		for _, section := range image.Sections {
			sections = append(sections, binarySection{section.Addr, section.Size, uint64(section.Offset)})
		}
		if image.Symtab == nil {
			return nil, fmt.Errorf("Mach-O has no symbol table: %s", filename)
		}
		for _, symbol := range image.Symtab.Syms {
			addresses[strings.TrimPrefix(symbol.Name, "_")] = symbol.Value
		}
	case "windows":
		image, err := pe.NewFile(file)
		if err != nil {
			return nil, err
		}
		header, ok := image.OptionalHeader.(*pe.OptionalHeader64)
		if !ok {
			return nil, fmt.Errorf("PE is not 64-bit: %s", filename)
		}
		for _, section := range image.Sections {
			sections = append(sections, binarySection{header.ImageBase + uint64(section.VirtualAddress), uint64(section.Size), uint64(section.Offset)})
		}
		for _, symbol := range image.Symbols {
			if symbol.SectionNumber > 0 && int(symbol.SectionNumber) <= len(sections) {
				addresses[symbol.Name] = sections[symbol.SectionNumber-1].address + uint64(symbol.Value)
			}
		}
	}
	read := func(address, size uint64) ([]byte, error) {
		for _, section := range sections {
			if address >= section.address && address-section.address <= section.size && size <= section.size-(address-section.address) {
				offset := section.offset + address - section.address
				if offset < section.offset || offset > uint64(fileInfo.Size()) || size > uint64(fileInfo.Size())-offset {
					return nil, fmt.Errorf("binary string lies outside the file: offset %d size %d", offset, size)
				}
				raw := make([]byte, size)
				_, err := file.ReadAt(raw, int64(offset))
				return raw, err
			}
		}
		return nil, fmt.Errorf("binary address %#x size %d is outside file-backed sections", address, size)
	}
	values := map[string][]byte{}
	for _, name := range names {
		address, exists := addresses[name]
		if !exists {
			return nil, fmt.Errorf("missing linked symbol %s", name)
		}
		header, err := read(address, goStringHeaderSize)
		if err != nil {
			return nil, err
		}
		values[name], err = read(order.Uint64(header[:8]), order.Uint64(header[8:]))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
	}
	return values, nil
}

func checkEmbedded(binary, resource []byte, name string) error {
	if len(resource) == 0 || !bytes.Contains(binary, resource) {
		return fmt.Errorf("binary does not contain the exact packaged %s bytes", name)
	}
	return nil
}
