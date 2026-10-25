// Package topology parses a decoded Packet Tracer XML document into ordered
// devices, links, and endpoints. It does not touch the filesystem.
package topology

import (
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoNetwork reports an XML document without a NETWORK element.
var ErrNoNetwork = errors.New("topology: XML document has no NETWORK element")

var refPattern = regexp.MustCompile(`^save-ref-id:(\d+)$`)

// Endpoint is one side of a cable: a device name and a physical port.
type Endpoint struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

// Link is a cable between two topology endpoints.
type Link struct {
	LinkType string   `json:"link_type"`
	From     Endpoint `json:"from"`
	To       Endpoint `json:"to"`
}

// Device is a topology device and its CLI configuration.
type Device struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Model         string   `json:"model"`
	Ref           string   `json:"ref"`
	RunningConfig []string `json:"-"`
	StartupConfig []string `json:"-"`
}

// Document is the parsed topology in document order.
type Document struct {
	Root    string
	Devices []Device
	Links   []Link
}

type xmlDocument struct {
	XMLName xml.Name
	Network *xmlNetwork `xml:"NETWORK"`
}

type xmlNetwork struct {
	Devices struct {
		Device []xmlDevice `xml:"DEVICE"`
	} `xml:"DEVICES"`
	Links struct {
		Link []xmlLink `xml:"LINK"`
	} `xml:"LINKS"`
}

type xmlDevice struct {
	Engine *xmlEngine `xml:"ENGINE"`
}

type xmlEngine struct {
	Type          xmlType    `xml:"TYPE"`
	Name          string     `xml:"NAME"`
	SaveRefID     string     `xml:"SAVE_REF_ID"`
	RunningConfig *xmlConfig `xml:"RUNNINGCONFIG"`
	StartupConfig *xmlConfig `xml:"STARTUPCONFIG"`
}

type xmlType struct {
	Text  string `xml:",chardata"`
	Model string `xml:"model,attr"`
}

type xmlConfig struct {
	Lines []string `xml:"LINE"`
}

type xmlLink struct {
	Type  string    `xml:"TYPE"`
	Cable *xmlCable `xml:"CABLE"`
}

type xmlCable struct {
	From  string   `xml:"FROM"`
	To    string   `xml:"TO"`
	Ports []string `xml:"PORT"`
}

// Parse extracts devices and links from a decoded Packet Tracer XML document.
func Parse(data []byte) (*Document, error) {
	var root xmlDocument
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("topology: parse XML: %w", err)
	}
	if root.Network == nil {
		return nil, ErrNoNetwork
	}

	doc := &Document{Root: root.XMLName.Local}
	byRef := make(map[string]string, len(root.Network.Devices.Device))

	for _, xd := range root.Network.Devices.Device {
		if xd.Engine == nil {
			continue
		}
		engine := xd.Engine
		dev := Device{
			Name:          strings.TrimSpace(engine.Name),
			Type:          strings.TrimSpace(engine.Type.Text),
			Model:         engine.Type.Model,
			Ref:           strings.TrimSpace(engine.SaveRefID),
			RunningConfig: configLines(engine.RunningConfig),
			StartupConfig: configLines(engine.StartupConfig),
		}
		if dev.Name == "" {
			dev.Name = "device"
		}
		doc.Devices = append(doc.Devices, dev)
		if id := refID(dev.Ref); id != "" {
			byRef[id] = dev.Name
		}
	}

	for _, xl := range root.Network.Links.Link {
		if xl.Cable == nil {
			continue
		}
		doc.Links = append(doc.Links, Link{
			LinkType: strings.TrimSpace(xl.Type),
			From:     endpoint(xl.Cable.From, xl.Cable.Ports, 0, byRef),
			To:       endpoint(xl.Cable.To, xl.Cable.Ports, 1, byRef),
		})
	}

	return doc, nil
}

func configLines(cfg *xmlConfig) []string {
	if cfg == nil {
		return nil
	}
	return cfg.Lines
}

func refID(ref string) string {
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

func endpoint(ref string, ports []string, index int, byRef map[string]string) Endpoint {
	endpoint := Endpoint{}
	if match := refPattern.FindStringSubmatch(ref); match != nil {
		id := match[1]
		if name, ok := byRef[id]; ok {
			endpoint.Name = name
		} else {
			endpoint.Name = fallbackName(id)
		}
	}
	if index < len(ports) {
		endpoint.Port = ports[index]
	}
	return endpoint
}

func fallbackName(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[len(id)-8:]
}
