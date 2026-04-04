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
} from '@mui/material';
import {
  Edit as EditIcon,
  Check as CheckIcon,
  Close as CloseIcon,
  Refresh as RefreshIcon,
  ContentCopy as CopyIcon,
  Tune as TuneIcon,
} from '@mui/icons-material';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { Flag } from '@/types';

type ConfigFlag = Flag & { _editValue?: string };

const TYPE_COLORS: Record<string, 'info' | 'warning' | 'success' | 'default'> = {
  string: 'info',
  number: 'warning',
  json: 'success',
};

function displayValue(flag: Flag): string {
  const val = flag.default;
  if (val === null || val === undefined) return '—';
  if (typeof val === 'object') return JSON.stringify(val);
  return String(val);
}

function parseValue(raw: string, type: string): unknown {
  if (type === 'number') return Number(raw);
  if (type === 'json') {
    try { return JSON.parse(raw); } catch { return raw; }
  }
  return raw;
}

export default function RemoteConfigPage() {
  const params = useParams();
  const projectId = params.projectId as string;
  const { currentEnvironment } = useEnvironment();

  const [flags, setFlags] = useState<ConfigFlag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [saving, setSaving] = useState(false);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false, message: '', severity: 'success',
  });

  const loadFlags = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const all = await apiClient.getFlags(currentEnvironment, projectId);
      // Remote config = string, number, json only
      setFlags(all.filter(f => f.type !== 'boolean' && f.type !== 'variant'));
    } catch {
      setError('Failed to load remote config. Make sure the server is running.');
    } finally {
      setLoading(false);
    }
  }, [currentEnvironment, projectId]);

  useEffect(() => { loadFlags(); }, [loadFlags]);

  const startEdit = (flag: ConfigFlag) => {
    setEditingKey(flag.key);
    const val = flag.default;
    setEditValue(typeof val === 'object' ? JSON.stringify(val, null, 2) : String(val ?? ''));
  };

  const cancelEdit = () => {
    setEditingKey(null);
    setEditValue('');
  };

  const saveEdit = async (flag: ConfigFlag) => {
    setSaving(true);
    try {
      const parsed = parseValue(editValue, flag.type);
      await apiClient.updateFlag(flag.key, { ...flag, default: parsed }, currentEnvironment);
      setSnackbar({ open: true, message: `"${flag.key}" updated`, severity: 'success' });
      cancelEdit();
      await loadFlags();
    } catch {
      setSnackbar({ open: true, message: 'Failed to save — check the value format', severity: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const copySnippet = (flag: Flag) => {
    const snippet = `// React Native\nconst { value } = use${flag.type === 'number' ? 'Number' : flag.type === 'json' ? 'JSON' : 'String'}Flag('${flag.key}', ${displayValue(flag)});`;
    navigator.clipboard.writeText(snippet);
    setSnackbar({ open: true, message: 'SDK snippet copied', severity: 'success' });
  };

  const enabledCount = flags.filter(f => f.enabled).length;

  return (
    <Box>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <TuneIcon sx={{ color: 'primary.main', fontSize: 32 }} />
          <Box>
            <Typography variant="h4" fontWeight={700}>
              Remote Config
            </Typography>
            <Typography variant="body2" color="text.secondary">
              String, number, and JSON values served to your mobile apps — no app store release required
            </Typography>
          </Box>
        </Box>
        <Tooltip title="Refresh">
          <IconButton onClick={loadFlags} disabled={loading}>
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {/* Stats row */}
      <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
        {[
          { label: 'Total configs', value: flags.length, color: 'primary.main' },
          { label: 'Active', value: enabledCount, color: 'success.main' },
          { label: 'Disabled', value: flags.length - enabledCount, color: 'text.secondary' },
        ].map(stat => (
          <Paper key={stat.label} sx={{ px: 3, py: 2, flex: '0 0 auto', minWidth: 130 }}>
            <Typography variant="h5" fontWeight={700} sx={{ color: stat.color }}>
              {stat.value}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {stat.label}
            </Typography>
          </Paper>
        ))}

        <Paper sx={{ px: 3, py: 2, flex: 1 }}>
          <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
            SDK Endpoint
          </Typography>
          <Typography variant="body2" fontFamily="monospace" sx={{ color: 'primary.main' }}>
            GET /api/v1/config?environment={currentEnvironment}
          </Typography>
        </Paper>
      </Box>

      {/* SDK Usage hint */}
      <Alert
        severity="info"
        sx={{ mb: 3 }}
        action={
          <Button
            size="small"
            color="inherit"
            onClick={() => {
              const snippet = `import AsyncStorage from '@react-native-async-storage/async-storage';\nimport { FlexFlagProvider, useStringFlag, useNumberFlag, useJSONFlag } from '@flexflag/react-native';\n\n// Wrap your app:\n<FlexFlagProvider config={{ apiKey: 'YOUR_KEY', environment: '${currentEnvironment}' }} storage={AsyncStorage}>\n  <App />\n</FlexFlagProvider>\n\n// Inside any component:\nconst { value: theme } = useStringFlag('app-theme', 'light');\nconst { value: config } = useJSONFlag('onboarding-config', {});`;
              navigator.clipboard.writeText(snippet);
              setSnackbar({ open: true, message: 'Setup snippet copied', severity: 'success' });
            }}
          >
            Copy Setup
          </Button>
        }
      >
        <strong>React Native SDK</strong> — install <code>@flexflag/react-native</code> and use{' '}
        <code>useStringFlag</code>, <code>useNumberFlag</code>, <code>useJSONFlag</code> hooks.
        Values update automatically every 30s without an app store release.
      </Alert>

      {/* Error */}
      {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

      {/* Table */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : flags.length === 0 ? (
        <Paper sx={{ p: 8, textAlign: 'center' }}>
          <TuneIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No remote config entries yet
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Create string, number, or JSON flags in Feature Flags — they automatically appear here.
          </Typography>
          <Button variant="contained" href={`/projects/${projectId}/flags`}>
            Go to Feature Flags
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
                  <TableRow
                    key={flag.key}
                    sx={{
                      '&:hover': { bgcolor: 'grey.50' },
                      opacity: flag.enabled ? 1 : 0.6,
                    }}
                  >
                    {/* Key */}
                    <TableCell>
                      <Typography variant="body2" fontFamily="monospace" fontWeight={600}>
                        {flag.key}
                      </Typography>
                      {flag.description && (
                        <Typography variant="caption" color="text.secondary" display="block">
                          {flag.description}
                        </Typography>
                      )}
                    </TableCell>

                    {/* Type */}
                    <TableCell>
                      <Chip
                        label={flag.type}
                        size="small"
                        color={TYPE_COLORS[flag.type] ?? 'default'}
                        variant="outlined"
                        sx={{ fontFamily: 'monospace', fontSize: '0.7rem' }}
                      />
                    </TableCell>

                    {/* Value (inline edit) */}
                    <TableCell sx={{ maxWidth: 320 }}>
                      {isEditing ? (
                        <TextField
                          value={editValue}
                          onChange={e => setEditValue(e.target.value)}
                          size="small"
                          fullWidth
                          multiline={flag.type === 'json'}
                          minRows={flag.type === 'json' ? 3 : 1}
                          maxRows={8}
                          inputProps={{ style: { fontFamily: 'monospace', fontSize: '0.8rem' } }}
                          autoFocus
                        />
                      ) : (
                        <Typography
                          variant="body2"
                          fontFamily="monospace"
                          sx={{
                            maxWidth: 300,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: flag.type === 'json' ? 'pre' : 'nowrap',
                            fontSize: '0.8rem',
                          }}
                        >
                          {displayValue(flag)}
                        </Typography>
                      )}
                    </TableCell>

                    {/* Status */}
                    <TableCell>
                      <Chip
                        label={flag.enabled ? 'Active' : 'Disabled'}
                        size="small"
                        color={flag.enabled ? 'success' : 'default'}
                        variant={flag.enabled ? 'filled' : 'outlined'}
                      />
                    </TableCell>

                    {/* Actions */}
                    <TableCell align="right">
                      {isEditing ? (
                        <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
                          <Tooltip title="Save">
                            <span>
                              <IconButton
                                size="small"
                                color="success"
                                onClick={() => saveEdit(flag)}
                                disabled={saving}
                              >
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
                          <Tooltip title="Edit value">
                            <IconButton size="small" onClick={() => startEdit(flag)}>
                              <EditIcon fontSize="small" />
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

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar(s => ({ ...s, open: false }))}
          variant="filled"
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
