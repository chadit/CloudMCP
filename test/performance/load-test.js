/**
 * CloudMCP Performance Test Suite
 * Comprehensive load testing for CloudMCP container validation
 * 
 * This script uses k6 to perform load testing on CloudMCP containers
 * to validate performance characteristics under various scenarios.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
export let healthCheckDuration = new Trend('health_check_duration');
export let metricsCheckDuration = new Trend('metrics_check_duration');
export let errorRate = new Rate('errors');
export let totalRequests = new Counter('total_requests');

// Test configuration
export let options = {
  scenarios: {
    // Baseline performance test
    baseline: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 5 },   // Ramp up to 5 users
        { duration: '60s', target: 5 },   // Stay at 5 users
        { duration: '30s', target: 0 },   // Ramp down
      ],
      gracefulRampDown: '10s',
    },
    
    // Spike test for container resilience
    spike: {
      executor: 'ramping-vus',
      startTime: '2m',
      startVUs: 1,
      stages: [
        { duration: '10s', target: 20 },  // Quick spike
        { duration: '10s', target: 1 },   // Quick recovery
      ],
      gracefulRampDown: '5s',
    },
    
    // Stress test to find limits
    stress: {
      executor: 'ramping-vus',
      startTime: '3m',
      startVUs: 5,
      stages: [
        { duration: '60s', target: 15 },  // Ramp up
        { duration: '60s', target: 15 },  // Stay high
        { duration: '30s', target: 0 },   // Ramp down
      ],
      gracefulRampDown: '10s',
    }
  },
  
  thresholds: {
    // Health endpoint should respond within 500ms
    'health_check_duration': ['p(95)<500'],
    
    // Metrics endpoint should respond within 1s
    'metrics_check_duration': ['p(95)<1000'],
    
    // Error rate should be less than 1%
    'errors': ['rate<0.01'],
    
    // Overall response time should be reasonable
    'http_req_duration': ['p(95)<1000'],
    
    // HTTP success rate should be high
    'http_req_failed': ['rate<0.01'],
  }
};

// Configuration from environment
const baseURL = __ENV.CLOUDMCP_URL || 'http://localhost:8080';
const testTimeout = __ENV.TEST_TIMEOUT || '10s';

// Test data
const testPayloads = [
  {},  // Empty payload
  { test: 'basic' },  // Simple payload
  { complex: { nested: { data: 'test' } } },  // Complex payload
];

export function setup() {
  console.log(`🚀 Starting CloudMCP performance tests against: ${baseURL}`);
  
  // Verify container is responsive before starting tests
  let response = http.get(`${baseURL}/health`);
  if (response.status !== 200) {
    throw new Error(`Container not ready: ${response.status} ${response.statusText}`);
  }
  
  console.log('✅ Container is ready for performance testing');
  return { baseURL: baseURL };
}

export default function(data) {
  totalRequests.add(1);
  
  // Test health endpoint performance
  testHealthEndpoint(data.baseURL);
  
  // Test metrics endpoint performance (if accessible)
  testMetricsEndpoint(data.baseURL);
  
  // Brief pause between test iterations
  sleep(Math.random() * 2); // 0-2 second random pause
}

/**
 * Test health endpoint performance and reliability
 */
function testHealthEndpoint(baseURL) {
  let start = new Date();
  
  let response = http.get(`${baseURL}/health`, {
    timeout: testTimeout,
    tags: { endpoint: 'health' }
  });
  
  let duration = new Date() - start;
  healthCheckDuration.add(duration);
  
  let success = check(response, {
    'health endpoint status is 200': (r) => r.status === 200,
    'health endpoint responds quickly': (r) => r.timings.duration < 500,
    'health endpoint has valid response': (r) => {
      try {
        // Expect JSON response or plain text "OK"
        return r.body.length > 0;
      } catch (e) {
        return false;
      }
    },
    'health endpoint has correct headers': (r) => {
      return r.headers['Content-Type'] !== undefined;
    }
  });
  
  if (!success) {
    errorRate.add(1);
    console.error(`Health check failed: ${response.status} ${response.statusText}`);
  }
}

/**
 * Test metrics endpoint performance (may require authentication)
 */
function testMetricsEndpoint(baseURL) {
  let start = new Date();
  
  let response = http.get(`${baseURL}/metrics`, {
    timeout: testTimeout,
    tags: { endpoint: 'metrics' }
  });
  
  let duration = new Date() - start;
  metricsCheckDuration.add(duration);
  
  // Metrics endpoint may require authentication, so we're more lenient
  let success = check(response, {
    'metrics endpoint is reachable': (r) => r.status !== 0, // Any response is good
    'metrics endpoint responds quickly': (r) => r.timings.duration < 1000,
    'metrics endpoint returns data or auth challenge': (r) => {
      // Accept 200 (success), 401 (auth required), or 403 (forbidden)
      return r.status === 200 || r.status === 401 || r.status === 403;
    }
  });
  
  if (!success && response.status !== 401 && response.status !== 403) {
    console.warn(`Metrics endpoint unexpected response: ${response.status}`);
  }
}

/**
 * Test container under sustained load
 */
export function sustainedLoad() {
  group('Sustained Load Test', function() {
    let responses = [];
    
    // Make multiple requests in quick succession
    for (let i = 0; i < 10; i++) {
      responses.push(http.get(`${baseURL}/health`, {
        tags: { scenario: 'sustained_load', iteration: i }
      }));
    }
    
    // Verify all requests succeeded
    let successCount = 0;
    responses.forEach((response, index) => {
      if (check(response, {
        [`sustained load request ${index} succeeded`]: (r) => r.status === 200
      })) {
        successCount++;
      }
    });
    
    console.log(`Sustained load: ${successCount}/${responses.length} requests succeeded`);
  });
}

/**
 * Test container memory behavior under load
 */
export function memoryStressTest() {
  group('Memory Stress Test', function() {
    // Make requests with progressively larger (simulated) payloads
    testPayloads.forEach((payload, index) => {
      let response = http.get(`${baseURL}/health`, {
        headers: {
          'X-Test-Payload-Size': JSON.stringify(payload).length.toString(),
          'X-Test-Iteration': index.toString()
        },
        tags: { scenario: 'memory_stress', payload_size: JSON.stringify(payload).length }
      });
      
      check(response, {
        [`memory stress request ${index} succeeded`]: (r) => r.status === 200,
        [`memory stress request ${index} fast enough`]: (r) => r.timings.duration < 1000
      });
    });
  });
}

export function teardown(data) {
  console.log('🏁 Performance tests completed');
  
  // Final health check to ensure container is still responsive
  let response = http.get(`${data.baseURL}/health`);
  let healthy = check(response, {
    'container still healthy after tests': (r) => r.status === 200
  });
  
  if (healthy) {
    console.log('✅ Container remained healthy throughout performance testing');
  } else {
    console.error('❌ Container health degraded during performance testing');
  }
}