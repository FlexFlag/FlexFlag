'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Chip,
  Button,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Science as ExperimentIcon,
  TrendingUp as TrendingUpIcon,
  PlayArrow as PlayIcon,
  Pause as PauseIcon,
  Stop as StopIcon,
  Analytics as AnalyticsIcon,
} from '@mui/icons-material';

interface Experiment {
  id: string;
  name: string;
  status: 'draft' | 'active' | 'paused' | 'completed';
  type: 'experiment';
  config: {
    variations: Array<{ variation_id: string; weight: number }>;
    hypothesis_description?: string;
    success_metrics?: string[];
  };
  analytics?: {
    total_users: number;
    conversions: Record<string, number>;
    statistical_significance?: number;
  };
}

export default function ExperimentsPage() {
  const [experiments, setExperiments] = useState<Experiment[]>([
    {
      id: '1',
      name: 'Checkout Button Color Test',
      status: 'active',
      type: 'experiment',
      config: {
        variations: [
          { variation_id: 'red_button', weight: 50 },
          { variation_id: 'blue_button', weight: 50 },
        ],
        hypothesis_description: 'Red checkout button will increase conversion rate by 15%',
        success_metrics: ['click_through_rate', 'conversion_rate'],
      },
      analytics: {
        total_users: 2847,
        conversions: {
          red_button: 142,
          blue_button: 186,
        },
        statistical_significance: 0.95,
      },
    },
    {
      id: '2',
      name: 'Pricing Page Layout',
      status: 'draft',
      type: 'experiment',
      config: {
        variations: [
          { variation_id: 'layout_a', weight: 33 },
          { variation_id: 'layout_b', weight: 33 },
          { variation_id: 'layout_c', weight: 34 },
        ],
        hypothesis_description: 'New pricing layout will improve signup conversions',
        success_metrics: ['signup_rate', 'page_time'],
      },
    },
  ]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'paused': return 'warning';
      case 'completed': return 'default';
      case 'draft': return 'info';
      default: return 'default';
    }
  };

  const calculateConversionRate = (conversions: number, total: number) => {
    return total > 0 ? ((conversions / total) * 100).toFixed(2) : '0.00';
  };

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            A/B Test Experiments
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Monitor and analyze your A/B testing experiments with statistical significance
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<ExperimentIcon />}
          size="large"
          onClick={() => window.location.href = '/rollouts'}
        >
          Create Experiment
        </Button>
      </Box>

      <Grid container spacing={3}>
        {experiments.map((experiment) => (
          <Grid item xs={12} key={experiment.id}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <ExperimentIcon color="primary" />
                    <Box>
                      <Typography variant="h6" fontWeight="bold">
                        {experiment.name}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {experiment.config.hypothesis_description}
                      </Typography>
                    </Box>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Chip 
                      label={experiment.status} 
                      size="small" 
                      color={getStatusColor(experiment.status) as any}
                    />
                    {experiment.analytics?.statistical_significance && (
                      <Chip 
                        label={`${(experiment.analytics.statistical_significance * 100).toFixed(0)}% Confident`}
                        size="small"
                        color={experiment.analytics.statistical_significance > 0.9 ? 'success' : 'warning'}
                      />
                    )}
                  </Box>
                </Box>

                {experiment.analytics && (
                  <Box sx={{ mb: 3 }}>
                    <Typography variant="subtitle2" gutterBottom>
                      Experiment Results
                    </Typography>
                    <TableContainer component={Paper} variant="outlined">
                      <Table size="small">
                        <TableHead>
                          <TableRow>
                            <TableCell>Variation</TableCell>
                            <TableCell align="right">Users</TableCell>
                            <TableCell align="right">Conversions</TableCell>
                            <TableCell align="right">Conversion Rate</TableCell>
                            <TableCell align="right">Lift</TableCell>
                            <TableCell align="center">Performance</TableCell>
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {experiment.config.variations.map((variation, index) => {
                            const users = Math.floor(experiment.analytics!.total_users * (variation.weight / 100));
                            const conversions = experiment.analytics!.conversions[variation.variation_id] || 0;
                            const conversionRate = parseFloat(calculateConversionRate(conversions, users));
                            const baselineRate = index === 0 ? conversionRate : parseFloat(
                              calculateConversionRate(
                                experiment.analytics!.conversions[experiment.config.variations[0].variation_id] || 0,
                                Math.floor(experiment.analytics!.total_users * (experiment.config.variations[0].weight / 100))
                              )
                            );
                            const lift = index === 0 ? 0 : ((conversionRate - baselineRate) / baselineRate) * 100;

                            return (
                              <TableRow key={variation.variation_id}>
                                <TableCell>
                                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                    {variation.variation_id.replace(/_/g, ' ').toUpperCase()}
                                    {index === 0 && (
                                      <Chip label="Control" size="small" variant="outlined" />
                                    )}
                                  </Box>
                                </TableCell>
                                <TableCell align="right">{users.toLocaleString()}</TableCell>
                                <TableCell align="right">{conversions.toLocaleString()}</TableCell>
                                <TableCell align="right">{conversionRate}%</TableCell>
                                <TableCell align="right">
                                  {index === 0 ? '-' : (
                                    <span style={{ color: lift > 0 ? '#4caf50' : lift < 0 ? '#f44336' : 'inherit' }}>
                                      {lift > 0 ? '+' : ''}{lift.toFixed(1)}%
                                    </span>
                                  )}
                                </TableCell>
                                <TableCell align="center">
                                  <LinearProgress
                                    variant="determinate"
                                    value={Math.min((conversionRate / Math.max(...experiment.config.variations.map((v, i) => {
                                      const u = Math.floor(experiment.analytics!.total_users * (v.weight / 100));
                                      const c = experiment.analytics!.conversions[v.variation_id] || 0;
                                      return parseFloat(calculateConversionRate(c, u));
                                    }))) * 100, 100)}
                                    sx={{ width: 60, height: 8, borderRadius: 4 }}
                                    color={index === 0 ? 'primary' : lift > 0 ? 'success' : 'error'}
                                  />
                                </TableCell>
                              </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  </Box>
                )}

                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    {experiment.config.success_metrics?.map((metric) => (
                      <Chip
                        key={metric}
                        label={metric.replace(/_/g, ' ')}
                        size="small"
                        variant="outlined"
                        icon={<TrendingUpIcon />}
                      />
                    ))}
                  </Box>
                  <Box>
                    {experiment.status === 'draft' && (
                      <Tooltip title="Start Experiment">
                        <IconButton color="primary">
                          <PlayIcon />
                        </IconButton>
                      </Tooltip>
                    )}
                    {experiment.status === 'active' && (
                      <Tooltip title="Pause Experiment">
                        <IconButton color="warning">
                          <PauseIcon />
                        </IconButton>
                      </Tooltip>
                    )}
                    {(experiment.status === 'active' || experiment.status === 'paused') && (
                      <Tooltip title="Stop Experiment">
                        <IconButton>
                          <StopIcon />
                        </IconButton>
                      </Tooltip>
                    )}
                    <Tooltip title="View Analytics">
                      <IconButton>
                        <AnalyticsIcon />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}