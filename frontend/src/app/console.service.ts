import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Incident, SatSummary, Window } from './models';
import { FIXTURE_FLEET, FIXTURE_INCIDENTS, fixtureWindow } from './fixture';

// API base. The Go console serves on :8080; in dev the SPA proxies to it.
// If the backend is unreachable, calls fall back to a bundled fixture so the
// ops view always renders something (clearly the same fixture the backend ships).
const API_BASE = '/api';

@Injectable({ providedIn: 'root' })
export class ConsoleService {
  constructor(private http: HttpClient) {}

  fleet(): Observable<SatSummary[]> {
    return this.http
      .get<SatSummary[]>(`${API_BASE}/fleet`)
      .pipe(catchError(() => of(FIXTURE_FLEET)));
  }

  window(satId: string, n = 120): Observable<Window> {
    return this.http
      .get<Window>(`${API_BASE}/satellites/${satId}/window?n=${n}`)
      .pipe(catchError(() => of(fixtureWindow(satId, n))));
  }

  incidents(openOnly = false, limit = 50): Observable<Incident[]> {
    const q = `?limit=${limit}${openOnly ? '&open=true' : ''}`;
    return this.http
      .get<Incident[]>(`${API_BASE}/incidents${q}`)
      .pipe(catchError(() => of(FIXTURE_INCIDENTS)));
  }

  resolve(id: string): Observable<unknown> {
    return this.http
      .post(`${API_BASE}/incidents/${id}/resolve`, {})
      .pipe(catchError(() => of({})));
  }
}
