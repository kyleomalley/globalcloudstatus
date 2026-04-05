package coreweave

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/globalcloudstatus/globalcloudstatus/internal/types"
)

const statusAPIURL = "https://api.status.io/1.0/status/5e126e998f2f032e1f8f0f4b"

// Status.io status codes.
const (
	codeOperational = 100
	codeMaintenance = 200 // planned, but still impacts availability
	codeDegraded    = 300
	codePartial     = 400
	codeDisruption  = 500
	codeSecurity    = 600
)

var httpClient = &http.Client{Timeout: 12 * time.Second}

// statusIOResponse is the top-level shape of the Status.io v1 API.
type statusIOResponse struct {
	Result struct {
		Status []struct {
			Name       string `json:"name"`
			Containers []struct {
				Name       string `json:"name"`
				Status     string `json:"status"`
				StatusCode int    `json:"status_code"`
			} `json:"containers"`
		} `json:"status"`
	} `json:"result"`
}

// FetchAllRegions calls the Status.io API and returns a status entry for
// every region defined in regions.go. On error all regions are unknown.
func FetchAllRegions() []types.RegionStatusData {
	resp, err := httpClient.Get(statusAPIURL)
	if err != nil {
		log.Printf("coreweave: fetch error: %v", err)
		return unknownAll()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("coreweave: HTTP %d from status API", resp.StatusCode)
		return unknownAll()
	}

	var apiResp statusIOResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("coreweave: decode error: %v", err)
		return unknownAll()
	}

	// Aggregate across all services: worst status code + affected service names.
	type locationImpact struct {
		worstCode int
		services  map[string]struct{}
	}
	impacts := make(map[string]*locationImpact)

	for _, svc := range apiResp.Result.Status {
		for _, c := range svc.Containers {
			li := impacts[c.Name]
			if li == nil {
				li = &locationImpact{services: make(map[string]struct{})}
				impacts[c.Name] = li
			}
			if c.StatusCode > li.worstCode {
				li.worstCode = c.StatusCode
			}
			if c.StatusCode != codeOperational {
				li.services[fmt.Sprintf("%s (%s)", svc.Name, c.Status)] = struct{}{}
			}
		}
	}

	now := time.Now().UTC()
	results := make([]types.RegionStatusData, 0, len(Regions))

	for _, r := range Regions {
		li := impacts[r.ID]
		status := types.StatusOperational
		var services []types.ServiceStatus

		if li != nil && li.worstCode > codeOperational {
			status = codeToStatus(li.worstCode)
			for svcLabel := range li.services {
				services = append(services, types.ServiceStatus{
					Name:   svcLabel,
					Status: status,
				})
			}
		}

		results = append(results, types.RegionStatusData{
			RegionID:  r.ID,
			Name:      r.Name,
			Lat:       r.Lat,
			Lon:       r.Lon,
			Status:    status,
			Services:  services,
			UpdatedAt: now,
		})
	}

	return results
}

func codeToStatus(code int) types.RegionStatus {
	switch {
	case code == codeOperational:
		return types.StatusOperational
	case code <= codePartial: // 200 maintenance, 300 degraded, 400 partial
		return types.StatusDegraded
	default: // 500 disruption, 600 security
		return types.StatusOutage
	}
}

func unknownAll() []types.RegionStatusData {
	now := time.Now().UTC()
	results := make([]types.RegionStatusData, len(Regions))
	for i, r := range Regions {
		results[i] = types.RegionStatusData{
			RegionID:  r.ID,
			Name:      r.Name,
			Lat:       r.Lat,
			Lon:       r.Lon,
			Status:    types.StatusUnknown,
			UpdatedAt: now,
		}
	}
	return results
}
