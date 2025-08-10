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
} from '@mui/material';
import {
  Add as AddIcon,
  Science as ExperimentIcon,
  TrendingUp as TrendingUpIcon,
  Assessment as AnalyticsIcon,
} from '@mui/icons-material';

export default function ProjectExperimentsPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);

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
            Experiments
          </Typography>
          <Typography variant="body1" color="text.secondary">
            A/B testing and experimentation for {project.name} in {currentEnvironment} environment
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          size="large"
          disabled
        >
          Create Experiment
        </Button>
      </Box>

      {/* Info Alert */}
      <Alert severity="info" sx={{ mb: 3 }}>
        Experiments are managed through Rollouts. Create an "A/B Experiment" type rollout to run experiments.
      </Alert>

      {/* Feature Overview */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, textAlign: 'center' }}>
            <ExperimentIcon sx={{ fontSize: 48, color: 'primary.main', mb: 2 }} />
            <Typography variant="h6" fontWeight="bold" gutterBottom>
              A/B Testing
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Run controlled experiments to test different variations of your features
            </Typography>
          </Paper>
        </Grid>
        
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, textAlign: 'center' }}>
            <AnalyticsIcon sx={{ fontSize: 48, color: 'secondary.main', mb: 2 }} />
            <Typography variant="h6" fontWeight="bold" gutterBottom>
              Statistical Analysis
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Get statistical significance and confidence intervals for your experiments
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, textAlign: 'center' }}>
            <TrendingUpIcon sx={{ fontSize: 48, color: 'success.main', mb: 2 }} />
            <Typography variant="h6" fontWeight="bold" gutterBottom>
              Performance Tracking
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Monitor conversion rates and user engagement across variations
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Quick Actions */}
      <Card>
        <CardContent>
          <Typography variant="h6" fontWeight="bold" gutterBottom>
            Quick Actions
          </Typography>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={4}>
              <Button
                variant="outlined"
                fullWidth
                startIcon={<AddIcon />}
                component="a"
                href={`/projects/${projectId}/rollouts`}
                sx={{ py: 2 }}
              >
                Create A/B Experiment
              </Button>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Button
                variant="outlined"
                fullWidth
                startIcon={<AnalyticsIcon />}
                component="a"
                href={`/projects/${projectId}/rollouts`}
                sx={{ py: 2 }}
              >
                View Running Experiments
              </Button>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Button
                variant="outlined"
                fullWidth
                startIcon={<TrendingUpIcon />}
                component="a"
                href={`/projects/${projectId}/performance`}
                sx={{ py: 2 }}
              >
                View Analytics
              </Button>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Coming Soon Section */}
      <Box sx={{ mt: 4 }}>
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 4 }}>
            <ExperimentIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
            <Typography variant="h5" color="text.secondary" gutterBottom>
              Dedicated Experiments Interface Coming Soon
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
              We're building a dedicated experiments interface with advanced analytics, 
              statistical significance testing, and conversion tracking.
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center', flexWrap: 'wrap' }}>
              <Chip label="Statistical Testing" variant="outlined" />
              <Chip label="Conversion Tracking" variant="outlined" />
              <Chip label="Real-time Analytics" variant="outlined" />
              <Chip label="Experiment Templates" variant="outlined" />
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
}