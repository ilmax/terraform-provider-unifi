package legacy

import (
	"context"
	"fmt"
	"strings"
)

// ColorEntry is one EtherLighting color: key is a link speed ("GbE", "10GbE",
// ...) for speed colors and a legacy network id for network colors.
type ColorEntry struct {
	Key         string `json:"key"`
	RawColorHex string `json:"raw_color_hex"`
}

// EtherLightingSettings is the per-site "ether_lighting" setting. Only the
// override lists are writable; the defaults are controller-owned (spec:
// api/fields/Setting.json).
type EtherLightingSettings struct {
	ID               string       `json:"_id,omitempty"`
	NetworkDefaults  []ColorEntry `json:"network_defaults"`
	NetworkOverrides []ColorEntry `json:"network_overrides"`
	SpeedDefaults    []ColorEntry `json:"speed_defaults"`
	SpeedOverrides   []ColorEntry `json:"speed_overrides"`
}

func (c *Client) GetEtherLightingSettings(ctx context.Context, siteID string) (EtherLightingSettings, error) {
	var s EtherLightingSettings
	err := c.GetSetting(ctx, siteID, "ether_lighting", &s)
	return s, err
}

// UpdateEtherLightingOverrides replaces both override lists.
func (c *Client) UpdateEtherLightingOverrides(ctx context.Context, siteID, id string, speed, network []ColorEntry) (EtherLightingSettings, error) {
	if speed == nil {
		speed = []ColorEntry{}
	}
	if network == nil {
		network = []ColorEntry{}
	}
	var out EtherLightingSettings
	err := c.UpdateSetting(ctx, siteID, "ether_lighting", id, map[string]any{
		"speed_overrides":   speed,
		"network_overrides": network,
	}, &out)
	return out, err
}

// DeviceEtherLighting is a device's own EtherLighting configuration (spec:
// api/fields/Device.json). The controller stores it even when the device does
// not advertise the capability.
type DeviceEtherLighting struct {
	LEDMode    string `json:"led_mode"`
	Mode       string `json:"mode"`
	Brightness int    `json:"brightness"`
	Behavior   string `json:"behavior"`
}

// Device is the part of a device document the EtherLighting resource needs.
type Device struct {
	ID            string               `json:"_id"`
	MAC           string               `json:"mac"`
	Model         string               `json:"model"`
	EtherLighting *DeviceEtherLighting `json:"ether_lighting"`
}

// FindDevice returns the adopted device with the given MAC, or nil.
func (c *Client) FindDevice(ctx context.Context, siteID, mac string) (*Device, error) {
	var all []Device
	if err := c.Stat(ctx, siteID, "device", &all); err != nil {
		return nil, err
	}
	for i := range all {
		if strings.EqualFold(all[i].MAC, mac) {
			return &all[i], nil
		}
	}
	return nil, nil
}

// SetDeviceEtherLighting writes the device's EtherLighting block.
func (c *Client) SetDeviceEtherLighting(ctx context.Context, siteID, deviceID string, el DeviceEtherLighting) (Device, error) {
	var out Device
	err := c.Update(ctx, siteID, "device", deviceID, map[string]any{"ether_lighting": el}, &out)
	if err != nil {
		return out, fmt.Errorf("setting ether_lighting on device %s: %w", deviceID, err)
	}
	return out, nil
}
