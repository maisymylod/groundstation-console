import { bootstrapApplication } from '@angular/platform-browser';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { provideExperimentalZonelessChangeDetection } from '@angular/core';
import { AppComponent } from './app/app.component';

// Zoneless change detection. Every component is OnPush and state lives in signals,
// so the app does not need zone.js. Dropping the zone.js polyfill removes the
// entire polyfills chunk from the bundle; see docs/profiling.md for the measured
// before/after.
bootstrapApplication(AppComponent, {
  providers: [
    provideExperimentalZonelessChangeDetection(),
    provideHttpClient(withFetch()),
  ],
}).catch((err) => console.error(err));
