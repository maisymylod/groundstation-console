#!/usr/bin/env node
// Toggle the frontend between the pre-optimization (zone.js) and committed
// (zoneless) configuration so scripts/profile.sh can build and measure both.
// Usage: node scripts/toggle-zone.mjs on|off   (run from the frontend dir)
//   on  -> zone.js polyfill + zone change detection (the "before" baseline)
//   off -> zoneless (the committed state)
//
// It edits angular.json (polyfills) and src/main.ts (the bootstrap provider),
// then restores them to the committed zoneless state when called with "off".
import { readFileSync, writeFileSync } from 'node:fs';

const mode = process.argv[2];
if (mode !== 'on' && mode !== 'off') {
  console.error('usage: toggle-zone.mjs on|off');
  process.exit(1);
}

const angularPath = 'angular.json';
const mainPath = 'src/main.ts';

const angular = JSON.parse(readFileSync(angularPath, 'utf8'));
const opts = angular.projects.console.architect.build.options;
opts.polyfills = mode === 'on' ? ['zone.js'] : [];
writeFileSync(angularPath, JSON.stringify(angular, null, 2) + '\n');

let main = readFileSync(mainPath, 'utf8');
if (mode === 'on') {
  main = main
    .replace(
      /import \{ provideExperimentalZonelessChangeDetection \} from '@angular\/core';\n/,
      '',
    )
    .replace(/\s*provideExperimentalZonelessChangeDetection\(\),\n/, '\n');
} else {
  if (!main.includes('provideExperimentalZonelessChangeDetection')) {
    main = main.replace(
      "import { AppComponent } from './app/app.component';",
      "import { provideExperimentalZonelessChangeDetection } from '@angular/core';\nimport { AppComponent } from './app/app.component';",
    );
    main = main.replace(
      'providers: [\n    provideHttpClient(withFetch()),',
      'providers: [\n    provideExperimentalZonelessChangeDetection(),\n    provideHttpClient(withFetch()),',
    );
  }
}
writeFileSync(mainPath, main);
console.log(`toggled zone mode: ${mode}`);
