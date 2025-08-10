'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  TextField,
  Button,
  Select,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Chip,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  IconButton,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import {
  PlayArrow as PlayArrowIcon,
  ExpandMore as ExpandMoreIcon,
  ContentCopy as ContentCopyIcon,
  Speed as SpeedIcon,
  Assessment as AssessmentIcon,
  History as HistoryIcon,
} from '@mui/icons-material';
import { apiClient } from '@/lib/api';
import { EvaluationRequest, EvaluationResponse } from '@/types';
import { useProject } from '@/contexts/ProjectContext';
import { useEnvironment } from '@/contexts/EnvironmentContext';

interface EvaluationResult extends EvaluationResponse {
  id: string;
  endpoint: string;
  timestamp_local: string;
}

function EvaluationTester() {
  const { currentProject } = useProject();
  const { currentEnvironment } = useEnvironment();
  const [request, setRequest] = useState<EvaluationRequest>({
    flag_key: '',
    user_id: 'test-user-123',
    attributes: {},
  });
  const [endpoint, setEndpoint] = useState<'standard' | 'fast' | 'ultra'>('ultra');
  const [result, setResult] = useState<EvaluationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [attributeKey, setAttributeKey] = useState('');
  const [attributeValue, setAttributeValue] = useState('');
  const [flags, setFlags] = useState<any[]>([]);
  const [loadingFlags, setLoadingFlags] = useState(false);

  // Fetch flags for the current project and environment
  const fetchFlags = async () => {
    if (!currentProject) {
      setFlags([]);
      return;
    }

    setLoadingFlags(true);
    try {
      const flagsData = await apiClient.getFlags(currentEnvironment, currentProject.id);
      setFlags(flagsData);
      
      // If no flag is selected and there are flags available, select the first one
      if (!request.flag_key && flagsData.length > 0) {
        setRequest(prev => ({ ...prev, flag_key: flagsData[0].key }));
      }
    } catch (err) {
      console.error('Failed to fetch flags:', err);
      setFlags([]);
    } finally {
      setLoadingFlags(false);
    }
  };

  useEffect(() => {
    fetchFlags();
  }, [currentProject, currentEnvironment]);

  const addAttribute = () => {
    if (attributeKey && attributeValue) {
      setRequest({
        ...request,
        attributes: {
          ...request.attributes,
          [attributeKey]: attributeValue,
        },
      });
      setAttributeKey('');
      setAttributeValue('');
    }
  };

  const removeAttribute = (key: string) => {
    const newAttributes = { ...request.attributes };
    delete newAttributes[key];
    setRequest({
      ...request,
      attributes: newAttributes,
    });
  };

  const evaluateFlag = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      let response: EvaluationResponse;
      
      switch (endpoint) {
        case 'standard':
          response = await apiClient.evaluateFlag(request, currentEnvironment, currentProject?.id);
          break;
        case 'fast':
          response = await apiClient.evaluateFlagFast(request, currentEnvironment, currentProject?.id);
          break;
        case 'ultra':
          response = await apiClient.evaluateFlagUltraFast(request, currentEnvironment, currentProject?.id);
          break;
      }

      const resultWithMeta: EvaluationResult = {
        ...response,
        id: Math.random().toString(36).substr(2, 9),
        endpoint: endpoint,
        timestamp_local: new Date().toLocaleString(),
      };

      setResult(resultWithMeta);
    } catch (err: any) {
      setError(err.message || 'Failed to evaluate flag');
      console.error('Evaluation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getPerformanceColor = (time: number) => {
    if (time < 1) return 'success';
    if (time < 5) return 'warning';
    return 'error';
  };

  return (
    <Grid container spacing={3}>
      {/* Request Configuration */}
      <Grid item xs={12} lg={6}>
        <Card>
          <CardContent>
            <Typography variant="h6" fontWeight="600" gutterBottom>
              Flag Evaluation Request
            </Typography>
            
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <FormControl fullWidth>
                  <InputLabel>Flag</InputLabel>
                  <Select
                    value={request.flag_key}
                    label="Flag"
                    onChange={(e) => setRequest({ ...request, flag_key: e.target.value })}
                    disabled={loadingFlags || !currentProject || flags.length === 0}
                  >
                    {flags.map((flag) => (
                      <MenuItem key={flag.id} value={flag.key}>
                        <Box>
                          <Typography variant="body2">{flag.name}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {flag.key} • {flag.type} • {flag.enabled ? 'Enabled' : 'Disabled'}
                          </Typography>
                        </Box>
                      </MenuItem>
                    ))}
                    {flags.length === 0 && !loadingFlags && (
                      <MenuItem disabled>
                        {currentProject ? 'No flags in this project' : 'Select a project first'}
                      </MenuItem>
                    )}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  label="User ID"
                  value={request.user_id}
                  onChange={(e) => setRequest({ ...request, user_id: e.target.value })}
                  fullWidth
                  placeholder="user-12345"
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  label="User Key (Optional)"
                  value={request.user_key || ''}
                  onChange={(e) => setRequest({ ...request, user_key: e.target.value })}
                  fullWidth
                  placeholder="user@example.com"
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  label="Environment"
                  value={currentEnvironment}
                  fullWidth
                  disabled
                  helperText="Environment is controlled by the global environment selector"
                />
              </Grid>
            </Grid>

            {/* Attributes */}
            <Box mt={3}>
              <Typography variant="subtitle2" gutterBottom>
                User Attributes
              </Typography>
              <Grid container spacing={2} alignItems="center" mb={2}>
                <Grid item xs={4}>
                  <TextField
                    size="small"
                    label="Key"
                    value={attributeKey}
                    onChange={(e) => setAttributeKey(e.target.value)}
                    fullWidth
                  />
                </Grid>
                <Grid item xs={6}>
                  <TextField
                    size="small"
                    label="Value"
                    value={attributeValue}
                    onChange={(e) => setAttributeValue(e.target.value)}
                    fullWidth
                  />
                </Grid>
                <Grid item xs={2}>
                  <Button
                    variant="outlined"
                    onClick={addAttribute}
                    disabled={!attributeKey || !attributeValue}
                    fullWidth
                  >
                    Add
                  </Button>
                </Grid>
              </Grid>
              
              <Box display="flex" flexWrap="wrap" gap={1}>
                {Object.entries(request.attributes || {}).map(([key, value]) => (
                  <Chip
                    key={key}
                    label={`${key}: ${value}`}
                    onDelete={() => removeAttribute(key)}
                    size="small"
                  />
                ))}
              </Box>
            </Box>

            {/* Endpoint Selection */}
            <Box mt={3}>
              <FormControl fullWidth>
                <InputLabel>Evaluation Endpoint</InputLabel>
                <Select
                  value={endpoint}
                  label="Evaluation Endpoint"
                  onChange={(e) => setEndpoint(e.target.value as any)}
                >
                  <MenuItem value="standard">
                    <Box>
                      <Typography variant="body2">Standard Evaluation</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Basic evaluation with full feature set
                      </Typography>
                    </Box>
                  </MenuItem>
                  <MenuItem value="fast">
                    <Box>
                      <Typography variant="body2">Fast Evaluation</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Optimized with in-memory caching
                      </Typography>
                    </Box>
                  </MenuItem>
                  <MenuItem value="ultra">
                    <Box>
                      <Typography variant="body2">Ultra-Fast Evaluation</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Maximum performance with preloading
                      </Typography>
                    </Box>
                  </MenuItem>
                </Select>
              </FormControl>
            </Box>

            <Button
              variant="contained"
              startIcon={<PlayArrowIcon />}
              onClick={evaluateFlag}
              disabled={loading || !request.flag_key || !request.user_id}
              fullWidth
              sx={{ mt: 3 }}
            >
              {loading ? 'Evaluating...' : 'Evaluate Flag'}
            </Button>
          </CardContent>
        </Card>
      </Grid>

      {/* Results */}
      <Grid item xs={12} lg={6}>
        <Card>
          <CardContent>
            <Typography variant="h6" fontWeight="600" gutterBottom>
              Evaluation Result
            </Typography>
            
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            {result ? (
              <Box>
                {/* Performance Metrics */}
                <Paper sx={{ p: 2, mb: 2, bgcolor: 'grey.50' }}>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Response Time
                      </Typography>
                      <Box display="flex" alignItems="center" gap={1}>
                        <Typography 
                          variant="h6" 
                          fontWeight="bold" 
                          color={`${getPerformanceColor(result.evaluation_time_ms)}.main`}
                        >
                          {result.evaluation_time_ms.toFixed(3)}ms
                        </Typography>
                        <Chip
                          size="small"
                          label={result.endpoint.toUpperCase()}
                          color={result.endpoint === 'ultra' ? 'success' : result.endpoint === 'fast' ? 'primary' : 'default'}
                        />
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Timestamp
                      </Typography>
                      <Typography variant="body1">
                        {result.timestamp_local}
                      </Typography>
                    </Grid>
                  </Grid>
                </Paper>

                {/* Flag Result */}
                <TableContainer component={Paper} sx={{ mb: 2 }}>
                  <Table size="small">
                    <TableBody>
                      <TableRow>
                        <TableCell><strong>Flag Key</strong></TableCell>
                        <TableCell>{result.flag_key}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell><strong>Value</strong></TableCell>
                        <TableCell>
                          <Box display="flex" alignItems="center" gap={1}>
                            <code style={{ 
                              backgroundColor: '#f5f5f5', 
                              padding: '2px 6px', 
                              borderRadius: '4px' 
                            }}>
                              {typeof result.value === 'object' 
                                ? JSON.stringify(result.value) 
                                : String(result.value)
                              }
                            </code>
                            <IconButton
                              size="small"
                              onClick={() => copyToClipboard(
                                typeof result.value === 'object' 
                                  ? JSON.stringify(result.value, null, 2) 
                                  : String(result.value)
                              )}
                            >
                              <ContentCopyIcon fontSize="small" />
                            </IconButton>
                          </Box>
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell><strong>Reason</strong></TableCell>
                        <TableCell>
                          <Chip 
                            label={result.reason} 
                            size="small"
                            color={result.default ? 'default' : 'primary'}
                          />
                        </TableCell>
                      </TableRow>
                      {result.variation && (
                        <TableRow>
                          <TableCell><strong>Variation</strong></TableCell>
                          <TableCell>{result.variation}</TableCell>
                        </TableRow>
                      )}
                      {result.rule_id && (
                        <TableRow>
                          <TableCell><strong>Rule ID</strong></TableCell>
                          <TableCell>{result.rule_id}</TableCell>
                        </TableRow>
                      )}
                      <TableRow>
                        <TableCell><strong>Is Default</strong></TableCell>
                        <TableCell>
                          <Chip 
                            label={result.default ? 'Yes' : 'No'} 
                            size="small"
                            color={result.default ? 'warning' : 'success'}
                          />
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </TableContainer>

                {/* Raw JSON */}
                <Accordion>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography variant="body2">Raw JSON Response</Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Box 
                      component="pre" 
                      sx={{ 
                        bgcolor: 'grey.50', 
                        p: 2, 
                        borderRadius: 1, 
                        fontSize: '0.75rem',
                        overflow: 'auto',
                        fontFamily: 'monospace',
                      }}
                    >
                      {JSON.stringify(result, null, 2)}
                    </Box>
                  </AccordionDetails>
                </Accordion>
              </Box>
            ) : (
              <Box textAlign="center" py={4}>
                <AssessmentIcon sx={{ fontSize: 48, color: 'grey.300', mb: 2 }} />
                <Typography variant="body1" color="text.secondary">
                  Configure your request and click "Evaluate Flag" to see results
                </Typography>
              </Box>
            )}
          </CardContent>
        </Card>
      </Grid>

      {/* Quick Tests */}
      <Grid item xs={12}>
        <Card>
          <CardContent>
            <Typography variant="h6" fontWeight="600" gutterBottom>
              Quick Test Scenarios
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6} md={3}>
                <Button
                  variant="outlined"
                  fullWidth
                  disabled={!request.flag_key}
                  onClick={() => setRequest({
                    flag_key: request.flag_key,
                    user_id: 'premium-user',
                    attributes: { plan: 'premium', region: 'us-west-1' }
                  })}
                >
                  Premium User Test
                </Button>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Button
                  variant="outlined"
                  fullWidth
                  disabled={!request.flag_key}
                  onClick={() => setRequest({
                    flag_key: request.flag_key,
                    user_id: 'basic-user',
                    attributes: { plan: 'basic', region: 'us-east-1' }
                  })}
                >
                  Basic User Test
                </Button>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Button
                  variant="outlined"
                  fullWidth
                  disabled={!request.flag_key}
                  onClick={() => setRequest({
                    flag_key: request.flag_key,
                    user_id: 'beta-tester',
                    attributes: { beta: 'true', role: 'tester' }
                  })}
                >
                  Beta Feature Test
                </Button>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Button
                  variant="outlined"
                  fullWidth
                  disabled={!request.flag_key}
                  onClick={() => setRequest({
                    flag_key: request.flag_key,
                    user_id: `random-${Math.floor(Math.random() * 1000)}`,
                    attributes: {}
                  })}
                >
                  Random User Test
                </Button>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
}

export default function EvaluationsPage() {
  return (
    <Box>
      {/* Header */}
      <Box mb={4}>
        <Typography variant="h4" fontWeight="bold" gutterBottom>
          Flag Evaluation Testing
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Test flag evaluations and compare performance across different endpoints
        </Typography>
      </Box>

      <EvaluationTester />
    </Box>
  );
}