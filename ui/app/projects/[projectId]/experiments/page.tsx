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
  Button,
  Chip,
  Alert,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Tooltip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  LinearProgress,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Divider,
  FormControlLabel,
  Switch,
} from '@mui/material';
import {
  Add as AddIcon,
  Science as ExperimentIcon,
  TrendingUp as TrendingUpIcon,
  Assessment as AnalyticsIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  PlayArrow as TestIcon,
  ExpandMore as ExpandMoreIcon,
  Visibility as ViewIcon,
  PauseCircle as PauseIcon,
  PlayCircle as PlayIcon,
  ContentCopy as ContentCopyIcon,
} from '@mui/icons-material';

interface Variation {
  id: string;
  name: string;
  description: string;
  value: string;
  weight: number;
}

interface VariantFlag {
  id: string;
  key: string;
  name: string;
  description: string;
  type: string;
  enabled: boolean;
  default: any;
  variations: Variation[];
  targeting: {
    rules: any[];
    rollout?: {
      type: string;
      variations: Array<{ variation_id: string; weight: number }>;
      bucket_by: string;
      seed: number;
    };
  };
  environment: string;
  created_at: string;
  updated_at: string;
}

export default function ProjectExperimentsPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [experiments, setExperiments] = useState<VariantFlag[]>([]);
  const [loading, setLoading] = useState(false);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openTestDialog, setOpenTestDialog] = useState(false);
  const [selectedExperiment, setSelectedExperiment] = useState<VariantFlag | null>(null);
  const [testResult, setTestResult] = useState<any>(null);
  const [formData, setFormData] = useState({
    key: '',
    name: '',
    description: '',
    default: '',
    variations: [
      { id: 'control', name: 'Control', description: 'Original version', value: '', weight: 50000 },
      { id: 'variant_a', name: 'Variant A', description: 'Test version', value: '', weight: 50000 }
    ],
    seed: Math.floor(Math.random() * 10000),
    stickyBucketing: false,
  });
  const [testData, setTestData] = useState({
    user_id: 'test_user_001',
    email: 'test@example.com',
    attributes: '{}'
  });

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
    }
  }, [projectId]);

  useEffect(() => {
    fetchExperiments();
  }, [projectId, currentEnvironment]);

  const fetchExperiments = async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/flags?project_id=${projectId}&environment=${currentEnvironment}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        // Filter only variant type flags (experiments)
        const variantFlags = (data.flags || []).filter((flag: any) => flag.type === 'variant');
        setExperiments(variantFlags);
      }
    } catch (error) {
      console.error('Error fetching experiments:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateExperiment = async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const experimentData = {
        key: formData.key,
        name: formData.name,
        description: formData.description,
        type: 'variant',
        enabled: true,
        default: formData.default || formData.variations[0].id,
        project_id: projectId,
        variations: formData.variations.map(v => ({
          id: v.id,
          name: v.name,
          description: v.description,
          value: v.value || `{"variant": "${v.id}"}`,
          weight: v.weight
        })),
        targeting: {
          rollout: {
            type: 'percentage',
            bucket_by: 'user_id',
            seed: formData.seed,
            sticky_bucketing: formData.stickyBucketing,
            variations: formData.variations.map(v => ({
              variation_id: v.id,
              weight: v.weight
            }))
          }
        }
      };

      const response = await fetch('http://localhost:8080/api/v1/flags', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(experimentData),
      });

      if (response.ok) {
        setOpenCreateDialog(false);
        fetchExperiments();
        resetForm();
        alert('Experiment created successfully!');
      } else {
        const errorData = await response.json();
        alert(`Error creating experiment: ${errorData.error}`);
      }
    } catch (error) {
      console.error('Error creating experiment:', error);
      alert('Failed to create experiment. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleTestExperiment = async () => {
    if (!selectedExperiment) return;
    
    try {
      const token = localStorage.getItem('token');
      let attributes = {};
      try {
        attributes = JSON.parse(testData.attributes);
      } catch (e) {
        attributes = {};
      }
      
      const response = await fetch('http://localhost:8080/api/v1/evaluate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          flag_key: selectedExperiment.key,
          user_id: testData.user_id,
          user_key: testData.user_id,
          attributes: {
            email: testData.email,
            ...attributes
          },
          environment: currentEnvironment,
        }),
      });
      
      if (response.ok) {
        const result = await response.json();
        setTestResult(result);
      } else {
        const errorData = await response.json();
        alert(`Error testing experiment: ${errorData.error}`);
      }
    } catch (error) {
      console.error('Error testing experiment:', error);
      alert('Failed to test experiment. Please try again.');
    }
  };

  const handleToggleExperiment = async (experiment: VariantFlag) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/flags/${experiment.key}/toggle?project_id=${projectId}&environment=${currentEnvironment}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        fetchExperiments();
      } else {
        const errorData = await response.json();
        alert(`Error toggling experiment: ${errorData.error}`);
      }
    } catch (error) {
      console.error('Error toggling experiment:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      key: '',
      name: '',
      description: '',
      default: '',
      variations: [
        { id: 'control', name: 'Control', description: 'Original version', value: '', weight: 50000 },
        { id: 'variant_a', name: 'Variant A', description: 'Test version', value: '', weight: 50000 }
      ],
      seed: Math.floor(Math.random() * 10000),
      stickyBucketing: false,
    });
  };

  const addVariation = () => {
    const newVariation = {
      id: `variant_${formData.variations.length}`,
      name: `Variant ${String.fromCharCode(65 + formData.variations.length - 1)}`,
      description: 'New test variation',
      value: '',
      weight: Math.floor(100000 / (formData.variations.length + 1))
    };
    setFormData({
      ...formData,
      variations: [...formData.variations, newVariation]
    });
  };

  const updateVariation = (index: number, field: string, value: any) => {
    const newVariations = [...formData.variations];
    newVariations[index] = { ...newVariations[index], [field]: value };
    setFormData({ ...formData, variations: newVariations });
  };

  const removeVariation = (index: number) => {
    if (formData.variations.length <= 2) {
      alert('An experiment must have at least 2 variations.');
      return;
    }
    const newVariations = formData.variations.filter((_, i) => i !== index);
    setFormData({ ...formData, variations: newVariations });
  };

  const getTotalWeight = () => {
    return formData.variations.reduce((sum, v) => sum + v.weight, 0);
  };

  const getVariationPercentage = (weight: number) => {
    const total = getTotalWeight();
    return total > 0 ? ((weight / total) * 100).toFixed(1) : '0.0';
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // You could add a toast notification here
      console.log('Copied to clipboard');
    } catch (err) {
      console.error('Failed to copy: ', err);
    }
  };

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
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            A/B Tests & Experiments
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage variant flags and A/B testing for {project?.name} in {currentEnvironment} environment
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpenCreateDialog(true)}
          size="large"
        >
          Create Experiment
        </Button>
      </Box>

      {/* Stats Overview */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h4" color="primary" fontWeight="bold">
                {experiments.length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total Experiments
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h4" color="success.main" fontWeight="bold">
                {experiments.filter(e => e.enabled).length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Running
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h4" color="warning.main" fontWeight="bold">
                {experiments.filter(e => !e.enabled).length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Paused
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography variant="h4" color="info.main" fontWeight="bold">
                {experiments.reduce((sum, e) => sum + (e.variations?.length || 0), 0)}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total Variants
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      {/* Experiments List */}
      {experiments.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 8 }}>
            <ExperimentIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
            <Typography variant="h6" color="text.secondary" gutterBottom>
              No experiments found
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Create your first A/B test to start experimenting with different variations
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setOpenCreateDialog(true)}
            >
              Create First Experiment
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Grid container spacing={3}>
          {experiments.map((experiment) => (
            <Grid item xs={12} key={experiment.id}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                    <Box sx={{ flex: 1 }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                        <Typography variant="h6" fontWeight="bold" sx={{ mr: 2 }}>
                          {experiment.name}
                        </Typography>
                        <Chip 
                          label={experiment.enabled ? 'Running' : 'Paused'}
                          color={experiment.enabled ? 'success' : 'default'}
                          size="small"
                          icon={experiment.enabled ? <PlayIcon /> : <PauseIcon />}
                        />
                        <Chip 
                          label={experiment.key}
                          size="small"
                          variant="outlined"
                          sx={{ ml: 1 }}
                        />
                      </Box>
                      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        {experiment.description || 'No description provided'}
                      </Typography>
                      
                      {/* Variations Summary */}
                      <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 2 }}>
                        {experiment.variations?.map((variation, index) => {
                          const rolloutVar = experiment.targeting?.rollout?.variations?.find(
                            v => v.variation_id === variation.id
                          );
                          const weight = rolloutVar?.weight || 0;
                          const totalWeight = experiment.targeting?.rollout?.variations?.reduce(
                            (sum, v) => sum + v.weight, 0
                          ) || 1;
                          const percentage = ((weight / totalWeight) * 100).toFixed(1);
                          
                          return (
                            <Chip
                              key={variation.id}
                              label={`${variation.name} (${percentage}%)`}
                              variant="outlined"
                              size="small"
                              color={index === 0 ? 'primary' : 'secondary'}
                            />
                          );
                        })}
                      </Box>
                    </Box>
                    
                    <Box sx={{ display: 'flex', gap: 1 }}>
                      <Tooltip title="Test Experiment">
                        <IconButton
                          size="small"
                          onClick={() => {
                            setSelectedExperiment(experiment);
                            setTestResult(null);
                            setOpenTestDialog(true);
                          }}
                        >
                          <TestIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title={experiment.enabled ? 'Pause Experiment' : 'Resume Experiment'}>
                        <IconButton
                          size="small"
                          onClick={() => handleToggleExperiment(experiment)}
                          color={experiment.enabled ? 'warning' : 'success'}
                        >
                          {experiment.enabled ? <PauseIcon /> : <PlayIcon />}
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </Box>
                  
                  {/* Expandable Details */}
                  <Accordion>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                      <Typography variant="body2" fontWeight="medium">
                        View Details & Configuration
                      </Typography>
                    </AccordionSummary>
                    <AccordionDetails>
                      <Grid container spacing={2}>
                        <Grid item xs={12} md={6}>
                          <Typography variant="subtitle2" gutterBottom>
                            Experiment Configuration
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            <strong>Flag Key:</strong> {experiment.key}<br />
                            <strong>Type:</strong> {experiment.type}<br />
                            <strong>Default:</strong> {experiment.default}<br />
                            <strong>Environment:</strong> {experiment.environment}<br />
                            <strong>Bucket By:</strong> {experiment.targeting?.rollout?.bucket_by || 'user_id'}<br />
                            <strong>Seed:</strong> {experiment.targeting?.rollout?.seed || 'N/A'}<br />
                            <strong>Sticky Bucketing:</strong> {experiment.targeting?.rollout?.sticky_bucketing ? 'Enabled' : 'Disabled'}
                          </Typography>
                        </Grid>
                        <Grid item xs={12} md={6}>
                          <Typography variant="subtitle2" gutterBottom>
                            Variations Detail
                          </Typography>
                          {experiment.variations?.map((variation) => (
                            <Box key={variation.id} sx={{ mb: 1, p: 1, bgcolor: 'grey.50', borderRadius: 1 }}>
                              <Typography variant="body2" fontWeight="medium">
                                {variation.name} ({variation.id})
                              </Typography>
                              <Typography variant="caption" color="text.secondary">
                                {variation.description}
                              </Typography>
                              <Typography variant="caption" display="block" sx={{ mt: 0.5 }}>
                                <strong>Value:</strong>
                              </Typography>
                              <Box 
                                component="pre" 
                                sx={{ 
                                  fontSize: '0.7rem', 
                                  fontFamily: 'monospace',
                                  bgcolor: 'grey.50', 
                                  p: 0.5, 
                                  borderRadius: 0.5, 
                                  mt: 0.5,
                                  maxHeight: '100px',
                                  overflow: 'auto',
                                  whiteSpace: 'pre-wrap',
                                  wordBreak: 'break-word'
                                }}
                              >
                                {(() => {
                                  try {
                                    if (typeof variation.value === 'string' && 
                                        (variation.value.startsWith('{') || variation.value.startsWith('['))) {
                                      const parsedValue = JSON.parse(variation.value);
                                      return JSON.stringify(parsedValue, null, 2);
                                    }
                                    else if (typeof variation.value === 'object') {
                                      return JSON.stringify(variation.value, null, 2);
                                    }
                                    else {
                                      return variation.value;
                                    }
                                  } catch (e) {
                                    return variation.value;
                                  }
                                })()}
                              </Box>
                            </Box>
                          ))}
                        </Grid>
                      </Grid>
                    </AccordionDetails>
                  </Accordion>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Create Experiment Dialog */}
      <Dialog open={openCreateDialog} onClose={() => { setOpenCreateDialog(false); resetForm(); }} maxWidth="md" fullWidth>
        <DialogTitle>Create New A/B Test Experiment</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Experiment Key"
                value={formData.key}
                onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                helperText="Unique identifier for the experiment"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Experiment Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                multiline
                rows={2}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Default Variation"
                value={formData.default}
                onChange={(e) => setFormData({ ...formData, default: e.target.value })}
                helperText="Leave empty to use first variation"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Random Seed"
                type="number"
                value={formData.seed}
                onChange={(e) => setFormData({ ...formData, seed: parseInt(e.target.value) || 0 })}
                helperText="For consistent bucketing"
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.stickyBucketing}
                    onChange={(e) => setFormData({ ...formData, stickyBucketing: e.target.checked })}
                  />
                }
                label="Enable Sticky Bucketing"
              />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                When enabled, users will remain in the same variation even if experiment weights change.
                Assignments are stored in the database and persist across sessions.
              </Typography>
            </Grid>
          </Grid>

          <Divider sx={{ my: 3 }} />

          <Typography variant="h6" gutterBottom>
            Variations ({formData.variations.length})
          </Typography>
          
          {/* Weight Distribution Warning */}
          {getTotalWeight() !== 100000 && (
            <Alert severity="warning" sx={{ mb: 2 }}>
              Total weight is {getTotalWeight().toLocaleString()} but should be 100,000 for proper distribution.
            </Alert>
          )}

          {formData.variations.map((variation, index) => (
            <Box key={index} sx={{ mb: 2, p: 2, border: 1, borderColor: 'divider', borderRadius: 1 }}>
              <Grid container spacing={2}>
                <Grid item xs={12} md={3}>
                  <TextField
                    fullWidth
                    label="Variation ID"
                    value={variation.id}
                    onChange={(e) => updateVariation(index, 'id', e.target.value)}
                  />
                </Grid>
                <Grid item xs={12} md={3}>
                  <TextField
                    fullWidth
                    label="Name"
                    value={variation.name}
                    onChange={(e) => updateVariation(index, 'name', e.target.value)}
                  />
                </Grid>
                <Grid item xs={12} md={3}>
                  <TextField
                    fullWidth
                    label="Weight"
                    type="number"
                    value={variation.weight}
                    onChange={(e) => updateVariation(index, 'weight', parseInt(e.target.value) || 0)}
                    helperText={`${getVariationPercentage(variation.weight)}%`}
                  />
                </Grid>
                <Grid item xs={12} md={2}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, height: '100%' }}>
                    <IconButton 
                      onClick={() => removeVariation(index)} 
                      color="error"
                      disabled={formData.variations.length <= 2}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Description"
                    value={variation.description}
                    onChange={(e) => updateVariation(index, 'description', e.target.value)}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="JSON Value"
                    value={variation.value}
                    onChange={(e) => updateVariation(index, 'value', e.target.value)}
                    placeholder={`{"variant": "${variation.id}", "color": "#007bff"}`}
                    helperText="JSON object returned when this variation is selected"
                    multiline
                    rows={2}
                  />
                </Grid>
              </Grid>
            </Box>
          ))}
          
          <Button onClick={addVariation} startIcon={<AddIcon />} sx={{ mt: 1 }}>
            Add Variation
          </Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setOpenCreateDialog(false); resetForm(); }}>Cancel</Button>
          <Button onClick={handleCreateExperiment} variant="contained" disabled={loading}>
            {loading ? 'Creating...' : 'Create Experiment'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Test Experiment Dialog */}
      <Dialog open={openTestDialog} onClose={() => { setOpenTestDialog(false); setTestResult(null); }} maxWidth="md" fullWidth>
        <DialogTitle>Test Experiment: {selectedExperiment?.name}</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="User ID"
                value={testData.user_id}
                onChange={(e) => setTestData({ ...testData, user_id: e.target.value })}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Email"
                value={testData.email}
                onChange={(e) => setTestData({ ...testData, email: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Additional Attributes (JSON)"
                value={testData.attributes}
                onChange={(e) => setTestData({ ...testData, attributes: e.target.value })}
                placeholder='{"subscription_plan": "premium", "country": "US"}'
                multiline
                rows={3}
              />
            </Grid>
          </Grid>

          {testResult && (
            <Alert severity={testResult.default ? 'warning' : 'success'} sx={{ mt: 3 }}>
              <Typography variant="subtitle2" fontWeight="bold">
                {testResult.default ? '⚠️ Using Default Value' : '✅ Experiment Variation Assigned'}
              </Typography>
              <Grid container spacing={2} sx={{ mt: 1 }}>
                <Grid item xs={12} sm={6}>
                  <Typography variant="body2">
                    <strong>Variation:</strong> {testResult.variation || 'N/A'}<br />
                    <strong>Reason:</strong> {testResult.reason}<br />
                    <strong>Is Default:</strong> {testResult.default ? 'Yes' : 'No'}<br />
                    <strong>Evaluation Time:</strong> {testResult.evaluation_time_ms ? `${testResult.evaluation_time_ms.toFixed(3)}ms` : 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                    <Typography variant="body2">
                      <strong>Value:</strong>
                    </Typography>
                    <Tooltip title="Copy JSON value">
                      <IconButton 
                        size="small" 
                        onClick={() => {
                          const valueTosCopy = (() => {
                            try {
                              if (typeof testResult.value === 'string' && 
                                  (testResult.value.startsWith('{') || testResult.value.startsWith('['))) {
                                const parsedValue = JSON.parse(testResult.value);
                                return JSON.stringify(parsedValue, null, 2);
                              }
                              else if (typeof testResult.value === 'object') {
                                return JSON.stringify(testResult.value, null, 2);
                              }
                              else {
                                return testResult.value;
                              }
                            } catch (e) {
                              return testResult.value;
                            }
                          })();
                          copyToClipboard(valueTosCopy);
                        }}
                      >
                        <ContentCopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </Box>
                  <Box 
                    component="pre" 
                    sx={{ 
                      fontSize: '0.75rem', 
                      bgcolor: 'grey.100', 
                      p: 1, 
                      borderRadius: 1,
                      maxHeight: '200px',
                      overflow: 'auto',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                      fontFamily: 'monospace'
                    }}
                  >
                    {(() => {
                      try {
                        // If the value is a string that looks like JSON, parse it first
                        if (typeof testResult.value === 'string' && 
                            (testResult.value.startsWith('{') || testResult.value.startsWith('['))) {
                          const parsedValue = JSON.parse(testResult.value);
                          return JSON.stringify(parsedValue, null, 2);
                        }
                        // If it's already an object, stringify it
                        else if (typeof testResult.value === 'object') {
                          return JSON.stringify(testResult.value, null, 2);
                        }
                        // Otherwise, return as-is
                        else {
                          return testResult.value;
                        }
                      } catch (e) {
                        // If JSON parsing fails, return the original value
                        return testResult.value;
                      }
                    })()}
                  </Box>
                </Grid>
              </Grid>
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setOpenTestDialog(false); setTestResult(null); }}>Close</Button>
          <Button onClick={handleTestExperiment} variant="contained">
            Run Test
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}