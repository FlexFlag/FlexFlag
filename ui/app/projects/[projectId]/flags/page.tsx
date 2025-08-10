'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Button,
  Card,
  CardContent,
  Grid,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  FormControl,
  InputLabel,
  Switch,
  FormControlLabel,
  Alert,
  Tooltip,
  Paper,
} from '@mui/material';
import {
  Add as AddIcon,
  MoreVert as MoreVertIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ContentCopy as ContentCopyIcon,
  PlayArrow as PlayArrowIcon,
  Pause as PauseIcon,
  Flag as FlagIcon,
  Warning as AlertIcon,
  AccountTree as ProjectIcon,
  Download as DownloadIcon,
  Upload as UploadIcon,
} from '@mui/icons-material';
import { apiClient } from '@/lib/api';
import { Flag, CreateFlagRequest } from '@/types';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { useParams } from 'next/navigation';

interface FlagCardProps {
  flag: Flag;
  onEdit: (flag: Flag) => void;
  onDelete: (flag: Flag) => void;
  onToggle: (flag: Flag) => void;
  onDuplicate: (flag: Flag) => void;
}

function FlagCard({ flag, onEdit, onDelete, onToggle, onDuplicate }: FlagCardProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'boolean': return 'primary';
      case 'string': return 'secondary';
      case 'number': return 'warning';
      case 'json': return 'info';
      default: return 'default';
    }
  };

  return (
    <Card 
      sx={{ 
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        transition: 'all 0.2s ease',
        border: '1px solid',
        borderColor: 'divider',
        boxShadow: 0,
        '&:hover': {
          boxShadow: 2,
          transform: 'translateY(-1px)',
          borderColor: flag.enabled ? 'success.main' : 'grey.400',
        },
      }}
    >
      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', p: 2 }}>
        <Box display="flex" alignItems="flex-start" justifyContent="space-between" mb={1.5}>
          <Box display="flex" alignItems="center" gap={1}>
            <Box
              sx={{
                width: 12,
                height: 12,
                borderRadius: '50%',
                bgcolor: flag.enabled ? 'success.main' : 'grey.400',
                boxShadow: flag.enabled ? '0 0 0 2px rgba(76, 175, 80, 0.2)' : 'none',
              }}
            />
            <Typography variant="subtitle1" fontWeight="600" noWrap sx={{ fontSize: '0.95rem' }}>
              {flag.name}
            </Typography>
          </Box>
          <IconButton size="small" onClick={handleClick}>
            <MoreVertIcon />
          </IconButton>
        </Box>

        <Typography 
          variant="body2" 
          color="text.secondary" 
          sx={{ fontFamily: 'monospace', bgcolor: 'grey.50', p: 0.75, borderRadius: 1, mb: 1.5, fontSize: '0.75rem' }}
        >
          {flag.key}
        </Typography>

        {flag.description && (
          <Typography variant="body2" color="text.secondary" mb={1.5} sx={{ 
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            fontSize: '0.8rem',
            flexGrow: 1,
          }}>
            {flag.description}
          </Typography>
        )}

        <Box display="flex" gap={0.5} flexWrap="wrap" mb={1.5}>
          <Chip 
            label={flag.type} 
            size="small" 
            color={getTypeColor(flag.type) as any}
            variant="outlined" 
          />
          <Chip 
            label={flag.enabled ? 'Enabled' : 'Disabled'} 
            size="small" 
            color={flag.enabled ? 'success' : 'default'}
          />
          {flag.variations && flag.variations.length > 0 && (
            <Chip 
              label={`${flag.variations.length} variations`} 
              size="small" 
              variant="outlined"
            />
          )}
        </Box>

        <Box display="flex" alignItems="center" justifyContent="space-between" sx={{ mt: 'auto' }}>
          <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
            Updated {flag.updated_at ? new Date(flag.updated_at).toLocaleDateString() : 'Unknown'}
          </Typography>
          <Tooltip title={flag.enabled ? 'Disable flag' : 'Enable flag'}>
            <IconButton 
              size="small" 
              onClick={() => onToggle(flag)}
              color={flag.enabled ? 'error' : 'success'}
              sx={{ p: 0.5 }}
            >
              {flag.enabled ? <PauseIcon sx={{ fontSize: '1rem' }} /> : <PlayArrowIcon sx={{ fontSize: '1rem' }} />}
            </IconButton>
          </Tooltip>
        </Box>
      </CardContent>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <MenuItem onClick={() => { onEdit(flag); handleClose(); }}>
          <EditIcon sx={{ mr: 1 }} fontSize="small" />
          Edit
        </MenuItem>
        <MenuItem onClick={() => { onDuplicate(flag); handleClose(); }}>
          <ContentCopyIcon sx={{ mr: 1 }} fontSize="small" />
          Duplicate
        </MenuItem>
        <MenuItem onClick={() => { onDelete(flag); handleClose(); }} sx={{ color: 'error.main' }}>
          <DeleteIcon sx={{ mr: 1 }} fontSize="small" />
          Delete
        </MenuItem>
      </Menu>
    </Card>
  );
}

function EditFlagDialog({ 
  open, 
  onClose, 
  onSave, 
  flag 
}: { 
  open: boolean; 
  onClose: () => void; 
  onSave: (flag: Partial<Flag>) => void;
  flag: Flag | null;
}) {
  const [formData, setFormData] = useState<Partial<Flag>>({});

  useEffect(() => {
    if (flag) {
      setFormData({
        key: flag.key,
        name: flag.name,
        description: flag.description,
        type: flag.type,
        enabled: flag.enabled,
        default: flag.default,
      });
    }
  }, [flag]);

  const handleSave = () => {
    onSave(formData);
    setFormData({});
  };

  if (!flag) return null;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Feature Flag</DialogTitle>
      <DialogContent>
        <Box display="flex" flexDirection="column" gap={3} pt={1}>
          <TextField
            label="Flag Key"
            value={flag.key}
            disabled
            helperText="Flag key cannot be changed after creation"
            fullWidth
          />
          <TextField
            label="Display Name"
            value={formData.name || ''}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="New Awesome Feature"
            fullWidth
            required
          />
          <TextField
            label="Description"
            value={formData.description || ''}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Brief description of what this flag controls"
            multiline
            rows={3}
            fullWidth
          />
          <TextField
            label="Default Value"
            value={typeof formData.default === 'object' ? JSON.stringify(formData.default) : formData.default || ''}
            onChange={(e) => {
              let value: any = e.target.value;
              if (flag.type === 'number') {
                value = parseFloat(value) || 0;
              } else if (flag.type === 'boolean') {
                value = value === 'true';
              } else if (flag.type === 'json') {
                try {
                  value = JSON.parse(value);
                } catch {
                  value = {};
                }
              }
              setFormData({ ...formData, default: value });
            }}
            fullWidth
            helperText={flag.type === 'json' ? 'Enter valid JSON' : `Type: ${flag.type}`}
          />
          <FormControlLabel
            control={
              <Switch
                checked={formData.enabled || false}
                onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              />
            }
            label="Enable flag"
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button 
          variant="contained" 
          onClick={handleSave}
          disabled={!formData.name}
        >
          Update Flag
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function CreateFlagDialog({ open, onClose, onSave }: { 
  open: boolean; 
  onClose: () => void; 
  onSave: (flag: CreateFlagRequest) => void; 
}) {
  const [formData, setFormData] = useState<CreateFlagRequest>({
    key: '',
    name: '',
    description: '',
    type: 'boolean',
    enabled: false,
    default: false,
  });

  const handleSave = () => {
    onSave(formData);
    setFormData({
      key: '',
      name: '',
      description: '',
      type: 'boolean',
      enabled: false,
      default: false,
    });
  };

  const getDefaultValue = (type: string) => {
    switch (type) {
      case 'boolean': return false;
      case 'string': return '';
      case 'number': return 0;
      case 'json': return {};
      default: return false;
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Create New Feature Flag</DialogTitle>
      <DialogContent>
        <Box display="flex" flexDirection="column" gap={3} pt={1}>
          <TextField
            label="Flag Key"
            value={formData.key}
            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
            placeholder="new-awesome-feature"
            helperText="Unique identifier for your flag (kebab-case recommended)"
            fullWidth
            required
          />
          <TextField
            label="Display Name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="New Awesome Feature"
            fullWidth
            required
          />
          <TextField
            label="Description"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Brief description of what this flag controls"
            multiline
            rows={3}
            fullWidth
          />
          <FormControl fullWidth>
            <InputLabel>Flag Type</InputLabel>
            <Select
              value={formData.type}
              label="Flag Type"
              onChange={(e) => {
                const newType = e.target.value as Flag['type'];
                setFormData({ 
                  ...formData, 
                  type: newType,
                  default: getDefaultValue(newType)
                });
              }}
            >
              <MenuItem value="boolean">Boolean</MenuItem>
              <MenuItem value="string">String</MenuItem>
              <MenuItem value="number">Number</MenuItem>
              <MenuItem value="json">JSON</MenuItem>
            </Select>
          </FormControl>
          <TextField
            label="Default Value"
            value={typeof formData.default === 'object' ? JSON.stringify(formData.default) : formData.default}
            onChange={(e) => {
              let value: any = e.target.value;
              if (formData.type === 'number') {
                value = parseFloat(value) || 0;
              } else if (formData.type === 'boolean') {
                value = value === 'true';
              } else if (formData.type === 'json') {
                try {
                  value = JSON.parse(value);
                } catch {
                  value = {};
                }
              }
              setFormData({ ...formData, default: value });
            }}
            fullWidth
            helperText={formData.type === 'json' ? 'Enter valid JSON' : undefined}
          />
          <FormControlLabel
            control={
              <Switch
                checked={formData.enabled}
                onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              />
            }
            label="Enable flag immediately"
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button 
          variant="contained" 
          onClick={handleSave}
          disabled={!formData.key || !formData.name}
        >
          Create Flag
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function ProjectFlagsPage() {
  const { currentEnvironment } = useEnvironment();
  const params = useParams();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [flags, setFlags] = useState<Flag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingFlag, setEditingFlag] = useState<Flag | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    title: string;
    message: string;
    action: () => void;
  }>({
    open: false,
    title: '',
    message: '',
    action: () => {},
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

  const fetchFlags = async () => {
    try {
      setLoading(true);
      setError(null);
      const flagsData = await apiClient.getFlags(currentEnvironment, projectId);
      setFlags(flagsData);
    } catch (err) {
      setError('Failed to load flags');
      console.error('Flags error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (projectId) {
      fetchFlags();
    }
  }, [currentEnvironment, projectId]);

  const handleCreateFlag = async (flagData: CreateFlagRequest) => {
    if (!projectId) {
      setError('Project ID is required');
      return;
    }
    
    try {
      // Add current project ID to flag data
      const flagDataWithProject = {
        ...flagData,
        project_id: projectId,
      };
      await apiClient.createFlag(flagDataWithProject, currentEnvironment);
      setCreateDialogOpen(false);
      fetchFlags();
    } catch (err) {
      setError('Failed to create flag');
      console.error('Create flag error:', err);
    }
  };

  const handleToggleFlag = async (flag: Flag) => {
    if (!projectId) {
      setError('Project ID is required');
      return;
    }

    const action = flag.enabled ? 'disable' : 'enable';
    setConfirmDialog({
      open: true,
      title: `${action.charAt(0).toUpperCase() + action.slice(1)} Flag`,
      message: `Are you sure you want to ${action} "${flag.name}"? This will affect all users in the ${currentEnvironment} environment.`,
      action: async () => {
        try {
          await apiClient.toggleFlag(flag.key, currentEnvironment, projectId);
          fetchFlags();
          setConfirmDialog({ ...confirmDialog, open: false });
        } catch (err) {
          setError(`Failed to ${action} flag`);
          console.error(`${action} flag error:`, err);
          setConfirmDialog({ ...confirmDialog, open: false });
        }
      },
    });
  };

  const handleDeleteFlag = async (flag: Flag) => {
    setConfirmDialog({
      open: true,
      title: 'Delete Flag',
      message: `Are you sure you want to permanently delete "${flag.name}"? This action cannot be undone and will affect the ${currentEnvironment} environment.`,
      action: async () => {
        try {
          await apiClient.deleteFlag(flag.key, currentEnvironment);
          fetchFlags();
          setConfirmDialog({ ...confirmDialog, open: false });
        } catch (err) {
          setError('Failed to delete flag');
          console.error('Delete flag error:', err);
          setConfirmDialog({ ...confirmDialog, open: false });
        }
      },
    });
  };

  const handleEditFlag = (flag: Flag) => {
    setEditingFlag(flag);
    setEditDialogOpen(true);
  };

  const handleUpdateFlag = async (flagData: Partial<Flag>) => {
    if (!editingFlag) return;
    
    try {
      // Ensure required fields are included
      const updateData = {
        key: editingFlag.key,
        type: editingFlag.type,
        ...flagData,
      };
      await apiClient.updateFlag(editingFlag.key, updateData, currentEnvironment);
      setEditDialogOpen(false);
      setEditingFlag(null);
      fetchFlags();
    } catch (err) {
      setError('Failed to update flag');
      console.error('Update flag error:', err);
    }
  };

  const handleDuplicateFlag = (flag: Flag) => {
    const duplicatedFlag: CreateFlagRequest = {
      key: `${flag.key}-copy`,
      name: `${flag.name} (Copy)`,
      description: flag.description,
      type: flag.type,
      enabled: false, // Start disabled for safety
      default: flag.default,
      variations: flag.variations?.map(v => ({
        name: v.name,
        value: v.value,
        description: v.description,
        weight: v.weight,
      })),
      targeting: flag.targeting,
    };
    
    setCreateDialogOpen(true);
  };

  const enabledFlags = flags.filter(flag => flag.enabled);
  const disabledFlags = flags.filter(flag => !flag.enabled);

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
                mb: 0.5
              }}
            >
              Feature Flags
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ fontSize: '0.875rem' }}
            >
              Manage flags for {project.name} in {currentEnvironment} environment
            </Typography>
          </Box>
          <Box display="flex" gap={2}>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setCreateDialogOpen(true)}
              sx={{ 
                borderRadius: 1.5,
                px: 3,
                py: 1.25,
                textTransform: 'none',
                fontWeight: 600,
                fontSize: '0.95rem',
                boxShadow: 2,
                '&:hover': {
                  boxShadow: 4
                }
              }}
            >
              Create Flag
            </Button>
          </Box>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Stats */}
      <Grid container spacing={3} mb={5}>
        <Grid item xs={12} sm={4}>
          <Paper 
            sx={{ 
              p: 3, 
              textAlign: 'center',
              border: '1px solid',
              borderColor: 'divider',
              boxShadow: 0,
              bgcolor: 'background.paper',
              transition: 'all 0.2s ease',
              '&:hover': {
                boxShadow: 2,
                borderColor: 'primary.main'
              }
            }}
          >
            <Typography 
              variant="h3" 
              fontWeight="700" 
              color="primary.main"
              sx={{ mb: 1 }}
            >
              {flags.length}
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ 
                fontWeight: 500,
                textTransform: 'uppercase',
                fontSize: '0.75rem',
                letterSpacing: '0.5px'
              }}
            >
              Total Flags
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper 
            sx={{ 
              p: 3, 
              textAlign: 'center',
              border: '1px solid',
              borderColor: 'divider',
              boxShadow: 0,
              bgcolor: 'background.paper',
              transition: 'all 0.2s ease',
              '&:hover': {
                boxShadow: 2,
                borderColor: 'success.main'
              }
            }}
          >
            <Typography 
              variant="h3" 
              fontWeight="700" 
              color="success.main"
              sx={{ mb: 1 }}
            >
              {enabledFlags.length}
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ 
                fontWeight: 500,
                textTransform: 'uppercase',
                fontSize: '0.75rem',
                letterSpacing: '0.5px'
              }}
            >
              Enabled
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper 
            sx={{ 
              p: 3, 
              textAlign: 'center',
              border: '1px solid',
              borderColor: 'divider',
              boxShadow: 0,
              bgcolor: 'background.paper',
              transition: 'all 0.2s ease',
              '&:hover': {
                boxShadow: 2,
                borderColor: 'grey.400'
              }
            }}
          >
            <Typography 
              variant="h3" 
              fontWeight="700" 
              color="grey.600"
              sx={{ mb: 1 }}
            >
              {disabledFlags.length}
            </Typography>
            <Typography 
              variant="body2" 
              color="text.secondary"
              sx={{ 
                fontWeight: 500,
                textTransform: 'uppercase',
                fontSize: '0.75rem',
                letterSpacing: '0.5px'
              }}
            >
              Disabled
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Flags Grid */}
      {loading ? (
        <Box display="flex" justifyContent="center" py={8}>
          <Typography>Loading flags...</Typography>
        </Box>
      ) : flags.length === 0 ? (
        <Box textAlign="center" py={8}>
          <FlagIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No flags found
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Create your first feature flag to get started
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
          >
            Create First Flag
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {flags.map((flag) => (
            <Grid item xs={12} sm={6} md={4} lg={3} key={flag.key}>
              <FlagCard
                flag={flag}
                onEdit={handleEditFlag}
                onDelete={handleDeleteFlag}
                onToggle={handleToggleFlag}
                onDuplicate={handleDuplicateFlag}
              />
            </Grid>
          ))}
        </Grid>
      )}

      <CreateFlagDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onSave={handleCreateFlag}
      />

      <EditFlagDialog
        open={editDialogOpen}
        onClose={() => {
          setEditDialogOpen(false);
          setEditingFlag(null);
        }}
        onSave={handleUpdateFlag}
        flag={editingFlag}
      />

      {/* Confirmation Dialog */}
      <Dialog
        open={confirmDialog.open}
        onClose={() => setConfirmDialog({ ...confirmDialog, open: false })}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <AlertIcon color="warning" />
          {confirmDialog.title}
        </DialogTitle>
        <DialogContent>
          <Typography variant="body1">
            {confirmDialog.message}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={() => setConfirmDialog({ ...confirmDialog, open: false })}
            color="inherit"
          >
            Cancel
          </Button>
          <Button 
            onClick={confirmDialog.action}
            variant="contained"
            color="warning"
            autoFocus
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}