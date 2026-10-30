package legacy

import (
	"context"
	"fmt"
	"strings"
)

// Reservation is a client's fixed-IP (DHCP reservation) record: the controller
// keeps it on the client document in rest/user. Field names come from the
// controller's api/fields/User.json.
type Reservation struct {
	ID                    string `json:"_id,omitempty"`
	MAC                   string `json:"mac"`
	Name                  string `json:"name"`
	Note                  string `json:"note"`
	UseFixedIP            bool   `json:"use_fixedip"`
	FixedIP               string `json:"fixed_ip"`
	NetworkID             string `json:"network_id"`
	LocalDNSRecordEnabled bool   `json:"local_dns_record_enabled"`
	LocalDNSRecord        string `json:"local_dns_record"`
}

// FindReservation returns the client document with the given MAC, or nil.
func (c *Client) FindReservation(ctx context.Context, siteID, mac string) (*Reservation, error) {
	var all []Reservation
	if err := c.List(ctx, siteID, "user", &all); err != nil {
		return nil, err
	}
	for i := range all {
		if strings.EqualFold(all[i].MAC, mac) {
			return &all[i], nil
		}
	}
	return nil, nil
}

// SaveReservation creates the client document, or updates it when the console
// already knows the MAC (POST would fail with api.err.MacUsed).
func (c *Client) SaveReservation(ctx context.Context, siteID string, r Reservation) (Reservation, error) {
	r.MAC = strings.ToLower(r.MAC)
	r.UseFixedIP = true
	existing, err := c.FindReservation(ctx, siteID, r.MAC)
	if err != nil {
		return Reservation{}, err
	}
	var out Reservation
	if existing == nil {
		err = c.Create(ctx, siteID, "user", r, &out)
	} else {
		err = c.Update(ctx, siteID, "user", existing.ID, r, &out)
	}
	return out, err
}

// GetReservation returns the active reservation for a MAC. A client record
// without use_fixedip is not a reservation.
func (c *Client) GetReservation(ctx context.Context, siteID, mac string) (*Reservation, error) {
	r, err := c.FindReservation(ctx, siteID, mac)
	if err != nil || r == nil || !r.UseFixedIP {
		return nil, err
	}
	return r, nil
}

// ReleaseReservation turns the fixed IP off but keeps the client record: the
// REST DELETE is not supported and "forget-sta" would also erase the history
// of a client the console has actually seen.
func (c *Client) ReleaseReservation(ctx context.Context, siteID, id string) error {
	var out Reservation
	err := c.Update(ctx, siteID, "user", id, map[string]any{
		"use_fixedip":              false,
		"local_dns_record_enabled": false,
	}, &out)
	if err != nil {
		return fmt.Errorf("releasing reservation %s: %w", id, err)
	}
	return nil
}
