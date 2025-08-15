/**
 * FlexFlag React Integration Example
 */

import React, { useState } from 'react';
import {
  FlexFlagProvider,
  useFeatureFlag,
  useBooleanFlag,
  useStringFlag,
  useVariation,
  FeatureGate,
  withFeatureFlag
} from '@flexflag/client/react';

// App wrapper with FlexFlag provider
export default function App() {
  const [userId, setUserId] = useState('user_123');
  const [plan, setPlan] = useState('free');

  return (
    <FlexFlagProvider
      config={{
        apiKey: 'ff_production_your_api_key_here',
        baseUrl: 'http://localhost:8080',
        environment: 'production',
        cache: {
          enabled: true,
          ttl: 300000, // 5 minutes
          storage: 'localStorage'
        },
        events: {
          onReady: () => console.log('🚀 FlexFlag ready in React!'),
          onCacheHit: (flagKey) => console.log(`⚡ Cache hit: ${flagKey}`)
        }
      }}
      context={{
        userId,
        attributes: {
          plan,
          region: 'us-east',
          signupDate: '2023-01-15'
        }
      }}
    >
      <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
        <h1>🎛️ FlexFlag React Example</h1>
        
        {/* User controls */}
        <div style={{ marginBottom: '20px', padding: '15px', backgroundColor: '#f5f5f5', borderRadius: '8px' }}>
          <h3>👤 User Context</h3>
          <label>
            User ID: 
            <input 
              type="text" 
              value={userId} 
              onChange={(e) => setUserId(e.target.value)}
              style={{ marginLeft: '10px', padding: '5px' }}
            />
          </label>
          <br /><br />
          <label>
            Plan: 
            <select 
              value={plan} 
              onChange={(e) => setPlan(e.target.value)}
              style={{ marginLeft: '10px', padding: '5px' }}
            >
              <option value="free">Free</option>
              <option value="premium">Premium</option>
              <option value="enterprise">Enterprise</option>
            </select>
          </label>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '20px' }}>
          {/* Boolean flag example */}
          <FeatureFlagCard title="🌙 Dark Mode Feature">
            <DarkModeToggle />
          </FeatureFlagCard>

          {/* String flag example */}
          <FeatureFlagCard title="🎨 UI Theme">
            <UIThemeSelector />
          </FeatureFlagCard>

          {/* Conditional rendering example */}
          <FeatureFlagCard title="💎 Premium Features">
            <PremiumFeaturesSection />
          </FeatureFlagCard>

          {/* A/B testing example */}
          <FeatureFlagCard title="🧪 A/B Test Checkout">
            <CheckoutVariation />
          </FeatureFlagCard>

          {/* Feature gating example */}
          <FeatureFlagCard title="🚀 Beta Features">
            <BetaFeaturesSection />
          </FeatureFlagCard>
        </div>
      </div>
    </FlexFlagProvider>
  );
}

// Individual components using FlexFlag hooks

function DarkModeToggle() {
  const { enabled, loading, error } = useBooleanFlag('dark-mode', false);

  if (loading) return <div>🔄 Loading dark mode setting...</div>;
  if (error) return <div>❌ Error: {error.message}</div>;

  return (
    <div>
      <p>Dark mode is currently: <strong>{enabled ? 'ON' : 'OFF'}</strong></p>
      <div style={{ 
        padding: '10px', 
        backgroundColor: enabled ? '#333' : '#fff',
        color: enabled ? '#fff' : '#333',
        border: '1px solid #ccc',
        borderRadius: '4px'
      }}>
        {enabled ? '🌙 Dark theme active' : '☀️ Light theme active'}
      </div>
    </div>
  );
}

function UIThemeSelector() {
  const { value: theme, loading, error } = useStringFlag('ui-theme', 'default');

  if (loading) return <div>🔄 Loading theme...</div>;
  if (error) return <div>❌ Error: {error.message}</div>;

  const themes = {
    default: { bg: '#f0f0f0', text: '#333' },
    blue: { bg: '#e3f2fd', text: '#1976d2' },
    green: { bg: '#e8f5e8', text: '#4caf50' },
    purple: { bg: '#f3e5f5', text: '#9c27b0' }
  };

  const currentTheme = themes[theme as keyof typeof themes] || themes.default;

  return (
    <div>
      <p>Current theme: <strong>{theme}</strong></p>
      <div style={{
        padding: '15px',
        backgroundColor: currentTheme.bg,
        color: currentTheme.text,
        borderRadius: '4px',
        textAlign: 'center'
      }}>
        🎨 Theme Preview: {theme}
      </div>
    </div>
  );
}

function PremiumFeaturesSection() {
  const { value: premiumEnabled, loading } = useFeatureFlag('premium-features', false);

  if (loading) return <div>🔄 Checking premium access...</div>;

  return (
    <div>
      {premiumEnabled ? (
        <div style={{ color: '#4caf50' }}>
          <h4>✨ Premium Features Unlocked!</h4>
          <ul>
            <li>🚀 Advanced Analytics</li>
            <li>🔧 Custom Integrations</li>
            <li>👥 Team Collaboration</li>
            <li>📞 Priority Support</li>
          </ul>
        </div>
      ) : (
        <div style={{ color: '#ff9800' }}>
          <h4>🔒 Premium Features</h4>
          <p>Upgrade to unlock advanced features!</p>
          <button style={{ 
            padding: '10px 20px', 
            backgroundColor: '#2196f3', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px' 
          }}>
            Upgrade to Premium
          </button>
        </div>
      )}
    </div>
  );
}

function CheckoutVariation() {
  const { variation, loading, error } = useVariation('checkout-flow', 'control');

  if (loading) return <div>🔄 Loading checkout variation...</div>;
  if (error) return <div>❌ Error: {error.message}</div>;

  const variations = {
    control: {
      title: '🛒 Standard Checkout',
      description: 'Traditional checkout process',
      color: '#f5f5f5'
    },
    streamlined: {
      title: '⚡ Quick Checkout',
      description: 'Streamlined one-click process',
      color: '#e8f5e8'
    },
    premium: {
      title: '💎 Premium Checkout',
      description: 'Enhanced checkout with perks',
      color: '#f3e5f5'
    }
  };

  const currentVariation = variations[variation as keyof typeof variations] || variations.control;

  return (
    <div>
      <p>A/B Test Variation: <strong>{variation}</strong></p>
      <div style={{
        padding: '15px',
        backgroundColor: currentVariation.color,
        borderRadius: '4px',
        border: '2px solid #ddd'
      }}>
        <h4>{currentVariation.title}</h4>
        <p>{currentVariation.description}</p>
        <button style={{
          padding: '8px 16px',
          backgroundColor: '#4caf50',
          color: 'white',
          border: 'none',
          borderRadius: '4px'
        }}>
          Proceed with {variation} checkout
        </button>
      </div>
    </div>
  );
}

function BetaFeaturesSection() {
  return (
    <div>
      <h4>🧪 Beta Features</h4>
      
      {/* Using FeatureGate component for conditional rendering */}
      <FeatureGate
        flagKey="beta-features"
        defaultValue={false}
        loading={<div>🔄 Checking beta access...</div>}
        fallback={
          <div style={{ color: '#999' }}>
            <p>Beta features not available for your account.</p>
            <small>Join our beta program to get early access!</small>
          </div>
        }
      >
        <div style={{ 
          padding: '15px', 
          backgroundColor: '#fff3e0', 
          border: '2px dashed #ff9800',
          borderRadius: '4px'
        }}>
          <h5>🚧 Experimental Features</h5>
          <ul>
            <li>🤖 AI-Powered Insights</li>
            <li>📊 Real-time Dashboard</li>
            <li>🔍 Advanced Search</li>
          </ul>
          <small style={{ color: '#f57c00' }}>
            ⚠️ These features are in beta and may change
          </small>
        </div>
      </FeatureGate>
    </div>
  );
}

// Higher-order component example
const EnhancedFeatures = withFeatureFlag('enhanced-ui', {
  fallback: ({ title }: { title: string }) => (
    <div style={{ padding: '15px', backgroundColor: '#f5f5f5' }}>
      <h4>{title}</h4>
      <p>🔒 Enhanced features not available</p>
    </div>
  )
})(({ title }: { title: string }) => (
  <div style={{ 
    padding: '15px', 
    backgroundColor: '#e8f5e8',
    border: '2px solid #4caf50',
    borderRadius: '4px'
  }}>
    <h4>{title} ✨</h4>
    <p>🚀 Enhanced UI features are active!</p>
  </div>
));

// Helper component for consistent card styling
function FeatureFlagCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div style={{
      padding: '20px',
      border: '1px solid #ddd',
      borderRadius: '8px',
      backgroundColor: '#fff',
      boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
    }}>
      <h3 style={{ marginTop: 0, borderBottom: '1px solid #eee', paddingBottom: '10px' }}>
        {title}
      </h3>
      {children}
    </div>
  );
}