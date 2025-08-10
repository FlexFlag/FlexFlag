'use client';

import React, { useState, useEffect } from 'react';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Slider,
  Alert,
  LinearProgress,
  Paper,
  Switch,
  FormControlLabel,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import {
  Add as AddIcon,
  PlayArrow as PlayIcon,
  Pause as PauseIcon,
  Stop as StopIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DonutLarge as RolloutIcon,
  Science as ExperimentIcon,
  Analytics as AnalyticsIcon,
  Group as GroupIcon,
} from '@mui/icons-material';

interface Rollout {
  id: string;
  flag_id: string;
  environment: string;
  type: 'percentage' | 'experiment' | 'segment';
  name: string;
  description: string;
  config: {
    percentage?: number;
    variations?: Array<{ variation_id: string; weight: number }>;
    sticky_bucketing?: boolean;
    bucket_by?: string;
    traffic_allocation?: number;
  };
  status: 'draft' | 'active' | 'paused' | 'completed';
  created_at: string;
  updated_at: string;
}

interface Flag {
  id: string;
  key: string;
  name: string;
}

export default function ProjectRolloutsPage() {
  const { currentEnvironment } = useEnvironment();
  const params = useParams();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [rollouts, setRollouts] = useState<Rollout[]>([]);
  const [flags, setFlags] = useState<Flag[]>([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openAnalyticsDialog, setOpenAnalyticsDialog] = useState(false);
  const [selectedRollout, setSelectedRollout] = useState<Rollout | null>(null);
  const [editingRollout, setEditingRollout] = useState<Rollout | null>(null);
  const [formData, setFormData] = useState({
    flag_id: '',
    environment: 'production',
    type: 'percentage' as const,
    name: '',
    description: '',
    percentage: 25,
    variations: [
      { variation_id: 'variant_a', weight: 50 },
      { variation_id: 'variant_b', weight: 50 },
    ],
    sticky_bucketing: true,
    bucket_by: 'user_key',
    traffic_allocation: 100,
  });

  // Fetch project data
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
    if (projectId) {
      fetchRollouts();
      fetchFlags();
    }
  }, [projectId, currentEnvironment]);

  const fetchRollouts = async () => {
    if (!projectId) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/rollouts?project_id=${projectId}&environment=${currentEnvironment}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        setRollouts(data.rollouts || []);
      }
    } catch (error) {
      console.error('Error fetching rollouts:', error);
    }
  };

  const fetchFlags = async () => {
    if (!projectId) return;
    
    try {
      const flagsData = await apiClient.getFlags(currentEnvironment, projectId);
      setFlags(flagsData || []);
    } catch (error) {
      console.error('Error fetching flags:', error);
    }
  };

  const handleCreateRollout = async () => {
    try {
      const token = localStorage.getItem('token');
      const config: any = {
        sticky_bucketing: formData.sticky_bucketing,
        bucket_by: formData.bucket_by,
        traffic_allocation: formData.traffic_allocation,
      };

      if (formData.type === 'percentage') {
        config.percentage = formData.percentage;
      } else if (formData.type === 'experiment') {
        config.variations = formData.variations;
      }

      const response = await fetch('http://localhost:8080/api/v1/rollouts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          flag_id: formData.flag_id,
          environment: currentEnvironment,
          type: formData.type,
          name: formData.name,
          description: formData.description,
          config,
        }),
      });
      
      if (response.ok) {
        setOpenDialog(false);
        fetchRollouts();
        resetForm();
      }
    } catch (error) {
      console.error('Error creating rollout:', error);
    }
  };

  const handleStatusChange = async (rolloutId: string, action: 'activate' | 'pause' | 'complete') => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/rollouts/${rolloutId}/${action}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        fetchRollouts();
      }
    } catch (error) {
      console.error(`Error ${action}ing rollout:`, error);
    }
  };

  const handleDeleteRollout = async (rolloutId: string) => {
    if (!confirm('Are you sure you want to delete this rollout?')) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/rollouts/${rolloutId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        fetchRollouts();
      }
    } catch (error) {
      console.error('Error deleting rollout:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      flag_id: '',
      environment: currentEnvironment || 'production',
      type: 'percentage',
      name: '',
      description: '',
      percentage: 25,
      variations: [
        { variation_id: 'variant_a', weight: 50 },
        { variation_id: 'variant_b', weight: 50 },
      ],
      sticky_bucketing: true,
      bucket_by: 'user_key',
      traffic_allocation: 100,
    });
  };

  const updateVariation = (index: number, field: string, value: any) => {
    const newVariations = [...formData.variations];
    newVariations[index] = { ...newVariations[index], [field]: value };
    setFormData({ ...formData, variations: newVariations });
  };

  const handleEditRollout = (rollout: Rollout) => {
    setEditingRollout(rollout);
    // Pre-populate the form with existing rollout data
    setFormData({
      flag_id: rollout.flag_id,
      environment: rollout.environment,
      type: rollout.type,
      name: rollout.name,
      description: rollout.description,
      percentage: rollout.config.percentage || 25,
      variations: rollout.config.variations || [
        { variation_id: 'variant_a', weight: 50 },
        { variation_id: 'variant_b', weight: 50 },
      ],
      sticky_bucketing: rollout.config.sticky_bucketing || true,
      bucket_by: rollout.config.bucket_by || 'user_key',
      traffic_allocation: rollout.config.traffic_allocation || 100,
    });
    setOpenEditDialog(true);
  };

  const handleUpdateRollout = async () => {
    if (!editingRollout) return;

    try {
      const token = localStorage.getItem('token');
      const config: any = {
        sticky_bucketing: formData.sticky_bucketing,
        bucket_by: formData.bucket_by,
        traffic_allocation: formData.traffic_allocation,
      };

      if (formData.type === 'percentage') {
        config.percentage = formData.percentage;
      } else if (formData.type === 'experiment') {
        config.variations = formData.variations;
      }

      const response = await fetch(`http://localhost:8080/api/v1/rollouts/${editingRollout.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          config,
        }),
      });
      
      if (response.ok) {
        setOpenEditDialog(false);
        setEditingRollout(null);
        fetchRollouts();
        resetForm();
      }
    } catch (error) {
      console.error('Error updating rollout:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'paused': return 'warning';
      case 'completed': return 'default';
      case 'draft': return 'info';
      default: return 'default';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'percentage': return <RolloutIcon />;
      case 'experiment': return <ExperimentIcon />;
      case 'segment': return <GroupIcon />;
      default: return <RolloutIcon />;
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
      <Box 
        sx={{ 
          mb: 5,
          pb: 3,
          borderBottom: '1px solid',
          borderColor: 'divider'
        }}
      >
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Box>
            <Typography 
              variant="h5" 
              fontWeight="600" 
              gutterBottom
              sx={{ 
                fontSize: '1.5rem',
                letterSpacing: '-0.01em',
                color: 'text.primary',
                mb: 0.5
              }}
            >
              Rollouts & Experiments
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ fontSize: '0.875rem' }}
            >
              Manage percentage rollouts and A/B testing experiments for {project.name} in {currentEnvironment} environment
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
            sx={{ 
              borderRadius: 1.5,
              px: 3,
              py: 1.25,
              textTransform: 'none',
              fontWeight: 600,
              fontSize: '0.95rem',
              boxShadow: 2,
              '&:hover': {
                boxShadow: 4
              }
            }}
          >
            Create Rollout
          </Button>
        </Box>
      </Box>

      {rollouts.length === 0 ? (
        <Box textAlign="center" py={8}>
          <RolloutIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No rollouts found
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Create your first rollout to enable gradual feature deployment and A/B testing
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
          >
            Create First Rollout
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {rollouts.map((rollout) => (
            <Grid item xs={12} sm={6} md={4} lg={3} key={rollout.id}>
              <Card 
                sx={{ 
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  position: 'relative',
                  transition: 'all 0.2s ease',
                  border: '1px solid',
                  borderColor: 'divider',
                  boxShadow: 0,
                  '&:hover': {
                    boxShadow: 2,
                    transform: 'translateY(-1px)',
                    borderColor: rollout.status === 'active' ? 'success.main' : 'grey.400',
                  },
                }}
              >
                <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', p: 2 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Box sx={{ color: 'primary.main', display: 'flex', alignItems: 'center' }}>
                        {React.cloneElement(getTypeIcon(rollout.type) as React.ReactElement, { sx: { fontSize: '1.1rem' } })}
                      </Box>
                      <Typography variant="subtitle2" fontWeight="600" noWrap sx={{ fontSize: '0.9rem' }}>
                        {rollout.name}
                      </Typography>
                    </Box>
                  </Box>
                  <Chip 
                    label={rollout.status} 
                    size="small" 
                    color={getStatusColor(rollout.status) as any}
                    sx={{ mb: 1, alignSelf: 'flex-start' }}
                  />

                  <Typography 
                    variant="body2" 
                    color="text.secondary" 
                    sx={{ 
                      mb: 1.5,
                      fontSize: '0.75rem',
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                      flexGrow: 1,
                    }}
                  >
                    {rollout.description || 'No description'}
                  </Typography>


                  {rollout.type === 'percentage' && rollout.config.percentage && (
                    <Box sx={{ mb: 1.5 }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <LinearProgress
                          variant="determinate"
                          value={rollout.config.percentage}
                          sx={{ flexGrow: 1, height: 6, borderRadius: 3 }}
                        />
                        <Typography variant="caption" fontWeight="bold" sx={{ fontSize: '0.7rem' }}>
                          {rollout.config.percentage}%
                        </Typography>
                      </Box>
                    </Box>
                  )}

                  {rollout.type === 'experiment' && rollout.config.variations && (
                    <Box sx={{ mb: 1.5 }}>
                      <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                        {rollout.config.variations.slice(0, 2).map((variation, index) => (
                          <Chip
                            key={index}
                            label={`${variation.weight}%`}
                            size="small"
                            variant="outlined"
                            sx={{ fontSize: '0.65rem', height: 20 }}
                          />
                        ))}
                        {rollout.config.variations.length > 2 && (
                          <Chip
                            label={`+${rollout.config.variations.length - 2}`}
                            size="small"
                            variant="outlined"
                            sx={{ fontSize: '0.65rem', height: 20 }}
                          />
                        )}
                      </Box>
                    </Box>
                  )}

                  <Box sx={{ display: 'flex', gap: 0.5, mb: 1.5, flexWrap: 'wrap' }}>
                    <Chip 
                      label={rollout.type} 
                      size="small" 
                      variant="outlined"
                      sx={{ fontSize: '0.65rem', height: 20 }}
                    />
                    {rollout.config.sticky_bucketing && (
                      <Chip label="Sticky" size="small" variant="outlined" sx={{ fontSize: '0.65rem', height: 20 }} />
                    )}
                  </Box>

                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 'auto' }}>
                    <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
                      {rollout.updated_at ? new Date(rollout.updated_at).toLocaleDateString() : 'Unknown'}
                    </Typography>
                    <Box sx={{ display: 'flex', gap: 0.5 }}>
                      {rollout.status === 'draft' && (
                        <Tooltip title="Activate">
                          <IconButton 
                            size="small"
                            color="primary"
                            onClick={() => handleStatusChange(rollout.id, 'activate')}
                            sx={{ p: 0.5 }}
                          >
                            <PlayIcon sx={{ fontSize: '1rem' }} />
                          </IconButton>
                        </Tooltip>
                      )}
                      {rollout.status === 'active' && (
                        <Tooltip title="Pause">
                          <IconButton 
                            size="small"
                            color="warning"
                            onClick={() => handleStatusChange(rollout.id, 'pause')}
                            sx={{ p: 0.5 }}
                          >
                            <PauseIcon sx={{ fontSize: '1rem' }} />
                          </IconButton>
                        </Tooltip>
                      )}
                      {rollout.status === 'paused' && (
                        <Tooltip title="Resume">
                          <IconButton 
                            size="small"
                            color="primary"
                            onClick={() => handleStatusChange(rollout.id, 'activate')}
                            sx={{ p: 0.5 }}
                          >
                            <PlayIcon sx={{ fontSize: '1rem' }} />
                          </IconButton>
                        </Tooltip>
                      )}
                      <Tooltip title="Edit">
                        <IconButton 
                          size="small"
                          onClick={() => handleEditRollout(rollout)}
                          sx={{ p: 0.5 }}
                        >
                          <EditIcon sx={{ fontSize: '1rem' }} />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton 
                          size="small"
                          color="error"
                          onClick={() => handleDeleteRollout(rollout.id)}
                          sx={{ p: 0.5 }}
                        >
                          <DeleteIcon sx={{ fontSize: '1rem' }} />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Create Rollout Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Create New Rollout</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Feature Flag</InputLabel>
                <Select
                  value={formData.flag_id}
                  onChange={(e) => setFormData({ ...formData, flag_id: e.target.value })}
                  label="Feature Flag"
                >
                  {flags.map((flag) => (
                    <MenuItem key={flag.id} value={flag.id}>
                      {flag.name} ({flag.key})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Rollout Type</InputLabel>
                <Select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
                  label="Rollout Type"
                >
                  <MenuItem value="percentage">Percentage Rollout</MenuItem>
                  <MenuItem value="experiment">A/B Experiment</MenuItem>
                  <MenuItem value="segment">Segment-based</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Rollout Name"
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

            {formData.type === 'percentage' && (
              <Grid item xs={12}>
                <Typography gutterBottom>
                  Rollout Percentage: {formData.percentage}%
                </Typography>
                <Slider
                  value={formData.percentage}
                  onChange={(e, value) => setFormData({ ...formData, percentage: value as number })}
                  min={0}
                  max={100}
                  marks={[
                    { value: 0, label: '0%' },
                    { value: 25, label: '25%' },
                    { value: 50, label: '50%' },
                    { value: 75, label: '75%' },
                    { value: 100, label: '100%' },
                  ]}
                />
              </Grid>
            )}

            {formData.type === 'experiment' && (
              <Grid item xs={12}>
                <Typography gutterBottom>Variations</Typography>
                {formData.variations.map((variation, index) => (
                  <Box key={index} sx={{ mb: 2 }}>
                    <Grid container spacing={2}>
                      <Grid item xs={6}>
                        <TextField
                          fullWidth
                          label="Variation ID"
                          value={variation.variation_id}
                          onChange={(e) => updateVariation(index, 'variation_id', e.target.value)}
                        />
                      </Grid>
                      <Grid item xs={6}>
                        <TextField
                          fullWidth
                          label="Weight %"
                          type="number"
                          value={variation.weight}
                          onChange={(e) => updateVariation(index, 'weight', parseInt(e.target.value))}
                        />
                      </Grid>
                    </Grid>
                  </Box>
                ))}
              </Grid>
            )}

            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.sticky_bucketing}
                    onChange={(e) => setFormData({ ...formData, sticky_bucketing: e.target.checked })}
                  />
                }
                label="Enable Sticky Bucketing"
              />
            </Grid>

            <Grid item xs={12}>
              <Typography gutterBottom>
                Traffic Allocation: {formData.traffic_allocation}%
              </Typography>
              <Slider
                value={formData.traffic_allocation}
                onChange={(e, value) => setFormData({ ...formData, traffic_allocation: value as number })}
                min={0}
                max={100}
                step={5}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateRollout} variant="contained">
            Create Rollout
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Rollout Dialog */}
      <Dialog open={openEditDialog} onClose={() => setOpenEditDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Edit Rollout</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Feature Flag</InputLabel>
                <Select
                  value={formData.flag_id}
                  onChange={(e) => setFormData({ ...formData, flag_id: e.target.value })}
                  label="Feature Flag"
                  disabled
                >
                  {flags.map((flag) => (
                    <MenuItem key={flag.id} value={flag.id}>
                      {flag.name} ({flag.key})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Rollout Type</InputLabel>
                <Select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
                  label="Rollout Type"
                  disabled
                >
                  <MenuItem value="percentage">Percentage Rollout</MenuItem>
                  <MenuItem value="experiment">A/B Experiment</MenuItem>
                  <MenuItem value="segment">Segment-based</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Rollout Name"
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

            {formData.type === 'percentage' && (
              <Grid item xs={12}>
                <Typography gutterBottom>
                  Rollout Percentage: {formData.percentage}%
                </Typography>
                <Slider
                  value={formData.percentage}
                  onChange={(e, value) => setFormData({ ...formData, percentage: value as number })}
                  min={0}
                  max={100}
                  marks={[
                    { value: 0, label: '0%' },
                    { value: 25, label: '25%' },
                    { value: 50, label: '50%' },
                    { value: 75, label: '75%' },
                    { value: 100, label: '100%' },
                  ]}
                />
              </Grid>
            )}

            {formData.type === 'experiment' && (
              <Grid item xs={12}>
                <Typography gutterBottom>Variations</Typography>
                {formData.variations.map((variation, index) => (
                  <Box key={index} sx={{ mb: 2 }}>
                    <Grid container spacing={2}>
                      <Grid item xs={6}>
                        <TextField
                          fullWidth
                          label="Variation ID"
                          value={variation.variation_id}
                          onChange={(e) => updateVariation(index, 'variation_id', e.target.value)}
                        />
                      </Grid>
                      <Grid item xs={6}>
                        <TextField
                          fullWidth
                          label="Weight %"
                          type="number"
                          value={variation.weight}
                          onChange={(e) => updateVariation(index, 'weight', parseInt(e.target.value))}
                        />
                      </Grid>
                    </Grid>
                  </Box>
                ))}
              </Grid>
            )}

            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.sticky_bucketing}
                    onChange={(e) => setFormData({ ...formData, sticky_bucketing: e.target.checked })}
                  />
                }
                label="Enable Sticky Bucketing"
              />
            </Grid>

            <Grid item xs={12}>
              <Typography gutterBottom>
                Traffic Allocation: {formData.traffic_allocation}%
              </Typography>
              <Slider
                value={formData.traffic_allocation}
                onChange={(e, value) => setFormData({ ...formData, traffic_allocation: value as number })}
                min={0}
                max={100}
                step={5}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenEditDialog(false)}>Cancel</Button>
          <Button onClick={handleUpdateRollout} variant="contained">
            Update Rollout
          </Button>
        </DialogActions>
      </Dialog>

      {/* Analytics Dialog */}
      <Dialog open={openAnalyticsDialog} onClose={() => setOpenAnalyticsDialog(false)} maxWidth="lg" fullWidth>
        <DialogTitle>Rollout Analytics: {selectedRollout?.name}</DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 2 }}>
            Real-time analytics coming soon. This will show user distribution, conversion rates, and statistical significance.
          </Alert>
          
          {selectedRollout?.type === 'experiment' && (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Variation</TableCell>
                    <TableCell align="right">Users</TableCell>
                    <TableCell align="right">Conversions</TableCell>
                    <TableCell align="right">Conversion Rate</TableCell>
                    <TableCell align="right">Confidence</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {selectedRollout.config.variations?.map((variation, index) => (
                    <TableRow key={index}>
                      <TableCell>{variation.variation_id}</TableCell>
                      <TableCell align="right">-</TableCell>
                      <TableCell align="right">-</TableCell>
                      <TableCell align="right">-</TableCell>
                      <TableCell align="right">-</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenAnalyticsDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}