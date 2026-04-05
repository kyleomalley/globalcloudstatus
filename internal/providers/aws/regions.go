package aws

// Region represents an AWS region with geographic coordinates.
type Region struct {
	ID   string
	Name string
	Lat  float64
	Lon  float64
	AZs  int
}

// Regions is the list of all active AWS commercial regions.
var Regions = []Region{
	{ID: "us-east-1", Name: "US East (N. Virginia)", Lat: 38.9, Lon: -77.0, AZs: 6},
	{ID: "us-east-2", Name: "US East (Ohio)", Lat: 40.0, Lon: -82.8, AZs: 3},
	{ID: "us-west-1", Name: "US West (N. California)", Lat: 37.3, Lon: -122.0, AZs: 3},
	{ID: "us-west-2", Name: "US West (Oregon)", Lat: 45.8, Lon: -119.7, AZs: 4},
	{ID: "ca-central-1", Name: "Canada (Central)", Lat: 45.5, Lon: -73.6, AZs: 3},
	{ID: "ca-west-1", Name: "Canada West (Calgary)", Lat: 51.0, Lon: -114.1, AZs: 3},
	{ID: "eu-west-1", Name: "Europe (Ireland)", Lat: 53.3, Lon: -6.3, AZs: 3},
	{ID: "eu-west-2", Name: "Europe (London)", Lat: 51.5, Lon: -0.1, AZs: 3},
	{ID: "eu-west-3", Name: "Europe (Paris)", Lat: 48.9, Lon: 2.3, AZs: 3},
	{ID: "eu-central-1", Name: "Europe (Frankfurt)", Lat: 50.1, Lon: 8.7, AZs: 3},
	{ID: "eu-central-2", Name: "Europe (Zurich)", Lat: 47.4, Lon: 8.5, AZs: 3},
	{ID: "eu-north-1", Name: "Europe (Stockholm)", Lat: 59.3, Lon: 18.1, AZs: 3},
	{ID: "eu-south-1", Name: "Europe (Milan)", Lat: 45.5, Lon: 9.2, AZs: 3},
	{ID: "eu-south-2", Name: "Europe (Spain)", Lat: 40.4, Lon: -3.7, AZs: 3},
	{ID: "ap-northeast-1", Name: "Asia Pacific (Tokyo)", Lat: 35.7, Lon: 139.7, AZs: 4},
	{ID: "ap-northeast-2", Name: "Asia Pacific (Seoul)", Lat: 37.6, Lon: 126.9, AZs: 4},
	{ID: "ap-northeast-3", Name: "Asia Pacific (Osaka)", Lat: 34.7, Lon: 135.5, AZs: 3},
	{ID: "ap-southeast-1", Name: "Asia Pacific (Singapore)", Lat: 1.3, Lon: 103.8, AZs: 3},
	{ID: "ap-southeast-2", Name: "Asia Pacific (Sydney)", Lat: -33.9, Lon: 151.2, AZs: 3},
	{ID: "ap-southeast-3", Name: "Asia Pacific (Jakarta)", Lat: -6.2, Lon: 106.8, AZs: 3},
	{ID: "ap-southeast-4", Name: "Asia Pacific (Melbourne)", Lat: -37.8, Lon: 145.0, AZs: 3},
	{ID: "ap-southeast-5", Name: "Asia Pacific (Malaysia)", Lat: 3.1, Lon: 101.7, AZs: 3},
	{ID: "ap-southeast-7", Name: "Asia Pacific (Thailand)", Lat: 13.8, Lon: 100.5, AZs: 3},
	{ID: "ap-south-1", Name: "Asia Pacific (Mumbai)", Lat: 19.1, Lon: 72.9, AZs: 3},
	{ID: "ap-south-2", Name: "Asia Pacific (Hyderabad)", Lat: 17.4, Lon: 78.5, AZs: 3},
	{ID: "ap-east-1", Name: "Asia Pacific (Hong Kong)", Lat: 22.3, Lon: 114.2, AZs: 3},
	{ID: "sa-east-1", Name: "South America (São Paulo)", Lat: -23.5, Lon: -46.6, AZs: 3},
	{ID: "me-south-1", Name: "Middle East (Bahrain)", Lat: 26.2, Lon: 50.6, AZs: 3},
	{ID: "me-central-1", Name: "Middle East (UAE)", Lat: 24.5, Lon: 54.4, AZs: 3},
	{ID: "af-south-1", Name: "Africa (Cape Town)", Lat: -33.9, Lon: 18.4, AZs: 3},
	{ID: "il-central-1", Name: "Israel (Tel Aviv)", Lat: 32.1, Lon: 34.8, AZs: 3},
	{ID: "mx-central-1", Name: "Mexico (Central)", Lat: 20.6, Lon: -103.4, AZs: 3},
}
