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
  TextField,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Paper,
  Alert,
  Chip,
  Divider,
} from '@mui/material';
import {
  PlayArrow as PlayIcon,
  Assessment as EvaluationIcon,
  Speed as SpeedIcon,
  AccountTree as ProjectIcon,
} from '@mui/icons-material';

interface Flag {
  id: string;
  key: string;
  name: string;
  type: string;
  enabled: boolean;
}

interface EvaluationRequest {
  flag_key: string;
  user_id: string;
  user_key?: string;
  attributes?: Record<string, any>;
}

interface EvaluationResponse {
  flag_key: string;
  value: any;
  variation_id?: string;
  reason: string;
  match: boolean;
}

export default function ProjectEvaluationsPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [flags, setFlags] = useState<Flag[]>([]);
  const [selectedFlag, setSelectedFlag] = useState<string>('');
  const [evaluationMode, setEvaluationMode] = useState<'standard' | 'fast' | 'ultra'>('standard');
  const [userContext, setUserContext] = useState({
    user_id: '',
    user_key: '',
    email: '',
    country: '',
    plan: '',
    custom_attributes: '{}',
  });
  const [result, setResult] = useState<EvaluationResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch project and flags
  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch project
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);

        // Fetch flags
        const flagsData = await apiClient.getFlags(currentEnvironment, projectId);
        setFlags(flagsData || []);
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    if (projectId) {
      fetchData();
    }
  }, [projectId, currentEnvironment]);

  const handleEvaluate = async () => {
    if (!selectedFlag || !userContext.user_id) {
      setError('Please select a flag and provide a user ID');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      let customAttributes = {};
      try {
        customAttributes = JSON.parse(userContext.custom_attributes);
      } catch {
        // Invalid JSON, use empty object
      }

      const request: EvaluationRequest = {
        flag_key: selectedFlag,
        user_id: userContext.user_id,
        user_key: userContext.user_key || userContext.user_id,
        attributes: {
          email: userContext.email,
          country: userContext.country,
          plan: userContext.plan,
          ...customAttributes,
        },
      };

      let response: EvaluationResponse;
      
      switch (evaluationMode) {
        case 'fast':
          response = await apiClient.evaluateFlagFast(request, currentEnvironment, projectId);
          break;
        case 'ultra':
          response = await apiClient.evaluateFlagUltraFast(request, currentEnvironment, projectId);
          break;
        default:
          response = await apiClient.evaluateFlag(request, currentEnvironment, projectId);
      }

      setResult(response);
    } catch (err: any) {
      setError(err.message || 'Failed to evaluate flag');
    } finally {
      setLoading(false);
    }
  };

  const selectedFlagObj = flags.find(f => f.key === selectedFlag);

  if (!project) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <Typography>Loading project...</Typography>
      </Box>
    );
  }

  return (
    <Box>
      {/* Header */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" fontWeight="bold" gutterBottom>
          Flag Evaluations
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Test flag evaluation logic for {project.name} in {currentEnvironment} environment
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {/* Evaluation Form */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <EvaluationIcon />
                Evaluation Request
              </Typography>

              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <FormControl fullWidth>
                    <InputLabel>Select Flag</InputLabel>
                    <Select
                      value={selectedFlag}
                      onChange={(e) => setSelectedFlag(e.target.value)}
                      label="Select Flag"
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
                    <InputLabel>Evaluation Mode</InputLabel>
                    <Select
                      value={evaluationMode}
                      onChange={(e) => setEvaluationMode(e.target.value as any)}
                      label="Evaluation Mode"
                    >
                      <MenuItem value="standard">Standard</MenuItem>
                      <MenuItem value="fast">Fast (Cached)</MenuItem>
                      <MenuItem value="ultra">Ultra Fast (Memory)</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>

                <Grid item xs={12}>
                  <Divider sx={{ my: 1 }} />
                  <Typography variant="subtitle2" gutterBottom>
                    User Context
                  </Typography>
                </Grid>

                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="User ID *"
                    value={userContext.user_id}
                    onChange={(e) => setUserContext({ ...userContext, user_id: e.target.value })}
                    placeholder="user123"
                  />
                </Grid>

                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="User Key"
                    value={userContext.user_key}
                    onChange={(e) => setUserContext({ ...userContext, user_key: e.target.value })}
                    placeholder="Optional override"
                  />
                </Grid>

                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="Email"
                    value={userContext.email}
                    onChange={(e) => setUserContext({ ...userContext, email: e.target.value })}
                    placeholder="user@example.com"
                  />
                </Grid>

                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="Country"
                    value={userContext.country}
                    onChange={(e) => setUserContext({ ...userContext, country: e.target.value })}
                    placeholder="US"
                  />
                </Grid>

                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="Plan"
                    value={userContext.plan}
                    onChange={(e) => setUserContext({ ...userContext, plan: e.target.value })}
                    placeholder="premium"
                  />
                </Grid>

                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Custom Attributes (JSON)"
                    value={userContext.custom_attributes}
                    onChange={(e) => setUserContext({ ...userContext, custom_attributes: e.target.value })}
                    placeholder='{"age": 25, "beta_user": true}'
                    multiline
                    rows={2}
                  />
                </Grid>

                <Grid item xs={12}>
                  <Button
                    variant="contained"
                    fullWidth
                    onClick={handleEvaluate}
                    disabled={loading || !selectedFlag || !userContext.user_id}
                    startIcon={<PlayIcon />}
                    size="large"
                  >
                    {loading ? 'Evaluating...' : 'Evaluate Flag'}
                  </Button>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Results */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <SpeedIcon />
                Evaluation Result
              </Typography>

              {error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                  {error}
                </Alert>
              )}

              {selectedFlagObj && (
                <Paper sx={{ p: 2, mb: 2, bgcolor: 'grey.50' }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Selected Flag
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 1 }}>
                    <Typography variant="body1" fontWeight="bold">
                      {selectedFlagObj.name}
                    </Typography>
                    <Chip
                      label={selectedFlagObj.enabled ? 'Enabled' : 'Disabled'}
                      size="small"
                      color={selectedFlagObj.enabled ? 'success' : 'default'}
                    />
                  </Box>
                  <Typography variant="caption" sx={{ fontFamily: 'monospace' }}>
                    {selectedFlagObj.key}
                  </Typography>
                </Paper>
              )}

              {result && (
                <Paper sx={{ p: 2 }}>
                  <Grid container spacing={2}>
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary">
                        Result Value
                      </Typography>
                      <Typography 
                        variant="h6" 
                        fontWeight="bold"
                        sx={{ fontFamily: 'monospace', color: 'primary.main' }}
                      >
                        {typeof result.value === 'object' 
                          ? JSON.stringify(result.value) 
                          : String(result.value)
                        }
                      </Typography>
                    </Grid>

                    {result.variation_id && (
                      <Grid item xs={12}>
                        <Typography variant="subtitle2" color="text.secondary">
                          Variation ID
                        </Typography>
                        <Typography variant="body1">
                          {result.variation_id}
                        </Typography>
                      </Grid>
                    )}

                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary">
                        Reason
                      </Typography>
                      <Typography variant="body2">
                        {result.reason}
                      </Typography>
                    </Grid>

                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary">
                        Match Status
                      </Typography>
                      <Chip
                        label={result.match ? 'Matched' : 'No Match'}
                        size="small"
                        color={result.match ? 'success' : 'default'}
                      />
                    </Grid>
                  </Grid>
                </Paper>
              )}

              {!result && !error && (
                <Paper sx={{ p: 4, textAlign: 'center', bgcolor: 'grey.50' }}>
                  <EvaluationIcon sx={{ fontSize: 48, color: 'grey.300', mb: 2 }} />
                  <Typography color="text.secondary">
                    Run an evaluation to see results
                  </Typography>
                </Paper>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}