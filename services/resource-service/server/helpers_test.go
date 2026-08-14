package server

import (
	"testing"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
)

func strPtr(s string) *string { return &s }

func TestDerefString(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want string
	}{
		{"nil pointer", nil, ""},
		{"empty string", strPtr(""), ""},
		{"valid string", strPtr("test-asset"), "test-asset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := derefString(tt.s)
			if got != tt.want {
				t.Errorf("derefString(%v) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestToProtoStatus(t *testing.T) {
	tests := []struct {
		input string
		want  resourcev1.WorkUnitStatus
	}{
		{StatusAvailable, resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE},
		{StatusAllocated, resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED},
		{StatusInProduction, resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION},
		{StatusFaulted, resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED},
		{"unknown", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toProtoStatus(tt.input)
			if got != tt.want {
				t.Errorf("toProtoStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromProtoStatus(t *testing.T) {
	tests := []struct {
		input resourcev1.WorkUnitStatus
		want  string
	}{
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE, StatusAvailable},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED, StatusAllocated},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION, StatusInProduction},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED, StatusFaulted},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := fromProtoStatus(tt.input)
			if got != tt.want {
				t.Errorf("fromProtoStatus(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToProtoAssetStatus(t *testing.T) {
	tests := []struct {
		input string
		want  resourcev1.PhysicalAssetStatus
	}{
		{AssetStatusActive, resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE},
		{AssetStatusFaulted, resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED},
		{AssetStatusUnderMaintenance, resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE},
		{AssetStatusDecommissioned, resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED},
		{"unknown", resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toProtoAssetStatus(tt.input)
			if got != tt.want {
				t.Errorf("toProtoAssetStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromProtoAssetStatus(t *testing.T) {
	tests := []struct {
		input resourcev1.PhysicalAssetStatus
		want  string
	}{
		{resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE, AssetStatusActive},
		{resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED, AssetStatusFaulted},
		{resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE, AssetStatusUnderMaintenance},
		{resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED, AssetStatusDecommissioned},
		{resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := fromProtoAssetStatus(tt.input)
			if got != tt.want {
				t.Errorf("fromProtoAssetStatus(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToProtoWorkUnit_Nil(t *testing.T) {
	got := toProtoWorkUnit(nil)
	if got != nil {
		t.Errorf("toProtoWorkUnit(nil) = %v, want nil", got)
	}
}

func TestToProtoEquipmentClass_Nil(t *testing.T) {
	got := toProtoEquipmentClass(nil)
	if got != nil {
		t.Errorf("toProtoEquipmentClass(nil) = %v, want nil", got)
	}
}

func TestToProtoCapability_Nil(t *testing.T) {
	got := toProtoCapability(nil)
	if got != nil {
		t.Errorf("toProtoCapability(nil) = %v, want nil", got)
	}
}

func TestToProtoPhysicalAsset_Nil(t *testing.T) {
	got := toProtoPhysicalAsset(nil)
	if got != nil {
		t.Errorf("toProtoPhysicalAsset(nil) = %v, want nil", got)
	}
}

func TestToProtoInstallation_Nil(t *testing.T) {
	got := toProtoInstallation(nil)
	if got != nil {
		t.Errorf("toProtoInstallation(nil) = %v, want nil", got)
	}
}
