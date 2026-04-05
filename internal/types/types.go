// Package types defines the shared status data structures used across all
// cloud provider implementations and the worker output format.
package types

import "time"

// RegionStatus is the health state of a region or service.
type RegionStatus string

const (
	StatusOperational RegionStatus = "operational"
	StatusDegraded    RegionStatus = "degraded"
	StatusOutage      RegionStatus = "outage"
	StatusUnknown     RegionStatus = "unknown"
)

// ServiceStatus is the status of one service within a region.
type ServiceStatus struct {
	Name   string       `json:"name"`
	Status RegionStatus `json:"status"`
}

// RegionStatusData is the aggregated status for a single provider region.
type RegionStatusData struct {
	RegionID  string          `json:"region_id"`
	Name      string          `json:"name"`
	Lat       float64         `json:"lat"`
	Lon       float64         `json:"lon"`
	AZs       int             `json:"azs,omitempty"`
	Status    RegionStatus    `json:"status"`
	Services  []ServiceStatus `json:"services"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ProviderOutput is the status data for one cloud provider.
type ProviderOutput struct {
	Provider  string             `json:"provider"`
	UpdatedAt time.Time          `json:"updated_at"`
	Regions   []RegionStatusData `json:"regions"`
}

// StatusOutput is the top-level JSON written by the worker.
type StatusOutput struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Providers   []ProviderOutput `json:"providers"`
}
