import './style.css';
import { WorldMap } from './map';
import { StatusOutput } from './types';

const STATUS_URL  = '/data/status.json';
const REFRESH_MS  = 5 * 60 * 1000; // 5 minutes

async function fetchStatus(): Promise<StatusOutput> {
  const res = await fetch(`${STATUS_URL}?_=${Date.now()}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

function setLastUpdated(iso: string): void {
  const el = document.getElementById('last-updated');
  if (el) el.textContent = new Date(iso).toLocaleString();
}

function setLoadingState(loading: boolean): void {
  const el = document.getElementById('loading-overlay');
  if (el) el.style.display = loading ? 'flex' : 'none';
}

async function main(): Promise<void> {
  const map = new WorldMap('world-map', 'tooltip');
  setLoadingState(true);

  await map.initialize();

  async function refresh(): Promise<void> {
    try {
      const data = await fetchStatus();
      map.updateStatus(data);
      setLastUpdated(data.generated_at);
    } catch (err) {
      console.warn('Status fetch failed:', err);
    } finally {
      setLoadingState(false);
    }
  }

  await refresh();
  setInterval(refresh, REFRESH_MS);
}

main().catch(console.error);
