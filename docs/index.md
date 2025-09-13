# FlexFlag Documentation

Welcome to FlexFlag - High-Performance Feature Flag Management System

## 🚀 Quick Start

FlexFlag is a high-performance, developer-first feature flag management system with distributed edge servers, real-time synchronization, and sub-millisecond flag evaluation.

### Features

- 🚀 **Ultra-Fast Evaluation**: <1ms flag evaluation with edge servers
- 🌐 **Distributed Architecture**: Edge servers for global low-latency access
- ⚡ **Real-time Sync**: SSE/WS-based flag propagation to edge nodes
- 🎯 **Advanced Targeting**: User segments, rollouts, and A/B testing
- 🏢 **Multi-Project Support**: Project isolation with environment management

## 📚 Documentation Sections

### Server Documentation
- [**Getting Started**](./README.md) - Main project documentation
- [**API Documentation**](../api/) - Server API reference
- [**Deployment Guide**](./deployment.md) - Production deployment

### SDK Documentation

#### JavaScript/TypeScript SDK
- [**JavaScript SDK Overview**](./sdks/javascript/) - SDK main documentation
- [**Getting Started**](./sdks/javascript/getting-started.md) - Installation and basic usage
- [**React Integration**](./sdks/javascript/react-integration.md) - React hooks and components
- [**Vue Integration**](./sdks/javascript/vue-integration.md) - Vue composables and components
- [**API Reference**](./sdks/javascript/api-reference.md) - Complete API documentation

## 📦 npm Package

The FlexFlag JavaScript SDK is available on npm:

```bash
npm install flexflag-client
```

**Package**: [flexflag-client](https://www.npmjs.com/package/flexflag-client)

## 🔗 Links

- **GitHub Repository**: [github.com/flexflag/flexflag](https://github.com/flexflag/flexflag)
- **Issues & Support**: [github.com/flexflag/flexflag/issues](https://github.com/flexflag/flexflag/issues)
- **npm Package**: [npmjs.com/package/flexflag-client](https://www.npmjs.com/package/flexflag-client)
- **Discord Community**: [Join our Discord](https://discord.gg/fpewTJyx9S)

## 🏗️ Architecture

FlexFlag uses a modern, distributed architecture designed for performance:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Next.js UI    │    │  Edge Server    │    │  Edge Server    │
│  (Port 3000)    │    │  (Port 8081)    │    │  (Port 8082)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │          SSE/WebSocket│          SSE/WebSocket│
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                     Main FlexFlag Server                        │
│                        (Port 8080)                              │
└─────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
         ┌─────────────────────────────────────────────────────────┐
         │                Infrastructure                           │
         │  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐│
         │  │ PostgreSQL  │     │    Redis    │     │   Docker    ││
         │  │ (Port 5433) │     │ (Port 6379) │     │  Compose    ││
         │  └─────────────┘     └─────────────┘     └─────────────┘│
         └─────────────────────────────────────────────────────────┘
```

## 🤝 Contributing

We welcome contributions! Please check out our contributing guidelines and feel free to submit issues and pull requests.

---

Built with ❤️ by the FlexFlag team