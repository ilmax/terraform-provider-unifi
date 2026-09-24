// Package dns provides the typed DNS policy request/response shapes that
// github.com/ilmax/unifi-client-go generated automatically through v0.1.10
// by rewriting the UniFi Network API's "DNS policy" discriminated union into
// a oneOf schema. Starting with v0.1.11 the SDK switched to pulling its
// OpenAPI spec directly from UniFi's developer endpoint and dropped that
// rewrite, so the generated client only exposes a flattened DNSPolicy type
// with no record-specific fields. The wire format itself did not change, so
// these types (and the CreateOrUpdateDNSPolicy/DNSPolicy envelopes mirroring
// the SDK's former oneOf accessors) are reimplemented here by hand.
package dns

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Discriminator values for the DNS policy "type" field.
const (
	TypeARecord       = "A_RECORD"
	TypeAaaaRecord    = "AAAA_RECORD"
	TypeCnameRecord   = "CNAME_RECORD"
	TypeForwardDomain = "FORWARD_DOMAIN"
	TypeMxRecord      = "MX_RECORD"
	TypeSrvRecord     = "SRV_RECORD"
	TypeTxtRecord     = "TXT_RECORD"
)

// UserDefinedEntityMetadata defines model for "User defined entity metadata".
type UserDefinedEntityMetadata struct {
	Origin string `json:"origin"`
}

// DNSPolicyBase holds the fields common to every DNS policy record type, as
// returned by the API.
type DNSPolicyBase struct {
	Domain   *string                   `json:"domain,omitempty"`
	Enabled  bool                      `json:"enabled"`
	Id       uuid.UUID                 `json:"id"`
	Metadata UserDefinedEntityMetadata `json:"metadata"`
	Type     string                    `json:"type"`
}

// IntegrationDnsPolicyPageDto defines model for IntegrationDnsPolicyPageDto.
type IntegrationDnsPolicyPageDto struct {
	Count      int32       `json:"count"`
	Data       []DNSPolicy `json:"data"`
	Limit      int32       `json:"limit"`
	Offset     int64       `json:"offset"`
	TotalCount int64       `json:"totalCount"`
}

// IntegrationDnsARecordCreateUpdateDto defines model for IntegrationDnsARecordCreateUpdateDto.
type IntegrationDnsARecordCreateUpdateDto struct {
	Domain      *string `json:"domain,omitempty"`
	Enabled     bool    `json:"enabled"`
	Ipv4Address *string `json:"ipv4Address,omitempty"`
	TtlSeconds  *int32  `json:"ttlSeconds,omitempty"`
	Type        string  `json:"type"`
}

// IntegrationDnsARecordDto defines model for IntegrationDnsARecordDto.
type IntegrationDnsARecordDto struct {
	Domain      *string                   `json:"domain,omitempty"`
	Enabled     bool                      `json:"enabled"`
	Id          uuid.UUID                 `json:"id"`
	Ipv4Address *string                   `json:"ipv4Address,omitempty"`
	Metadata    UserDefinedEntityMetadata `json:"metadata"`
	TtlSeconds  *int32                    `json:"ttlSeconds,omitempty"`
	Type        string                    `json:"type"`
}

// IntegrationDnsAaaaRecordCreateUpdateDto defines model for IntegrationDnsAaaaRecordCreateUpdateDto.
type IntegrationDnsAaaaRecordCreateUpdateDto struct {
	Domain      *string `json:"domain,omitempty"`
	Enabled     bool    `json:"enabled"`
	Ipv6Address *string `json:"ipv6Address,omitempty"`
	TtlSeconds  *int32  `json:"ttlSeconds,omitempty"`
	Type        string  `json:"type"`
}

// IntegrationDnsAaaaRecordDto defines model for IntegrationDnsAaaaRecordDto.
type IntegrationDnsAaaaRecordDto struct {
	Domain      *string                   `json:"domain,omitempty"`
	Enabled     bool                      `json:"enabled"`
	Id          uuid.UUID                 `json:"id"`
	Ipv6Address *string                   `json:"ipv6Address,omitempty"`
	Metadata    UserDefinedEntityMetadata `json:"metadata"`
	TtlSeconds  *int32                    `json:"ttlSeconds,omitempty"`
	Type        string                    `json:"type"`
}

// IntegrationDnsCnameRecordCreateUpdateDto defines model for IntegrationDnsCnameRecordCreateUpdateDto.
type IntegrationDnsCnameRecordCreateUpdateDto struct {
	Domain       *string `json:"domain,omitempty"`
	Enabled      bool    `json:"enabled"`
	TargetDomain *string `json:"targetDomain,omitempty"`
	TtlSeconds   *int32  `json:"ttlSeconds,omitempty"`
	Type         string  `json:"type"`
}

// IntegrationDnsCnameRecordDto defines model for IntegrationDnsCnameRecordDto.
type IntegrationDnsCnameRecordDto struct {
	Domain       *string                   `json:"domain,omitempty"`
	Enabled      bool                      `json:"enabled"`
	Id           uuid.UUID                 `json:"id"`
	Metadata     UserDefinedEntityMetadata `json:"metadata"`
	TargetDomain *string                   `json:"targetDomain,omitempty"`
	TtlSeconds   *int32                    `json:"ttlSeconds,omitempty"`
	Type         string                    `json:"type"`
}

// IntegrationDnsForwardDomainPolicyCreateUpdateDto defines model for IntegrationDnsForwardDomainPolicyCreateUpdateDto.
type IntegrationDnsForwardDomainPolicyCreateUpdateDto struct {
	Domain    *string `json:"domain,omitempty"`
	Enabled   bool    `json:"enabled"`
	IpAddress *string `json:"ipAddress,omitempty"`
	Type      string  `json:"type"`
}

// IntegrationDnsForwardDomainPolicyDto defines model for IntegrationDnsForwardDomainPolicyDto.
type IntegrationDnsForwardDomainPolicyDto struct {
	Domain    *string                   `json:"domain,omitempty"`
	Enabled   bool                      `json:"enabled"`
	Id        uuid.UUID                 `json:"id"`
	IpAddress *string                   `json:"ipAddress,omitempty"`
	Metadata  UserDefinedEntityMetadata `json:"metadata"`
	Type      string                    `json:"type"`
}

// IntegrationDnsMxRecordCreateUpdateDto defines model for IntegrationDnsMxRecordCreateUpdateDto.
type IntegrationDnsMxRecordCreateUpdateDto struct {
	Domain           *string `json:"domain,omitempty"`
	Enabled          bool    `json:"enabled"`
	MailServerDomain *string `json:"mailServerDomain,omitempty"`
	Priority         *int32  `json:"priority,omitempty"`
	Type             string  `json:"type"`
}

// IntegrationDnsMxRecordDto defines model for IntegrationDnsMxRecordDto.
type IntegrationDnsMxRecordDto struct {
	Domain           *string                   `json:"domain,omitempty"`
	Enabled          bool                      `json:"enabled"`
	Id               uuid.UUID                 `json:"id"`
	MailServerDomain *string                   `json:"mailServerDomain,omitempty"`
	Metadata         UserDefinedEntityMetadata `json:"metadata"`
	Priority         *int32                    `json:"priority,omitempty"`
	Type             string                    `json:"type"`
}

// IntegrationDnsSrvRecordCreateUpdateDto defines model for IntegrationDnsSrvRecordCreateUpdateDto.
type IntegrationDnsSrvRecordCreateUpdateDto struct {
	Domain       *string `json:"domain,omitempty"`
	Enabled      bool    `json:"enabled"`
	Port         *int32  `json:"port,omitempty"`
	Priority     *int32  `json:"priority,omitempty"`
	Protocol     *string `json:"protocol,omitempty"`
	ServerDomain *string `json:"serverDomain,omitempty"`
	Service      *string `json:"service,omitempty"`
	Type         string  `json:"type"`
	Weight       *int32  `json:"weight,omitempty"`
}

// IntegrationDnsSrvRecordDto defines model for IntegrationDnsSrvRecordDto.
type IntegrationDnsSrvRecordDto struct {
	Domain       *string                   `json:"domain,omitempty"`
	Enabled      bool                      `json:"enabled"`
	Id           uuid.UUID                 `json:"id"`
	Metadata     UserDefinedEntityMetadata `json:"metadata"`
	Port         *int32                    `json:"port,omitempty"`
	Priority     *int32                    `json:"priority,omitempty"`
	Protocol     *string                   `json:"protocol,omitempty"`
	ServerDomain *string                   `json:"serverDomain,omitempty"`
	Service      *string                   `json:"service,omitempty"`
	Type         string                    `json:"type"`
	Weight       *int32                    `json:"weight,omitempty"`
}

// IntegrationDnsTxtRecordCreateUpdateDto defines model for IntegrationDnsTxtRecordCreateUpdateDto.
type IntegrationDnsTxtRecordCreateUpdateDto struct {
	Domain  *string `json:"domain,omitempty"`
	Enabled bool    `json:"enabled"`
	Text    *string `json:"text,omitempty"`
	Type    string  `json:"type"`
}

// IntegrationDnsTxtRecordDto defines model for IntegrationDnsTxtRecordDto.
type IntegrationDnsTxtRecordDto struct {
	Domain   *string                   `json:"domain,omitempty"`
	Enabled  bool                      `json:"enabled"`
	Id       uuid.UUID                 `json:"id"`
	Metadata UserDefinedEntityMetadata `json:"metadata"`
	Text     *string                   `json:"text,omitempty"`
	Type     string                    `json:"type"`
}

// CreateOrUpdateDNSPolicy is a oneOf-style envelope for the request body of
// the create/update DNS policy endpoints, holding whichever concrete record
// type was set via one of its FromX methods.
type CreateOrUpdateDNSPolicy struct {
	union json.RawMessage
}

func (t CreateOrUpdateDNSPolicy) MarshalJSON() ([]byte, error) {
	if t.union == nil {
		return []byte("null"), nil
	}
	return t.union, nil
}

func (t *CreateOrUpdateDNSPolicy) UnmarshalJSON(b []byte) error {
	t.union = append(json.RawMessage(nil), b...)
	return nil
}

func (t CreateOrUpdateDNSPolicy) Discriminator() (string, error) {
	var d struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(t.union, &d)
	return d.Type, err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsARecordCreateUpdateDto() (IntegrationDnsARecordCreateUpdateDto, error) {
	var body IntegrationDnsARecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsARecordCreateUpdateDto(v IntegrationDnsARecordCreateUpdateDto) error {
	v.Type = TypeARecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsAaaaRecordCreateUpdateDto() (IntegrationDnsAaaaRecordCreateUpdateDto, error) {
	var body IntegrationDnsAaaaRecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsAaaaRecordCreateUpdateDto(v IntegrationDnsAaaaRecordCreateUpdateDto) error {
	v.Type = TypeAaaaRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsCnameRecordCreateUpdateDto() (IntegrationDnsCnameRecordCreateUpdateDto, error) {
	var body IntegrationDnsCnameRecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsCnameRecordCreateUpdateDto(v IntegrationDnsCnameRecordCreateUpdateDto) error {
	v.Type = TypeCnameRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsForwardDomainPolicyCreateUpdateDto() (IntegrationDnsForwardDomainPolicyCreateUpdateDto, error) {
	var body IntegrationDnsForwardDomainPolicyCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsForwardDomainPolicyCreateUpdateDto(v IntegrationDnsForwardDomainPolicyCreateUpdateDto) error {
	v.Type = TypeForwardDomain
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsMxRecordCreateUpdateDto() (IntegrationDnsMxRecordCreateUpdateDto, error) {
	var body IntegrationDnsMxRecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsMxRecordCreateUpdateDto(v IntegrationDnsMxRecordCreateUpdateDto) error {
	v.Type = TypeMxRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsSrvRecordCreateUpdateDto() (IntegrationDnsSrvRecordCreateUpdateDto, error) {
	var body IntegrationDnsSrvRecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsSrvRecordCreateUpdateDto(v IntegrationDnsSrvRecordCreateUpdateDto) error {
	v.Type = TypeSrvRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t CreateOrUpdateDNSPolicy) AsIntegrationDnsTxtRecordCreateUpdateDto() (IntegrationDnsTxtRecordCreateUpdateDto, error) {
	var body IntegrationDnsTxtRecordCreateUpdateDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *CreateOrUpdateDNSPolicy) FromIntegrationDnsTxtRecordCreateUpdateDto(v IntegrationDnsTxtRecordCreateUpdateDto) error {
	v.Type = TypeTxtRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// DNSPolicy is a oneOf-style envelope for the response body of the DNS
// policy endpoints, holding whichever concrete record type matches the
// "type" discriminator.
type DNSPolicy struct {
	union json.RawMessage
}

func (t DNSPolicy) MarshalJSON() ([]byte, error) {
	if t.union == nil {
		return []byte("null"), nil
	}
	return t.union, nil
}

func (t *DNSPolicy) UnmarshalJSON(b []byte) error {
	t.union = append(json.RawMessage(nil), b...)
	return nil
}

func (t DNSPolicy) Discriminator() (string, error) {
	var d struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(t.union, &d)
	return d.Type, err
}

func (t DNSPolicy) AsIntegrationDnsARecordDto() (IntegrationDnsARecordDto, error) {
	var body IntegrationDnsARecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsARecordDto(v IntegrationDnsARecordDto) error {
	v.Type = TypeARecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsAaaaRecordDto() (IntegrationDnsAaaaRecordDto, error) {
	var body IntegrationDnsAaaaRecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsAaaaRecordDto(v IntegrationDnsAaaaRecordDto) error {
	v.Type = TypeAaaaRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsCnameRecordDto() (IntegrationDnsCnameRecordDto, error) {
	var body IntegrationDnsCnameRecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsCnameRecordDto(v IntegrationDnsCnameRecordDto) error {
	v.Type = TypeCnameRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsForwardDomainPolicyDto() (IntegrationDnsForwardDomainPolicyDto, error) {
	var body IntegrationDnsForwardDomainPolicyDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsForwardDomainPolicyDto(v IntegrationDnsForwardDomainPolicyDto) error {
	v.Type = TypeForwardDomain
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsMxRecordDto() (IntegrationDnsMxRecordDto, error) {
	var body IntegrationDnsMxRecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsMxRecordDto(v IntegrationDnsMxRecordDto) error {
	v.Type = TypeMxRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsSrvRecordDto() (IntegrationDnsSrvRecordDto, error) {
	var body IntegrationDnsSrvRecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsSrvRecordDto(v IntegrationDnsSrvRecordDto) error {
	v.Type = TypeSrvRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DNSPolicy) AsIntegrationDnsTxtRecordDto() (IntegrationDnsTxtRecordDto, error) {
	var body IntegrationDnsTxtRecordDto
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DNSPolicy) FromIntegrationDnsTxtRecordDto(v IntegrationDnsTxtRecordDto) error {
	v.Type = TypeTxtRecord
	b, err := json.Marshal(v)
	t.union = b
	return err
}
