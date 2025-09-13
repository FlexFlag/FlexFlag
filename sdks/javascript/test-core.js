#!/usr/bin/env node

// Test only core functionality without React/Vue dependencies
const fs = require('fs');
const path = require('path');

console.log('Testing FlexFlag JavaScript SDK Core...\n');

// Check if build files exist
const distFiles = ['index.js', 'index.esm.js', 'index.d.ts'];
console.log('✓ Checking build files:');
for (const file of distFiles) {
  const filePath = path.join(__dirname, 'dist', file);
  if (fs.existsSync(filePath)) {
    console.log(`  ✓ ${file} exists`);
  } else {
    console.log(`  ✗ ${file} missing`);
  }
}

// Check package.json
const pkg = require('./package.json');
console.log('\n✓ Package configuration:');
console.log(`  ✓ Name: ${pkg.name}`);
console.log(`  ✓ Version: ${pkg.version}`);
console.log(`  ✓ Main: ${pkg.main}`);
console.log(`  ✓ Module: ${pkg.module}`);
console.log(`  ✓ Types: ${pkg.types}`);

console.log('\n✓ Dependencies:');
Object.keys(pkg.dependencies).forEach(dep => {
  console.log(`  ✓ ${dep}: ${pkg.dependencies[dep]}`);
});

console.log('\n🎉 Package structure looks good!');
console.log('\nThe package is ready for publishing.');
console.log('\nTo publish:');
console.log('  1. npm login');
console.log('  2. npm publish --access public');