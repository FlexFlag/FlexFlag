'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Paper,
  Chip,
  Button,
  Alert,
  LinearProgress,
} from '@mui/material';
import {
  Speed as SpeedIcon,
  Timeline as TimelineIcon,
  Memory as MemoryIcon,
  Refresh as RefreshIcon,
  Assessment as AssessmentIcon,
  TrendingUp as TrendingUpIcon,
  Storage as StorageIcon,
} from '@mui/icons-material';

interface PerformanceStats {
  cache_hits?: number;
  cache_misses?: number;
  total_requests?: number;
  average_evaluation_time_ms?: number;
  p95_evaluation_time_ms?: number;
  p99_evaluation_time_ms?: number;
}

interface UltraFastStats {
  cached_flags?: number;
  total_evaluations?: number;
  cache_hit_rate?: number;
  average_response_time_ns?: number;
}

export default function ProjectPerformancePage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [performanceStats, setPerformanceStats] = useState<PerformanceStats | null>(null);
  const [ultraFastStats, setUltraFastStats] = useState<UltraFastStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);
      } catch (error) {
        console.error('Error fetching project:', error);
      }
    };

    if (projectId) {
      fetchProject();
      fetchStats();
    }
  }, [projectId]);

  const fetchStats = async () => {
    try {
      setLoading(true);
      
      // Fetch performance stats
      try {
        const perfStats = await apiClient.getCacheStats();
        setPerformanceStats(perfStats);
      } catch (error) {
        console.error('Error fetching performance stats:', error);
      }

      // Fetch ultra-fast stats
      try {
        const ultraStats = await apiClient.getUltraFastStats();
        setUltraFastStats(ultraStats);
      } catch (error) {
        console.error('Error fetching ultra-fast stats:', error);
      }
    } catch (error) {
      console.error('Error fetching stats:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClearCache = async () => {
    try {
      await apiClient.clearCache();
      fetchStats(); // Refresh stats
    } catch (error) {
      console.error('Error clearing cache:', error);
    }
  };

  if (!project) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <Typography>Loading project...</Typography>
      </Box>
    );
  }

  const cacheHitRate = performanceStats?.total_requests 
    ? ((performanceStats.cache_hits || 0) / performanceStats.total_requests * 100)
    : 0;

  return (
    <Box>
      {/* Header */}
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Performance
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Monitor evaluation performance and caching for {project.name}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={fetchStats}
            disabled={loading}
          >
            Refresh
          </Button>
          <Button
            variant="outlined"
            startIcon={<StorageIcon />}
            onClick={handleClearCache}
            color="warning"
          >
            Clear Cache
          </Button>
        </Box>
      </Box>

      {/* Performance Overview */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <SpeedIcon sx={{ fontSize: 32, color: 'primary.main', mb: 1 }} />
            <Typography variant="h4" fontWeight="bold">
              {performanceStats?.average_evaluation_time_ms?.toFixed(2) || '-'}ms
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Avg Response Time
            </Typography>
          </Paper>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <AssessmentIcon sx={{ fontSize: 32, color: 'success.main', mb: 1 }} />
            <Typography variant="h4" fontWeight="bold">
              {performanceStats?.total_requests || 0}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Total Requests
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <MemoryIcon sx={{ fontSize: 32, color: 'secondary.main', mb: 1 }} />
            <Typography variant="h4" fontWeight="bold">
              {cacheHitRate.toFixed(1)}%
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Cache Hit Rate
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <TrendingUpIcon sx={{ fontSize: 32, color: 'warning.main', mb: 1 }} />
            <Typography variant="h4" fontWeight="bold">
              {ultraFastStats?.cached_flags || 0}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Cached Flags
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Detailed Performance Metrics */}
      <Grid container spacing={3}>
        {/* Standard Evaluation Performance */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <SpeedIcon />
                Standard Evaluation
              </Typography>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Box sx={{ mb: 2 }}>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Cache Hit Rate: {cacheHitRate.toFixed(1)}%
                    </Typography>
                    <LinearProgress
                      variant="determinate"
                      value={cacheHitRate}
                      sx={{ height: 8, borderRadius: 4 }}
                    />
                  </Box>
                </Grid>

                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Cache Hits
                  </Typography>
                  <Typography variant="h6" color="success.main">
                    {performanceStats?.cache_hits || 0}
                  </Typography>
                </Grid>

                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Cache Misses
                  </Typography>
                  <Typography variant="h6" color="error.main">
                    {performanceStats?.cache_misses || 0}
                  </Typography>
                </Grid>

                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    P95 Response Time
                  </Typography>
                  <Typography variant="h6">
                    {performanceStats?.p95_evaluation_time_ms?.toFixed(2) || '-'}ms
                  </Typography>
                </Grid>

                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    P99 Response Time
                  </Typography>
                  <Typography variant="h6">
                    {performanceStats?.p99_evaluation_time_ms?.toFixed(2) || '-'}ms
                  </Typography>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Ultra-Fast Evaluation Performance */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <MemoryIcon />
                Ultra-Fast Evaluation
              </Typography>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    Total Evaluations
                  </Typography>
                  <Typography variant="h6">
                    {ultraFastStats?.total_evaluations || 0}
                  </Typography>
                </Grid>

                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    Cached Flags
                  </Typography>
                  <Typography variant="h6">
                    {ultraFastStats?.cached_flags || 0}
                  </Typography>
                </Grid>

                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    Cache Hit Rate
                  </Typography>
                  <Typography variant="h6">
                    {ultraFastStats?.cache_hit_rate ? 
                      `${(ultraFastStats.cache_hit_rate * 100).toFixed(1)}%` : '-'
                    }
                  </Typography>
                </Grid>

                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    Avg Response Time
                  </Typography>
                  <Typography variant="h6">
                    {ultraFastStats?.average_response_time_ns ? 
                      `${(ultraFastStats.average_response_time_ns / 1000000).toFixed(2)}ms` : '-'
                    }
                  </Typography>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Performance Tips */}
      <Box sx={{ mt: 4 }}>
        <Alert severity="info">
          <Typography variant="subtitle2" gutterBottom>
            Performance Tips:
          </Typography>
          <ul style={{ margin: 0, paddingLeft: '20px' }}>
            <li>Use Ultra-Fast evaluation for high-frequency requests</li>
            <li>Cache hit rates above 80% indicate good performance</li>
            <li>Clear cache after significant flag configuration changes</li>
            <li>Monitor P95/P99 response times for user experience impact</li>
          </ul>
        </Alert>
      </Box>

      {/* Coming Soon */}
      <Box sx={{ mt: 4 }}>
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 3 }}>
            <TimelineIcon sx={{ fontSize: 48, color: 'grey.300', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>
              Advanced Analytics Coming Soon
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Real-time performance dashboards, alerts, and historical trend analysis
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center', flexWrap: 'wrap' }}>
              <Chip label="Real-time Dashboards" variant="outlined" size="small" />
              <Chip label="Performance Alerts" variant="outlined" size="small" />
              <Chip label="Historical Trends" variant="outlined" size="small" />
              <Chip label="Custom Metrics" variant="outlined" size="small" />
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
}