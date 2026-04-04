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
  LinearProgress,
} from '@mui/material';
import {
  Flag as FlagIcon,
  DonutLarge as RolloutIcon,
  Segment as SegmentIcon,
  Assessment as EvaluationIcon,
  Bolt as BoltIcon,
  People as PeopleIcon,
} from '@mui/icons-material';

interface ProjectStats {
  flags: number;
  segments: number;
  rollouts: number;
}

interface FlagStat {
  total_evaluations: number;
  unique_users: number;
  last_evaluated_at?: string;
  variation_counts: Record<string, number>;
}

export default function ProjectOverview() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const [analytics, setAnalytics] = useState<Record<string, FlagStat>>({});
  const [totalEvaluations, setTotalEvaluations] = useState(0);
  const [totalUsers, setTotalUsers] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProjectData = async () => {
      try {
        setLoading(true);

        const projects = await apiClient.getProjects();
        const foundProject = projects.find((p: any) => p.id === projectId);
        setProject(foundProject);

        if (foundProject) {
          const [projectStats, analyticsData] = await Promise.allSettled([
            apiClient.getProjectStats(foundProject.id),
            apiClient.getFlagAnalytics(currentEnvironment),
          ]);

          if (projectStats.status === 'fulfilled') setStats(projectStats.value);

          if (analyticsData.status === 'fulfilled') {
            const flags = analyticsData.value.flags ?? {};
            setAnalytics(flags);
            const evals = Object.values(flags).reduce((s, f) => s + f.total_evaluations, 0);
            const users = Object.values(flags).reduce((s, f) => s + f.unique_users, 0);
            setTotalEvaluations(evals);
            setTotalUsers(users);
          }
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
  }, [projectId, currentEnvironment]);

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
              {totalEvaluations > 999 ? `${(totalEvaluations / 1000).toFixed(1)}k` : totalEvaluations}
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: '0.7rem'
            }}>
              Evaluations
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

      {/* Evaluation Analytics */}
      <Box sx={{ mt: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
          <BoltIcon sx={{ color: 'warning.main', fontSize: 20 }} />
          <Typography variant="h6" fontWeight="600" sx={{ fontSize: '1.1rem' }}>
            Flag Evaluations
          </Typography>
          <Chip label={currentEnvironment} size="small" variant="outlined" sx={{ ml: 'auto', fontSize: '0.7rem' }} />
        </Box>

        {Object.keys(analytics).length === 0 ? (
          <Paper sx={{ p: 4, textAlign: 'center', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
            <EvaluationIcon sx={{ fontSize: 36, color: 'grey.400', mb: 1 }} />
            <Typography variant="body2" color="text.secondary">
              No evaluations yet. Evaluations appear here as your SDK calls are made.
            </Typography>
          </Paper>
        ) : (
          <Grid container spacing={2}>
            {Object.entries(analytics)
              .sort(([, a], [, b]) => b.total_evaluations - a.total_evaluations)
              .slice(0, 6)
              .map(([flagKey, stat]) => {
                const maxEvals = Math.max(...Object.values(analytics).map(s => s.total_evaluations), 1);
                const pct = Math.round((stat.total_evaluations / maxEvals) * 100);
                const lastEval = stat.last_evaluated_at
                  ? new Date(stat.last_evaluated_at).toLocaleString()
                  : null;
                return (
                  <Grid item xs={12} md={6} key={flagKey}>
                    <Paper sx={{ p: 2.5, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1.5 }}>
                        <Typography variant="body2" fontWeight="600" sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
                          {flagKey}
                        </Typography>
                        <Box sx={{ display: 'flex', gap: 0.5 }}>
                          <Chip
                            icon={<BoltIcon sx={{ fontSize: '12px !important' }} />}
                            label={stat.total_evaluations.toLocaleString()}
                            size="small"
                            color="warning"
                            variant="outlined"
                            sx={{ fontSize: '0.65rem', height: 20 }}
                          />
                          <Chip
                            icon={<PeopleIcon sx={{ fontSize: '12px !important' }} />}
                            label={stat.unique_users.toLocaleString()}
                            size="small"
                            variant="outlined"
                            sx={{ fontSize: '0.65rem', height: 20 }}
                          />
                        </Box>
                      </Box>
                      <LinearProgress
                        variant="determinate"
                        value={pct}
                        sx={{ mb: 1.5, borderRadius: 1, height: 4, bgcolor: 'action.hover' }}
                        color="warning"
                      />
                      {Object.keys(stat.variation_counts).length > 0 && (
                        <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', mb: 1 }}>
                          {Object.entries(stat.variation_counts).map(([v, count]) => (
                            <Chip
                              key={v}
                              label={`${v}: ${count}`}
                              size="small"
                              sx={{ fontSize: '0.6rem', height: 18, bgcolor: 'action.selected' }}
                            />
                          ))}
                        </Box>
                      )}
                      {lastEval && (
                        <Typography variant="caption" color="text.disabled" sx={{ fontSize: '0.65rem' }}>
                          Last: {lastEval}
                        </Typography>
                      )}
                    </Paper>
                  </Grid>
                );
              })}
          </Grid>
        )}

        {/* Summary row */}
        {Object.keys(analytics).length > 0 && (
          <Box sx={{ mt: 2, display: 'flex', gap: 2 }}>
            <Paper sx={{ p: 2, flex: 1, border: '1px solid', borderColor: 'divider', boxShadow: 0, display: 'flex', alignItems: 'center', gap: 1.5 }}>
              <BoltIcon sx={{ color: 'warning.main' }} />
              <Box>
                <Typography variant="h6" fontWeight="700">{totalEvaluations.toLocaleString()}</Typography>
                <Typography variant="caption" color="text.secondary">Total evaluations</Typography>
              </Box>
            </Paper>
            <Paper sx={{ p: 2, flex: 1, border: '1px solid', borderColor: 'divider', boxShadow: 0, display: 'flex', alignItems: 'center', gap: 1.5 }}>
              <PeopleIcon sx={{ color: 'primary.main' }} />
              <Box>
                <Typography variant="h6" fontWeight="700">{totalUsers.toLocaleString()}</Typography>
                <Typography variant="caption" color="text.secondary">Unique users tracked</Typography>
              </Box>
            </Paper>
            <Paper sx={{ p: 2, flex: 1, border: '1px solid', borderColor: 'divider', boxShadow: 0, display: 'flex', alignItems: 'center', gap: 1.5 }}>
              <FlagIcon sx={{ color: 'success.main' }} />
              <Box>
                <Typography variant="h6" fontWeight="700">{Object.keys(analytics).length}</Typography>
                <Typography variant="caption" color="text.secondary">Flags with activity</Typography>
              </Box>
            </Paper>
          </Box>
        )}
      </Box>
    </Box>
  );
}