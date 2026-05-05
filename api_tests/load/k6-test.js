import http from 'k6/http';
import { check, sleep } from 'k6';

const FIXTURE_BODY_TYPES = {
  summarize: { default: 5, heavy: 3, light: 10 },
  structurize: { default: 5, heavy: 3, light: 10 },
  template: { default: 5, heavy: 3, light: 10 },
};

const FIXTURES = Object.entries(FIXTURE_BODY_TYPES).flatMap(([category, types]) => {
  return Object.entries(types).flatMap(([bodyType, count]) => {
    return Array.from({ length: count }, (_, index) => {
      const path = `../request_fixtures/${category}_${bodyType}_${index}.json`;
      return {
        endpoint: `/${category}`,
        payload: open(path),
        source: path,
      };
    });
  });
});

function randomChoice(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export const options = {
  stages: [
    { duration: '30s', target: 20 },  
    { duration: '1m', target: 100 },  
    { duration: '30s', target: 0 },   
  ],
  
  thresholds: {
    http_req_duration: ['p(95)<1500'],
    http_req_failed: ['rate<0.01'],  
  },
  
  noConnectionReuse: false,  
  userAgent: 'k6-AI-Service-LoadTest/1.0',
};

export default function () {
  const BASE_URL = 'http://localhost:8080';
  
  const fixture = randomChoice(FIXTURES);
  const payload = fixture.payload;
  const expectedCheck = (r) => {
    try {
      const body = r.json();
      return body && typeof body.requestId === 'string' && body.requestId.length > 0;
    } catch {
      return false;
    }
  };

  const res = http.post(
    `${BASE_URL}${fixture.endpoint}`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Test-Source': 'k6-load-test',
        'X-Fixture-Source': fixture.source,
      },
      timeout: '50s',  
    }
  );
  
  check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    // 'has valid requestId': expectedCheck,
    // 'response has content': (r) => r.body && r.body.length > 0,
  });
  
  sleep(0.1 + Math.random() * 0.4);
}