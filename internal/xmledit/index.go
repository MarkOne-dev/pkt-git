// Package xmledit indexes a Packet Tracer XML snapshot by byte offset so edits
// can replace only the requested span and leave every other byte untouched.
package xmledit

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	// ErrDeviceNotFound reports an exact device name that is absent.
	ErrDeviceNotFound = errors.New("xmledit: device not found")
	// ErrDeviceAmbiguous reports more than one device with the same name.
	ErrDeviceAmbiguous = errors.New("xmledit: device name is ambiguous")
	// ErrNoRunningConfig reports a device without a RUNNINGCONFIG block.
	ErrNoRunningConfig = errors.New("xmledit: device has no RUNNINGCONFIG")
	// ErrInterfaceNotFound reports an interface line that is absent.
	ErrInterfaceNotFound = errors.New("xmledit: interface not found")
	// ErrInterfaceUnclosed reports an interface block without a following "!".
	ErrInterfaceUnclosed = errors.New("xmledit: interface block has no closing ! line")
)

// Index is a byte-offset view of one XML document.
type Index struct {
	Data    []byte
	Devices []Device
	byName  map[string][]int
}

// Device is a DEVICE element with its name and running config.
type Device struct {
	Name          string
	Start         int
	End           int
	NameStart     int
	NameEnd       int
	RunningConfig *Config
}

// Config is a RUNNINGCONFIG element and its ordered LINE elements.
type Config struct {
	Start        int
	End          int
	ContentStart int
	ContentEnd   int
	Lines        []Line
}

// Line is one LINE element, including its physical line bounds.
type Line struct {
	Text         string
	Start        int
	End          int
	ContentStart int
	ContentEnd   int
	LineStart    int
	LineEnd      int
	Indent       string
	Newline      string
	SelfClosing  bool
}

// Block is an interface block: the interface line, the closing "!" line, and
// the lines between them.
type Block struct {
	Interface    Line
	Close        Line
	Lines        []Line
	ContentStart int
	ContentEnd   int
	Indent       string
	Newline      string
}

// BuildIndex parses data and records device, config, and line byte spans.
func BuildIndex(data []byte) (*Index, error) {
	index := &Index{Data: data}
	decoder := xml.NewDecoder(bytes.NewReader(data))

	var stack []string
	var current *deviceBuilder
	var config *configBuilder
	var line *lineBuilder
	var previous int64

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xmledit: parse XML: %w", err)
		}
		end := int(decoder.InputOffset())

		switch token := token.(type) {
		case xml.StartElement:
			parent := ""
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			stack = append(stack, token.Name.Local)

			switch token.Name.Local {
			case "DEVICE":
				if parent == "DEVICES" {
					current = &deviceBuilder{start: int(previous)}
				}
			case "NAME":
				if parent == "ENGINE" && current != nil && !current.nameOpen {
					current.nameOpen = true
					current.nameStart = int(previous)
					current.nameContentStart = end
				}
			case "RUNNINGCONFIG":
				if parent == "ENGINE" && current != nil && config == nil {
					config = &configBuilder{start: int(previous), contentStart: end}
				}
			case "LINE":
				if parent == "RUNNINGCONFIG" && config != nil {
					line = &lineBuilder{start: int(previous), contentStart: end}
				}
			}

		case xml.CharData:
			if line != nil {
				line.text.Write(token)
			} else if current != nil && current.nameOpen {
				current.nameText.Write(token)
			}

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("xmledit: unexpected end element %s", token.Name.Local)
			}
			stack = stack[:len(stack)-1]

			switch token.Name.Local {
			case "LINE":
				if line != nil && config != nil {
					config.lines = append(config.lines, line.finish(data, int(previous), end))
					line = nil
				}
			case "RUNNINGCONFIG":
				if config != nil && current != nil {
					config.end = end
					config.contentEnd = int(previous)
					current.config = config
					config = nil
				}
			case "NAME":
				if current != nil && current.nameOpen {
					current.nameOpen = false
					current.name = current.nameText.String()
					current.nameEnd = end
				}
			case "DEVICE":
				if current != nil {
					current.end = end
					index.addDevice(current.build())
					current = nil
				}
			}
		}

		previous = int64(end)
	}

	index.byName = make(map[string][]int, len(index.Devices))
	for i := range index.Devices {
		name := index.Devices[i].Name
		if name != "" {
			index.byName[name] = append(index.byName[name], i)
		}
	}
	return index, nil
}

// FindDevice returns the unique device with the exact name.
func (index *Index) FindDevice(name string) (*Device, error) {
	matches := index.byName[name]
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrDeviceNotFound, name)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("%w: %q", ErrDeviceAmbiguous, name)
	}
	return &index.Devices[matches[0]], nil
}

// InterfaceBlock finds an interface block by exact "interface <iface>" line.
func (device *Device) InterfaceBlock(iface string) (*Block, error) {
	if device.RunningConfig == nil {
		return nil, fmt.Errorf("%w: %q", ErrNoRunningConfig, device.Name)
	}

	lines := device.RunningConfig.Lines
	openIndex := -1
	for i, line := range lines {
		if line.Text == "interface "+iface {
			openIndex = i
			break
		}
	}
	if openIndex < 0 {
		return nil, fmt.Errorf("%w: %q in device %q", ErrInterfaceNotFound, iface, device.Name)
	}

	closeIndex := -1
	for i := openIndex + 1; i < len(lines); i++ {
		if lines[i].Text == "!" {
			closeIndex = i
			break
		}
	}
	if closeIndex < 0 {
		return nil, fmt.Errorf("%w: %q in device %q", ErrInterfaceUnclosed, iface, device.Name)
	}

	open := lines[openIndex]
	closeLine := lines[closeIndex]
	block := &Block{
		Interface:    open,
		Close:        closeLine,
		Lines:        lines[openIndex+1 : closeIndex],
		ContentStart: open.End,
		ContentEnd:   closeLine.Start,
		Indent:       open.Indent,
		Newline:      open.Newline,
	}
	if len(block.Lines) > 0 {
		block.Indent = block.Lines[0].Indent
	}
	return block, nil
}

type deviceBuilder struct {
	start            int
	end              int
	name             string
	nameStart        int
	nameEnd          int
	nameContentStart int
	nameOpen         bool
	nameText         strings.Builder
	config           *configBuilder
}

func (index *Index) addDevice(device Device) {
	index.Devices = append(index.Devices, device)
}

func (builder *deviceBuilder) build() Device {
	device := Device{
		Name:      builder.name,
		Start:     builder.start,
		End:       builder.end,
		NameStart: builder.nameStart,
		NameEnd:   builder.nameEnd,
	}
	if builder.config != nil {
		device.RunningConfig = &Config{
			Start:        builder.config.start,
			End:          builder.config.end,
			ContentStart: builder.config.contentStart,
			ContentEnd:   builder.config.contentEnd,
			Lines:        builder.config.lines,
		}
	}
	return device
}

type configBuilder struct {
	start        int
	end          int
	contentStart int
	contentEnd   int
	lines        []Line
}

type lineBuilder struct {
	start        int
	contentStart int
	text         strings.Builder
}

func (builder *lineBuilder) finish(data []byte, contentEnd, end int) Line {
	lineStart, lineEnd := physicalLine(data, builder.start, end)
	indentStart := builder.start
	for indentStart > 0 && (data[indentStart-1] == ' ' || data[indentStart-1] == '\t') {
		indentStart--
	}
	return Line{
		Text:         builder.text.String(),
		Start:        builder.start,
		End:          end,
		ContentStart: builder.contentStart,
		ContentEnd:   contentEnd,
		LineStart:    lineStart,
		LineEnd:      lineEnd,
		Indent:       string(data[indentStart:builder.start]),
		Newline:      newlineAt(data, lineEnd),
		SelfClosing:  bytes.HasSuffix(data[builder.start:end], []byte("/>")),
	}
}

func physicalLine(data []byte, start, end int) (int, int) {
	lineStart := bytes.LastIndexByte(data[:start], '\n') + 1
	lineEnd := len(data)
	if offset := bytes.IndexByte(data[end:], '\n'); offset >= 0 {
		lineEnd = end + offset + 1
	}
	return lineStart, lineEnd
}

func newlineAt(data []byte, lineEnd int) string {
	if lineEnd >= 2 && data[lineEnd-2] == '\r' && data[lineEnd-1] == '\n' {
		return "\r\n"
	}
	return "\n"
}
