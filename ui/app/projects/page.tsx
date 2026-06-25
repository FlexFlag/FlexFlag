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
  PersonAdd as PersonAddIcon,
  Close as CloseIcon,
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

interface ProjectMember {
  member_id: string;
  role: string;
  joined_at: string;
  user_id: string;
  email: string;
  full_name: string;
  global_role: string;
}

export default function ProjectsPage() {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectStats, setProjectStats] = useState<Record<string, ProjectStats>>({});
  const [projectMembers, setProjectMembers] = useState<Record<string, ProjectMember[]>>({});
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openManageDialog, setOpenManageDialog] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [managingProject, setManagingProject] = useState<Project | null>(null);
  const [allUsers, setAllUsers] = useState<any[]>([]);
  const [selectedUser, setSelectedUser] = useState('');
  const [selectedRole, setSelectedRole] = useState('viewer');
  const [formData, setFormData] = useState({
    slug: '',
    name: '',
    description: '',
  });

  useEffect(() => {
    fetchProjects();
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const users = await apiClient.getUsers();
      setAllUsers(users);
    } catch (error) {
      console.error('Error fetching users:', error);
    }
  };

  const fetchProjects = async () => {
    try {
      const data = await apiClient.getProjects();
      setProjects(data || []);

      // Fetch stats and members for each project
      const dataPromises = (data || []).map(async (project: Project) => {
        try {
          const [stats, members] = await Promise.all([
            apiClient.getProjectStats(project.id),
            apiClient.getProjectMembers(project.id)
          ]);
          return { projectId: project.id, stats, members };
        } catch (error) {
          console.error(`Error fetching data for project ${project.id}:`, error);
          return {
            projectId: project.id,
            stats: { flags: 0, segments: 0, rollouts: 0 },
            members: []
          };
        }
      });

      const results = await Promise.all(dataPromises);
      const statsMap = results.reduce((acc, { projectId, stats }) => {
        acc[projectId] = stats;
        return acc;
      }, {} as Record<string, ProjectStats>);

      const membersMap = results.reduce((acc, { projectId, members }) => {
        acc[projectId] = members;
        return acc;
      }, {} as Record<string, ProjectMember[]>);

      setProjectStats(statsMap);
      setProjectMembers(membersMap);
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
      const response = await fetch(`/api/v1/projects/${editingProject.slug}`, {
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
    setSelectedUser('');
    setSelectedRole('viewer');
    setOpenManageDialog(true);
  };

  const handleAddMember = async () => {
    if (!managingProject || !selectedUser) return;

    try {
      await apiClient.addProjectMember(managingProject.id, selectedUser, selectedRole);
      // Refresh project members
      const members = await apiClient.getProjectMembers(managingProject.id);
      setProjectMembers(prev => ({ ...prev, [managingProject.id]: members }));
      setSelectedUser('');
      setSelectedRole('viewer');
    } catch (error) {
      console.error('Error adding member:', error);
      alert('Failed to add member');
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!managingProject) return;

    if (!confirm('Are you sure you want to remove this member?')) return;

    try {
      await apiClient.removeProjectMember(managingProject.id, userId);
      // Refresh project members
      const members = await apiClient.getProjectMembers(managingProject.id);
      setProjectMembers(prev => ({ ...prev, [managingProject.id]: members }));
    } catch (error) {
      console.error('Error removing member:', error);
      alert('Failed to remove member');
    }
  };

  const handleUpdateMemberRole = async (userId: string, newRole: string) => {
    if (!managingProject) return;

    try {
      await apiClient.updateProjectMemberRole(managingProject.id, userId, newRole);
      // Refresh project members
      const members = await apiClient.getProjectMembers(managingProject.id);
      setProjectMembers(prev => ({ ...prev, [managingProject.id]: members }));
    } catch (error) {
      console.error('Error updating member role:', error);
      alert('Failed to update member role');
    }
  };

  const handleDeleteProject = async (slug: string) => {
    if (!confirm('Are you sure you want to delete this project? This action cannot be undone.')) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/projects/${slug}`, {
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
                    {projectMembers[project.id] && projectMembers[project.id].length > 0 ? (
                      <AvatarGroup max={4} sx={{ '& .MuiAvatar-root': { width: 24, height: 24, fontSize: '0.75rem' } }}>
                        {projectMembers[project.id].map((member) => (
                          <Tooltip key={member.member_id} title={`${member.full_name} (${member.role})`}>
                            <Avatar sx={{ bgcolor: 'primary.main' }}>
                              {member.full_name.charAt(0).toUpperCase()}
                            </Avatar>
                          </Tooltip>
                        ))}
                      </AvatarGroup>
                    ) : (
                      <Typography variant="caption" color="text.secondary">
                        No members
                      </Typography>
                    )}
                  </Box>
                  <Box onClick={(e) => e.stopPropagation()}>
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
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <TeamIcon />
            Manage Team: {managingProject?.name}
          </Box>
        </DialogTitle>
        <DialogContent>
          {/* Add Member Section */}
          <Box sx={{ mb: 4, p: 2, bgcolor: 'background.default', borderRadius: 1 }}>
            <Typography variant="subtitle1" fontWeight="bold" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <PersonAddIcon />
              Add Team Member
            </Typography>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              <Grid item xs={12} sm={6}>
                <TextField
                  select
                  fullWidth
                  label="Select User"
                  value={selectedUser}
                  onChange={(e) => setSelectedUser(e.target.value)}
                  SelectProps={{ native: true }}
                >
                  <option value="">-- Select User --</option>
                  {allUsers
                    .filter(user => !managingProject || !projectMembers[managingProject.id]?.some(m => m.user_id === user.id))
                    .map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.full_name} ({user.email})
                      </option>
                    ))}
                </TextField>
              </Grid>
              <Grid item xs={12} sm={4}>
                <TextField
                  select
                  fullWidth
                  label="Role"
                  value={selectedRole}
                  onChange={(e) => setSelectedRole(e.target.value)}
                  SelectProps={{ native: true }}
                >
                  <option value="viewer">Viewer</option>
                  <option value="editor">Editor</option>
                  <option value="admin">Admin</option>
                </TextField>
              </Grid>
              <Grid item xs={12} sm={2}>
                <Button
                  fullWidth
                  variant="contained"
                  onClick={handleAddMember}
                  disabled={!selectedUser}
                  sx={{ height: '56px' }}
                >
                  Add
                </Button>
              </Grid>
            </Grid>
          </Box>

          {/* Current Members Section */}
          <Typography variant="subtitle1" fontWeight="bold" gutterBottom>
            Current Members ({managingProject && projectMembers[managingProject.id] ? projectMembers[managingProject.id].length : 0})
          </Typography>
          <Box sx={{ mt: 2 }}>
            {managingProject && projectMembers[managingProject.id] && projectMembers[managingProject.id].length > 0 ? (
              projectMembers[managingProject.id].map((member) => (
                <Box
                  key={member.member_id}
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    p: 2,
                    mb: 1,
                    border: 1,
                    borderColor: 'divider',
                    borderRadius: 1,
                    '&:hover': {
                      bgcolor: 'action.hover',
                    },
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Avatar sx={{ bgcolor: 'primary.main' }}>
                      {member.full_name.charAt(0).toUpperCase()}
                    </Avatar>
                    <Box>
                      <Typography variant="body1" fontWeight="medium">
                        {member.full_name}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {member.email}
                      </Typography>
                    </Box>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <TextField
                      select
                      size="small"
                      value={member.role}
                      onChange={(e) => handleUpdateMemberRole(member.user_id, e.target.value)}
                      SelectProps={{ native: true }}
                      sx={{ minWidth: 120 }}
                    >
                      <option value="viewer">Viewer</option>
                      <option value="editor">Editor</option>
                      <option value="admin">Admin</option>
                    </TextField>
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() => handleRemoveMember(member.user_id)}
                    >
                      <CloseIcon />
                    </IconButton>
                  </Box>
                </Box>
              ))
            ) : (
              <Box sx={{ textAlign: 'center', py: 4, color: 'text.secondary' }}>
                <Typography variant="body2">
                  No members yet. Add users to this project to get started.
                </Typography>
              </Box>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenManageDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}