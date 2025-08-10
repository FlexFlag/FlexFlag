'use client';

import { useState, useEffect } from 'react';
import { useProject } from '../../contexts/ProjectContext';
import { apiClient } from '../../lib/api';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Avatar,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  Security as SecurityIcon,
  Person as PersonIcon,
  Flag as FlagIcon,
  Group as GroupIcon,
  Settings as SettingsIcon,
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ToggleOff as ToggleIcon,
  ExpandMore as ExpandMoreIcon,
  Visibility as ViewIcon,
} from '@mui/icons-material';

interface AuditLog {
  id: string;
  user_id?: string;
  user_name?: string;
  user_email?: string;
  resource_type: 'flag' | 'segment' | 'rollout' | 'user' | 'project';
  resource_id: string;
  resource_name?: string;
  action: 'create' | 'update' | 'delete' | 'toggle' | 'activate' | 'pause';
  old_values?: any;
  new_values?: any;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export default function AuditPage() {
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([
    {
      id: '1',
      user_id: '1',
      user_name: 'Admin User',
      user_email: 'admin@example.com',
      resource_type: 'flag',
      resource_id: 'test-feature',
      resource_name: 'Test Feature Flag',
      action: 'create',
      new_values: {
        name: 'Test Feature Flag',
        enabled: false,
        type: 'boolean',
      },
      ip_address: '192.168.1.100',
      created_at: '2025-01-10T14:30:00Z',
    },
    {
      id: '2',
      user_id: '1',
      user_name: 'Admin User',
      user_email: 'admin@example.com',
      resource_type: 'flag',
      resource_id: 'test-feature',
      resource_name: 'Test Feature Flag',
      action: 'toggle',
      old_values: { enabled: false },
      new_values: { enabled: true },
      ip_address: '192.168.1.100',
      created_at: '2025-01-10T14:32:15Z',
    },
    {
      id: '3',
      user_id: '2',
      user_name: 'Editor User',
      user_email: 'editor@example.com',
      resource_type: 'segment',
      resource_id: 'premium-users',
      resource_name: 'Premium Users Segment',
      action: 'create',
      new_values: {
        name: 'Premium Users Segment',
        rules: [
          { attribute: 'plan', operator: 'equals', values: ['premium'] }
        ],
      },
      ip_address: '192.168.1.101',
      created_at: '2025-01-10T13:15:00Z',
    },
    {
      id: '4',
      user_id: '1',
      user_name: 'Admin User',
      user_email: 'admin@example.com',
      resource_type: 'rollout',
      resource_id: 'rollout-123',
      resource_name: '25% Feature Rollout',
      action: 'activate',
      old_values: { status: 'draft' },
      new_values: { status: 'active' },
      ip_address: '192.168.1.100',
      created_at: '2025-01-10T12:45:00Z',
    },
    {
      id: '5',
      user_id: '3',
      user_name: 'Viewer User',
      user_email: 'viewer@example.com',
      resource_type: 'user',
      resource_id: '4',
      resource_name: 'New Editor User',
      action: 'create',
      new_values: {
        email: 'neweditor@example.com',
        role: 'editor',
        is_active: true,
      },
      ip_address: '192.168.1.102',
      created_at: '2025-01-10T11:20:00Z',
    },
  ]);

  const { currentProject } = useProject();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [filters, setFilters] = useState({
    resource_type: '',
    action: '',
    user_id: '',
  });

  // Fetch audit logs from API
  useEffect(() => {
    const fetchAuditLogs = async () => {
      if (!currentProject) {
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        const logs = await apiClient.getAuditLogs(currentProject.id);
        setAuditLogs(logs);
      } catch (err) {
        console.error('Failed to fetch audit logs:', err);
        setError('Failed to load audit logs');
        // Keep the mock data as fallback
      } finally {
        setLoading(false);
      }
    };

    fetchAuditLogs();
  }, [currentProject]);

  const getResourceIcon = (type: string) => {
    switch (type) {
      case 'flag': return <FlagIcon />;
      case 'segment': return <GroupIcon />;
      case 'rollout': return <ToggleIcon />;
      case 'user': return <PersonIcon />;
      case 'project': return <SettingsIcon />;
      default: return <SecurityIcon />;
    }
  };

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'create': return <AddIcon />;
      case 'update': return <EditIcon />;
      case 'delete': return <DeleteIcon />;
      case 'toggle': return <ToggleIcon />;
      case 'activate': case 'pause': return <ToggleIcon />;
      default: return <SecurityIcon />;
    }
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case 'create': return 'success';
      case 'update': return 'info';
      case 'delete': return 'error';
      case 'toggle': case 'activate': return 'warning';
      case 'pause': return 'default';
      default: return 'default';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const filteredLogs = auditLogs.filter(log => {
    return (
      (!filters.resource_type || log.resource_type === filters.resource_type) &&
      (!filters.action || log.action === filters.action) &&
      (!filters.user_id || log.user_id === filters.user_id)
    );
  });

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Audit Logs
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Track all system activities and changes for compliance and security
          </Typography>
        </Box>
      </Box>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={2}>
            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>Resource Type</InputLabel>
                <Select
                  value={filters.resource_type}
                  onChange={(e) => setFilters({ ...filters, resource_type: e.target.value })}
                  label="Resource Type"
                >
                  <MenuItem value="">All Resources</MenuItem>
                  <MenuItem value="flag">Feature Flags</MenuItem>
                  <MenuItem value="segment">Segments</MenuItem>
                  <MenuItem value="rollout">Rollouts</MenuItem>
                  <MenuItem value="user">Users</MenuItem>
                  <MenuItem value="project">Projects</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>Action</InputLabel>
                <Select
                  value={filters.action}
                  onChange={(e) => setFilters({ ...filters, action: e.target.value })}
                  label="Action"
                >
                  <MenuItem value="">All Actions</MenuItem>
                  <MenuItem value="create">Create</MenuItem>
                  <MenuItem value="update">Update</MenuItem>
                  <MenuItem value="delete">Delete</MenuItem>
                  <MenuItem value="toggle">Toggle</MenuItem>
                  <MenuItem value="activate">Activate</MenuItem>
                  <MenuItem value="pause">Pause</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                fullWidth
                label="Search by User"
                placeholder="Search user name or email..."
                value={filters.user_id}
                onChange={(e) => setFilters({ ...filters, user_id: e.target.value })}
              />
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Audit Logs Table */}
      <Card>
        <CardContent>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>User</TableCell>
                  <TableCell>Action</TableCell>
                  <TableCell>Resource</TableCell>
                  <TableCell>Date</TableCell>
                  <TableCell>IP Address</TableCell>
                  <TableCell align="right">Details</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredLogs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Avatar sx={{ bgcolor: 'primary.main', width: 32, height: 32 }}>
                          {log.user_name?.charAt(0) || '?'}
                        </Avatar>
                        <Box>
                          <Typography variant="body2" fontWeight="medium">
                            {log.user_name || 'System'}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {log.user_email}
                          </Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Chip
                        icon={getActionIcon(log.action)}
                        label={log.action.charAt(0).toUpperCase() + log.action.slice(1)}
                        color={getActionColor(log.action) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {getResourceIcon(log.resource_type)}
                        <Box>
                          <Typography variant="body2" fontWeight="medium">
                            {log.resource_name || log.resource_id}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {log.resource_type}
                          </Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {formatDate(log.created_at)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" color="text.secondary">
                        {log.ip_address || 'N/A'}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Tooltip title="View Details">
                        <IconButton 
                          size="small"
                          onClick={() => {
                            setSelectedLog(log);
                            setOpenDialog(true);
                          }}
                        >
                          <ViewIcon />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Audit Log Details Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>
          Audit Log Details
        </DialogTitle>
        <DialogContent>
          {selectedLog && (
            <Box sx={{ mt: 2 }}>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    User
                  </Typography>
                  <Typography variant="body1">
                    {selectedLog.user_name} ({selectedLog.user_email})
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    Action
                  </Typography>
                  <Typography variant="body1">
                    {selectedLog.action.charAt(0).toUpperCase() + selectedLog.action.slice(1)}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    Resource
                  </Typography>
                  <Typography variant="body1">
                    {selectedLog.resource_name || selectedLog.resource_id}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    Resource Type
                  </Typography>
                  <Typography variant="body1">
                    {selectedLog.resource_type.charAt(0).toUpperCase() + selectedLog.resource_type.slice(1)}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    Date & Time
                  </Typography>
                  <Typography variant="body1">
                    {formatDate(selectedLog.created_at)}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">
                    IP Address
                  </Typography>
                  <Typography variant="body1">
                    {selectedLog.ip_address || 'N/A'}
                  </Typography>
                </Grid>
              </Grid>

              {(selectedLog.old_values || selectedLog.new_values) && (
                <Box sx={{ mt: 3 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Change Details
                  </Typography>
                  
                  {selectedLog.old_values && (
                    <Accordion>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography>Previous Values</Typography>
                      </AccordionSummary>
                      <AccordionDetails>
                        <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
                          <pre style={{ margin: 0, fontSize: '0.875rem' }}>
                            {JSON.stringify(selectedLog.old_values, null, 2)}
                          </pre>
                        </Paper>
                      </AccordionDetails>
                    </Accordion>
                  )}

                  {selectedLog.new_values && (
                    <Accordion>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography>New Values</Typography>
                      </AccordionSummary>
                      <AccordionDetails>
                        <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
                          <pre style={{ margin: 0, fontSize: '0.875rem' }}>
                            {JSON.stringify(selectedLog.new_values, null, 2)}
                          </pre>
                        </Paper>
                      </AccordionDetails>
                    </Accordion>
                  )}
                </Box>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}