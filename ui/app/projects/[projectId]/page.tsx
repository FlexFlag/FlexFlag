'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  Paper,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Flag as FlagIcon,
  DonutLarge as RolloutIcon,
  Segment as SegmentIcon,
  Assessment as EvaluationIcon,
  Speed as PerformanceIcon,
  Science as ExperimentIcon,
  TrendingUp as TrendingUpIcon,
  People as PeopleIcon,
} from '@mui/icons-material';

interface ProjectStats {
  flags: number;
  segments: number;
  rollouts: number;
}

export default function ProjectOverview() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProjectData = async () => {
      try {
        setLoading(true);
        
        // Fetch project details
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);

        // Fetch project statistics
        if (foundProject) {
          const projectStats = await apiClient.getProjectStats(foundProject.id);
          setStats(projectStats);
        }
      } catch (error) {
        console.error('Error fetching project data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (projectId) {
      fetchProjectData();
    }
  }, [projectId]);

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <Typography>Loading project...</Typography>
      </Box>
    );
  }

  if (!project) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <Typography color="error">Project not found</Typography>
      </Box>
    );
  }

  const quickActions = [
    {
      title: 'Feature Flags',
      description: 'Manage feature flags',
      icon: <FlagIcon />,
      href: `/projects/${projectId}/flags`,
      count: stats?.flags || 0,
      color: 'primary.main',
    },
    {
      title: 'User Segments',
      description: 'Define user segments',
      icon: <SegmentIcon />,
      href: `/projects/${projectId}/segments`,
      count: stats?.segments || 0,
      color: 'secondary.main',
    },
    {
      title: 'Rollouts',
      description: 'Manage rollouts & experiments',
      icon: <RolloutIcon />,
      href: `/projects/${projectId}/rollouts`,
      count: stats?.rollouts || 0,
      color: 'success.main',
    },
    {
      title: 'Evaluations',
      description: 'Test flag evaluations',
      icon: <EvaluationIcon />,
      href: `/projects/${projectId}/evaluations`,
      count: '-',
      color: 'warning.main',
    },
  ];

  return (
    <Box>
      {/* Project Header */}
      <Box sx={{ mb: 5, pb: 3, borderBottom: '1px solid', borderColor: 'divider' }}>
        <Typography variant="h5" fontWeight="600" gutterBottom sx={{ mb: 1 }}>
          Project Overview
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {project.description || 'Monitor and manage your feature flags and configurations'}
        </Typography>
      </Box>

      {/* Quick Stats */}
      <Grid container spacing={3} sx={{ mb: 5 }}>
        <Grid item xs={12} sm={3}>
          <Paper sx={{ 
            p: 3, 
            textAlign: 'center', 
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: 0,
            bgcolor: 'background.paper',
            minHeight: 120,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center'
          }}>
            <Typography variant="h3" fontWeight="700" color="primary.main" sx={{ mb: 1 }}>
              {stats?.flags || 0}
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ 
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: '0.7rem'
            }}>
              Feature Flags
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={3}>
          <Paper sx={{ 
            p: 3, 
            textAlign: 'center', 
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: 0,
            bgcolor: 'background.paper',
            minHeight: 120,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center'
          }}>
            <Typography variant="h3" fontWeight="700" color="secondary.main" sx={{ mb: 1 }}>
              {stats?.segments || 0}
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ 
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: '0.7rem'
            }}>
              User Segments
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={3}>
          <Paper sx={{ 
            p: 3, 
            textAlign: 'center', 
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: 0,
            bgcolor: 'background.paper',
            minHeight: 120,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center'
          }}>
            <Typography variant="h3" fontWeight="700" color="success.main" sx={{ mb: 1 }}>
              {stats?.rollouts || 0}
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ 
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: '0.7rem'
            }}>
              Rollouts
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={3}>
          <Paper sx={{ 
            p: 3, 
            textAlign: 'center', 
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: 0,
            bgcolor: 'background.paper',
            minHeight: 120,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center'
          }}>
            <Typography variant="h3" fontWeight="700" color="warning.main" sx={{ mb: 1 }}>
              -
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ 
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: '0.7rem'
            }}>
              Active Users
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Quick Actions */}
      <Typography variant="h6" fontWeight="600" gutterBottom sx={{ mb: 3, fontSize: '1.1rem' }}>
        Quick Actions
      </Typography>
      <Grid container spacing={3}>
        {quickActions.map((action, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <Card 
              component="a"
              href={action.href}
              sx={{ 
                textDecoration: 'none',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                border: '1px solid',
                borderColor: 'divider',
                boxShadow: 0,
                '&:hover': {
                  transform: 'translateY(-2px)',
                  boxShadow: 2,
                  borderColor: action.color,
                },
              }}
            >
              <CardContent sx={{ textAlign: 'center', py: 3 }}>
                <Box sx={{ color: action.color, mb: 2 }}>
                  {action.icon}
                </Box>
                <Typography variant="subtitle2" fontWeight="600" gutterBottom sx={{ mb: 1 }}>
                  {action.title}
                </Typography>
                <Chip 
                  label={`${action.count} ${action.count === '-' ? 'items' : action.count === 1 ? 'item' : 'items'}`} 
                  size="small" 
                  variant="outlined"
                  sx={{ fontSize: '0.7rem' }}
                />
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Recent Activity */}
      <Box sx={{ mt: 4 }}>
        <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ mb: 2 }}>
          Recent Activity
        </Typography>
        <Paper sx={{ p: 2, textAlign: 'center', minHeight: 120, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
          <TrendingUpIcon sx={{ fontSize: 32, color: 'grey.400', mb: 1 }} />
          <Typography variant="subtitle1" color="text.secondary" gutterBottom>
            Activity Feed Coming Soon
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Track changes, evaluations, and team activities
          </Typography>
        </Paper>
      </Box>
    </Box>
  );
}