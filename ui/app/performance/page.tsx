'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  Chip,
  LinearProgress,
  Paper,
  Alert,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Tab,
  Select,
  FormControl,
  InputLabel,
  MenuItem,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  TrendingUp as TrendingUpIcon,
  ClearAll as ClearAllIcon,
  PlayArrow as PlayArrowIcon,
} from '@mui/icons-material';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, LineChart, Line, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { apiClient } from '@/lib/api';
import { PerformanceStats, UltraFastStats, EvaluationRequest } from '@/types';

interface PerformanceMetric {
  label: string;
  value: number;
  unit: string;
  color: 'success' | 'warning' | 'error' | 'primary';
  benchmark: number;
}

function MetricCard({ metric }: { metric: PerformanceMetric }) {
  const getPerformanceLevel = (value: number, benchmark: number) => {
    if (value <= benchmark) return 'excellent';
    if (value <= benchmark * 2) return 'good';
    if (value <= benchmark * 5) return 'warning';
    return 'poor';
  };

  const level = getPerformanceLevel(metric.value, metric.benchmark);
  
  const levelColors = {
    excellent: 'success.main',
    good: 'primary.main',
    warning: 'warning.main',
    poor: 'error.main',
  };

  const levelLabels = {
    excellent: 'Excellent',
    good: 'Good',
    warning: 'Needs Attention',
    poor: 'Poor',
  };

  return (
    <Card>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Typography variant="h6" fontWeight="600">
            {metric.label}
          </Typography>
          <Chip
            label={levelLabels[level]}
            size="small"
            sx={{
              bgcolor: levelColors[level],
              color: 'white',
              fontWeight: 600,
            }}
          />
        </Box>
        <Typography variant="h4" fontWeight="bold" color={levelColors[level]} gutterBottom>
          {metric.value.toFixed(3)}{metric.unit}
        </Typography>
        <Box mt={2}>
          <Box display="flex" justifyContent="space-between" mb={1}>
            <Typography variant="caption" color="text.secondary">
              Benchmark: {metric.benchmark}{metric.unit}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {((metric.value / metric.benchmark) * 100).toFixed(0)}% of benchmark
            </Typography>
          </Box>
          <LinearProgress
            variant="determinate"
            value={Math.min((metric.benchmark / metric.value) * 100, 100)}
            sx={{
              height: 6,
              borderRadius: 3,
              bgcolor: 'grey.200',
              '& .MuiLinearProgress-bar': {
                borderRadius: 3,
                bgcolor: levelColors[level],
              },
            }}
          />
        </Box>
      </CardContent>
    </Card>
  );
}

function LoadTestRunner({ onTestComplete }: { onTestComplete: () => void }) {
  const [isRunning, setIsRunning] = useState(false);
  const [testType, setTestType] = useState<'standard' | 'fast' | 'ultra'>('ultra');
  const [requestCount, setRequestCount] = useState(100);
  const [testResults, setTestResults] = useState<any>(null);

  const runLoadTest = async () => {
    setIsRunning(true);
    setTestResults(null);
    
    try {
      const startTime = performance.now();
      const promises: Promise<any>[] = [];
      
      const endpoint = testType === 'standard' ? 'evaluate' : 
                     testType === 'fast' ? 'evaluate/fast' : 'evaluate/ultra';
      
      // Create test requests
      for (let i = 0; i < requestCount; i++) {
        const request: EvaluationRequest = {
          flag_key: 'new-feature',
          user_id: `test-user-${i}`,
          attributes: {
            plan: i % 2 === 0 ? 'premium' : 'basic',
            region: `us-west-${(i % 3) + 1}`,
          },
        };
        
        const promise = testType === 'standard' 
          ? apiClient.evaluateFlag(request)
          : testType === 'fast' 
          ? apiClient.evaluateFlagFast(request)
          : apiClient.evaluateFlagUltraFast(request);
        
        promises.push(promise);
      }
      
      const responses = await Promise.all(promises);
      const endTime = performance.now();
      
      const totalTime = endTime - startTime;
      const evaluationTimes = responses.map(r => r.evaluation_time_ms);
      evaluationTimes.sort((a, b) => a - b);
      
      const results = {
        totalTime,
        totalRequests: requestCount,
        successfulRequests: responses.length,
        averageTime: evaluationTimes.reduce((a, b) => a + b, 0) / evaluationTimes.length,
        medianTime: evaluationTimes[Math.floor(evaluationTimes.length / 2)],
        p95Time: evaluationTimes[Math.floor(evaluationTimes.length * 0.95)],
        p99Time: evaluationTimes[Math.floor(evaluationTimes.length * 0.99)],
        minTime: Math.min(...evaluationTimes),
        maxTime: Math.max(...evaluationTimes),
        throughput: (requestCount / totalTime) * 1000, // requests per second
      };
      
      setTestResults(results);
      onTestComplete();
    } catch (error) {
      console.error('Load test failed:', error);
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" fontWeight="600" gutterBottom>
          Performance Load Test
        </Typography>
        <Typography variant="body2" color="text.secondary" mb={3}>
          Run a load test to measure evaluation performance
        </Typography>
        
        <Grid container spacing={2} mb={3}>
          <Grid item xs={12} sm={4}>
            <FormControl fullWidth>
              <InputLabel>Test Type</InputLabel>
              <Select
                value={testType}
                label="Test Type"
                onChange={(e) => setTestType(e.target.value as any)}
                disabled={isRunning}
              >
                <MenuItem value="standard">Standard Evaluation</MenuItem>
                <MenuItem value="fast">Fast Evaluation (Cached)</MenuItem>
                <MenuItem value="ultra">Ultra-Fast Evaluation</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} sm={4}>
            <FormControl fullWidth>
              <InputLabel>Request Count</InputLabel>
              <Select
                value={requestCount}
                label="Request Count"
                onChange={(e) => setRequestCount(e.target.value as number)}
                disabled={isRunning}
              >
                <MenuItem value={50}>50 requests</MenuItem>
                <MenuItem value={100}>100 requests</MenuItem>
                <MenuItem value={500}>500 requests</MenuItem>
                <MenuItem value={1000}>1000 requests</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} sm={4}>
            <Button
              variant="contained"
              fullWidth
              startIcon={isRunning ? null : <PlayArrowIcon />}
              onClick={runLoadTest}
              disabled={isRunning}
              sx={{ height: '100%' }}
            >
              {isRunning ? 'Running Test...' : 'Start Load Test'}
            </Button>
          </Grid>
        </Grid>

        {isRunning && (
          <Box mb={3}>
            <LinearProgress />
            <Typography variant="body2" color="text.secondary" textAlign="center" mt={1}>
              Running {requestCount} {testType} evaluations...
            </Typography>
          </Box>
        )}

        {testResults && (
          <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
            <Typography variant="subtitle1" fontWeight="600" mb={2}>
              Test Results
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">Average</Typography>
                <Typography variant="h6" fontWeight="bold">
                  {testResults.averageTime.toFixed(3)}ms
                </Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">95th Percentile</Typography>
                <Typography variant="h6" fontWeight="bold">
                  {testResults.p95Time.toFixed(3)}ms
                </Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">99th Percentile</Typography>
                <Typography variant="h6" fontWeight="bold">
                  {testResults.p99Time.toFixed(3)}ms
                </Typography>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Typography variant="body2" color="text.secondary">Throughput</Typography>
                <Typography variant="h6" fontWeight="bold">
                  {testResults.throughput.toFixed(0)} req/s
                </Typography>
              </Grid>
            </Grid>
          </Paper>
        )}
      </CardContent>
    </Card>
  );
}

export default function PerformancePage() {
  const [performanceStats, setPerformanceStats] = useState<PerformanceStats | null>(null);
  const [ultraFastStats, setUltraFastStats] = useState<UltraFastStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);

  const fetchPerformanceData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [perfStats, ultraStats] = await Promise.allSettled([
        apiClient.getCacheStats(),
        apiClient.getUltraFastStats(),
      ]);

      if (perfStats.status === 'fulfilled') {
        setPerformanceStats(perfStats.value);
      }
      if (ultraStats.status === 'fulfilled') {
        setUltraFastStats(ultraStats.value);
      }
    } catch (err) {
      setError('Failed to load performance data');
      console.error('Performance error:', err);
    } finally {
      setLoading(false);
    }
  };

  const clearCache = async () => {
    try {
      await apiClient.clearCache();
      fetchPerformanceData();
    } catch (err) {
      setError('Failed to clear cache');
    }
  };

  useEffect(() => {
    fetchPerformanceData();
    
    // Auto-refresh every 10 seconds
    const interval = setInterval(fetchPerformanceData, 10000);
    return () => clearInterval(interval);
  }, []);

  const metrics: PerformanceMetric[] = performanceStats ? [
    {
      label: 'Average Response Time',
      value: performanceStats.average_evaluation_time_ms,
      unit: 'ms',
      color: 'primary',
      benchmark: 1.0, // 1ms benchmark
    },
    {
      label: '95th Percentile',
      value: performanceStats.p95_evaluation_time_ms,
      unit: 'ms',
      color: 'warning',
      benchmark: 5.0, // 5ms benchmark
    },
    {
      label: '99th Percentile',
      value: performanceStats.p99_evaluation_time_ms,
      unit: 'ms',
      color: 'error',
      benchmark: 10.0, // 10ms benchmark
    },
  ] : [];

  const cacheHitRate = performanceStats && performanceStats.total_requests > 0 ? 
    Math.round((performanceStats.cache_hits / performanceStats.total_requests) * 100) : 0;

  const chartData = [
    {
      name: 'Cache Hits',
      value: performanceStats?.cache_hits || 0,
      fill: '#10b981',
    },
    {
      name: 'Cache Misses',
      value: performanceStats?.cache_misses || 0,
      fill: '#ef4444',
    },
  ];

  return (
    <Box>
      {/* Header */}
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={4}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Performance Monitoring
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Monitor evaluation performance and system metrics
          </Typography>
        </Box>
        <Box display="flex" gap={2}>
          <Button
            variant="outlined"
            startIcon={<ClearAllIcon />}
            onClick={clearCache}
          >
            Clear Cache
          </Button>
          <IconButton
            onClick={fetchPerformanceData}
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
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={tabValue} onChange={(e, newValue) => setTabValue(newValue)}>
          <Tab label="Overview" />
          <Tab label="Load Testing" />
          <Tab label="Cache Analysis" />
        </Tabs>
      </Box>

      {tabValue === 0 && (
        <>
          {/* Key Metrics */}
          <Grid container spacing={3} mb={4}>
            {metrics.map((metric, index) => (
              <Grid item xs={12} md={4} key={index}>
                <MetricCard metric={metric} />
              </Grid>
            ))}
          </Grid>

          {/* System Stats */}
          <Grid container spacing={3} mb={4}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" fontWeight="600" mb={2}>
                    Cache Performance
                  </Typography>
                  {performanceStats && (
                    <Box>
                      <Box display="flex" justifyContent="space-between" mb={2}>
                        <Typography variant="body2">Hit Rate</Typography>
                        <Typography variant="body2" fontWeight="600">
                          {cacheHitRate}%
                        </Typography>
                      </Box>
                      <LinearProgress
                        variant="determinate"
                        value={cacheHitRate}
                        sx={{
                          height: 8,
                          borderRadius: 4,
                          bgcolor: 'grey.200',
                          '& .MuiLinearProgress-bar': {
                            borderRadius: 4,
                            bgcolor: cacheHitRate > 90 ? 'success.main' : 
                                    cacheHitRate > 70 ? 'warning.main' : 'error.main',
                          },
                        }}
                      />
                      <Box display="flex" justifyContent="space-between" mt={2}>
                        <Box textAlign="center">
                          <Typography variant="h6" fontWeight="bold" color="success.main">
                            {performanceStats.cache_hits.toLocaleString()}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            Cache Hits
                          </Typography>
                        </Box>
                        <Box textAlign="center">
                          <Typography variant="h6" fontWeight="bold" color="error.main">
                            {performanceStats.cache_misses.toLocaleString()}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            Cache Misses
                          </Typography>
                        </Box>
                        <Box textAlign="center">
                          <Typography variant="h6" fontWeight="bold" color="primary.main">
                            {performanceStats.total_requests.toLocaleString()}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            Total Requests
                          </Typography>
                        </Box>
                      </Box>
                    </Box>
                  )}
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" fontWeight="600" mb={2}>
                    Ultra-Fast Handler Stats
                  </Typography>
                  {ultraFastStats && (
                    <Box>
                      <Box display="flex" justifyContent="space-between" py={1}>
                        <Typography variant="body2">Preloaded Flags</Typography>
                        <Typography variant="body2" fontWeight="600">
                          {ultraFastStats.preloaded_flags}
                        </Typography>
                      </Box>
                      <Box display="flex" justifyContent="space-between" py={1}>
                        <Typography variant="body2">Cached Responses</Typography>
                        <Typography variant="body2" fontWeight="600">
                          {ultraFastStats.cached_responses.toLocaleString()}
                        </Typography>
                      </Box>
                      <Box display="flex" justifyContent="space-between" py={1}>
                        <Typography variant="body2">Preload Status</Typography>
                        <Chip
                          label={ultraFastStats.preload_complete ? 'Complete' : 'Loading'}
                          color={ultraFastStats.preload_complete ? 'success' : 'warning'}
                          size="small"
                        />
                      </Box>
                    </Box>
                  )}
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </>
      )}

      {tabValue === 1 && (
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <LoadTestRunner onTestComplete={fetchPerformanceData} />
          </Grid>
        </Grid>
      )}

      {tabValue === 2 && performanceStats && (
        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Card>
              <CardContent>
                <Typography variant="h6" fontWeight="600" mb={2}>
                  Cache Hit/Miss Distribution
                </Typography>
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={chartData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({name, percent}) => `${name} ${(percent * 100).toFixed(0)}%`}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {chartData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.fill} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Typography variant="h6" fontWeight="600" mb={2}>
                  Cache Recommendations
                </Typography>
                <Box display="flex" flexDirection="column" gap={2}>
                  {cacheHitRate > 95 && (
                    <Alert severity="success">
                      Excellent cache performance! Your hit rate is above 95%.
                    </Alert>
                  )}
                  {cacheHitRate >= 80 && cacheHitRate <= 95 && (
                    <Alert severity="info">
                      Good cache performance. Consider optimizing frequently accessed flags.
                    </Alert>
                  )}
                  {cacheHitRate < 80 && (
                    <Alert severity="warning">
                      Cache hit rate is below 80%. Consider reviewing your caching strategy.
                    </Alert>
                  )}
                  
                  <Typography variant="body2" color="text.secondary">
                    <strong>Tips for optimization:</strong>
                    <br />
                    • Use ultra-fast evaluation for high-frequency flags
                    <br />
                    • Enable flag preloading for commonly used flags
                    <br />
                    • Monitor cache TTL settings
                  </Typography>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}
    </Box>
  );
}