# FlexFlag Organization Setup Guide

This guide will help you transfer the FlexFlag project to the FlexFlag organization and set up proper GitHub repository structure.

## 🏢 Organization Setup Steps

### 1. Transfer Repository to FlexFlag Organization

You'll need to transfer the current repository from `reez-personal/FlexFlag` to `flexflag/flexflag`:

1. **Go to Repository Settings**
   - Navigate to https://github.com/reez-personal/FlexFlag/settings
   - Scroll down to "Danger Zone"

2. **Transfer Repository**
   - Click "Transfer ownership"
   - Enter destination: `flexflag` (organization name)
   - Enter new repository name: `flexflag` (lowercase)
   - Confirm the transfer

3. **Update Local Git Remote**
   ```bash
   git remote set-url origin https://github.com/flexflag/flexflag.git
   git remote -v  # Verify the change
   ```

### 2. Update Repository Settings

After transfer, configure the repository:

1. **Repository Description**
   ```
   High-performance feature flag management system with distributed edge servers and sub-millisecond evaluation
   ```

2. **Topics/Tags** (Add these topics)
   ```
   feature-flags, feature-toggles, ab-testing, go, nextjs, postgres, redis, 
   distributed-systems, edge-computing, real-time, performance, developer-tools
   ```

3. **Enable Features**
   - ✅ Issues
   - ✅ Projects  
   - ✅ Wiki
   - ✅ Discussions
   - ✅ Actions (CI/CD)

4. **Branch Protection**
   - Set up branch protection for `main`
   - Require PR reviews
   - Require status checks

### 3. GitHub Pages Documentation

Enable GitHub Pages for documentation:

1. **Go to Repository Settings → Pages**
2. **Source**: Deploy from a branch
3. **Branch**: `main`
4. **Folder**: `/docs`
5. **Custom Domain** (optional): `docs.flexflag.io`

This will make documentation available at:
- https://flexflag.github.io/flexflag/
- Or https://docs.flexflag.io/ (if custom domain is set)

### 4. Update npm Package

The npm package has already been updated to point to the new repository. Version 1.0.1 includes:

- ✅ Correct repository URL: `https://github.com/flexflag/flexflag.git`
- ✅ Correct homepage: `https://github.com/flexflag/flexflag`  
- ✅ Correct issues URL: `https://github.com/flexflag/flexflag/issues`

### 5. Setup Documentation Website

We've created comprehensive documentation structure:

```
docs/
├── sdks/
│   └── javascript/
│       ├── README.md              # Main SDK documentation
│       ├── getting-started.md     # Installation & basic usage
│       ├── react-integration.md   # React hooks & components
│       ├── vue-integration.md     # Vue composables & components
│       └── api-reference.md       # Complete API documentation
└── SETUP_ORGANIZATION.md          # This file
```

### 6. CI/CD Setup

Consider setting up GitHub Actions for:

1. **Automated Testing**
   ```yaml
   # .github/workflows/test.yml
   name: Tests
   on: [push, pull_request]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v4
         - uses: actions/setup-node@v4
         - run: make test
         - run: cd ui && npm test
   ```

2. **Automated SDK Publishing**
   ```yaml
   # .github/workflows/publish-sdk.yml
   name: Publish SDK
   on:
     push:
       tags: ['v*']
   jobs:
     publish:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-node@v4
         - run: cd sdks/javascript && npm publish
   ```

### 7. Community Setup

1. **Create Issue Templates**
   ```
   .github/ISSUE_TEMPLATE/
   ├── bug_report.md
   ├── feature_request.md
   └── question.md
   ```

2. **Add Contributing Guidelines**
   ```
   CONTRIBUTING.md
   ```

3. **Add Code of Conduct**
   ```
   CODE_OF_CONDUCT.md
   ```

4. **Add Security Policy**
   ```
   SECURITY.md
   ```

### 8. Custom Domain Setup (Optional)

If you want to use custom domains:

1. **Buy Domain**: `flexflag.io`
2. **DNS Setup**:
   ```
   docs.flexflag.io  CNAME  flexflag.github.io
   www.flexflag.io   CNAME  flexflag.github.io  
   flexflag.io       A      185.199.108.153 (GitHub Pages IP)
   ```

3. **Update Repository Settings**:
   - GitHub Pages custom domain: `docs.flexflag.io`
   - Update package.json homepage to `https://flexflag.io`

## 📋 Post-Transfer Checklist

After transferring to the FlexFlag organization:

- [ ] Repository transferred to `flexflag/flexflag`
- [ ] Local git remote updated
- [ ] Repository description and topics added
- [ ] Branch protection enabled
- [ ] GitHub Pages enabled for documentation
- [ ] npm package points to correct repository ✅ (already done)
- [ ] Documentation is comprehensive ✅ (already done)
- [ ] CI/CD workflows configured
- [ ] Community files added
- [ ] Custom domain configured (optional)

## 🔗 Updated Links

After transfer, all links will be:

- **Repository**: https://github.com/flexflag/flexflag
- **Issues**: https://github.com/flexflag/flexflag/issues
- **Documentation**: https://flexflag.github.io/flexflag/ (or https://docs.flexflag.io/)
- **npm Package**: https://www.npmjs.com/package/flexflag-client

## 💡 Benefits of Organization

1. **Professional Branding**: Dedicated organization for FlexFlag
2. **Team Collaboration**: Easy to add team members
3. **Multiple Repositories**: Space for additional repos (SDKs, examples, etc.)
4. **Organization Pages**: Can create https://flexflag.github.io/
5. **Better SEO**: Professional appearance for users and contributors

## 🚀 Next Steps

1. **Transfer the repository** to FlexFlag organization
2. **Update local git remotes** on all development machines
3. **Configure repository settings** as outlined above
4. **Set up GitHub Pages** for documentation
5. **Add team members** to the organization
6. **Consider CI/CD setup** for automated testing and deployment

The npm package (flexflag-client v1.0.1) is already live and ready to use with the correct repository references!