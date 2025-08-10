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
import { DataGrid, GridColDef, GridRowParams } from '@mui/x-data-grid';
import { apiClient } from '@/lib/api';
import { Flag, CreateFlagRequest } from '@/types';
import { useProject } from '@/contexts/ProjectContext';
import { useEnvironment } from '@/contexts/EnvironmentContext';

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
        position: 'relative',
        transition: 'all 0.2s ease',
        '&:hover': {
          boxShadow: 4,
          transform: 'translateY(-2px)',
        },
      }}
    >
      <CardContent>
        <Box display="flex" alignItems="flex-start" justifyContent="space-between" mb={2}>
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
            <Typography variant="h6" fontWeight="600" noWrap>
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
          sx={{ fontFamily: 'monospace', bgcolor: 'grey.50', p: 1, borderRadius: 1, mb: 2 }}
        >
          {flag.key}
        </Typography>

        {flag.description && (
          <Typography variant="body2" color="text.secondary" mb={2} sx={{ 
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}>
            {flag.description}
          </Typography>
        )}

        <Box display="flex" gap={1} flexWrap="wrap" mb={2}>
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

        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Typography variant="caption" color="text.secondary">
            Updated {flag.updated_at ? new Date(flag.updated_at).toLocaleDateString() : 'Unknown'}
          </Typography>
          <Tooltip title={flag.enabled ? 'Disable flag' : 'Enable flag'}>
            <IconButton 
              size="small" 
              onClick={() => onToggle(flag)}
              color={flag.enabled ? 'error' : 'success'}
            >
              {flag.enabled ? <PauseIcon /> : <PlayArrowIcon />}
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

function ImportFlagDialog({ 
  open, 
  onClose, 
  onImport 
}: { 
  open: boolean; 
  onClose: () => void; 
  onImport: (data: any) => void; 
}) {
  const [importData, setImportData] = useState<string>('');
  const [parseError, setParseError] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);

  const handleImport = () => {
    try {
      const data = JSON.parse(importData);
      onImport(data);
      setImportData('');
      setParseError(null);
    } catch (err) {
      setParseError('Invalid JSON format');
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setImportData(content);
        setParseError(null);
        try {
          JSON.parse(content);
        } catch {
          setParseError('Invalid JSON format');
        }
      };
      reader.readAsText(file);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      const file = files[0];
      const reader = new FileReader();
      reader.onload = (event) => {
        const content = event.target?.result as string;
        setImportData(content);
        setParseError(null);
        try {
          JSON.parse(content);
        } catch {
          setParseError('Invalid JSON format');
        }
      };
      reader.readAsText(file);
    }
  };

  const isValidJson = () => {
    try {
      if (!importData.trim()) return false;
      const data = JSON.parse(importData);
      return data.flags && Array.isArray(data.flags);
    } catch {
      return false;
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Import Feature Flags</DialogTitle>
      <DialogContent>
        <Box display="flex" flexDirection="column" gap={3} pt={1}>
          <Typography variant="body2" color="text.secondary">
            Import flags from a JSON file exported from FlexFlag or in the compatible format.
          </Typography>

          {/* File Upload Area */}
          <Box
            sx={{
              border: `2px dashed ${dragOver ? 'primary.main' : 'grey.300'}`,
              borderRadius: 2,
              p: 3,
              textAlign: 'center',
              bgcolor: dragOver ? 'primary.50' : 'grey.50',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => document.getElementById('file-upload')?.click()}
          >
            <UploadIcon sx={{ fontSize: 48, color: 'grey.400', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Drop JSON file here or click to upload
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Supports .json files
            </Typography>
            <input
              id="file-upload"
              type="file"
              accept=".json"
              style={{ display: 'none' }}
              onChange={handleFileUpload}
            />
          </Box>

          {/* Text Area */}
          <TextField
            label="Or paste JSON content"
            multiline
            rows={10}
            value={importData}
            onChange={(e) => {
              setImportData(e.target.value);
              setParseError(null);
            }}
            placeholder={`{
  "project": "My Project",
  "environment": "production",
  "exportDate": "2023-12-01T00:00:00.000Z",
  "flags": [
    {
      "key": "example-flag",
      "name": "Example Flag",
      "description": "An example feature flag",
      "type": "boolean",
      "enabled": true,
      "default": false
    }
  ]
}`}
            fullWidth
            error={!!parseError}
            helperText={parseError || 'Paste your JSON export data here'}
          />

          {parseError && (
            <Alert severity="error">
              {parseError}
            </Alert>
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button 
          variant="contained" 
          onClick={handleImport}
          disabled={!isValidJson()}
        >
          Import Flags
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function FlagsPage() {
  const { currentProject } = useProject();
  const { currentEnvironment } = useEnvironment();
  const [flags, setFlags] = useState<Flag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingFlag, setEditingFlag] = useState<Flag | null>(null);
  const [importDialogOpen, setImportDialogOpen] = useState(false);
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
  const [viewMode, setViewMode] = useState<'cards' | 'table'>('cards');

  const fetchFlags = async () => {
    try {
      setLoading(true);
      setError(null);
      const flagsData = await apiClient.getFlags(currentEnvironment, currentProject?.id);
      setFlags(flagsData);
    } catch (err) {
      setError('Failed to load flags');
      console.error('Flags error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFlags();
  }, [currentEnvironment, currentProject?.id]);

  const handleCreateFlag = async (flagData: CreateFlagRequest) => {
    if (!currentProject) {
      setError('Please select a project before creating a flag');
      return;
    }
    
    try {
      // Add current project ID to flag data
      const flagDataWithProject = {
        ...flagData,
        project_id: currentProject.id,
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
    if (!currentProject) {
      setError('Please select a project before toggling flags');
      return;
    }

    const action = flag.enabled ? 'disable' : 'enable';
    setConfirmDialog({
      open: true,
      title: `${action.charAt(0).toUpperCase() + action.slice(1)} Flag`,
      message: `Are you sure you want to ${action} "${flag.name}"? This will affect all users in the ${currentEnvironment} environment.`,
      action: async () => {
        try {
          await apiClient.toggleFlag(flag.key, currentEnvironment, currentProject.id);
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
    
    // Auto-fill the create dialog with duplicated data
    setCreateDialogOpen(true);
    // We'll need to modify the CreateFlagDialog to accept initial data
  };

  const handleExportFlags = () => {
    if (flags.length === 0) {
      setError('No flags to export');
      return;
    }

    // Prepare export data
    const exportData = {
      project: currentProject?.name || 'Unknown Project',
      environment: currentEnvironment,
      exportDate: new Date().toISOString(),
      flags: flags.map(flag => ({
        key: flag.key,
        name: flag.name,
        description: flag.description,
        type: flag.type,
        enabled: flag.enabled,
        default: flag.default,
        variations: flag.variations,
        targeting: flag.targeting,
        tags: flag.tags,
        metadata: flag.metadata,
      })),
    };

    // Create and download file
    const dataStr = JSON.stringify(exportData, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `flexflag-${currentProject?.name || 'project'}-${currentEnvironment}-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    URL.revokeObjectURL(url);
  };

  const handleImportFlags = async (importData: any) => {
    if (!currentProject) {
      setError('Please select a project before importing flags');
      return;
    }

    try {
      const { flags: importedFlags } = importData;
      
      if (!importedFlags || !Array.isArray(importedFlags)) {
        throw new Error('Invalid import format: flags array is required');
      }

      let successCount = 0;
      let errorCount = 0;
      const errors: string[] = [];

      for (const flagData of importedFlags) {
        try {
          const createFlagData: CreateFlagRequest = {
            key: flagData.key,
            name: flagData.name,
            description: flagData.description || '',
            type: flagData.type || 'boolean',
            enabled: flagData.enabled || false,
            default: flagData.default,
            variations: flagData.variations,
            targeting: flagData.targeting,
            tags: flagData.tags,
            metadata: flagData.metadata,
          };

          // Add current project ID to flag data
          const flagDataWithProject = {
            ...createFlagData,
            project_id: currentProject.id,
          };

          await apiClient.createFlag(flagDataWithProject, currentEnvironment);
          successCount++;
        } catch (err: any) {
          errorCount++;
          errors.push(`${flagData.key}: ${err.response?.data?.error || err.message}`);
        }
      }

      if (successCount > 0) {
        setError(null);
        fetchFlags();
      }

      if (errorCount > 0) {
        const errorMessage = `Import completed with ${successCount} successes and ${errorCount} errors:\n${errors.join('\n')}`;
        setError(errorMessage);
      } else {
        setError(null);
      }

      setImportDialogOpen(false);
    } catch (err: any) {
      setError(`Import failed: ${err.message}`);
      console.error('Import error:', err);
    }
  };

  const enabledFlags = flags.filter(flag => flag.enabled);
  const disabledFlags = flags.filter(flag => !flag.enabled);

  return (
    <Box>
      {/* Header */}
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={4}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Feature Flags
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {currentProject 
              ? `Manage flags for ${currentProject.name}`
              : 'Manage and configure your feature flags'
            }
          </Typography>
        </Box>
        <Box display="flex" gap={2}>
          <Button
            variant="outlined"
            startIcon={<UploadIcon />}
            onClick={() => setImportDialogOpen(true)}
            disabled={!currentProject}
            sx={{ borderRadius: 2 }}
          >
            Import Flags
          </Button>
          <Button
            variant="outlined"
            startIcon={<DownloadIcon />}
            onClick={handleExportFlags}
            disabled={!currentProject || flags.length === 0}
            sx={{ borderRadius: 2 }}
          >
            Export Flags
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
            disabled={!currentProject}
            sx={{ borderRadius: 2 }}
          >
            Create Flag
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Stats */}
      <Grid container spacing={2} mb={4}>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" fontWeight="bold" color="primary.main">
              {flags.length}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Total Flags
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" fontWeight="bold" color="success.main">
              {enabledFlags.length}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Enabled
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" fontWeight="bold" color="grey.600">
              {disabledFlags.length}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Disabled
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* Flags Grid */}
      {!currentProject ? (
        <Box textAlign="center" py={8}>
          <ProjectIcon sx={{ fontSize: 64, color: 'grey.300', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No project selected
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Please select a project to view its feature flags
          </Typography>
        </Box>
      ) : loading ? (
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
            disabled={!currentProject}
          >
            Create First Flag
          </Button>
        </Box>
      ) : (
        <Grid container spacing={3}>
          {flags.map((flag) => (
            <Grid item xs={12} sm={6} lg={4} key={flag.key}>
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

      <ImportFlagDialog
        open={importDialogOpen}
        onClose={() => setImportDialogOpen(false)}
        onImport={handleImportFlags}
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