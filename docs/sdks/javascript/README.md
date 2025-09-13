# FlexFlag JavaScript SDK Documentation

Welcome to the FlexFlag JavaScript SDK documentation. This SDK provides high-performance feature flag evaluation for JavaScript and TypeScript applications with advanced caching, offline support, and framework integrations.

## 🚀 Quick Start

```bash
npm install flexflag-client
```

```javascript
import { FlexFlagClient } from 'flexflag-client';

const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://your-flexflag-server.com',
  environment: 'production'
});

await client.waitForReady();
const isEnabled = await client.evaluate('new-feature', false);
```

## 📚 Documentation

### Getting Started
- [**Getting Started**](./getting-started.md) - Installation, basic usage, and configuration
- [**API Reference**](./api-reference.md) - Complete API documentation

### Framework Integrations
- [**React Integration**](./react-integration.md) - Hooks, components, and React patterns
- [**Vue Integration**](./vue-integration.md) - Composables, components, and Vue patterns

### Advanced Topics
- [**Advanced Usage**](./advanced-usage.md) - Batch evaluation, metrics, custom cache providers
- [**Performance Guide**](./performance.md) - Optimization tips and best practices
- [**Migration Guide**](./migration.md) - Upgrading from older versions

## ✨ Key Features

- **🚀 High Performance**: Sub-millisecond flag evaluation with intelligent caching
- **🔄 Real-time Updates**: WebSocket/SSE support for instant flag updates
- **📱 Offline Support**: Works offline with cached flags and localStorage persistence
- **⚛️ React Integration**: Hooks and components for seamless React integration
- **🔧 Vue Integration**: Composables and components for Vue 3 applications
- **📊 Performance Metrics**: Built-in analytics and performance monitoring
- **🎯 Advanced Targeting**: User segments, rollouts, and A/B testing support
- **🛡️ TypeScript Support**: Full type definitions included
- **🔧 Configurable**: Extensive configuration options for all use cases

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Your App      │    │  FlexFlag SDK   │
│                 │    │                 │
│  ┌─────────────┐│    │ ┌─────────────┐ │
│  │ React/Vue   ││◄──►│ │   Client    │ │
│  │ Components  ││    │ │             │ │
│  └─────────────┘│    │ └─────────────┘ │
│                 │    │         │       │
│                 │    │ ┌─────────────┐ │
│                 │    │ │    Cache    │ │
│                 │    │ │   Provider  │ │
│                 │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘
                                │
                                ▼
                    ┌─────────────────────┐
                    │   FlexFlag Server   │
                    │                     │
                    │  WebSocket/HTTP     │
                    │     API             │
                    └─────────────────────┘
```

## 🎯 Use Cases

### A/B Testing
```javascript
const variant = await client.getVariation('checkout-flow', context);
if (variant === 'new') {
  showNewCheckout();
} else {
  showOldCheckout();
}
```

### Feature Rollouts
```javascript
const isEnabled = await client.evaluate('beta-feature', false, {
  userId: user.id,
  attributes: { plan: user.plan }
});
```

### Configuration Management
```javascript
const config = await client.evaluate('app-config', {
  theme: 'light',
  maxItems: 10
});
```

### Kill Switches
```javascript
const allowPayments = await client.evaluate('payments-enabled', true);
if (!allowPayments) {
  showMaintenanceMessage();
  return;
}
```

## 🛠️ Installation & Setup

### NPM
```bash
npm install flexflag-client
```

### Yarn
```bash
yarn add flexflag-client
```

### CDN
```html
<script src="https://cdn.jsdelivr.net/npm/flexflag-client@latest/dist/index.js"></script>
```

## ⚙️ Configuration Examples

### Basic Configuration
```javascript
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.yourapp.com',
  environment: 'production'
});
```

### Advanced Configuration
```javascript
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.yourapp.com',
  environment: 'production',
  
  cache: {
    storage: 'localStorage',
    ttl: 300000,
    maxSize: 1000
  },
  
  connection: {
    mode: 'streaming',
    retryAttempts: 3
  },
  
  offline: {
    enabled: true,
    defaultFlags: {
      'feature-1': true,
      'feature-2': false
    }
  },
  
  logging: {
    level: 'warn'
  }
});
```

## 📊 Performance Benchmarks

- **Cache Hit**: ~0.1ms average response time
- **Cache Miss**: ~2-5ms average response time  
- **Batch Evaluation**: ~0.5ms per flag
- **Memory Usage**: <1MB baseline
- **Bundle Size**: ~28KB gzipped

## 🔗 Links

- **NPM Package**: [npmjs.com/package/flexflag-client](https://www.npmjs.com/package/flexflag-client)
- **GitHub Repository**: [github.com/flexflag/flexflag](https://github.com/flexflag/flexflag)
- **Issues & Support**: [github.com/flexflag/flexflag/issues](https://github.com/flexflag/flexflag/issues)
- **FlexFlag Server**: [github.com/flexflag/flexflag](https://github.com/flexflag/flexflag)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../../CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](../../../LICENSE) for details.

---

**Need help?** 
- 📖 Check the [API Reference](./api-reference.md)
- 🐛 [Report issues](https://github.com/flexflag/flexflag/issues)
- 💬 [Join our Discord](https://discord.gg/fpewTJyx9S)
- 🗣️ [GitHub Discussions](https://github.com/flexflag/flexflag/discussions)