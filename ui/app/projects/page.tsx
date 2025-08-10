'use client';

import { useState, useEffect } from 'react';
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
  Avatar,
  AvatarGroup,
  Tooltip,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Settings as SettingsIcon,
  AccountTree as ProjectIcon,
  Group as TeamIcon,
  Launch as LaunchIcon,
} from '@mui/icons-material';
import { useRouter } from 'next/navigation';

interface Project {
  id: string;
  slug: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

interface ProjectStats {
  flags: number;
  segments: number;
  rollouts: number;
}

export default function ProjectsPage() {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectStats, setProjectStats] = useState<Record<string, ProjectStats>>({});
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openManageDialog, setOpenManageDialog] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [managingProject, setManagingProject] = useState<Project | null>(null);
  const [formData, setFormData] = useState({
    slug: '',
    name: '',
    description: '',
  });

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      const data = await apiClient.getProjects();
      setProjects(data || []);
      
      // Fetch stats for each project
      const statsPromises = (data || []).map(async (project: Project) => {
        try {
          const stats = await apiClient.getProjectStats(project.id);
          return { projectId: project.id, stats };
        } catch (error) {
          console.error(`Error fetching stats for project ${project.id}:`, error);
          return { projectId: project.id, stats: { flags: 0, segments: 0, rollouts: 0 } };
        }
      });
      
      const statsResults = await Promise.all(statsPromises);
      const statsMap = statsResults.reduce((acc, { projectId, stats }) => {
        acc[projectId] = stats;
        return acc;
      }, {} as Record<string, ProjectStats>);
      
      setProjectStats(statsMap);
    } catch (error) {
      console.error('Error fetching projects:', error);
    }
  };

  const handleCreateProject = async () => {
    try {
      await apiClient.createProject(formData);
      setOpenDialog(false);
      fetchProjects();
      resetForm();
    } catch (error) {
      console.error('Error creating project:', error);
    }
  };

  const handleEditProject = (project: Project) => {
    setEditingProject(project);
    setFormData({
      slug: project.slug,
      name: project.name,
      description: project.description,
    });
    setOpenEditDialog(true);
  };

  const handleUpdateProject = async () => {
    if (!editingProject) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/projects/${editingProject.slug}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(formData),
      });
      
      if (response.ok) {
        setOpenEditDialog(false);
        setEditingProject(null);
        fetchProjects();
        resetForm();
      }
    } catch (error) {
      console.error('Error updating project:', error);
    }
  };

  const handleManageProject = (project: Project) => {
    setManagingProject(project);
    setOpenManageDialog(true);
  };

  const handleDeleteProject = async (slug: string) => {
    if (!confirm('Are you sure you want to delete this project? This action cannot be undone.')) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v1/projects/${slug}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        fetchProjects();
      }
    } catch (error) {
      console.error('Error deleting project:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      slug: '',
      name: '',
      description: '',
    });
  };

  const handleOpenProject = (project: Project) => {
    router.push(`/projects/${project.id}`);
  };

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Projects
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage your feature flag projects and team access
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpenDialog(true)}
          size="large"
        >
          Create Project
        </Button>
      </Box>

      <Grid container spacing={3}>
        {projects.map((project) => (
          <Grid item xs={12} md={6} lg={4} key={project.id}>
            <Card 
              sx={{ 
                height: '100%', 
                position: 'relative',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                '&:hover': {
                  transform: 'translateY(-4px)',
                  boxShadow: 4,
                },
              }}
              onClick={() => handleOpenProject(project)}
            >
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Avatar sx={{ bgcolor: 'primary.main', width: 40, height: 40 }}>
                      <ProjectIcon />
                    </Avatar>
                    <Box>
                      <Typography variant="h6" fontWeight="bold">
                        {project.name}
                      </Typography>
                      <Chip 
                        label={project.slug} 
                        size="small" 
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                  <Box onClick={(e) => e.stopPropagation()}>
                    <IconButton 
                      size="small"
                      onClick={() => handleManageProject(project)}
                    >
                      <SettingsIcon fontSize="small" />
                    </IconButton>
                    <IconButton 
                      size="small"
                      onClick={() => handleEditProject(project)}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton 
                      size="small"
                      onClick={() => handleDeleteProject(project.slug)}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Box>

                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                  {project.description || 'No description provided'}
                </Typography>

                {/* Click to open indicator */}
                <Box sx={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: 1, 
                  mb: 2,
                  opacity: 0.7,
                  fontSize: '0.75rem',
                  color: 'text.secondary'
                }}>
                  <LaunchIcon sx={{ fontSize: 16 }} />
                  <Typography variant="caption">Click card to open project</Typography>
                </Box>

                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="caption" color="text.secondary">
                      Team Members
                    </Typography>
                    <AvatarGroup max={4} sx={{ '& .MuiAvatar-root': { width: 24, height: 24, fontSize: '0.75rem' } }}>
                      <Avatar>A</Avatar>
                      <Avatar>B</Avatar>
                      <Avatar>C</Avatar>
                      <Avatar>+5</Avatar>
                    </AvatarGroup>
                  </Box>
                  <Box onClick={(e) => e.stopPropagation()}>
                    <Tooltip title="Open Project">
                      <IconButton 
                        size="small"
                        color="primary"
                        onClick={() => handleOpenProject(project)}
                        sx={{ mr: 1 }}
                      >
                        <LaunchIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Button 
                      size="small" 
                      startIcon={<TeamIcon />}
                      onClick={() => handleManageProject(project)}
                    >
                      Manage
                    </Button>
                  </Box>
                </Box>

                <Box sx={{ mt: 2, pt: 2, borderTop: 1, borderColor: 'divider' }}>
                  <Grid container spacing={2}>
                    <Grid item xs={4}>
                      <Typography variant="caption" color="text.secondary">
                        Flags
                      </Typography>
                      <Typography variant="h6">{projectStats[project.id]?.flags || 0}</Typography>
                    </Grid>
                    <Grid item xs={4}>
                      <Typography variant="caption" color="text.secondary">
                        Segments
                      </Typography>
                      <Typography variant="h6">{projectStats[project.id]?.segments || 0}</Typography>
                    </Grid>
                    <Grid item xs={4}>
                      <Typography variant="caption" color="text.secondary">
                        Rollouts
                      </Typography>
                      <Typography variant="h6">{projectStats[project.id]?.rollouts || 0}</Typography>
                    </Grid>
                  </Grid>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}

        {/* Add New Project Card */}
        <Grid item xs={12} md={6} lg={4}>
          <Card 
            sx={{ 
              height: '100%', 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center',
              minHeight: 300,
              border: '2px dashed',
              borderColor: 'divider',
              cursor: 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                borderColor: 'primary.main',
                bgcolor: 'primary.50',
              }
            }}
            onClick={() => setOpenDialog(true)}
          >
            <CardContent sx={{ textAlign: 'center' }}>
              <AddIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 2 }} />
              <Typography variant="h6" color="text.secondary">
                Create New Project
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Create Project Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Project</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Project Slug"
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                helperText="URL-friendly identifier for the project (e.g., my-project)"
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Project Name"
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
                rows={3}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateProject} variant="contained">
            Create Project
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Project Dialog */}
      <Dialog open={openEditDialog} onClose={() => setOpenEditDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Project</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Project Slug"
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                helperText="URL-friendly identifier"
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Project Name"
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
                rows={3}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenEditDialog(false)}>Cancel</Button>
          <Button onClick={handleUpdateProject} variant="contained">
            Update Project
          </Button>
        </DialogActions>
      </Dialog>

      {/* Manage Project Dialog */}
      <Dialog open={openManageDialog} onClose={() => setOpenManageDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Manage Project: {managingProject?.name}</DialogTitle>
        <DialogContent>
          <Typography variant="body1" sx={{ mb: 3 }}>
            User management functionality will be implemented here. This will include:
          </Typography>
          <Box component="ul" sx={{ pl: 2 }}>
            <li>Add/remove team members</li>
            <li>Assign roles and permissions</li>
            <li>View user activity</li>
            <li>Manage project access levels</li>
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            Coming soon...
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenManageDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}