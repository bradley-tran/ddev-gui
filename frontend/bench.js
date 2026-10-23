import { performance } from 'perf_hooks';

const iterations = 5_000_000;

function benchCurrent() {
  const runtimeListeners = new Map();
  // pre-populate
  for (let i = 0; i < 100; i++) {
    runtimeListeners.set(`event-${i}`, new Set());
  }

  const start = performance.now();

  for (let i = 0; i < iterations; i++) {
    const event = `event-${i % 100}`;
    const cb = () => {};
    const listeners = runtimeListeners.get(event) ?? new Set();
    runtimeListeners.set(event, listeners);
    listeners.add(cb);
  }

  const end = performance.now();
  return end - start;
}

function benchOptimized() {
  const runtimeListeners = new Map();
  // pre-populate
  for (let i = 0; i < 100; i++) {
    runtimeListeners.set(`event-${i}`, new Set());
  }

  const start = performance.now();

  for (let i = 0; i < iterations; i++) {
    const event = `event-${i % 100}`;
    const cb = () => {};
    let listeners = runtimeListeners.get(event);
    if (!listeners) {
      listeners = new Set();
      runtimeListeners.set(event, listeners);
    }
    listeners.add(cb);
  }

  const end = performance.now();
  return end - start;
}

// Warmup
benchCurrent();
benchOptimized();

let currentTotal = 0;
let optTotal = 0;
const runs = 5;

for (let i = 0; i < runs; i++) {
    currentTotal += benchCurrent();
    optTotal += benchOptimized();
}

console.log(`Current: ${(currentTotal/runs).toFixed(2)} ms`);
console.log(`Optimized: ${(optTotal/runs).toFixed(2)} ms`);
