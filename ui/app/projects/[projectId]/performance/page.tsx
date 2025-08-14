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
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  Speed as SpeedIcon,
  Timeline as TimelineIcon,
  Memory as MemoryIcon,
  Refresh as RefreshIcon,
  Assessment as AssessmentIcon,
  TrendingUp as TrendingUpIcon,
  Storage as StorageIcon,
  PlayArrow as PlayIcon,
  ExpandMore as ExpandMoreIcon,
  Analytics as AnalyticsIcon,
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

interface LoadTestResult {
  test_type: string;
  requests: number;
  concurrent_users: number;
  duration_ms: number;
  requests_per_second: number;
  avg_response_time_ms: number;
  p95_response_time_ms: number;
  p99_response_time_ms: number;
  error_rate: number;
  success_count: number;
  error_count: number;
}

export default function ProjectPerformancePage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [performanceStats, setPerformanceStats] = useState<PerformanceStats | null>(null);
  const [ultraFastStats, setUltraFastStats] = useState<UltraFastStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadTestResults, setLoadTestResults] = useState<LoadTestResult[]>([]);
  const [loadTestConfig, setLoadTestConfig] = useState({
    test_type: 'standard',
    requests: 100,
    concurrent_users: 10,
    duration_seconds: 30,
    flag_key: '',
  });
  const [loadTestRunning, setLoadTestRunning] = useState(false);
  const [flags, setFlags] = useState<any[]>([]);

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);
        
        // Fetch flags for load testing
        const flagsData = await apiClient.getFlags(currentEnvironment, projectId);
        setFlags(flagsData || []);
        
        // Set first flag as default if available
        if (flagsData && flagsData.length > 0) {
          setLoadTestConfig(prev => ({ ...prev, flag_key: flagsData[0].key }));
        }
      } catch (error) {
        console.error('Error fetching project:', error);
      }
    };

    if (projectId) {
      fetchProject();
      fetchStats();
    }
  }, [projectId, currentEnvironment]);

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

  const runLoadTest = async () => {
    if (!loadTestConfig.flag_key) {
      alert('Please select a flag to test');
      return;
    }

    try {
      setLoadTestRunning(true);
      
      const startTime = Date.now();
      const results: number[] = [];
      const errors: number[] = [];
      
      // Simulate concurrent load test
      const promises = Array.from({ length: loadTestConfig.concurrent_users }, async (_, userIndex) => {
        const userResults: number[] = [];
        const userErrors: number[] = [];
        
        const requestsPerUser = Math.floor(loadTestConfig.requests / loadTestConfig.concurrent_users);
        
        for (let i = 0; i < requestsPerUser; i++) {
          try {
            const evalStart = Date.now();
            
            const request = {
              flag_key: loadTestConfig.flag_key,
              user_id: `loadtest_user_${userIndex}_${i}`,
              user_key: `loadtest_user_${userIndex}_${i}`,
              attributes: {
                email: `loadtest_user_${userIndex}_${i}@example.com`,
                plan: 'premium',
                loadtest: true,
              },
            };

            let response;
            switch (loadTestConfig.test_type) {
              case 'fast':
                response = await apiClient.evaluateFlagFast(request, currentEnvironment, projectId);
                break;
              case 'ultra':
                response = await apiClient.evaluateFlagUltraFast(request, currentEnvironment, projectId);
                break;
              default:
                response = await apiClient.evaluateFlag(request, currentEnvironment, projectId);
            }
            
            const evalTime = Date.now() - evalStart;
            userResults.push(evalTime);
          } catch (error) {
            userErrors.push(1);
          }
        }
        
        return { results: userResults, errors: userErrors };
      });

      const allResults = await Promise.all(promises);
      
      // Aggregate results
      allResults.forEach(({ results: userResults, errors: userErrors }) => {
        results.push(...userResults);
        errors.push(...userErrors);
      });

      const endTime = Date.now();
      const totalDuration = endTime - startTime;
      
      // Calculate statistics
      const sortedResults = results.sort((a, b) => a - b);
      const avgResponseTime = results.reduce((sum, time) => sum + time, 0) / results.length;
      const p95Index = Math.floor(sortedResults.length * 0.95);
      const p99Index = Math.floor(sortedResults.length * 0.99);
      
      const testResult: LoadTestResult = {
        test_type: loadTestConfig.test_type,
        requests: loadTestConfig.requests,
        concurrent_users: loadTestConfig.concurrent_users,
        duration_ms: totalDuration,
        requests_per_second: (loadTestConfig.requests / totalDuration) * 1000,
        avg_response_time_ms: avgResponseTime,
        p95_response_time_ms: sortedResults[p95Index] || 0,
        p99_response_time_ms: sortedResults[p99Index] || 0,
        error_rate: (errors.length / loadTestConfig.requests) * 100,
        success_count: results.length,
        error_count: errors.length,
      };

      setLoadTestResults(prev => [testResult, ...prev.slice(0, 4)]); // Keep last 5 results
      
    } catch (error) {
      console.error('Load test failed:', error);
      alert('Load test failed. Please try again.');
    } finally {
      setLoadTestRunning(false);
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

      {/* Load Testing */}
      <Box sx={{ mt: 4 }}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <AnalyticsIcon />
              Load Testing
            </Typography>
            
            <Grid container spacing={3}>
              {/* Load Test Configuration */}
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" gutterBottom>
                  Test Configuration
                </Typography>
                
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <FormControl fullWidth>
                      <InputLabel>Flag to Test</InputLabel>
                      <Select
                        value={loadTestConfig.flag_key}
                        onChange={(e) => setLoadTestConfig(prev => ({ ...prev, flag_key: e.target.value }))}
                        label="Flag to Test"
                      >
                        {flags.map((flag) => (
                          <MenuItem key={flag.key} value={flag.key}>
                            {flag.name} ({flag.key})
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <FormControl fullWidth>
                      <InputLabel>Test Type</InputLabel>
                      <Select
                        value={loadTestConfig.test_type}
                        onChange={(e) => setLoadTestConfig(prev => ({ ...prev, test_type: e.target.value }))}
                        label="Test Type"
                      >
                        <MenuItem value="standard">Standard Evaluation</MenuItem>
                        <MenuItem value="fast">Fast Evaluation (Cached)</MenuItem>
                        <MenuItem value="ultra">Ultra Fast (Memory)</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>
                  
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      label="Total Requests"
                      type="number"
                      value={loadTestConfig.requests}
                      onChange={(e) => setLoadTestConfig(prev => ({ ...prev, requests: parseInt(e.target.value) || 100 }))}
                      inputProps={{ min: 1, max: 10000 }}
                    />
                  </Grid>
                  
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      label="Concurrent Users"
                      type="number"
                      value={loadTestConfig.concurrent_users}
                      onChange={(e) => setLoadTestConfig(prev => ({ ...prev, concurrent_users: parseInt(e.target.value) || 10 }))}
                      inputProps={{ min: 1, max: 100 }}
                    />
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Button
                      variant="contained"
                      fullWidth
                      startIcon={loadTestRunning ? <CircularProgress size={20} /> : <PlayIcon />}
                      onClick={runLoadTest}
                      disabled={loadTestRunning || !loadTestConfig.flag_key}
                      size="large"
                    >
                      {loadTestRunning ? 'Running Load Test...' : 'Start Load Test'}
                    </Button>
                  </Grid>
                </Grid>
              </Grid>
              
              {/* Load Test Results */}
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" gutterBottom>
                  Recent Test Results
                </Typography>
                
                {loadTestResults.length === 0 ? (
                  <Paper sx={{ p: 3, textAlign: 'center', bgcolor: 'grey.50' }}>
                    <AnalyticsIcon sx={{ fontSize: 32, color: 'grey.300', mb: 1 }} />
                    <Typography color="text.secondary">
                      No load tests run yet
                    </Typography>
                  </Paper>
                ) : (
                  <Box sx={{ maxHeight: '400px', overflow: 'auto' }}>
                    {loadTestResults.map((result, index) => (
                      <Accordion key={index} sx={{ mb: 1 }}>
                        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, width: '100%' }}>
                            <Chip 
                              label={result.test_type.toUpperCase()} 
                              size="small" 
                              color={result.test_type === 'ultra' ? 'success' : result.test_type === 'fast' ? 'warning' : 'primary'}
                            />
                            <Typography variant="body2">
                              {result.requests} reqs, {result.concurrent_users} users
                            </Typography>
                            <Typography variant="body2" color="primary" sx={{ ml: 'auto' }}>
                              {result.avg_response_time_ms.toFixed(2)}ms avg
                            </Typography>
                          </Box>
                        </AccordionSummary>
                        <AccordionDetails>
                          <Grid container spacing={2}>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Requests/sec
                              </Typography>
                              <Typography variant="body2" fontWeight="bold">
                                {result.requests_per_second.toFixed(1)}
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Duration
                              </Typography>
                              <Typography variant="body2" fontWeight="bold">
                                {(result.duration_ms / 1000).toFixed(2)}s
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                P95 Response
                              </Typography>
                              <Typography variant="body2" fontWeight="bold">
                                {result.p95_response_time_ms.toFixed(2)}ms
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                P99 Response
                              </Typography>
                              <Typography variant="body2" fontWeight="bold">
                                {result.p99_response_time_ms.toFixed(2)}ms
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Success Rate
                              </Typography>
                              <Typography variant="body2" fontWeight="bold" color="success.main">
                                {((result.success_count / result.requests) * 100).toFixed(1)}%
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Errors
                              </Typography>
                              <Typography variant="body2" fontWeight="bold" color={result.error_count > 0 ? 'error.main' : 'success.main'}>
                                {result.error_count}
                              </Typography>
                            </Grid>
                          </Grid>
                        </AccordionDetails>
                      </Accordion>
                    ))}
                  </Box>
                )}
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
}