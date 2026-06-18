import { ChangeDetectionStrategy, Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ConsoleService } from './console.service';
import { Incident, SatSummary, severityClass, SEVERITY_LABELS } from './models';
import { SparklineComponent } from './sparkline.component';

@Component({
  selector: 'gsc-root',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [CommonModule, SparklineComponent],
  styles: [
    `
      .wrap { max-width: 1200px; margin: 0 auto; padding: 1.25rem; }
      header { display: flex; align-items: baseline; gap: 1rem; border-bottom: 1px solid var(--line); padding-bottom: .75rem; }
      h1 { font-size: 1.15rem; margin: 0; }
      .sub { color: var(--muted); font-size: .8rem; }
      .grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: .75rem; margin: 1rem 0; }
      .card { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: .75rem; cursor: pointer; }
      .card.sel { border-color: var(--accent); }
      .card h3 { margin: 0 0 .25rem; font-size: .95rem; }
      .row { display: flex; justify-content: space-between; font-size: .78rem; color: var(--muted); }
      .badge { font-size: .7rem; padding: .1rem .4rem; border-radius: 4px; }
      .nominal { color: var(--nominal); } .warning { color: var(--warning); } .critical { color: var(--critical); }
      .bg-nominal { background: rgba(63,185,80,.15); color: var(--nominal); }
      .bg-warning { background: rgba(210,153,34,.15); color: var(--warning); }
      .bg-critical { background: rgba(248,81,73,.15); color: var(--critical); }
      .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
      .charts { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
      .chart-label { font-size: .75rem; color: var(--muted); margin-bottom: .25rem; }
      table { width: 100%; border-collapse: collapse; font-size: .8rem; }
      th, td { text-align: left; padding: .4rem .5rem; border-bottom: 1px solid var(--line); }
      th { color: var(--muted); font-weight: normal; }
      button { background: var(--accent); color: #06101f; border: 0; border-radius: 4px; padding: .25rem .6rem; cursor: pointer; font-family: inherit; }
    `,
  ],
  template: `
    <div class="wrap">
      <header>
        <h1>Heliosnet Ops Console</h1>
        <span class="sub">live telemetry and incident view over the ground-system data plane</span>
      </header>

      <section class="grid">
        <div
          class="card"
          *ngFor="let s of fleet()"
          [class.sel]="s.sat_id === selected()"
          (click)="select(s.sat_id)"
        >
          <h3>
            {{ s.sat_id }}
            <span class="badge" [ngClass]="'bg-' + sevClass(s.worst_severity)">
              {{ sevLabel(s.worst_severity) }}
            </span>
          </h3>
          <div class="row"><span>open incidents</span><span [ngClass]="s.open_incidents ? 'warning' : 'nominal'">{{ s.open_incidents }}</span></div>
          <div class="row"><span>battery</span><span>{{ s.last_battery_v | number: '1.2-2' }} V</span></div>
          <div class="row"><span>temp</span><span>{{ s.last_temp_c | number: '1.1-1' }} C</span></div>
          <div class="row"><span>signal</span><span>{{ s.last_signal_dbm | number: '1.1-1' }} dBm</span></div>
        </div>
      </section>

      <section class="panel" *ngIf="selected() as sel">
        <div class="chart-label">{{ sel }} telemetry window (last {{ windowFrames().length }} frames)</div>
        <div class="charts">
          <div>
            <div class="chart-label">battery (V)</div>
            <gsc-sparkline [values]="series('battery_v')" color="#3fb950"></gsc-sparkline>
          </div>
          <div>
            <div class="chart-label">temperature (C)</div>
            <gsc-sparkline [values]="series('temp_c')" color="#f85149"></gsc-sparkline>
          </div>
          <div>
            <div class="chart-label">signal (dBm)</div>
            <gsc-sparkline [values]="series('signal_dbm')" color="#58a6ff"></gsc-sparkline>
          </div>
          <div>
            <div class="chart-label">attitude error</div>
            <gsc-sparkline [values]="series('attitude_err')" color="#d29922"></gsc-sparkline>
          </div>
        </div>
      </section>

      <section class="panel">
        <div class="chart-label">open incidents</div>
        <table>
          <thead>
            <tr><th>satellite</th><th>type</th><th>severity</th><th>summary</th><th></th></tr>
          </thead>
          <tbody>
            <tr *ngFor="let i of incidents()">
              <td>{{ i.sat_id }}</td>
              <td>{{ i.type }}</td>
              <td [ngClass]="sevClass(i.severity)">{{ sevLabel(i.severity) }}</td>
              <td>{{ i.summary }}</td>
              <td><button (click)="resolve(i)">resolve</button></td>
            </tr>
            <tr *ngIf="incidents().length === 0"><td colspan="5" class="nominal">no open incidents</td></tr>
          </tbody>
        </table>
      </section>
    </div>
  `,
})
export class AppComponent implements OnInit {
  fleet = signal<SatSummary[]>([]);
  incidents = signal<Incident[]>([]);
  selected = signal<string>('');
  windowFrames = signal<Record<string, number>[]>([]);

  constructor(private api: ConsoleService) {}

  ngOnInit(): void {
    this.api.fleet().subscribe((f) => {
      this.fleet.set(f);
      if (f.length && !this.selected()) {
        this.select(f[0].sat_id);
      }
    });
    this.loadIncidents();
  }

  select(satId: string): void {
    this.selected.set(satId);
    this.api.window(satId, 120).subscribe((w) => {
      this.windowFrames.set(w.frames as unknown as Record<string, number>[]);
    });
  }

  series(field: string): number[] {
    return this.windowFrames().map((f) => Number(f[field]));
  }

  resolve(i: Incident): void {
    this.api.resolve(i.id).subscribe(() => this.loadIncidents());
  }

  private loadIncidents(): void {
    this.api.incidents(true, 50).subscribe((i) => this.incidents.set(i));
  }

  sevClass = severityClass;
  sevLabel = (s: number): string => SEVERITY_LABELS[s] ?? 'nominal';
}
