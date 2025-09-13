# Demo Mode Implementation Guide

This guide explains the demo mode features that need to be implemented in FlexFlag for the live demo deployment.

## 🎯 Demo Mode Features to Implement

### 1. Demo Mode Configuration

Add these configuration options to `internal/config/config.go`:

```go
type DemoConfig struct {
    Enabled         bool          `mapstructure:"enabled" yaml:"enabled"`
    ResetInterval   time.Duration `mapstructure:"reset_interval" yaml:"reset_interval"`
    MaxFlags        int          `mapstructure:"max_flags" yaml:"max_flags"`
    MaxProjects     int          `mapstructure:"max_projects" yaml:"max_projects"`
    Title           string       `mapstructure:"title" yaml:"title"`
    Subtitle        string       `mapstructure:"subtitle" yaml:"subtitle"`
    ShowBanner      bool         `mapstructure:"show_banner" yaml:"show_banner"`
}

type Config struct {
    // ... existing fields
    Demo DemoConfig `mapstructure:"demo" yaml:"demo"`
}
```

### 2. Demo Middleware

Create `internal/api/middleware/demo.go`:

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func DemoMode(demoConfig DemoConfig) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        if demoConfig.Enabled {
            // Add demo headers
            c.Header("X-Demo-Mode", "true")
            c.Header("X-Demo-Reset-Interval", demoConfig.ResetInterval.String())
            
            // Block destructive operations for demo users
            if isDemoUser(c) && isDestructiveOperation(c) {
                c.JSON(http.StatusForbidden, gin.H{
                    "error": "This action is not allowed in demo mode",
                    "code": "DEMO_RESTRICTED"
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    })
}

func isDemoUser(c *gin.Context) bool {
    // Check if user is demo user based on email or role
    user := getUserFromContext(c)
    return user != nil && strings.Contains(user.Email, "demo@")
}

func isDestructiveOperation(c *gin.Context) bool {
    method := c.Request.Method
    path := c.Request.URL.Path
    
    // Block certain operations for demo users
    destructiveOperations := []string{
        "DELETE",
        "PUT /api/v1/users",
        "DELETE /api/v1/projects",
    }
    
    for _, op := range destructiveOperations {
        if method == op || (method+" "+path) == op {
            return true
        }
    }
    
    return false
}
```

### 3. Rate Limiting

Create `internal/api/middleware/ratelimit.go`:

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "net/http"
    "sync"
)

var (
    visitors = make(map[string]*rate.Limiter)
    mu       sync.RWMutex
)

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        ip := c.ClientIP()
        
        mu.Lock()
        if _, exists := visitors[ip]; !exists {
            visitors[ip] = rate.NewLimiter(rate.Limit(requestsPerMinute)/60, requestsPerMinute)
        }
        limiter := visitors[ip]
        mu.Unlock()
        
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded. Please try again later.",
                "code": "RATE_LIMITED"
            })
            c.Abort()
            return
        }
        
        c.Next()
    })
}
```

### 4. Demo Data Reset Service

Create `internal/services/demo_service.go`:

```go
package services

import (
    "context"
    "database/sql"
    "time"
)

type DemoService struct {
    db     *sql.DB
    config DemoConfig
    ticker *time.Ticker
}

func NewDemoService(db *sql.DB, config DemoConfig) *DemoService {
    return &DemoService{
        db:     db,
        config: config,
    }
}

func (ds *DemoService) Start(ctx context.Context) {
    if !ds.config.Enabled {
        return
    }
    
    ds.ticker = time.NewTicker(ds.config.ResetInterval)
    
    go func() {
        for {
            select {
            case <-ctx.Done():
                ds.ticker.Stop()
                return
            case <-ds.ticker.C:
                ds.resetDemoData(ctx)
            }
        }
    }()
}

func (ds *DemoService) resetDemoData(ctx context.Context) error {
    tx, err := ds.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Reset demo data - keep demo users but reset their projects/flags
    queries := []string{
        "DELETE FROM flags WHERE project_id IN (SELECT id FROM projects WHERE created_by IN (SELECT id FROM users WHERE email LIKE '%demo%'))",
        "DELETE FROM projects WHERE created_by IN (SELECT id FROM users WHERE email LIKE '%demo%')",
        "DELETE FROM audit_logs WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%demo%')",
    }
    
    for _, query := range queries {
        if _, err := tx.ExecContext(ctx, query); err != nil {
            return err
        }
    }
    
    // Re-insert demo data
    if err := ds.insertDemoData(ctx, tx); err != nil {
        return err
    }
    
    return tx.Commit()
}

func (ds *DemoService) insertDemoData(ctx context.Context, tx *sql.Tx) error {
    // Insert demo projects and flags
    // This would contain the same data as docker/demo-data.sql
    return nil
}
```

### 5. Demo Banner Component (UI)

Create `ui/components/DemoBanner.tsx`:

```tsx
import React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { InfoIcon, RefreshCw } from 'lucide-react';

interface DemoBannerProps {
  resetInterval?: string;
  nextReset?: string;
}

export function DemoBanner({ resetInterval, nextReset }: DemoBannerProps) {
  if (!process.env.NEXT_PUBLIC_DEMO_MODE) {
    return null;
  }

  return (
    <Alert className="mb-4 border-blue-200 bg-blue-50">
      <InfoIcon className="h-4 w-4" />
      <AlertDescription className="flex items-center justify-between">
        <span>
          <strong>Demo Mode:</strong> This is a live demo of FlexFlag. 
          Data resets every {resetInterval} for a fresh experience.
        </span>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <RefreshCw className="h-3 w-3" />
          Next reset: {nextReset}
        </div>
      </AlertDescription>
    </Alert>
  );
}
```

### 6. Demo Analytics

Create `internal/analytics/demo.go`:

```go
package analytics

import (
    "context"
    "encoding/json"
    "net/http"
)

type DemoAnalytics struct {
    endpoint string
    apiKey   string
}

type DemoEvent struct {
    Event      string            `json:"event"`
    Properties map[string]any    `json:"properties"`
    Timestamp  string           `json:"timestamp"`
}

func (da *DemoAnalytics) Track(event string, properties map[string]any) {
    if da.endpoint == "" {
        return
    }
    
    demoEvent := DemoEvent{
        Event:      event,
        Properties: properties,
        Timestamp:  time.Now().UTC().Format(time.RFC3339),
    }
    
    go da.sendEvent(demoEvent)
}

func (da *DemoAnalytics) sendEvent(event DemoEvent) {
    data, _ := json.Marshal(event)
    
    req, _ := http.NewRequest("POST", da.endpoint, bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+da.apiKey)
    
    client := &http.Client{Timeout: 5 * time.Second}
    client.Do(req)
}
```

## 🔧 Integration Points

### 1. Update Main Server

In `cmd/server/main.go`:

```go
func main() {
    // ... existing code
    
    // Demo mode setup
    if cfg.Demo.Enabled {
        // Start demo data reset service
        demoService := services.NewDemoService(db, cfg.Demo)
        demoService.Start(ctx)
        
        // Add demo middleware
        router.Use(middleware.DemoMode(cfg.Demo))
        router.Use(middleware.RateLimit(100)) // 100 requests per minute
    }
    
    // ... rest of setup
}
```

### 2. Update UI Configuration

In `ui/next.config.js`:

```javascript
const nextConfig = {
  env: {
    NEXT_PUBLIC_DEMO_MODE: process.env.FLEXFLAG_DEMO_MODE || 'false',
    NEXT_PUBLIC_DEMO_TITLE: process.env.FLEXFLAG_DEMO_TITLE || 'FlexFlag Demo',
    NEXT_PUBLIC_DEMO_SUBTITLE: process.env.FLEXFLAG_DEMO_SUBTITLE || '',
  },
}
```

### 3. Environment Variables

```bash
# Demo mode configuration
FLEXFLAG_DEMO_MODE=true
FLEXFLAG_DEMO_RESET_INTERVAL=1h
FLEXFLAG_DEMO_MAX_FLAGS=50
FLEXFLAG_DEMO_MAX_PROJECTS=5
FLEXFLAG_DEMO_TITLE="FlexFlag Interactive Demo"
FLEXFLAG_DEMO_SUBTITLE="Experience high-performance feature flags"
FLEXFLAG_DEMO_SHOW_BANNER=true

# Analytics (optional)
FLEXFLAG_DEMO_ANALYTICS_ENDPOINT=https://api.analytics.com/track
FLEXFLAG_DEMO_ANALYTICS_KEY=your-analytics-key
```

## 📊 Demo Metrics to Track

- Page views and user interactions
- Feature flag evaluations
- API endpoint usage
- Demo session duration
- Most used features
- Conversion from demo to signup

## 🔒 Security Considerations

1. **Isolated Database**: Demo uses separate database
2. **Rate Limiting**: Prevent abuse
3. **Input Sanitization**: Clean all user inputs
4. **Resource Limits**: Prevent resource exhaustion
5. **Auto-cleanup**: Regular data purging
6. **Monitoring**: Track unusual activity

## 📋 Implementation Checklist

- [ ] Add demo configuration struct
- [ ] Implement demo middleware
- [ ] Create rate limiting middleware
- [ ] Build demo data reset service
- [ ] Add demo banner to UI
- [ ] Set up analytics tracking
- [ ] Update deployment configuration
- [ ] Test demo functionality
- [ ] Configure monitoring
- [ ] Set up automated deployment

This implementation will provide a professional, secure, and engaging demo experience for FlexFlag! 🚀