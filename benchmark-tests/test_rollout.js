#!/usr/bin/env node

// Test script for percentage rollouts
// Usage: node test_rollout.js

const BASE_URL = 'http://localhost:8080/api/v1';

async function testRollout() {
  console.log('🧪 Testing Percentage Rollout (25%)\n');

  // Test with different user keys to see percentage distribution
  const userKeys = [
    'user_001', 'user_002', 'user_003', 'user_004', 'user_005',
    'user_006', 'user_007', 'user_008', 'user_009', 'user_010',
    'user_011', 'user_012', 'user_013', 'user_014', 'user_015',
    'user_016', 'user_017', 'user_018', 'user_019', 'user_020'
  ];

  let enabledCount = 0;
  let totalCount = userKeys.length;

  console.log('Testing flag "eg1" with different users:\n');

  for (const userKey of userKeys) {
    try {
      const response = await fetch(`${BASE_URL}/evaluate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          flag_key: 'eg1',
          user_key: userKey,
          attributes: {}
        })
      });

      if (response.ok) {
        const result = await response.json();
        const status = result.value ? '✅ ENABLED' : '❌ DISABLED';
        console.log(`${userKey}: ${status} (value: ${result.value})`);
        
        if (result.value) {
          enabledCount++;
        }
      } else {
        console.log(`${userKey}: ⚠️  ERROR - ${response.status}`);
      }
    } catch (error) {
      console.log(`${userKey}: ⚠️  ERROR - ${error.message}`);
    }
  }

  const actualPercentage = (enabledCount / totalCount * 100).toFixed(1);
  
  console.log('\n📊 Results:');
  console.log(`Total users tested: ${totalCount}`);
  console.log(`Users with flag enabled: ${enabledCount}`);
  console.log(`Actual percentage: ${actualPercentage}%`);
  console.log(`Expected percentage: ~25%`);

  if (Math.abs(parseFloat(actualPercentage) - 25) <= 15) {
    console.log('✅ Rollout is working correctly! (within acceptable range)');
  } else {
    console.log('❌ Rollout might not be working as expected');
  }
}

// Test individual rollout evaluation if rollout ID is provided
async function testRolloutDirect(rolloutId) {
  console.log(`\n🎯 Testing rollout ${rolloutId} directly:\n`);

  const testUsers = ['alice', 'bob', 'charlie', 'diana', 'eve'];
  
  for (const userKey of testUsers) {
    try {
      const response = await fetch(`${BASE_URL}/rollouts/${rolloutId}/evaluate?user_key=${userKey}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        }
      });

      if (response.ok) {
        const result = await response.json();
        const status = result.matched ? '✅ MATCHED' : '❌ NO MATCH';
        console.log(`${userKey}: ${status} (variation: ${result.variation_id || 'none'})`);
      } else {
        console.log(`${userKey}: ⚠️  ERROR - ${response.status}`);
      }
    } catch (error) {
      console.log(`${userKey}: ⚠️  ERROR - ${error.message}`);
    }
  }
}

// Run tests
testRollout().then(() => {
  // If you want to test a specific rollout directly, uncomment and provide rollout ID:
  // return testRolloutDirect('your-rollout-id-here');
}).catch(console.error);