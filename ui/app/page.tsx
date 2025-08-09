'use client';

import { useState, useEffect } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Chip,
  LinearProgress,
  IconButton,
  Alert,
  Paper,
} from '@mui/material';
import {
  Flag as FlagIcon,
  Speed as SpeedIcon,
  TrendingUp as TrendingUpIcon,
  Refresh as RefreshIcon,
  CheckCircle as CheckCircleIcon,
} from '@mui/icons-material';
import { apiClient } from '@/lib/api';
import { Flag, UltraFastStats, PerformanceStats } from '@/types';

// Stats Card Component
interface StatsCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  color: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
}

function StatsCard({ title, value, subtitle, icon, color }: StatsCardProps) {
  return (
    <Card sx={{ height: '100%', position: 'relative', overflow: 'visible' }}>
      <CardContent sx={{ pb: 2 }}>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Box
            sx={{
              p: 1.5,
              borderRadius: 2,
              bgcolor: `${color}.50`,
              color: `${color}.main`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {icon}
          </Box>
        </Box>
        <Typography variant="h4" fontWeight="bold" color="text.primary" gutterBottom>
          {value}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {title}
        </Typography>
        {subtitle && (
          <Typography variant="caption" color="text.secondary" mt={1}>
            {subtitle}
          </Typography>
        )}
      </CardContent>
    </Card>
  );
}

export default function Dashboard() {
  const [flags, setFlags] = useState<Flag[]>([]);
  const [ultraFastStats, setUltraFastStats] = useState<UltraFastStats | null>(null);
  const [performanceStats, setPerformanceStats] = useState<PerformanceStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Try to fetch flags first
      try {
        const flagsData = await apiClient.getFlags('production');
        setFlags(flagsData);
      } catch (err) {
        console.warn('Failed to load flags:', err);
        setFlags([]); // Set empty array as fallback
      }

      // Try to fetch ultra-fast stats
      try {
        const ultraStats = await apiClient.getUltraFastStats();
        setUltraFastStats(ultraStats);
      } catch (err) {
        console.warn('Failed to load ultra-fast stats:', err);
        setUltraFastStats({
          preloaded_flags: 0,
          cached_responses: 0,
          preload_complete: false,
        });
      }

      // Try to fetch performance stats
      try {
        const perfStats = await apiClient.getCacheStats();
        setPerformanceStats(perfStats);
      } catch (err) {
        console.warn('Failed to load performance stats:', err);
        setPerformanceStats({
          cache_hits: 0,
          cache_misses: 0,
          total_requests: 0,
          average_evaluation_time_ms: 0,
          p95_evaluation_time_ms: 0,
          p99_evaluation_time_ms: 0,
        });
      }
    } catch (err) {
      setError('Failed to load dashboard data');
      console.error('Dashboard error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    
    // Set up auto-refresh every 30 seconds
    const interval = setInterval(fetchDashboardData, 30000);
    return () => clearInterval(interval);
  }, []);

  const enabledFlags = flags.filter((flag: Flag) => flag.enabled);
  const disabledFlags = flags.filter((flag: Flag) => !flag.enabled);

  return (
    <Box>
      {/* Header */}
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={4}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Dashboard
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Monitor your feature flags and system performance
          </Typography>
        </Box>
        <IconButton
          onClick={fetchDashboardData}
          disabled={loading}
          sx={{
            bgcolor: 'primary.50',
            color: 'primary.main',
            '&:hover': { bgcolor: 'primary.100' },
          }}
        >
          <RefreshIcon />
        </IconButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Stats Cards */}
      <Grid container spacing={3} mb={4}>
        <Grid item xs={12} sm={6} lg={3}>
          <StatsCard
            title="Total Flags"
            value={flags.length}
            subtitle={`${enabledFlags.length} enabled, ${disabledFlags.length} disabled`}
            icon={<FlagIcon />}
            color="primary"
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatsCard
            title="Cache Performance"
            value={ultraFastStats?.cached_responses || 0}
            subtitle="Cached responses"
            icon={<SpeedIcon />}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatsCard
            title="Preloaded Flags"
            value={ultraFastStats?.preloaded_flags || 0}
            subtitle={ultraFastStats?.preload_complete ? 'Preload complete' : 'Loading...'}
            icon={<TrendingUpIcon />}
            color="secondary"
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatsCard
            title="System Status"
            value="Healthy"
            subtitle="All systems operational"
            icon={<CheckCircleIcon />}
            color="success"
          />
        </Grid>
      </Grid>

      {/* Recent Flags */}
      <Grid container spacing={3}>
        <Grid item xs={12} lg={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" fontWeight="600" gutterBottom>
                Recent Flags
              </Typography>
              {loading ? (
                <LinearProgress />
              ) : flags.length === 0 ? (
                <Box textAlign="center" py={4}>
                  <Typography variant="body1" color="text.secondary">
                    No flags found. Create your first feature flag to get started.
                  </Typography>
                </Box>
              ) : (
                <Box>
                  {flags.slice(0, 5).map((flag) => (
                    <Box
                      key={flag.key}
                      display="flex"
                      alignItems="center"
                      justifyContent="space-between"
                      py={2}
                      borderBottom={1}
                      borderColor="divider"
                      sx={{ '&:last-child': { borderBottom: 0 } }}
                    >
                      <Box display="flex" alignItems="center" gap={2}>
                        <Box
                          sx={{
                            width: 8,
                            height: 8,
                            borderRadius: '50%',
                            bgcolor: flag.enabled ? 'success.main' : 'grey.400',
                          }}
                        />
                        <Box>
                          <Typography variant="body1" fontWeight="500">
                            {flag.name}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {flag.key} • {flag.type}
                          </Typography>
                        </Box>
                      </Box>
                      <Chip
                        label={flag.enabled ? 'Enabled' : 'Disabled'}
                        color={flag.enabled ? 'success' : 'default'}
                        size="small"
                      />
                    </Box>
                  ))}
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} lg={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" fontWeight="600" gutterBottom>
                Quick Actions
              </Typography>
              <Box display="flex" flexDirection="column" gap={2}>
                <Paper
                  elevation={0}
                  sx={{
                    p: 2,
                    bgcolor: 'primary.50',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                    '&:hover': { bgcolor: 'primary.100' },
                  }}
                >
                  <Typography variant="body1" fontWeight="500" color="primary.main">
                    Create New Flag
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Add a new feature flag
                  </Typography>
                </Paper>
                <Paper
                  elevation={0}
                  sx={{
                    p: 2,
                    bgcolor: 'secondary.50',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                    '&:hover': { bgcolor: 'secondary.100' },
                  }}
                >
                  <Typography variant="body1" fontWeight="500" color="secondary.main">
                    Test Evaluation
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Test flag evaluation performance
                  </Typography>
                </Paper>
                <Paper
                  elevation={0}
                  sx={{
                    p: 2,
                    bgcolor: 'success.50',
                    cursor: 'pointer',
                    transition: 'all 0.2s',
                    '&:hover': { bgcolor: 'success.100' },
                  }}
                >
                  <Typography variant="body1" fontWeight="500" color="success.main">
                    View Performance
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Detailed performance metrics
                  </Typography>
                </Paper>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}