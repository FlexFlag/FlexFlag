# Publishing Checklist for @flexflag/client

## ✅ Pre-publish Checklist

- [x] Package structure is correct
- [x] Build files exist (index.js, index.esm.js, index.d.ts)
- [x] Package.json is configured properly
- [x] Dependencies are correct
- [x] README.md is comprehensive
- [x] TypeScript definitions are available
- [x] Core functionality is working

## 📋 Publishing Steps

### 1. Login to npm
```bash
npm login
```
Enter your npm credentials when prompted.

### 2. Verify package contents
```bash
npm pack --dry-run
```
This shows what files will be included in the package.

### 3. Test the package locally (optional)
```bash
npm pack
# This creates a .tgz file you can test in another project
```

### 4. Publish to npm
```bash
npm publish --access public
```

**Note**: Since this is a scoped package (`@flexflag/client`), you need `--access public` for the first publish.

### 5. Verify the publication
```bash
npm view @flexflag/client
```

## 🔄 Version Updates

For future updates:

1. Update version in package.json:
   ```bash
   npm version patch  # for bug fixes
   npm version minor  # for new features  
   npm version major  # for breaking changes
   ```

2. Rebuild:
   ```bash
   npm run build
   ```

3. Publish:
   ```bash
   npm publish
   ```

## 📦 Package Contents

The published package will include:
- `dist/` - Built JavaScript files and type definitions
- `package.json` - Package metadata
- `README.md` - Documentation

## 🚀 Post-publish

After publishing, the package will be available at:
- npm: https://www.npmjs.com/package/@flexflag/client
- Installation: `npm install @flexflag/client`

## 📊 Package Status

- **Name**: @flexflag/client
- **Version**: 1.0.0
- **Main**: dist/index.js
- **Module**: dist/index.esm.js
- **Types**: dist/index.d.ts
- **Files**: dist directory only
- **Dependencies**: axios, eventemitter3, lru-cache