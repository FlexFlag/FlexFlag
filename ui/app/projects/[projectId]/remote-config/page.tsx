'use client';

import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  TextField,
  CircularProgress,
  Alert,
  Button,
  Tooltip,
  Snackbar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormControlLabel,
  Switch,
  Divider,
} from '@mui/material';
import {
  Edit as EditIcon,
  Check as CheckIcon,
  Close as CloseIcon,
  Refresh as RefreshIcon,
  ContentCopy as CopyIcon,
  Tune as TuneIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  PowerSettingsNew as ToggleIcon,
} from '@mui/icons-material';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { Flag } from '@/types';

type ConfigType = 'string' | 'number' | 'json';
type ConfigFlag = Flag;

const TYPE_COLORS: Record<string, 'info' | 'warning' | 'success' | 'default'> = {
  string: 'info',
  number: 'warning',
  json: 'success',
};

const TYPE_PLACEHOLDERS: Record<ConfigType, string> = {
  string: 'e.g. light',
  number: 'e.g. 5000',
  json: 'e.g. {"key": "value"}',
};

const DEFAULT_NEW_CONFIG = {
  key: '',
  name: '',
  description: '',
  type: 'string' as ConfigType,
  value: '',
  enabled: true,
};

function displayValue(flag: Flag): string {
  const val = flag.default;
  if (val === null || val === undefined) return '—';
  if (typeof val === 'object') return JSON.stringify(val);
  return String(val);
}

function parseValue(raw: string, type: string): unknown {
  if (type === 'number') {
    const n = Number(raw);
    return isNaN(n) ? 0 : n;
  }
  if (type === 'json') {
    try { return JSON.parse(raw); } catch { return raw; }
  }
  return raw;
}

function defaultForType(type: ConfigType): string {
  if (type === 'number') return '0';
  if (type === 'json') return '{}';
  return '';
}

export default function RemoteConfigPage() {
  const params = useParams();
  const projectId = params.projectId as string;
  const { currentEnvironment } = useEnvironment();

  const [flags, setFlags] = useState<ConfigFlag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Inline edit state
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [saving, setSaving] = useState(false);

  // Create dialog state
  const [createOpen, setCreateOpen] = useState(false);
  const [newConfig, setNewConfig] = useState({ ...DEFAULT_NEW_CONFIG });
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  // Delete confirm state
  const [deleteTarget, setDeleteTarget] = useState<ConfigFlag | null>(null);
  const [deleting, setDeleting] = useState(false);

  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false, message: '', severity: 'success',
  });

  const showSnack = (message: string, severity: 'success' | 'error' = 'success') =>
    setSnackbar({ open: true, message, severity });

  const loadFlags = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const all = await apiClient.getFlags(currentEnvironment, projectId);
      setFlags(all.filter(f => f.type !== 'boolean' && f.type !== 'variant'));
    } catch {
      setError('Failed to load remote config. Make sure the server is running.');
    } finally {
      setLoading(false);
    }
  }, [currentEnvironment, projectId]);

  useEffect(() => { loadFlags(); }, [loadFlags]);

  // ── Inline edit ──────────────────────────────────────────────────────────

  const startEdit = (flag: ConfigFlag) => {
    setEditingKey(flag.key);
    const val = flag.default;
    setEditValue(typeof val === 'object' ? JSON.stringify(val, null, 2) : String(val ?? ''));
  };

  const cancelEdit = () => { setEditingKey(null); setEditValue(''); };

  const saveEdit = async (flag: ConfigFlag) => {
    setSaving(true);
    try {
      const parsed = parseValue(editValue, flag.type);
      await apiClient.updateFlag(flag.key, { ...flag, default: parsed }, currentEnvironment);
      showSnack(`"${flag.key}" updated`);
      cancelEdit();
      await loadFlags();
    } catch {
      showSnack('Failed to save — check the value format', 'error');
    } finally {
      setSaving(false);
    }
  };

  // ── Toggle enabled ───────────────────────────────────────────────────────

  const toggleFlag = async (flag: ConfigFlag) => {
    try {
      await apiClient.toggleFlag(flag.key, currentEnvironment, projectId);
      showSnack(`"${flag.key}" ${flag.enabled ? 'disabled' : 'enabled'}`);
      await loadFlags();
    } catch {
      showSnack('Failed to toggle config', 'error');
    }
  };

  // ── Delete ───────────────────────────────────────────────────────────────

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await apiClient.deleteFlag(deleteTarget.key, currentEnvironment);
      showSnack(`"${deleteTarget.key}" deleted`);
      setDeleteTarget(null);
      await loadFlags();
    } catch {
      showSnack('Failed to delete', 'error');
    } finally {
      setDeleting(false);
    }
  };

  // ── Create ───────────────────────────────────────────────────────────────

  const openCreate = () => {
    setNewConfig({ ...DEFAULT_NEW_CONFIG });
    setCreateError(null);
    setCreateOpen(true);
  };

  const handleTypeChange = (type: ConfigType) => {
    setNewConfig(prev => ({ ...prev, type, value: defaultForType(type) }));
  };

  const handleCreate = async () => {
    setCreateError(null);
    if (!newConfig.key.trim()) { setCreateError('Key is required'); return; }
    if (!newConfig.name.trim()) { setCreateError('Name is required'); return; }
    if (newConfig.type === 'json') {
      try { JSON.parse(newConfig.value); } catch { setCreateError('Invalid JSON value'); return; }
    }

    setCreating(true);
    try {
      await apiClient.createFlag({
        key: newConfig.key.trim(),
        name: newConfig.name.trim(),
        description: newConfig.description.trim() || undefined,
        type: newConfig.type,
        enabled: newConfig.enabled,
        default: parseValue(newConfig.value, newConfig.type),
        project_id: projectId,
      }, currentEnvironment);
      showSnack(`"${newConfig.key}" created`);
      setCreateOpen(false);
      await loadFlags();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create';
      setCreateError(msg.includes('409') || msg.includes('already') ? 'A config with this key already exists' : msg);
    } finally {
      setCreating(false);
    }
  };

  // ── Copy snippet ─────────────────────────────────────────────────────────

  const copySnippet = (flag: Flag) => {
    const hookName = flag.type === 'number' ? 'useNumberFlag' : flag.type === 'json' ? 'useJSONFlag' : 'useStringFlag';
    const defVal = displayValue(flag);
    const snippet = `const { value } = ${hookName}('${flag.key}', ${defVal});`;
    navigator.clipboard.writeText(snippet);
    showSnack('SDK snippet copied');
  };

  const enabledCount = flags.filter(f => f.enabled).length;

  return (
    <Box>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <TuneIcon sx={{ color: 'primary.main', fontSize: 32 }} />
          <Box>
            <Typography variant="h4" fontWeight={700}>Remote Config</Typography>
            <Typography variant="body2" color="text.secondary">
              String, number, and JSON values served to your mobile apps — no app store release required
            </Typography>
          </Box>
        </Box>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="Refresh">
            <IconButton onClick={loadFlags} disabled={loading}><RefreshIcon /></IconButton>
          </Tooltip>
          <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>
            Add Config
          </Button>
        </Box>
      </Box>

      {/* Stats row */}
      <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
        {[
          { label: 'Total configs', value: flags.length, color: 'primary.main' },
          { label: 'Active', value: enabledCount, color: 'success.main' },
          { label: 'Disabled', value: flags.length - enabledCount, color: 'text.secondary' },
        ].map(stat => (
          <Paper key={stat.label} sx={{ px: 3, py: 2, flex: '0 0 auto', minWidth: 130 }}>
            <Typography variant="h5" fontWeight={700} sx={{ color: stat.color }}>{stat.value}</Typography>
            <Typography variant="caption" color="text.secondary">{stat.label}</Typography>
          </Paper>
        ))}
        <Paper sx={{ px: 3, py: 2, flex: 1 }}>
          <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>SDK Endpoint</Typography>
          <Typography variant="body2" fontFamily="monospace" sx={{ color: 'primary.main' }}>
            GET /api/v1/config?environment={currentEnvironment}
          </Typography>
        </Paper>
      </Box>

      {/* SDK hint */}
      <Alert severity="info" sx={{ mb: 3 }}
        action={
          <Button size="small" color="inherit" onClick={() => {
            const snippet = `import AsyncStorage from '@react-native-async-storage/async-storage';\nimport { FlexFlagProvider, useStringFlag, useNumberFlag, useJSONFlag } from '@flexflag/react-native';\n\n<FlexFlagProvider config={{ apiKey: 'YOUR_KEY', environment: '${currentEnvironment}' }} storage={AsyncStorage}>\n  <App />\n</FlexFlagProvider>`;
            navigator.clipboard.writeText(snippet);
            showSnack('Setup snippet copied');
          }}>Copy Setup</Button>
        }
      >
        <strong>React Native SDK</strong> — install <code>@flexflag/react-native</code> and use{' '}
        <code>useStringFlag</code>, <code>useNumberFlag</code>, <code>useJSONFlag</code> hooks.
        Values update every 30s without an app store release.
      </Alert>

      {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

      {/* Table */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}><CircularProgress /></Box>
      ) : flags.length === 0 ? (
        <Paper sx={{ p: 8, textAlign: 'center' }}>
          <TuneIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>No remote config entries yet</Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Add your first config value to start serving dynamic parameters to your mobile app.
          </Typography>
          <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>
            Add Config
          </Button>
        </Paper>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2 }}>
          <Table>
            <TableHead>
              <TableRow sx={{ '& th': { fontWeight: 700, bgcolor: 'grey.50' } }}>
                <TableCell>Key</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Value ({currentEnvironment})</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {flags.map(flag => {
                const isEditing = editingKey === flag.key;
                return (
                  <TableRow key={flag.key} sx={{ '&:hover': { bgcolor: 'grey.50' }, opacity: flag.enabled ? 1 : 0.6 }}>
                    <TableCell>
                      <Typography variant="body2" fontFamily="monospace" fontWeight={600}>{flag.key}</Typography>
                      {flag.description && (
                        <Typography variant="caption" color="text.secondary" display="block">{flag.description}</Typography>
                      )}
                    </TableCell>

                    <TableCell>
                      <Chip label={flag.type} size="small" color={TYPE_COLORS[flag.type] ?? 'default'} variant="outlined"
                        sx={{ fontFamily: 'monospace', fontSize: '0.7rem' }} />
                    </TableCell>

                    <TableCell sx={{ maxWidth: 320 }}>
                      {isEditing ? (
                        <TextField
                          value={editValue}
                          onChange={e => setEditValue(e.target.value)}
                          size="small" fullWidth
                          multiline={flag.type === 'json'}
                          minRows={flag.type === 'json' ? 3 : 1} maxRows={8}
                          inputProps={{ style: { fontFamily: 'monospace', fontSize: '0.8rem' } }}
                          autoFocus
                        />
                      ) : (
                        <Typography variant="body2" fontFamily="monospace" sx={{
                          maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis',
                          whiteSpace: flag.type === 'json' ? 'pre' : 'nowrap', fontSize: '0.8rem',
                        }}>
                          {displayValue(flag)}
                        </Typography>
                      )}
                    </TableCell>

                    <TableCell>
                      <Chip label={flag.enabled ? 'Active' : 'Disabled'} size="small"
                        color={flag.enabled ? 'success' : 'default'}
                        variant={flag.enabled ? 'filled' : 'outlined'} />
                    </TableCell>

                    <TableCell align="right">
                      {isEditing ? (
                        <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
                          <Tooltip title="Save">
                            <span>
                              <IconButton size="small" color="success" onClick={() => saveEdit(flag)} disabled={saving}>
                                {saving ? <CircularProgress size={16} /> : <CheckIcon fontSize="small" />}
                              </IconButton>
                            </span>
                          </Tooltip>
                          <Tooltip title="Cancel">
                            <IconButton size="small" onClick={cancelEdit} disabled={saving}>
                              <CloseIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      ) : (
                        <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
                          <Tooltip title="Copy SDK snippet">
                            <IconButton size="small" onClick={() => copySnippet(flag)}>
                              <CopyIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title={flag.enabled ? 'Disable' : 'Enable'}>
                            <IconButton size="small" onClick={() => toggleFlag(flag)}
                              color={flag.enabled ? 'default' : 'success'}>
                              <ToggleIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Edit value">
                            <IconButton size="small" onClick={() => startEdit(flag)}>
                              <EditIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Delete">
                            <IconButton size="small" color="error" onClick={() => setDeleteTarget(flag)}>
                              <DeleteIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* ── Create Dialog ──────────────────────────────────────────────────── */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ pb: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <TuneIcon color="primary" />
            <Typography variant="h6" fontWeight={700}>New Remote Config</Typography>
          </Box>
          <Typography variant="body2" color="text.secondary" mt={0.5}>
            String, number, and JSON parameters served to your app at runtime.
          </Typography>
        </DialogTitle>

        <Divider />

        <DialogContent sx={{ pt: 3, display: 'flex', flexDirection: 'column', gap: 2.5 }}>
          {createError && <Alert severity="error">{createError}</Alert>}

          {/* Type selector — choose first, it determines default value */}
          <FormControl fullWidth size="small">
            <InputLabel>Type</InputLabel>
            <Select
              value={newConfig.type}
              label="Type"
              onChange={e => handleTypeChange(e.target.value as ConfigType)}
            >
              <MenuItem value="string">
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Chip label="string" size="small" color="info" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Text values — themes, labels, URLs, copy</Typography>
                </Box>
              </MenuItem>
              <MenuItem value="number">
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Chip label="number" size="small" color="warning" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Numeric values — timeouts, limits, percentages</Typography>
                </Box>
              </MenuItem>
              <MenuItem value="json">
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Chip label="json" size="small" color="success" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Structured data — layouts, menus, config objects</Typography>
                </Box>
              </MenuItem>
            </Select>
          </FormControl>

          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="Key"
              value={newConfig.key}
              onChange={e => setNewConfig(p => ({ ...p, key: e.target.value.toLowerCase().replace(/\s+/g, '-') }))}
              size="small"
              fullWidth
              placeholder="e.g. app-theme"
              helperText="Lowercase, hyphens only. Used in SDK calls."
              inputProps={{ style: { fontFamily: 'monospace' } }}
            />
            <TextField
              label="Name"
              value={newConfig.name}
              onChange={e => setNewConfig(p => ({ ...p, name: e.target.value }))}
              size="small"
              fullWidth
              placeholder="e.g. App Theme"
            />
          </Box>

          <TextField
            label="Description"
            value={newConfig.description}
            onChange={e => setNewConfig(p => ({ ...p, description: e.target.value }))}
            size="small"
            fullWidth
            placeholder="What does this config control?"
          />

          <TextField
            label={`Default value (${newConfig.type})`}
            value={newConfig.value}
            onChange={e => setNewConfig(p => ({ ...p, value: e.target.value }))}
            size="small"
            fullWidth
            multiline={newConfig.type === 'json'}
            minRows={newConfig.type === 'json' ? 4 : 1}
            placeholder={TYPE_PLACEHOLDERS[newConfig.type]}
            helperText={newConfig.type === 'json' ? 'Must be valid JSON' : undefined}
            inputProps={{ style: { fontFamily: 'monospace' } }}
          />

          <FormControlLabel
            control={
              <Switch
                checked={newConfig.enabled}
                onChange={e => setNewConfig(p => ({ ...p, enabled: e.target.checked }))}
                color="success"
              />
            }
            label={
              <Box>
                <Typography variant="body2" fontWeight={500}>
                  {newConfig.enabled ? 'Active' : 'Disabled'}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {newConfig.enabled ? 'SDK clients will receive this value immediately' : 'Not served until enabled'}
                </Typography>
              </Box>
            }
          />
        </DialogContent>

        <Divider />

        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={() => setCreateOpen(false)} disabled={creating}>Cancel</Button>
          <Button variant="contained" onClick={handleCreate} disabled={creating}
            startIcon={creating ? <CircularProgress size={16} /> : <AddIcon />}>
            {creating ? 'Creating…' : 'Create Config'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* ── Delete Confirm Dialog ──────────────────────────────────────────── */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>Delete config?</DialogTitle>
        <DialogContent>
          <Typography variant="body2">
            <strong>{deleteTarget?.key}</strong> will be permanently deleted from{' '}
            <strong>{currentEnvironment}</strong>. SDK clients will stop receiving this value.
          </Typography>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setDeleteTarget(null)} disabled={deleting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={confirmDelete} disabled={deleting}
            startIcon={deleting ? <CircularProgress size={16} /> : <DeleteIcon />}>
            {deleting ? 'Deleting…' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={snackbar.open} autoHideDuration={3000}
        onClose={() => setSnackbar(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert severity={snackbar.severity} onClose={() => setSnackbar(s => ({ ...s, open: false }))} variant="filled">
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
