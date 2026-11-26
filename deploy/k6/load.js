// k6 — нагрузочный тест WebLab #5.
//
// Гипотеза (до запуска): узкое место — БД/бэкенд на чтение; балансировка
// 2:1:1 должна распределить GET по трём инстансам (2:1:1), а запись — идти
// только на основной инстанс. Целевые показатели: доля ошибок < 1%,
// p95 < 500 мс на локальном стенде.
//
// Запуск: BASE_URL=http://nginx:80 k6 run deploy/k6/load.js
//   (внутри compose-сети; nginx слушает :80 внутренне).

import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    reads: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 100 },
        { duration: '60s', target: 100 },
        { duration: '20s', target: 0 },
      ],
      exec: 'read',
    },
    writes: {
      executor: 'constant-vus',
      vus: 10,
      duration: '60s',
      startTime: '20s',
      exec: 'write',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE = __ENV.BASE_URL || 'http://nginx:80';

// Пути чтения (GET) — те самые, что балансируются 2:1:1.
const READ_PATHS = [
  '/api/v2/subscription-plans',
  '/api/v2/services',
  '/api/v2/promocodes',
];

export function read() {
  const path = READ_PATHS[Math.floor(Math.random() * READ_PATHS.length)];
  const res = http.get(`${BASE}${path}`);
  check(res, { 'GET 200': (r) => r.status === 200 });
}

export function write() {
  const payload = JSON.stringify({
    email: `k6-${__VU}-${__ITER}-${Math.random().toString(36).slice(2, 10)}@example.com`,
    password: 'secret',
    user_name: 'k6',
    age: 25,
    balance: 1000,
  });
  const res = http.post(`${BASE}/api/v2/users`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'POST 201': (r) => r.status === 201 });
}
