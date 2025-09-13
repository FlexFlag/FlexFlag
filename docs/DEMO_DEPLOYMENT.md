# FlexFlag Demo Deployment Guide

This guide explains how to deploy FlexFlag as a live demo with restricted access for showcasing purposes.

## 🎯 Demo Features

- **Restricted Access**: Read-only demo account + admin preview
- **Auto-Reset**: Demo data resets every hour to maintain clean state
- **Pre-loaded Data**: Sample projects, flags, and segments
- **Rate Limited**: Prevents abuse with API rate limiting
- **Monitoring**: Built-in analytics and usage tracking

## 🚀 Deployment Options

### Option 1: Railway (Recommended)

Railway provides the easiest deployment with automatic GitHub integration.

#### Steps:
1. **Sign up** at [railway.app](https://railway.app)
2. **Connect GitHub** repository
3. **Add PostgreSQL database**:
   ```
   railway add postgresql
   ```
4. **Set environment variables**:
   ```
   FLEXFLAG_DEMO_MODE=true
   FLEXFLAG_DEMO_RESET_INTERVAL=1h
   FLEXFLAG_DEMO_MAX_FLAGS=50
   FLEXFLAG_DEMO_MAX_PROJECTS=5
   JWT_SECRET=your-secure-jwt-secret
   ```
5. **Deploy**: Railway auto-deploys from your main branch

#### Railway Configuration (`railway.json`):
Already created in the repository root.

### Option 2: Docker on VPS

Deploy to any VPS (DigitalOcean, Linode, etc.) using Docker Compose.

#### VPS Setup:
```bash
# 1. Create VPS (minimum 2GB RAM, 1 CPU)
# 2. Install Docker and Docker Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 3. Clone repository
git clone https://github.com/FlexFlag/FlexFlag.git
cd FlexFlag

# 4. Create environment file
cat > .env.demo << EOF
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USERNAME=flexflag
DATABASE_PASSWORD=$(openssl rand -base64 32)
DATABASE_NAME=flexflag_demo
REDIS_HOST=redis
REDIS_PORT=6379
JWT_SECRET=$(openssl rand -base64 32)
FLEXFLAG_DEMO_MODE=true
FLEXFLAG_DEMO_RESET_INTERVAL=1h
EOF

# 5. Deploy
docker-compose -f docker-compose.demo.yml up -d
```

### Option 3: Cloud Platform (AWS/GCP/Azure)

For production-ready deployment with auto-scaling.

#### AWS ECS Deployment:
```bash
# Use the provided Dockerfile.demo
# Deploy with:
# - ECS Fargate (2 vCPU, 4GB RAM)
# - RDS PostgreSQL
# - ElastiCache Redis
# - Application Load Balancer
# - Route 53 for DNS
```

## 🔐 Demo Access Control

### Demo User Accounts:
- **Public Demo**: `demo@flexflag.io` / `demo123` (read-only)
- **Admin Preview**: `admin@flexflag.io` / `admin123` (full access, resets hourly)

### Restrictions:
- **Rate Limiting**: 100 requests/minute per IP
- **Resource Limits**: Max 5 projects, 50 flags per project
- **Auto-Reset**: All data resets every hour
- **API Restrictions**: Some endpoints disabled in demo mode

## 🎨 Demo Customization

### Environment Variables:
```bash
# Demo mode settings
FLEXFLAG_DEMO_MODE=true
FLEXFLAG_DEMO_RESET_INTERVAL=1h
FLEXFLAG_DEMO_MAX_FLAGS=50
FLEXFLAG_DEMO_MAX_PROJECTS=5

# Custom branding
FLEXFLAG_DEMO_TITLE="FlexFlag Demo"
FLEXFLAG_DEMO_SUBTITLE="Experience high-performance feature flags"
FLEXFLAG_DEMO_BANNER=true

# Analytics
FLEXFLAG_ANALYTICS_ENABLED=true
FLEXFLAG_ANALYTICS_KEY=your-analytics-key
```

### Custom Demo Data:
Edit `docker/demo-data.sql` to customize:
- Sample projects and flags
- User segments
- Demo accounts
- Feature showcase examples

## 📊 Monitoring & Analytics

### Built-in Metrics:
- Page views and user interactions
- API usage statistics  
- Performance metrics
- Error rates

### Custom Analytics:
```javascript
// Add to UI for tracking
gtag('event', 'demo_action', {
  'event_category': 'engagement',
  'event_label': 'flag_created'
});
```

## 🌐 Custom Domain Setup

### DNS Configuration:
```
Type: A
Name: demo
Value: YOUR_SERVER_IP
TTL: 300
```

### SSL Certificate (Let's Encrypt):
```bash
# Using Traefik (included in docker-compose.demo.yml)
# SSL certificates are automatically generated
```

## 📱 Demo Features to Highlight

### 1. Real-time Flag Updates
- Toggle flags and see instant updates
- Demonstrate WebSocket/SSE connectivity

### 2. A/B Testing
- Show percentage rollouts
- User targeting and segments

### 3. Performance Monitoring
- Display evaluation metrics
- Show sub-millisecond response times

### 4. Multi-environment Support
- Production/Staging/Development environments
- Environment-specific flag values

### 5. SDK Integration
- Live code examples
- Real-time flag evaluation

## 🔒 Security Best Practices

### Demo Environment:
- **Isolated Database**: Separate from production
- **Limited Permissions**: Demo users can't access sensitive data
- **Rate Limiting**: Prevent abuse and DDoS
- **Auto-cleanup**: Regular data purging
- **Monitoring**: Track unusual activity

### Production Considerations:
- **HTTPS Only**: Force SSL/TLS
- **CORS Configuration**: Restrict origins
- **API Authentication**: Require valid keys
- **Input Validation**: Sanitize all inputs

## 🚀 Deployment Checklist

- [ ] Choose deployment platform
- [ ] Set up database (PostgreSQL)
- [ ] Configure Redis for caching
- [ ] Set environment variables
- [ ] Deploy application
- [ ] Configure domain/SSL
- [ ] Test demo functionality
- [ ] Set up monitoring
- [ ] Configure auto-reset
- [ ] Add analytics tracking

## 📞 Demo URLs

Once deployed, your demo will be available at:
- **Railway**: `https://your-app.railway.app`
- **Custom Domain**: `https://demo.flexflag.io`
- **VPS**: `https://your-domain.com`

## 🎯 Marketing Integration

### Landing Page Features:
```html
<div class="demo-cta">
  <h2>Try FlexFlag Live Demo</h2>
  <p>Experience sub-millisecond feature flag evaluation</p>
  <a href="https://demo.flexflag.io" class="demo-button">
    Launch Interactive Demo
  </a>
</div>
```

### Social Proof:
- Real performance metrics
- Usage statistics
- Developer testimonials

---

## 📋 Quick Start Commands

### Railway:
```bash
# Connect to Railway
railway login
railway link
railway up
```

### Docker (Local Testing):
```bash
docker-compose -f docker-compose.demo.yml up
```

### Production VPS:
```bash
git clone https://github.com/FlexFlag/FlexFlag.git
cd FlexFlag
./deploy-demo.sh
```

Your FlexFlag demo will showcase the power of high-performance feature flags with a professional, interactive experience! 🎉