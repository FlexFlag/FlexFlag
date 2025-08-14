'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Button,
  Card,
  CardContent,
  CardActions,
  Grid,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Fab,
  IconButton,
  Chip,
  Alert,
  Snackbar,
} from '@mui/material';
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';

interface Environment {
  id: string;
  key: string;
  name: string;
  description: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

interface CreateEnvironmentRequest {
  name: string;
  key: string;
  description: string;
  sort_order: number;
}

export default function EnvironmentsPage() {
  const params = useParams();
  const projectId = params.projectId as string;
  const { refreshEnvironments } = useEnvironment();

  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedEnvironment, setSelectedEnvironment] = useState<Environment | null>(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' as 'success' | 'error' });
  const [projectSlug, setProjectSlug] = useState<string>('');
  
  const [formData, setFormData] = useState<CreateEnvironmentRequest>({
    name: '',
    key: '',
    description: '',
    sort_order: 0,
  });

  useEffect(() => {
    fetchProjectAndEnvironments();
  }, [projectId]);

  const fetchProjectAndEnvironments = async () => {
    try {
      // First, get all projects to find the one with matching ID
      const projects = await apiClient.getProjects();
      const project = projects.find(p => p.id === projectId);
      
      if (!project) {
        throw new Error('Project not found');
      }
      
      setProjectSlug(project.slug);
      
      // Then fetch environments using the project slug
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/projects/${project.slug}/environments`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch environments');
      }

      const data = await response.json();
      setEnvironments(data.environments || []);
    } catch (error: any) {
      console.error('Error fetching environments:', error);
      showSnackbar(error.message || 'Failed to load environments', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateEnvironment = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/projects/${projectSlug}/environments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create environment');
      }

      showSnackbar('Environment created successfully', 'success');
      setCreateDialogOpen(false);
      resetForm();
      await fetchProjectAndEnvironments();
      // Refresh the environment context so the selector updates
      await refreshEnvironments();
    } catch (error: any) {
      console.error('Error creating environment:', error);
      showSnackbar(error.message || 'Failed to create environment', 'error');
    }
  };

  const handleEditEnvironment = async () => {
    if (!selectedEnvironment) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/environments/${selectedEnvironment.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          sort_order: formData.sort_order,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update environment');
      }

      showSnackbar('Environment updated successfully', 'success');
      setEditDialogOpen(false);
      resetForm();
      await fetchProjectAndEnvironments();
      // Refresh the environment context so the selector updates
      await refreshEnvironments();
    } catch (error: any) {
      console.error('Error updating environment:', error);
      showSnackbar(error.message || 'Failed to update environment', 'error');
    }
  };

  const handleDeleteEnvironment = async (environment: Environment) => {
    if (!confirm(`Are you sure you want to delete the environment "${environment.name}"?`)) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/environments/${environment.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete environment');
      }

      showSnackbar('Environment deleted successfully', 'success');
      await fetchProjectAndEnvironments();
      // Refresh the environment context so the selector updates
      await refreshEnvironments();
    } catch (error: any) {
      console.error('Error deleting environment:', error);
      showSnackbar(error.message || 'Failed to delete environment', 'error');
    }
  };

  const openEditDialog = (environment: Environment) => {
    setSelectedEnvironment(environment);
    setFormData({
      name: environment.name,
      key: environment.key,
      description: environment.description,
      sort_order: environment.sort_order,
    });
    setEditDialogOpen(true);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      key: '',
      description: '',
      sort_order: 0,
    });
    setSelectedEnvironment(null);
  };

  const showSnackbar = (message: string, severity: 'success' | 'error') => {
    setSnackbar({ open: true, message, severity });
  };

  const generateKeyFromName = (name: string) => {
    return name.toLowerCase().replace(/[^a-z0-9]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
  };

  const handleNameChange = (value: string) => {
    setFormData(prev => ({
      ...prev,
      name: value,
      key: !editDialogOpen ? generateKeyFromName(value) : prev.key, // Only auto-generate for new environments
    }));
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography>Loading environments...</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Environments
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setCreateDialogOpen(true)}
          disabled={!projectSlug}
        >
          Add Environment
        </Button>
      </Box>

      <Grid container spacing={3}>
        {environments.map((environment) => (
          <Grid item xs={12} md={6} lg={4} key={environment.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6" component="h2">
                    {environment.name}
                  </Typography>
                  <Chip
                    label={environment.is_active ? 'Active' : 'Inactive'}
                    color={environment.is_active ? 'success' : 'default'}
                    size="small"
                  />
                </Box>
                <Typography color="text.secondary" gutterBottom>
                  Key: <code>{environment.key}</code>
                </Typography>
                <Box sx={{ minHeight: '2.5rem', mb: 2 }}>
                  <Typography variant="body2" color="text.secondary">
                    {environment.description || 'No description provided'}
                  </Typography>
                </Box>
                <Typography variant="body2" color="text.secondary">
                  Sort Order: {environment.sort_order}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Created: {new Date(environment.created_at).toLocaleDateString()}
                </Typography>
              </CardContent>
              <CardActions>
                <IconButton
                  size="small"
                  onClick={() => openEditDialog(environment)}
                  color="primary"
                >
                  <EditIcon />
                </IconButton>
                <IconButton
                  size="small"
                  onClick={() => handleDeleteEnvironment(environment)}
                  color="error"
                  disabled={['production', 'staging', 'development'].includes(environment.key)}
                >
                  <DeleteIcon />
                </IconButton>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Create Environment Dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Environment</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={formData.name}
            onChange={(e) => handleNameChange(e.target.value)}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Key"
            fullWidth
            variant="outlined"
            value={formData.key}
            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
            helperText="URL-friendly identifier (e.g., qa, pre-prod)"
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Sort Order"
            type="number"
            fullWidth
            variant="outlined"
            value={formData.sort_order}
            onChange={(e) => setFormData({ ...formData, sort_order: parseInt(e.target.value) || 0 })}
            helperText="Lower numbers appear first"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateEnvironment} variant="contained">Create</Button>
        </DialogActions>
      </Dialog>

      {/* Edit Environment Dialog */}
      <Dialog open={editDialogOpen} onClose={() => setEditDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Environment</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Key"
            fullWidth
            variant="outlined"
            value={formData.key}
            disabled
            helperText="Environment key cannot be changed"
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Sort Order"
            type="number"
            fullWidth
            variant="outlined"
            value={formData.sort_order}
            onChange={(e) => setFormData({ ...formData, sort_order: parseInt(e.target.value) || 0 })}
            helperText="Lower numbers appear first"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleEditEnvironment} variant="contained">Update</Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}