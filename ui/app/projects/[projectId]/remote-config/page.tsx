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
  Slider,
  FormGroup,
  Checkbox,
  FormLabel,
  Stack,
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
  PhoneAndroid as AndroidIcon,
  Apple as AppleIcon,
  Web as WebIcon,
  Percent as PercentIcon,
} from '@mui/icons-material';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { Flag } from '@/types';

type ConfigType = 'string' | 'number' | 'json';
type Platform = 'ios' | 'android' | 'web';
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

const ALL_PLATFORMS: { value: Platform; label: string; icon: React.ReactNode }[] = [
  { value: 'ios', label: 'iOS', icon: <AppleIcon sx={{ fontSize: 16 }} /> },
  { value: 'android', label: 'Android', icon: <AndroidIcon sx={{ fontSize: 16 }} /> },
  { value: 'web', label: 'Web', icon: <WebIcon sx={{ fontSize: 16 }} /> },
];

const DEFAULT_NEW_CONFIG = {
  key: '',
  name: '',
  description: '',
  type: 'string' as ConfigType,
  value: '',
  enabled: true,
  platforms: [] as Platform[],   // empty = all platforms
  rolloutPercentage: 100,
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

function getPlatforms(flag: Flag): Platform[] {
  return (flag.metadata?.platforms as Platform[]) ?? [];
}

function getRolloutPct(flag: Flag): number {
  return (flag.metadata?.rollout_percentage as number) ?? 100;
}

function PlatformBadges({ platforms }: { platforms: Platform[] }) {
  if (platforms.length === 0) return <Chip label="All platforms" size="small" variant="outlined" color="default" sx={{ fontSize: '0.7rem' }} />;
  return (
    <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
      {platforms.map(p => (
        <Chip
          key={p}
          label={p}
          size="small"
          variant="outlined"
          icon={p === 'ios' ? <AppleIcon /> : p === 'android' ? <AndroidIcon /> : <WebIcon />}
          sx={{ fontSize: '0.7rem' }}
        />
      ))}
    </Box>
  );
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

  // Create / Edit dialog state
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formData, setFormData] = useState({ ...DEFAULT_NEW_CONFIG });
  const [editTarget, setEditTarget] = useState<ConfigFlag | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

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

  // ── Inline value edit ────────────────────────────────────────────────────

  const startInlineEdit = (flag: ConfigFlag) => {
    setEditingKey(flag.key);
    const val = flag.default;
    setEditValue(typeof val === 'object' ? JSON.stringify(val, null, 2) : String(val ?? ''));
  };

  const cancelInlineEdit = () => { setEditingKey(null); setEditValue(''); };

  const saveInlineEdit = async (flag: ConfigFlag) => {
    setSaving(true);
    try {
      const parsed = parseValue(editValue, flag.type);
      await apiClient.updateFlag(flag.key, { ...flag, default: parsed }, currentEnvironment);
      showSnack(`"${flag.key}" updated`);
      cancelInlineEdit();
      await loadFlags();
    } catch {
      showSnack('Failed to save — check the value format', 'error');
    } finally {
      setSaving(false);
    }
  };

  // ── Toggle ────────────────────────────────────────────────────────────────

  const toggleFlag = async (flag: ConfigFlag) => {
    try {
      await apiClient.toggleFlag(flag.key, currentEnvironment, projectId);
      showSnack(`"${flag.key}" ${flag.enabled ? 'disabled' : 'enabled'}`);
      await loadFlags();
    } catch {
      showSnack('Failed to toggle config', 'error');
    }
  };

  // ── Delete ────────────────────────────────────────────────────────────────

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

  // ── Create / Edit dialog ─────────────────────────────────────────────────

  const openCreate = () => {
    setDialogMode('create');
    setFormData({ ...DEFAULT_NEW_CONFIG });
    setFormError(null);
    setEditTarget(null);
    setDialogOpen(true);
  };

  const openEdit = (flag: ConfigFlag) => {
    setDialogMode('edit');
    setFormError(null);
    setEditTarget(flag);
    setFormData({
      key: flag.key,
      name: flag.name,
      description: flag.description ?? '',
      type: flag.type as ConfigType,
      value: typeof flag.default === 'object' ? JSON.stringify(flag.default, null, 2) : String(flag.default ?? ''),
      enabled: flag.enabled,
      platforms: getPlatforms(flag),
      rolloutPercentage: getRolloutPct(flag),
    });
    setDialogOpen(true);
  };

  const handleTypeChange = (type: ConfigType) => {
    setFormData(prev => ({ ...prev, type, value: defaultForType(type) }));
  };

  const togglePlatform = (p: Platform) => {
    setFormData(prev => ({
      ...prev,
      platforms: prev.platforms.includes(p)
        ? prev.platforms.filter(x => x !== p)
        : [...prev.platforms, p],
    }));
  };

  const buildMetadata = () => {
    const meta: Record<string, unknown> = {};
    if (formData.platforms.length > 0) meta.platforms = formData.platforms;
    if (formData.rolloutPercentage < 100) meta.rollout_percentage = formData.rolloutPercentage;
    return Object.keys(meta).length > 0 ? meta : undefined;
  };

  const validateForm = (): string | null => {
    if (!formData.key.trim()) return 'Key is required';
    if (!formData.name.trim()) return 'Name is required';
    if (formData.type === 'json') {
      try { JSON.parse(formData.value); } catch { return 'Invalid JSON value'; }
    }
    if (formData.type === 'number' && isNaN(Number(formData.value))) return 'Value must be a number';
    return null;
  };

  const handleSubmit = async () => {
    const err = validateForm();
    if (err) { setFormError(err); return; }
    setFormError(null);
    setSubmitting(true);

    const payload = {
      key: formData.key.trim(),
      name: formData.name.trim(),
      description: formData.description.trim() || undefined,
      type: formData.type,
      enabled: formData.enabled,
      default: parseValue(formData.value, formData.type),
      project_id: projectId,
      metadata: buildMetadata(),
    };

    try {
      if (dialogMode === 'create') {
        await apiClient.createFlag(payload, currentEnvironment);
        showSnack(`"${payload.key}" created`);
      } else {
        await apiClient.updateFlag(payload.key, payload, currentEnvironment);
        showSnack(`"${payload.key}" updated`);
      }
      setDialogOpen(false);
      await loadFlags();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed';
      setFormError(msg.includes('409') || msg.includes('already') ? 'A config with this key already exists' : msg);
    } finally {
      setSubmitting(false);
    }
  };

  // ── SDK snippet ───────────────────────────────────────────────────────────

  const copySnippet = (flag: Flag) => {
    const hookName = flag.type === 'number' ? 'useNumberFlag' : flag.type === 'json' ? 'useJSONFlag' : 'useStringFlag';
    navigator.clipboard.writeText(`const { value } = ${hookName}('${flag.key}', ${displayValue(flag)});`);
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

      {/* Stats */}
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
            POST /api/v1/config/evaluate?environment={currentEnvironment}
          </Typography>
        </Paper>
      </Box>

      {/* SDK hint */}
      <Alert severity="info" sx={{ mb: 3 }}
        action={
          <Button size="small" color="inherit" onClick={() => {
            const s = `import AsyncStorage from '@react-native-async-storage/async-storage';\nimport { FlexFlagProvider, useStringFlag, useNumberFlag, useJSONFlag } from '@flexflag/react-native';\n\n<FlexFlagProvider\n  config={{ apiKey: 'YOUR_KEY', environment: '${currentEnvironment}' }}\n  storage={AsyncStorage}\n>\n  <App />\n</FlexFlagProvider>\n\n// In a component (platform + user context for targeting):\nconst client = useFlexFlagClient();\nawait client.getConfig({ userId: 'user-123', platform: 'ios' });`;
            navigator.clipboard.writeText(s);
            showSnack('Setup snippet copied');
          }}>Copy Setup</Button>
        }
      >
        <strong>React Native SDK</strong> — install <code>@flexflag/react-native</code>.
        Pass <code>platform: &quot;ios&quot;</code> or <code>&quot;android&quot;</code> in the context to respect platform targeting and rollout percentages.
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
          <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>Add Config</Button>
        </Paper>
      ) : (
        <TableContainer component={Paper} sx={{ borderRadius: 2 }}>
          <Table>
            <TableHead>
              <TableRow sx={{ '& th': { fontWeight: 700, bgcolor: 'grey.50' } }}>
                <TableCell>Key</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Value ({currentEnvironment})</TableCell>
                <TableCell>Platforms</TableCell>
                <TableCell>Rollout</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {flags.map(flag => {
                const isEditing = editingKey === flag.key;
                const platforms = getPlatforms(flag);
                const rolloutPct = getRolloutPct(flag);

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

                    <TableCell sx={{ maxWidth: 240 }}>
                      {isEditing ? (
                        <TextField
                          value={editValue}
                          onChange={e => setEditValue(e.target.value)}
                          size="small" fullWidth
                          multiline={flag.type === 'json'}
                          minRows={flag.type === 'json' ? 3 : 1} maxRows={6}
                          inputProps={{ style: { fontFamily: 'monospace', fontSize: '0.8rem' } }}
                          autoFocus
                        />
                      ) : (
                        <Typography variant="body2" fontFamily="monospace" sx={{
                          maxWidth: 230, overflow: 'hidden', textOverflow: 'ellipsis',
                          whiteSpace: flag.type === 'json' ? 'pre' : 'nowrap', fontSize: '0.8rem',
                        }}>
                          {displayValue(flag)}
                        </Typography>
                      )}
                    </TableCell>

                    <TableCell>
                      <PlatformBadges platforms={platforms} />
                    </TableCell>

                    <TableCell>
                      {rolloutPct < 100 ? (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          <PercentIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                          <Typography variant="body2" fontWeight={600} color={rolloutPct < 50 ? 'warning.main' : 'success.main'}>
                            {rolloutPct}%
                          </Typography>
                        </Box>
                      ) : (
                        <Typography variant="body2" color="text.disabled">100%</Typography>
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
                              <IconButton size="small" color="success" onClick={() => saveInlineEdit(flag)} disabled={saving}>
                                {saving ? <CircularProgress size={16} /> : <CheckIcon fontSize="small" />}
                              </IconButton>
                            </span>
                          </Tooltip>
                          <Tooltip title="Cancel">
                            <IconButton size="small" onClick={cancelInlineEdit} disabled={saving}>
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
                          <Tooltip title="Edit settings">
                            <IconButton size="small" onClick={() => openEdit(flag)}>
                              <EditIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Quick edit value">
                            <IconButton size="small" color="primary" onClick={() => startInlineEdit(flag)}>
                              <TuneIcon fontSize="small" />
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

      {/* ── Create / Edit Dialog ──────────────────────────────────────────── */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ pb: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <TuneIcon color="primary" />
            <Typography variant="h6" fontWeight={700}>
              {dialogMode === 'create' ? 'New Remote Config' : `Edit "${editTarget?.key}"`}
            </Typography>
          </Box>
          <Typography variant="body2" color="text.secondary" mt={0.5}>
            {dialogMode === 'create'
              ? 'String, number, and JSON parameters served to your app at runtime.'
              : 'Update value, platform targeting, or rollout percentage.'}
          </Typography>
        </DialogTitle>

        <Divider />

        <DialogContent sx={{ pt: 3, display: 'flex', flexDirection: 'column', gap: 2.5 }}>
          {formError && <Alert severity="error">{formError}</Alert>}

          {/* Type */}
          <FormControl fullWidth size="small" disabled={dialogMode === 'edit'}>
            <InputLabel>Type</InputLabel>
            <Select value={formData.type} label="Type" onChange={e => handleTypeChange(e.target.value as ConfigType)}>
              <MenuItem value="string">
                <Stack direction="row" alignItems="center" gap={1}>
                  <Chip label="string" size="small" color="info" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Text — themes, labels, URLs, copy</Typography>
                </Stack>
              </MenuItem>
              <MenuItem value="number">
                <Stack direction="row" alignItems="center" gap={1}>
                  <Chip label="number" size="small" color="warning" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Numeric — timeouts, limits, percentages</Typography>
                </Stack>
              </MenuItem>
              <MenuItem value="json">
                <Stack direction="row" alignItems="center" gap={1}>
                  <Chip label="json" size="small" color="success" variant="outlined" sx={{ fontSize: '0.7rem' }} />
                  <Typography variant="body2">Structured data — layouts, menus, objects</Typography>
                </Stack>
              </MenuItem>
            </Select>
          </FormControl>

          {/* Key + Name */}
          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="Key"
              value={formData.key}
              onChange={e => setFormData(p => ({ ...p, key: e.target.value.toLowerCase().replace(/\s+/g, '-') }))}
              size="small" fullWidth
              placeholder="e.g. app-theme"
              helperText="Lowercase, hyphens only."
              inputProps={{ style: { fontFamily: 'monospace' } }}
              disabled={dialogMode === 'edit'}
            />
            <TextField
              label="Name"
              value={formData.name}
              onChange={e => setFormData(p => ({ ...p, name: e.target.value }))}
              size="small" fullWidth placeholder="e.g. App Theme"
            />
          </Box>

          <TextField
            label="Description"
            value={formData.description}
            onChange={e => setFormData(p => ({ ...p, description: e.target.value }))}
            size="small" fullWidth placeholder="What does this config control?"
          />

          {/* Default value */}
          <TextField
            label={`Default value (${formData.type})`}
            value={formData.value}
            onChange={e => setFormData(p => ({ ...p, value: e.target.value }))}
            size="small" fullWidth
            multiline={formData.type === 'json'} minRows={formData.type === 'json' ? 4 : 1}
            placeholder={TYPE_PLACEHOLDERS[formData.type]}
            helperText={formData.type === 'json' ? 'Must be valid JSON' : undefined}
            inputProps={{ style: { fontFamily: 'monospace' } }}
          />

          <Divider />

          {/* Platform targeting */}
          <Box>
            <FormLabel component="legend" sx={{ fontSize: '0.875rem', fontWeight: 600, mb: 1 }}>
              Platform targeting
            </FormLabel>
            <Typography variant="caption" color="text.secondary" display="block" mb={1.5}>
              Leave all unchecked to serve to all platforms.
            </Typography>
            <FormGroup row>
              {ALL_PLATFORMS.map(({ value, label, icon }) => (
                <FormControlLabel
                  key={value}
                  control={
                    <Checkbox
                      checked={formData.platforms.includes(value)}
                      onChange={() => togglePlatform(value)}
                      size="small"
                    />
                  }
                  label={
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      {icon}
                      <Typography variant="body2">{label}</Typography>
                    </Box>
                  }
                />
              ))}
            </FormGroup>
          </Box>

          {/* Rollout percentage */}
          <Box>
            <FormLabel component="legend" sx={{ fontSize: '0.875rem', fontWeight: 600, mb: 0.5 }}>
              Rollout percentage — {formData.rolloutPercentage}%
            </FormLabel>
            <Typography variant="caption" color="text.secondary" display="block" mb={1.5}>
              Each user gets a consistent value based on their user ID (sticky bucketing).
              100% serves all users.
            </Typography>
            <Slider
              value={formData.rolloutPercentage}
              onChange={(_, v) => setFormData(p => ({ ...p, rolloutPercentage: v as number }))}
              min={0} max={100} step={5}
              marks={[
                { value: 0, label: '0%' },
                { value: 25, label: '25%' },
                { value: 50, label: '50%' },
                { value: 75, label: '75%' },
                { value: 100, label: '100%' },
              ]}
              valueLabelDisplay="auto"
              color={formData.rolloutPercentage < 50 ? 'warning' : 'primary'}
            />
          </Box>

          <Divider />

          {/* Enabled toggle */}
          <FormControlLabel
            control={
              <Switch checked={formData.enabled} onChange={e => setFormData(p => ({ ...p, enabled: e.target.checked }))} color="success" />
            }
            label={
              <Box>
                <Typography variant="body2" fontWeight={500}>{formData.enabled ? 'Active' : 'Disabled'}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {formData.enabled ? 'SDK clients will receive this value' : 'Not served until enabled'}
                </Typography>
              </Box>
            }
          />
        </DialogContent>

        <Divider />

        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={() => setDialogOpen(false)} disabled={submitting}>Cancel</Button>
          <Button variant="contained" onClick={handleSubmit} disabled={submitting}
            startIcon={submitting ? <CircularProgress size={16} /> : (dialogMode === 'create' ? <AddIcon /> : <CheckIcon />)}>
            {submitting ? (dialogMode === 'create' ? 'Creating…' : 'Saving…') : (dialogMode === 'create' ? 'Create Config' : 'Save Changes')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* ── Delete Confirm ────────────────────────────────────────────────── */}
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
