'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
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
  Grid,
  Alert,
  Tooltip,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Group as GroupIcon,
  ContentCopy as CopyIcon,
  Science as TestIcon,
} from '@mui/icons-material';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';

interface TargetingRule {
  attribute: string;
  operator: string;
  values: string[];
}

interface Segment {
  id: string;
  key: string;
  name: string;
  description: string;
  environment: string;
  rules: TargetingRule[];
  created_at: string;
  updated_at: string;
}

const operators = [
  { value: 'equals', label: 'Equals' },
  { value: 'not_equals', label: 'Not Equals' },
  { value: 'contains', label: 'Contains' },
  { value: 'not_contains', label: 'Not Contains' },
  { value: 'starts_with', label: 'Starts With' },
  { value: 'ends_with', label: 'Ends With' },
  { value: 'in', label: 'In List' },
  { value: 'not_in', label: 'Not In List' },
  { value: 'greater_than', label: 'Greater Than' },
  { value: 'less_than', label: 'Less Than' },
  { value: 'regex', label: 'Regex Match' },
];

export default function ProjectSegmentsPage() {
  const { currentEnvironment } = useEnvironment();
  const params = useParams();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [segments, setSegments] = useState<Segment[]>([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [openTestDialog, setOpenTestDialog] = useState(false);
  const [selectedSegment, setSelectedSegment] = useState<Segment | null>(null);
  const [testResult, setTestResult] = useState<any>(null);
  const [formData, setFormData] = useState({
    key: '',
    name: '',
    description: '',
    environment: 'production',
    rules: [{ attribute: '', operator: 'equals', values: [''] }],
  });
  const [testData, setTestData] = useState({
    user_id: '',
    email: '',
    country: '',
    plan: '',
    custom_attributes: {},
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
    fetchSegments();
  }, [currentEnvironment, projectId]);

  const fetchSegments = async () => {
    if (!projectId) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/segments?project_id=${projectId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        setSegments(data.segments || []);
      }
    } catch (error) {
      console.error('Error fetching segments:', error);
    }
  };

  const handleCreateSegment = async () => {
    if (!projectId) {
      console.error('Project ID is required');
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const segmentData = {
        ...formData,
        project_id: projectId,
        environment: currentEnvironment,
      };
      
      const response = await fetch('http://localhost:8080/api/v1/segments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
          'X-Environment': currentEnvironment,
        },
        body: JSON.stringify(segmentData),
      });
      
      if (response.ok) {
        setOpenDialog(false);
        fetchSegments();
        resetForm();
      }
    } catch (error) {
      console.error('Error creating segment:', error);
    }
  };

  const handleTestSegment = async () => {
    if (!selectedSegment) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('http://localhost:8080/api/v1/segments/evaluate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          project_id: projectId,
          segment_key: selectedSegment.key,
          user_key: testData.user_id || 'test_user',
          user_id: testData.user_id,
          attributes: testData,
          environment: currentEnvironment,
        }),
      });
      
      if (response.ok) {
        const result = await response.json();
        setTestResult(result);
      }
    } catch (error) {
      console.error('Error testing segment:', error);
    }
  };

  const handleDeleteSegment = async (key: string) => {
    if (!confirm('Are you sure you want to delete this segment?')) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/segments/${key}?project_id=${projectId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        fetchSegments();
      }
    } catch (error) {
      console.error('Error deleting segment:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      key: '',
      name: '',
      description: '',
      environment: currentEnvironment,
      rules: [{ attribute: '', operator: 'equals', values: [''] }],
    });
  };

  const addRule = () => {
    setFormData({
      ...formData,
      rules: [...formData.rules, { attribute: '', operator: 'equals', values: [''] }],
    });
  };

  const updateRule = (index: number, field: string, value: any) => {
    const newRules = [...formData.rules];
    newRules[index] = { ...newRules[index], [field]: value };
    setFormData({ ...formData, rules: newRules });
  };

  const removeRule = (index: number) => {
    const newRules = formData.rules.filter((_, i) => i !== index);
    setFormData({ ...formData, rules: newRules });
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
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            User Segments
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Create and manage user segments for {project.name} in {currentEnvironment} environment
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpenDialog(true)}
          size="large"
        >
          Create Segment
        </Button>
      </Box>

      {segments.length === 0 ? (
        <Box textAlign="center" py={8}>
          <GroupIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No segments found
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Create your first user segment to enable targeted feature rollouts
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
          >
            Create First Segment
          </Button>
        </Box>
      ) : (
        <Grid container spacing={3}>
          {segments.map((segment) => (
            <Grid item xs={12} md={6} lg={4} key={segment.id}>
              <Card sx={{ height: '100%' }}>
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                    <Typography variant="h6" fontWeight="bold">
                      {segment.name}
                    </Typography>
                    <Box>
                      <Tooltip title="Test Segment">
                        <IconButton 
                          size="small" 
                          onClick={() => {
                            setSelectedSegment(segment);
                            setOpenTestDialog(true);
                          }}
                        >
                          <TestIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit">
                        <IconButton size="small">
                          <EditIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton 
                          size="small" 
                          onClick={() => handleDeleteSegment(segment.key)}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </Box>
                  
                  <Box sx={{ mb: 2 }}>
                    <Chip 
                      label={segment.key} 
                      size="small" 
                      variant="outlined" 
                      icon={<CopyIcon fontSize="small" />}
                      sx={{ mr: 1 }}
                    />
                    <Chip 
                      label={segment.environment} 
                      size="small" 
                      color="primary"
                    />
                  </Box>

                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    {segment.description || 'No description provided'}
                  </Typography>

                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Targeting Rules
                    </Typography>
                    {segment.rules.map((rule, index) => (
                      <Box key={index} sx={{ mt: 1 }}>
                        <Chip
                          label={`${rule.attribute} ${rule.operator} ${rule.values.join(', ')}`}
                          size="small"
                          sx={{ mb: 0.5 }}
                        />
                      </Box>
                    ))}
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Create/Edit Segment Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Create New Segment</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Segment Key"
                value={formData.key}
                onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                helperText="Unique identifier for the segment"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Segment Name"
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
            
            <Grid item xs={12}>
              <Typography variant="subtitle1" gutterBottom>
                Targeting Rules
              </Typography>
              {formData.rules.map((rule, index) => (
                <Box key={index} sx={{ mb: 2, p: 2, border: 1, borderColor: 'divider', borderRadius: 1 }}>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={4}>
                      <TextField
                        fullWidth
                        label="Attribute"
                        value={rule.attribute}
                        onChange={(e) => updateRule(index, 'attribute', e.target.value)}
                        placeholder="e.g., email, country, plan"
                      />
                    </Grid>
                    <Grid item xs={12} md={3}>
                      <FormControl fullWidth>
                        <InputLabel>Operator</InputLabel>
                        <Select
                          value={rule.operator}
                          onChange={(e) => updateRule(index, 'operator', e.target.value)}
                          label="Operator"
                        >
                          {operators.map((op) => (
                            <MenuItem key={op.value} value={op.value}>
                              {op.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <TextField
                        fullWidth
                        label="Values"
                        value={rule.values.join(', ')}
                        onChange={(e) => updateRule(index, 'values', e.target.value.split(',').map(v => v.trim()))}
                        placeholder="Comma-separated values"
                      />
                    </Grid>
                    <Grid item xs={12} md={1}>
                      <IconButton onClick={() => removeRule(index)} color="error">
                        <DeleteIcon />
                      </IconButton>
                    </Grid>
                  </Grid>
                </Box>
              ))}
              <Button onClick={addRule} startIcon={<AddIcon />}>
                Add Rule
              </Button>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateSegment} variant="contained">
            Create Segment
          </Button>
        </DialogActions>
      </Dialog>

      {/* Test Segment Dialog */}
      <Dialog open={openTestDialog} onClose={() => setOpenTestDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Test Segment: {selectedSegment?.name}</DialogTitle>
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
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Country"
                value={testData.country}
                onChange={(e) => setTestData({ ...testData, country: e.target.value })}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Plan"
                value={testData.plan}
                onChange={(e) => setTestData({ ...testData, plan: e.target.value })}
              />
            </Grid>
          </Grid>

          {testResult && (
            <Alert 
              severity={testResult.matched ? 'success' : 'info'} 
              sx={{ mt: 3 }}
            >
              <Typography variant="subtitle2">
                Match Result: {testResult.matched ? 'User matches segment' : 'User does not match segment'}
              </Typography>
              {testResult.matched_rules && (
                <Typography variant="body2" sx={{ mt: 1 }}>
                  Matched Rules: {testResult.matched_rules.length}
                </Typography>
              )}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => {
            setOpenTestDialog(false);
            setTestResult(null);
          }}>
            Close
          </Button>
          <Button onClick={handleTestSegment} variant="contained">
            Test Segment
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}