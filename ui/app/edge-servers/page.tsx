'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Chip,
  Alert,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Tooltip,
  LinearProgress,
  Avatar,
} from '@mui/material';
import {
  Storage as StorageIcon,
  Refresh as RefreshIcon,
  CheckCircle as ConnectedIcon,
  Warning as WarningIcon,
  Error as DisconnectedIcon,
  LocationOn as LocationIcon,
  Computer as ComputerIcon,
  Schedule as ScheduleIcon,
  NetworkCheck as NetworkIcon,
} from '@mui/icons-material';

interface EdgeServer {
  id: string;
  client_id: string;
  project_id: string;
  environment: string;
  status: string;
  connected_at: string;
  last_ping: string;
  uptime_seconds: number;
  uptime_human: string;
}

interface EdgeServersStatusResponse {
  servers: EdgeServer[];
  total: number;
}

export default function EdgeServersPage() {
  const [serversData, setServersData] = useState<EdgeServersStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchEdgeServers = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/v1/edge/servers', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch edge servers');
      }

      const data = await response.json();
      setServersData(data);
      setLastUpdated(new Date());
    } catch (err) {
      setError('Failed to load edge servers status');
      console.error('Edge servers error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEdgeServers();
    
    // Set up auto-refresh every 30 seconds
    const interval = setInterval(fetchEdgeServers, 30000);
    return () => clearInterval(interval);
  }, []);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'connected':
        return <ConnectedIcon color="success" />;
      case 'unhealthy':
        return <WarningIcon color="warning" />;
      case 'disconnected':
        return <DisconnectedIcon color="error" />;
      default:
        return <ComputerIcon color="disabled" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected':
        return 'success';
      case 'unhealthy':
        return 'warning';
      case 'disconnected':
        return 'error';
      default:
        return 'default';
    }
  };

  const connectedCount = serversData?.servers.filter(s => s.status === 'connected').length || 0;
  const disconnectedCount = serversData?.total ? serversData.total - connectedCount : 0;
  
  // Group servers by environment
  const environments = serversData?.servers.reduce((acc, server) => {
    acc[server.environment] = (acc[server.environment] || 0) + 1;
    return acc;
  }, {} as Record<string, number>) || {};

  return (
    <Box>
      {/* Header */}
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
                mb: 0.5,
                display: 'flex',
                alignItems: 'center',
                gap: 1
              }}
            >
              <StorageIcon sx={{ fontSize: '1.5rem' }} />
              Edge Servers
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ fontSize: '0.875rem' }}
            >
              Monitor distributed edge server health, connections, and performance
              {lastUpdated && (
                <> • Last updated {lastUpdated.toLocaleTimeString()}</>
              )}
            </Typography>
          </Box>
          <Box display="flex" gap={2}>
            <Tooltip title="Refresh data">
              <IconButton
                onClick={fetchEdgeServers}
                disabled={loading}
                sx={{ 
                  bgcolor: 'primary.50',
                  border: '1px solid',
                  borderColor: 'primary.200',
                  '&:hover': {
                    bgcolor: 'primary.100',
                  }
                }}
              >
                <RefreshIcon color="primary" />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Stats Cards */}
      <Grid container spacing={3} mb={5}>
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ height: '100%', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
            <CardContent sx={{ textAlign: 'center', p: 3 }}>
              <Box display="flex" alignItems="center" justifyContent="center" mb={2}>
                <Avatar sx={{ bgcolor: 'primary.main', width: 48, height: 48 }}>
                  <ComputerIcon />
                </Avatar>
              </Box>
              <Typography variant="h4" fontWeight="700" color="primary.main" gutterBottom>
                {serversData?.total || 0}
              </Typography>
              <Typography variant="body2" color="text.secondary" fontWeight={500}>
                Total Servers
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ height: '100%', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
            <CardContent sx={{ textAlign: 'center', p: 3 }}>
              <Box display="flex" alignItems="center" justifyContent="center" mb={2}>
                <Avatar sx={{ bgcolor: 'success.main', width: 48, height: 48 }}>
                  <ConnectedIcon />
                </Avatar>
              </Box>
              <Typography variant="h4" fontWeight="700" color="success.main" gutterBottom>
                {connectedCount}
              </Typography>
              <Typography variant="body2" color="text.secondary" fontWeight={500}>
                Connected
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ height: '100%', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
            <CardContent sx={{ textAlign: 'center', p: 3 }}>
              <Box display="flex" alignItems="center" justifyContent="center" mb={2}>
                <Avatar sx={{ bgcolor: 'error.main', width: 48, height: 48 }}>
                  <DisconnectedIcon />
                </Avatar>
              </Box>
              <Typography variant="h4" fontWeight="700" color="error.main" gutterBottom>
                {disconnectedCount}
              </Typography>
              <Typography variant="body2" color="text.secondary" fontWeight={500}>
                Disconnected
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ height: '100%', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
            <CardContent sx={{ textAlign: 'center', p: 3 }}>
              <Box display="flex" alignItems="center" justifyContent="center" mb={2}>
                <Avatar sx={{ bgcolor: 'info.main', width: 48, height: 48 }}>
                  <LocationIcon />
                </Avatar>
              </Box>
              <Typography variant="h4" fontWeight="700" color="info.main" gutterBottom>
                {Object.keys(environments).length}
              </Typography>
              <Typography variant="body2" color="text.secondary" fontWeight={500}>
                Environments
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Environment Distribution */}
      {serversData && Object.keys(environments).length > 0 && (
        <Card sx={{ mb: 4, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
          <CardContent sx={{ p: 3 }}>
            <Typography variant="h6" fontWeight="600" gutterBottom display="flex" alignItems="center" gap={1}>
              <LocationIcon />
              Environment Distribution
            </Typography>
            <Box display="flex" gap={2} flexWrap="wrap" mt={2}>
              {Object.entries(environments).map(([env, count]) => (
                <Chip
                  key={env}
                  label={`${env.charAt(0).toUpperCase() + env.slice(1)}: ${count}`}
                  variant="outlined"
                  color="primary"
                  size="medium"
                  sx={{ fontWeight: 500 }}
                />
              ))}
            </Box>
          </CardContent>
        </Card>
      )}

      {/* Edge Servers Table */}
      <Card sx={{ border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
        <CardContent sx={{ p: 0 }}>
          <Box sx={{ p: 3, pb: 0 }}>
            <Typography variant="h6" fontWeight="600" display="flex" alignItems="center" gap={1}>
              <NetworkIcon />
              Server Status
            </Typography>
          </Box>
          
          {loading && <LinearProgress />}
          
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Status</TableCell>
                  <TableCell>Server ID</TableCell>
                  <TableCell>Client ID</TableCell>
                  <TableCell>Project ID</TableCell>
                  <TableCell>Environment</TableCell>
                  <TableCell>Uptime</TableCell>
                  <TableCell>Connected At</TableCell>
                  <TableCell>Last Ping</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {(!serversData?.servers || serversData?.servers.length === 0) ? (
                  <TableRow>
                    <TableCell colSpan={8} align="center" sx={{ py: 6 }}>
                      <Box display="flex" flexDirection="column" alignItems="center" gap={2}>
                        <StorageIcon sx={{ fontSize: 48, color: 'grey.300' }} />
                        <Typography variant="body1" color="text.secondary">
                          No edge servers connected
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          Deploy edge servers to see them here
                        </Typography>
                      </Box>
                    </TableCell>
                  </TableRow>
                ) : (
                  serversData?.servers.map((server) => (
                    <TableRow key={server.client_id} hover>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={1}>
                          {getStatusIcon(server.status)}
                          <Chip
                            label={server.status}
                            size="small"
                            color={getStatusColor(server.status) as any}
                            variant={server.status === 'connected' ? 'filled' : 'outlined'}
                          />
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" fontFamily="monospace">
                          {server.id}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" fontFamily="monospace" color="text.secondary" fontSize="0.75rem">
                          {server.client_id}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" fontFamily="monospace" color="text.secondary" fontSize="0.75rem">
                          {server.project_id}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={server.environment}
                          size="small"
                          variant="outlined"
                          color="default"
                        />
                      </TableCell>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={0.5}>
                          <ScheduleIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                          <Typography variant="body2" color="text.secondary">
                            {server.uptime_human}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {new Date(server.connected_at).toLocaleString()}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {new Date(server.last_ping).toLocaleString()}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Box>
  );
}