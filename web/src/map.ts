import * as d3 from 'd3';
import * as topojson from 'topojson-client';
import type { Topology, Objects } from 'topojson-specification';
import { RegionStatus, RegionStatusData, StatusOutput } from './types';

const MARKER_SIZE = 12;    // px — 20% smaller than original 15px
const MARKER_GAP  = 1;     // px between markers in a cluster row
const CLUSTER_RADIUS = 32; // px — larger radius collapses same-metro DCs across providers

// Stroke color per status (wireframe style)
const STATUS_STROKE: Record<RegionStatus, string> = {
  operational: '#22c55e',
  degraded:    '#f59e0b',
  outage:      '#ef4444',
  unknown:     '#555e6e',
};

// Very faint fill so the rect interior still captures pointer events
const STATUS_FILL: Record<RegionStatus, string> = {
  operational: 'rgba(34,197,94,0.08)',
  degraded:    'rgba(245,158,11,0.08)',
  outage:      'rgba(239,68,68,0.12)',
  unknown:     'rgba(107,114,128,0.05)',
};

const PROVIDER_NAMES: Record<string, string> = {
  A: 'Amazon Web Services',
  M: 'Microsoft Azure',
  G: 'Google Cloud',
  O: 'Oracle Cloud',
  C: 'CoreWeave',
};

const PROVIDER_BADGE_COLOR: Record<string, string> = {
  A: '#FF9900',
  M: '#0078D4',
  G: '#4285F4',
  O: '#F80000',
  C: '#7B68EE',
};

const PROVIDER_LETTERS: Record<string, string> = {
  aws:       'A',
  azure:     'M',
  gcp:       'G',
  oracle:    'O',
  coreweave: 'C',
};

interface PlacedMarker {
  region: RegionStatusData;
  letter: string;
  x: number;  // center x
  y: number;  // center y
}

export class WorldMap {
  private svg: d3.Selection<SVGSVGElement, unknown, HTMLElement, unknown>;
  private tooltip: d3.Selection<HTMLDivElement, unknown, HTMLElement, unknown>;
  private markersGroup: d3.Selection<SVGGElement, unknown, HTMLElement, unknown> | null = null;
  private projection!: d3.GeoProjection;
  private pendingStatus: StatusOutput | null = null;
  private lastStatus: StatusOutput | null = null;
  private width = 0;
  private height = 0;

  constructor(svgId: string, tooltipId: string) {
    this.svg = d3.select<SVGSVGElement, unknown>(`#${svgId}`);
    this.tooltip = d3.select<HTMLDivElement, unknown>(`#${tooltipId}`);
  }

  async initialize(): Promise<void> {
    this.width = this.svg.node()!.parentElement!.clientWidth;
    this.height = Math.round(this.width * 0.508); // Natural Earth approx aspect

    this.svg.attr('viewBox', `0 0 ${this.width} ${this.height}`)
      .attr('width', '100%');

    this.projection = d3.geoNaturalEarth1()
      .rotate([30, 0])  // center on ~30°W (mid-Atlantic)
      .fitSize([this.width, this.height], { type: 'Sphere' });

    const path = d3.geoPath(this.projection);

    // Ocean
    this.svg.append('rect')
      .attr('width', this.width)
      .attr('height', this.height)
      .attr('fill', '#0a1628');

    // Graticule
    this.svg.append('path')
      .datum(d3.geoGraticule()())
      .attr('d', path)
      .attr('stroke', '#111e38')
      .attr('stroke-width', 0.5)
      .attr('fill', 'none');

    // Load world atlas
    const world = await fetch(
      'https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json'
    ).then(r => r.json()) as Topology<Objects>;

    const countries = topojson.feature(world, world.objects['countries'] as any);
    const borders   = topojson.mesh(world, world.objects['countries'] as any, (a, b) => a !== b);

    this.svg.append('g')
      .selectAll<SVGPathElement, unknown>('path')
      .data((countries as any).features)
      .join('path')
      .attr('d', path as any)
      .attr('fill', '#1a2744');

    this.svg.append('path')
      .datum(borders)
      .attr('d', path as any)
      .attr('stroke', '#2a3d6e')
      .attr('stroke-width', 0.5)
      .attr('fill', 'none');

    this.markersGroup = this.svg.append('g').attr('class', 'markers');

    if (this.pendingStatus) {
      this.updateStatus(this.pendingStatus);
      this.pendingStatus = null;
    }

    window.addEventListener('resize', () => this.onResize());
  }

  updateStatus(data: StatusOutput): void {
    if (!this.markersGroup) {
      this.pendingStatus = data;
      return;
    }

    // Collect regions from all providers, tagging each with its display letter.
    const allItems: Array<{ region: RegionStatusData; letter: string }> = [];
    for (const provider of data.providers) {
      const letter = PROVIDER_LETTERS[provider.provider]
        ?? provider.provider[0].toUpperCase();
      for (const region of provider.regions) {
        allItems.push({ region, letter });
      }
    }

    this.lastStatus = data;
    const placed = this.placeMarkers(allItems);
    this.renderMarkers(placed);
  }

  // placeMarkers projects regions to screen coords and clusters nearby ones into rows.
  private placeMarkers(items: Array<{ region: RegionStatusData; letter: string }>): PlacedMarker[] {
    // Project every region to [x, y].
    const pts = items.map(({ region, letter }) => {
      const c = this.projection([region.lon, region.lat]);
      return { region, letter, x: c ? c[0] : -9999, y: c ? c[1] : -9999 };
    }).filter(p => p.x > 0 && p.x < this.width && p.y > 0 && p.y < this.height);

    // Greedy clustering: group points within CLUSTER_RADIUS of the first unclustered point.
    const assigned = new Set<number>();
    const clusters: typeof pts[number][][] = [];

    for (let i = 0; i < pts.length; i++) {
      if (assigned.has(i)) continue;
      const cluster = [pts[i]];
      assigned.add(i);
      for (let j = i + 1; j < pts.length; j++) {
        if (assigned.has(j)) continue;
        const dx = pts[i].x - pts[j].x;
        const dy = pts[i].y - pts[j].y;
        if (Math.hypot(dx, dy) < CLUSTER_RADIUS) {
          cluster.push(pts[j]);
          assigned.add(j);
        }
      }
      clusters.push(cluster);
    }

    // Build cluster boxes: grid layout centred on the geographic anchor.
    const MAX_COLS = 3;
    const step = MARKER_SIZE + MARKER_GAP;

    const boxes = clusters.map(cluster => {
      const ax = cluster.reduce((s, p) => s + p.x, 0) / cluster.length;
      const ay = cluster.reduce((s, p) => s + p.y, 0) / cluster.length;
      const cols = Math.min(cluster.length, MAX_COLS);
      const rows = Math.ceil(cluster.length / cols);
      return {
        anchorX: ax, anchorY: ay,   // geographic centre — never moves
        x: ax,       y: ay,          // current centre — nudged by simulation
        w: cols * MARKER_SIZE + (cols - 1) * MARKER_GAP,
        h: rows * MARKER_SIZE + (rows - 1) * MARKER_GAP,
        cols,
        cluster,
      };
    });

    // Iterative separation: push overlapping boxes apart, spring back to anchor.
    const BOX_PAD  = 3;   // minimum gap between bounding boxes (px)
    const SPRING   = 0.2; // attraction strength toward geographic anchor per iteration
    const ITERS    = 50;

    for (let iter = 0; iter < ITERS; iter++) {
      // Pairwise repulsion
      for (let i = 0; i < boxes.length; i++) {
        for (let j = i + 1; j < boxes.length; j++) {
          const a = boxes[i], b = boxes[j];
          const dx = b.x - a.x;
          const dy = b.y - a.y;
          const overlapX = (a.w + b.w) / 2 + BOX_PAD - Math.abs(dx);
          const overlapY = (a.h + b.h) / 2 + BOX_PAD - Math.abs(dy);

          if (overlapX > 0 && overlapY > 0) {
            // Resolve along the axis of least penetration
            if (overlapX <= overlapY) {
              const push = overlapX / 2;
              const dir  = dx >= 0 ? 1 : -1;
              a.x -= dir * push;
              b.x += dir * push;
            } else {
              const push = overlapY / 2;
              const dir  = dy >= 0 ? 1 : -1;
              a.y -= dir * push;
              b.y += dir * push;
            }
          }
        }
      }

      // Spring each box back toward its geographic anchor
      for (const box of boxes) {
        box.x += (box.anchorX - box.x) * SPRING;
        box.y += (box.anchorY - box.y) * SPRING;
      }
    }

    // Emit final marker positions from settled box centres
    const result: PlacedMarker[] = [];
    for (const box of boxes) {
      const originX = box.x - box.w / 2 + MARKER_SIZE / 2;
      const originY = box.y - box.h / 2 + MARKER_SIZE / 2;

      box.cluster.forEach((p, idx) => {
        const col = idx % box.cols;
        const row = Math.floor(idx / box.cols);
        result.push({
          region: p.region,
          letter: p.letter,
          x: originX + col * step,
          y: originY + row * step,
        });
      });
    }

    return result;
  }

  private renderMarkers(placed: PlacedMarker[]): void {
    this.markersGroup!
      .selectAll<SVGGElement, PlacedMarker>('.region-marker')
      .data(placed, d => d.region.region_id)
      .join(
        enter => {
          const g = enter.append('g')
            .attr('class', 'region-marker')
            .attr('transform', d => `translate(${d.x - MARKER_SIZE / 2},${d.y - MARKER_SIZE / 2})`)
            .style('cursor', 'pointer');

          g.append('rect')
            .attr('width', MARKER_SIZE)
            .attr('height', MARKER_SIZE)
            .attr('rx', 1)
            .attr('fill', d => STATUS_FILL[d.region.status])
            .attr('stroke', d => STATUS_STROKE[d.region.status])
            .attr('stroke-width', 1.5)
            // fill:none kills pointer events for the rect interior in SVG,
            // so we use a faint fill above and explicitly enable pointer-events.
            .attr('pointer-events', 'all');

          g.append('text')
            .attr('x', MARKER_SIZE / 2)
            .attr('y', MARKER_SIZE / 2)
            .attr('text-anchor', 'middle')
            .attr('dominant-baseline', 'central')
            .attr('font-size', '9px')
            .attr('font-weight', '900')
            .attr('font-family', 'ui-monospace, monospace')
            .attr('fill', d => STATUS_STROKE[d.region.status])
            .attr('pointer-events', 'none')
            .text(d => d.letter);

          g.on('mousemove', (event: MouseEvent, d) => this.showTooltip(event, d.region, d.letter))
           .on('mouseleave', () => this.hideTooltip());

          return g;
        },
        update => update
          .attr('transform', d => `translate(${d.x - MARKER_SIZE / 2},${d.y - MARKER_SIZE / 2})`)
          .call(u => u.select('rect')
            .attr('fill',   d => STATUS_FILL[d.region.status])
            .attr('stroke', d => STATUS_STROKE[d.region.status]))
          .call(u => u.select('text')
            .attr('fill', d => STATUS_STROKE[d.region.status])
            .text(d => d.letter))
            ,
      );
  }

  private showTooltip(event: MouseEvent, r: RegionStatusData, letter: string): void {
    const statusLabel = r.status.charAt(0).toUpperCase() + r.status.slice(1);
    const providerName = PROVIDER_NAMES[letter] ?? letter;
    const svcLines = (r.services ?? [])
      .map(s => `<div class="tt-service">
        <span class="tt-svc-name">${s.name}</span>
        <span class="tt-svc-dot" style="color:${STATUS_STROKE[s.status]}">●</span>
      </div>`)
      .join('');

    this.tooltip
      .style('display', 'block')
      .style('left', `${event.clientX + 14}px`)
      .style('top',  `${event.clientY - 8}px`)
      .html(`
        <div class="tt-provider">
          <span class="badge" style="background:${PROVIDER_BADGE_COLOR[letter] ?? '#888'}">${letter}</span>
          ${providerName}
        </div>
        <div class="tt-name">${r.name}</div>
        <div class="tt-id">${r.region_id}</div>
        <div class="tt-status status-${r.status}">● ${statusLabel}</div>
        ${r.azs ? `<div class="tt-azs">${r.azs} Availability Zones</div>` : ''}
        ${svcLines ? `<div class="tt-services">${svcLines}</div>` : ''}
      `);
  }

  private hideTooltip(): void {
    this.tooltip.style('display', 'none');
  }

  private onResize(): void {
    const newWidth = this.svg.node()!.parentElement!.clientWidth;
    if (Math.abs(newWidth - this.width) < 10) return;
    this.width  = newWidth;
    this.height = Math.round(newWidth * 0.508);
    this.svg.attr('viewBox', `0 0 ${this.width} ${this.height}`);
    this.projection.fitSize([this.width, this.height], { type: 'Sphere' });
    if (this.lastStatus) this.updateStatus(this.lastStatus);
  }
}
