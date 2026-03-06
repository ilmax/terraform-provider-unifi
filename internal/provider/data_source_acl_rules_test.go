package provider

import (
	"errors"
	"testing"

	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/acl_rules"
)

func TestCollectACLRulesPagesReadsAllPages(t *testing.T) {
	var offsets []int64

	rules, err := collectACLRulesPages(2, 10, func(offset, limit int64) (acl_rules.ListACLRulesResponse, error) {
		offsets = append(offsets, offset)
		switch offset {
		case 0:
			return acl_rules.ListACLRulesResponse{
				PaginatedResponse: acl_rules.PaginatedResponse{TotalCount: 3},
				Data: []acl_rules.ListACLRulesData{
					{Id: "r1", Name: "one", Enabled: true, Index: 1},
					{Id: "r2", Name: "two", Enabled: true, Index: 2},
				},
			}, nil
		case 2:
			return acl_rules.ListACLRulesResponse{
				PaginatedResponse: acl_rules.PaginatedResponse{TotalCount: 3},
				Data: []acl_rules.ListACLRulesData{
					{Id: "r3", Name: "three", Enabled: true, Index: 3},
				},
			}, nil
		default:
			return acl_rules.ListACLRulesResponse{}, nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 2 {
		t.Fatalf("unexpected offsets: %#v", offsets)
	}
}

func TestCollectACLRulesPagesPropagatesError(t *testing.T) {
	expectedErr := errors.New("boom")
	_, err := collectACLRulesPages(2, 10, func(offset, limit int64) (acl_rules.ListACLRulesResponse, error) {
		return acl_rules.ListACLRulesResponse{}, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestACLRulesFromAPISortsByIndex(t *testing.T) {
	items := []acl_rules.ListACLRulesData{
		{Id: "r2", Name: "second", Enabled: true, Index: 20},
		{Id: "r1", Name: "first", Enabled: true, Index: 10},
	}

	out := aclRulesFromAPI(items)
	if len(out) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(out))
	}
	if out[0].ID.ValueString() != "r1" || out[1].ID.ValueString() != "r2" {
		t.Fatalf("unexpected order: %s, %s", out[0].ID.ValueString(), out[1].ID.ValueString())
	}
}
