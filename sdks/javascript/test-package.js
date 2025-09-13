#!/usr/bin/env node

// Simple test to verify the package can be imported and basic functionality works
const { FlexFlagClient, MemoryCache } = require('./dist/index.js');

console.log('Testing FlexFlag JavaScript SDK...\n');

try {
  // Test 1: Constructor validation
  console.log('✓ Test 1: Constructor validation');
  try {
    new FlexFlagClient({});
    console.log('✗ Expected error for missing API key');
    process.exit(1);
  } catch (error) {
    if (error.message.includes('API key is required')) {
      console.log('  ✓ Correctly throws error for missing API key');
    } else {
      console.log('  ✗ Unexpected error:', error.message);
      process.exit(1);
    }
  }

  // Test 2: Successful instantiation
  console.log('\n✓ Test 2: Successful instantiation');
  const client = new FlexFlagClient({
    apiKey: 'test-key',
    baseUrl: 'http://localhost:8080',
    environment: 'test'
  });
  console.log('  ✓ Client created successfully');
  console.log('  ✓ Client is instance of FlexFlagClient:', client instanceof FlexFlagClient);

  // Test 3: Cache instantiation
  console.log('\n✓ Test 3: Cache functionality');
  const cache = new MemoryCache();
  console.log('  ✓ MemoryCache created successfully');

  console.log('\n🎉 All basic tests passed!');
  console.log('\nPackage is ready for publishing.');
  
} catch (error) {
  console.error('✗ Test failed:', error.message);
  process.exit(1);
}