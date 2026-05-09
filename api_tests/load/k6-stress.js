import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const FIXTURE_BODY_TYPES = {
  summarize:   { heavy: 200 },
  structurize: { heavy: 200 },
  template:    { heavy: 200 },
};

const ALL_FIXTURES = Object.entries(FIXTURE_BODY_TYPES).flatMap(([category, types]) =>
  Object.entries(types).flatMap(([bodyType, count]) =>
    Array.from({ length: count }, (_, index) => ({
      endpoint: `/${category}`,
      payload:  open(`../request_fixtures/${category}_${bodyType}_${index}.json`),
      source:   `${category}_${bodyType}_${index}.json`,
      isHeavy:  bodyType === 'heavy',
      category,
      bodyType,
    }))
  )
);

const HEAVY_FIXTURES = ALL_FIXTURES.filter(f => f.isHeavy);

const errorsByEndpoint  = new Counter('errors_by_endpoint');
const latencyByEndpoint = new Trend('latency_by_endpoint', true);
const heavyRequestRate  = new Rate('heavy_request_rate');
const timeoutRate       = new Rate('timeout_rate');

export const options = {
  stages: [
    { duration: '2m',  target: 12 }, 
    { duration: '5m',  target: 14 }, 
    { duration: '5m',  target: 14 },  
    { duration: '5m',  target: 16 },  
    { duration: '5m',  target: 18 },  
    { duration: '5m',  target: 16 },  
    { duration: '1m',  target: 12 },  
    { duration: '2m',  target: 0 },  
  ],

  thresholds: {
    http_req_duration:   ['p(95)<8000', 'p(99)<15000'],
    http_req_failed:     ['rate<0.05'],
    timeout_rate:        ['rate<0.02'],
    latency_by_endpoint: ['p(95)<8000'],
    heavy_request_rate:  ['rate>0'],
  },

  noConnectionReuse: false,

  userAgent: 'k6-AI-Service-StressTest/3.0-cpu-safe',

  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

const BASE_URL = 'http://localhost:8080';

function randomChoice(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function adaptiveSleep(vuCount) {
  if (vuCount >= 7) return 0.1 + Math.random() * 1.0;  // 1.5–2.5s
  if (vuCount >= 4) return 0.1 + Math.random() * 1.5;  // 2.0–3.5s
  return 0.1 + Math.random() * 2.0;                     // 3.0–5.0s
}

function buildHeaders(fixture) {
  return {
    'Content-Type':     'application/json',
    'X-Test-Source':    'k6-stress-test',
    'X-Fixture-Source': fixture.source,
    'X-Body-Type':      fixture.bodyType,
  };
}

function sendRequest(fixture) {
  const res = http.post(
    `${BASE_URL}${fixture.endpoint}`,
    fixture.payload,
    { headers: buildHeaders(fixture), timeout: '90s' }
  );

  latencyByEndpoint.add(res.timings.duration, { endpoint: fixture.endpoint });
  timeoutRate.add(res.timings.duration >= 90000 ? 1 : 0);

  return res;
}

function scenarioRandom() {
  const fixture = randomChoice(ALL_FIXTURES);
  heavyRequestRate.add(fixture.isHeavy);

  const res = sendRequest(fixture);

  const ok = check(res, {
    'status is 2xx':   (r) => r.status >= 200 && r.status < 300,
    'no server error': (r) => r.status !== 500 && r.status !== 503,
  });

  if (!ok) errorsByEndpoint.add(1, { endpoint: fixture.endpoint });
}

function scenarioHeavy() {
  const fixture = randomChoice(HEAVY_FIXTURES);
  heavyRequestRate.add(1);

  const res = sendRequest(fixture);

  check(res, {
    'heavy: status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  sleep(1.0 + Math.random() * 1.0);
}

export default function () {
  if (__VU % 10 < 6) {
    group('random', scenarioRandom);  
  } else {
    group('heavy', scenarioHeavy);    
  }

  sleep(adaptiveSleep(__VU));
}

export function handleSummary(data) {
  const report = {
    timestamp:    new Date().toISOString(),
    duration_min: 30,
    metrics: {
      total_requests:  data.metrics.http_reqs?.values?.count ?? 0,
      failed_requests: data.metrics.http_req_failed?.values?.passes ?? 0,
      p95_ms:          data.metrics.http_req_duration?.values?.['p(95)'] ?? 0,
      p99_ms:          data.metrics.http_req_duration?.values?.['p(99)'] ?? 0,
      avg_ms:          data.metrics.http_req_duration?.values?.avg ?? 0,
      max_ms:          data.metrics.http_req_duration?.values?.max ?? 0,
    },
  };

  return {
    'stdout':              JSON.stringify(report, null, 2),
    'stress-summary.json': JSON.stringify(report, null, 2),
  };
}