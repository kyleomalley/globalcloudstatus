export type RegionStatus = 'operational' | 'degraded' | 'outage' | 'unknown';

export interface ServiceStatus {
  name: string;
  status: RegionStatus;
}

export interface RegionStatusData {
  region_id: string;
  name: string;
  lat: number;
  lon: number;
  azs: number;
  status: RegionStatus;
  services: ServiceStatus[];
  updated_at: string;
}

export interface ProviderOutput {
  provider: string;
  updated_at: string;
  regions: RegionStatusData[];
}

export interface StatusOutput {
  generated_at: string;
  providers: ProviderOutput[];
}
