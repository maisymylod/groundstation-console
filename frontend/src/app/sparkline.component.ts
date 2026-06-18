import { ChangeDetectionStrategy, Component, Input } from '@angular/core';

// A dependency-free SVG sparkline. Drawing this inline (rather than pulling in a
// charting library) keeps the production bundle small; see docs/profiling.md.
@Component({
  selector: 'gsc-sparkline',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <svg [attr.viewBox]="'0 0 ' + width + ' ' + height" [attr.width]="width" [attr.height]="height">
      <polyline [attr.points]="points" fill="none" [attr.stroke]="color" stroke-width="1.5" />
    </svg>
  `,
})
export class SparklineComponent {
  @Input() width = 220;
  @Input() height = 44;
  @Input() color = '#58a6ff';

  private _values: number[] = [];
  points = '';

  @Input() set values(v: number[]) {
    this._values = v ?? [];
    this.points = this.buildPoints();
  }

  private buildPoints(): string {
    const v = this._values;
    if (v.length < 2) {
      return '';
    }
    const min = Math.min(...v);
    const max = Math.max(...v);
    const span = max - min || 1;
    const stepX = this.width / (v.length - 1);
    const pad = 3;
    const usable = this.height - pad * 2;
    return v
      .map((y, i) => {
        const px = (i * stepX).toFixed(2);
        const py = (pad + usable - ((y - min) / span) * usable).toFixed(2);
        return `${px},${py}`;
      })
      .join(' ');
  }
}
