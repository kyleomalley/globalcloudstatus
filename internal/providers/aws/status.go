package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/globalcloudstatus/globalcloudstatus/internal/types"
)

const (
	currentEventsURL  = "https://health.aws.amazon.com/public/currentevents"
	serviceCatalogURL = "https://servicedata-us-west-2-prod.s3.amazonaws.com/services.json"
)

// Event status codes returned by the currentevents API.
const (
	eventInvestigating = 1 // AWS is looking into it
	eventMonitoring    = 2 // Issue persists, being watched (resolving)
	eventInProgress    = 3 // Active / ongoing outage
)

// currentEvent is one entry from /public/currentevents.
// Note: the API returns numeric fields (date, status) as JSON strings.
type currentEvent struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	ServiceName string `json:"service_name"`
	RegionName  string `json:"region_name"`
	Summary     string `json:"summary"`
	// ImpactedServices maps service+region keys (e.g. "rds-me-south-1") to AZ impact data.
	ImpactedServices map[string]struct {
		ServiceName string `json:"service_name"`
		Current     string `json:"current"`
		Max         string `json:"max"`
	} `json:"impacted_services"`
}

// catalogEntry is one row from the service catalog S3 file.
type catalogEntry struct {
	Service     string `json:"service"`      // composite key: "{service-slug}-{region-id}"
	ServiceName string `json:"service_name"`
	RegionID    string `json:"region_id"`   // AWS region ID, empty for global services
	RegionName  string `json:"region_name"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func getJSON(url string, dest any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	body, err = normalizeEncoding(body)
	if err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}

	return json.Unmarshal(body, dest)
}

// normalizeEncoding converts UTF-16 BE/LE (with BOM) to UTF-8, or strips a
// UTF-8 BOM. Returns the input unchanged if none of those markers are present.
func normalizeEncoding(b []byte) ([]byte, error) {
	if len(b) < 2 {
		return b, nil
	}

	var bigEndian bool
	switch {
	case b[0] == 0xFE && b[1] == 0xFF:
		bigEndian = true
		b = b[2:]
	case b[0] == 0xFF && b[1] == 0xFE:
		bigEndian = false
		b = b[2:]
	default:
		// Strip UTF-8 BOM if present.
		return bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF}), nil
	}

	if len(b)%2 != 0 {
		return nil, fmt.Errorf("odd byte count in UTF-16 payload")
	}

	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		if bigEndian {
			u16[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
		} else {
			u16[i] = uint16(b[2*i+1])<<8 | uint16(b[2*i])
		}
	}

	runes := utf16.Decode(u16)
	out := make([]byte, 0, len(runes)*3)
	var buf [utf8.UTFMax]byte
	for _, r := range runes {
		n := utf8.EncodeRune(buf[:], r)
		out = append(out, buf[:n]...)
	}
	return out, nil
}

// FetchAllRegions fetches active AWS health events and returns a status entry
// for every region defined in regions.go. Unaffected regions are operational.
func FetchAllRegions() []types.RegionStatusData {
	var (
		events  []currentEvent
		catalog []catalogEntry
		wg      sync.WaitGroup
		evErr   error
		catErr  error
	)

	wg.Add(2)
	go func() { defer wg.Done(); evErr = getJSON(currentEventsURL, &events) }()
	go func() { defer wg.Done(); catErr = getJSON(serviceCatalogURL, &catalog) }()
	wg.Wait()

	if evErr != nil {
		log.Printf("warn: currentevents: %v", evErr)
	}
	if catErr != nil {
		log.Printf("warn: service catalog: %v", catErr)
	}

	// Build service-key → region_id lookup from the catalog.
	keyToRegion := make(map[string]string, len(catalog))
	for _, e := range catalog {
		if e.RegionID != "" {
			keyToRegion[e.Service] = e.RegionID
		}
	}

	// Aggregate impacts per region: worst numeric status + affected service names.
	type regionImpact struct {
		worstStatus int
		services    map[string]struct{} // deduplicated service names
	}
	impacted := make(map[string]*regionImpact)

	recordImpact := func(regionID string, eventStatus int, serviceName string) {
		ri := impacted[regionID]
		if ri == nil {
			ri = &regionImpact{services: make(map[string]struct{})}
			impacted[regionID] = ri
		}
		if eventStatus > ri.worstStatus {
			ri.worstStatus = eventStatus
		}
		if serviceName != "" {
			ri.services[serviceName] = struct{}{}
		}
	}

	for _, ev := range events {
		statusCode := parseStatus(ev.Status)
		// Prefer the granular impacted_services map when available.
		if len(ev.ImpactedServices) > 0 {
			for serviceKey, svc := range ev.ImpactedServices {
				if regionID, ok := keyToRegion[serviceKey]; ok {
					recordImpact(regionID, statusCode, svc.ServiceName)
				}
			}
		} else if ev.Service != "" {
			// Fall back to the top-level service key.
			if regionID, ok := keyToRegion[ev.Service]; ok {
				recordImpact(regionID, statusCode, ev.ServiceName)
			}
		}
	}

	now := time.Now().UTC()
	results := make([]types.RegionStatusData, len(Regions))

	for i, r := range Regions {
		ri := impacted[r.ID]
		status := types.StatusOperational

		var services []types.ServiceStatus
		if ri != nil {
			status = toRegionStatus(ri.worstStatus)
			for svcName := range ri.services {
				services = append(services, types.ServiceStatus{
					Name:   svcName,
					Status: status,
				})
			}
		}

		results[i] = types.RegionStatusData{
			RegionID:  r.ID,
			Name:      r.Name,
			Lat:       r.Lat,
			Lon:       r.Lon,
			AZs:       r.AZs,
			Status:    status,
			Services:  services,
			UpdatedAt: now,
		}
	}

	return results
}

func parseStatus(s string) int {
	switch s {
	case "1":
		return eventInvestigating
	case "2":
		return eventMonitoring
	case "3":
		return eventInProgress
	default:
		return 0
	}
}

func toRegionStatus(eventStatus int) types.RegionStatus {
	switch eventStatus {
	case eventInvestigating, eventMonitoring:
		return types.StatusDegraded
	case eventInProgress:
		return types.StatusOutage
	default:
		return types.StatusUnknown
	}
}
