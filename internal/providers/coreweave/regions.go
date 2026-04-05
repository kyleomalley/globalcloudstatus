package coreweave

// Region maps a Status.io container name to geographic coordinates.
// The ID must match the "name" field returned by the Status.io API exactly.
type Region struct {
	ID   string
	Name string
	Lat  float64
	Lon  float64
}

// Regions is the list of all known CoreWeave locations.
// US-EAST locations are in the Northern Virginia / DC metro and will cluster.
// US-WEST locations are in the Pacific Northwest (Portland/Hillsboro, OR area).
// EU-SOUTH is Amsterdam, EU-NORTH is Helsinki/Stockholm.
var Regions = []Region{
	// Legacy named DCs
	{ID: "US-ORD1", Name: "US Central (Chicago)", Lat: 41.88, Lon: -87.63},
	{ID: "US-LAS1", Name: "US West (Las Vegas)", Lat: 36.17, Lon: -115.14},
	{ID: "US-LGA1", Name: "US East (New Jersey)", Lat: 40.77, Lon: -74.05},
	{ID: "US-RNO2", Name: "US West (Reno)", Lat: 39.53, Lon: -119.81},
	{ID: "RDU1", Name: "US East (Research Triangle)", Lat: 35.78, Lon: -78.64},

	// US East — all in the Northern Virginia / DC metro; same anchor, cluster handles layout
	{ID: "US-EAST-01", Name: "US East 01 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-02", Name: "US East 02 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-03", Name: "US East 03 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-04", Name: "US East 04 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-06", Name: "US East 06 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-08", Name: "US East 08 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-13", Name: "US East 13 (Virginia)", Lat: 38.95, Lon: -77.45},
	{ID: "US-EAST-14", Name: "US East 14 (Virginia)", Lat: 38.95, Lon: -77.45},

	// US West — Pacific Northwest (Portland / Hillsboro, OR area)
	{ID: "US-WEST-01", Name: "US West 01 (Pacific NW)", Lat: 45.52, Lon: -122.68},
	{ID: "US-WEST-03", Name: "US West 03 (Pacific NW)", Lat: 45.52, Lon: -122.68},
	{ID: "US-WEST-04", Name: "US West 04 (Pacific NW)", Lat: 45.52, Lon: -122.68},
	{ID: "US-WEST-07", Name: "US West 07 (Pacific NW)", Lat: 45.52, Lon: -122.68},
	{ID: "US-WEST-09", Name: "US West 09 (Pacific NW)", Lat: 45.52, Lon: -122.68},

	// US Central — Dallas / Fort Worth area
	{ID: "US-CENTRAL-02", Name: "US Central 02 (Dallas)", Lat: 32.90, Lon: -97.04},
	{ID: "US-CENTRAL-07", Name: "US Central 07 (Dallas)", Lat: 32.90, Lon: -97.04},

	// US Lab
	{ID: "US-LAB-01", Name: "US Lab 01", Lat: 40.72, Lon: -74.00},

	// Europe
	{ID: "EU-SOUTH-03", Name: "EU South 03 (Amsterdam)", Lat: 52.37, Lon: 4.89},
	{ID: "EU-SOUTH-04", Name: "EU South 04 (Amsterdam)", Lat: 52.30, Lon: 4.82},
	{ID: "EU-NORTH-02", Name: "EU North 02 (Helsinki)", Lat: 60.17, Lon: 24.94},

	// Canada
	{ID: "CA-EAST-01", Name: "Canada East (Montreal)", Lat: 45.50, Lon: -73.57},
}
