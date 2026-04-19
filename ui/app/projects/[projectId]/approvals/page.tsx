'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import {
  Box, Typography, Paper, Chip, Button, Tabs, Tab,
  Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, Divider, Grid, Alert, Skeleton,
  Switch, FormControlLabel, IconButton, Tooltip,
} from '@mui/material';
import {
  CheckCircle as ApproveIcon,
  Cancel as RejectIcon,
  HowToVote as ApprovalIcon,
  Settings as SettingsIcon,
  Slack as SlackIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';

interface ApprovalRequest {
  id: string;
  project_id: string;
  flag_key: string;
  environment: string;
  change_type: string;
  payload: Record<string, any>;
  requested_by_name: string;
  status: 'pending' | 'approved' | 'rejected';
  review_note?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
}

const STATUS_COLORS = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
} as const;

function describeChange(req: ApprovalRequest) {
  switch (req.change_type) {
    case 'toggle':
      return req.payload.new_enabled ? 'Enable flag' : 'Disable flag';
    case 'update':
      return 'Update flag configuration';
    case 'delete':
      return 'Delete flag';
    default:
      return req.change_type;
  }
}

function ReviewDialog({
  open, action, onClose, onConfirm,
}: {
  open: boolean;
  action: 'approve' | 'reject';
  onClose: () => void;
  onConfirm: (note: string) => void;
}) {
  const [note, setNote] = useState('');
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{action === 'approve' ? 'Approve Change' : 'Reject Change'}</DialogTitle>
      <DialogContent>
        <TextField
          fullWidth
          multiline
          rows={3}
          label="Note (optional)"
          placeholder={action === 'approve' ? 'Looks good, approved.' : 'Reason for rejection…'}
          value={note}
          onChange={e => setNote(e.target.value)}
          sx={{ mt: 1 }}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button
          variant="contained"
          color={action === 'approve' ? 'success' : 'error'}
          onClick={() => { onConfirm(note); setNote(''); }}
        >
          {action === 'approve' ? 'Approve & Apply' : 'Reject'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function SettingsPanel({ projectId }: { projectId: string }) {
  const [requireApproval, setRequireApproval] = useState(false);
  const [webhookURL, setWebhookURL] = useState('');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient.getApprovalSettings(projectId)
      .then(s => {
        setRequireApproval(s.require_approval);
        setWebhookURL(s.slack_webhook_url || '');
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [projectId]);

  const save = async () => {
    setSaving(true);
    try {
      await apiClient.updateApprovalSettings(projectId, {
        require_approval: requireApproval,
        slack_webhook_url: webhookURL,
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch {
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <Skeleton height={200} />;

  return (
    <Paper sx={{ p: 3, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
      <Typography variant="subtitle1" fontWeight="600" gutterBottom>
        Approval Workflow Settings
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
        When enabled, flag toggles and updates will require approval before taking effect.
      </Typography>

      <FormControlLabel
        control={
          <Switch
            checked={requireApproval}
            onChange={e => setRequireApproval(e.target.checked)}
            color="primary"
          />
        }
        label={
          <Box>
            <Typography variant="body2" fontWeight="600">Require approval for flag changes</Typography>
            <Typography variant="caption" color="text.secondary">
              All toggle and update actions will create a pending approval request
            </Typography>
          </Box>
        }
        sx={{ mb: 3, alignItems: 'flex-start', ml: 0 }}
      />

      <Divider sx={{ mb: 3 }} />

      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <Typography variant="body2" fontWeight="600">Slack Webhook URL</Typography>
        <Chip label="optional" size="small" sx={{ fontSize: '0.6rem', height: 18 }} />
      </Box>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5 }}>
        When set, approval requests are sent to this Slack channel automatically.
        <br />Get it from: Slack → Your App → Incoming Webhooks
      </Typography>
      <TextField
        fullWidth
        size="small"
        placeholder="https://hooks.slack.com/services/T.../B.../..."
        value={webhookURL}
        onChange={e => setWebhookURL(e.target.value)}
        InputProps={{
          endAdornment: webhookURL && (
            <Tooltip title="Copy">
              <IconButton size="small" onClick={() => navigator.clipboard.writeText(webhookURL)}>
                <CopyIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          ),
        }}
        sx={{ mb: 3 }}
      />

      {saved && <Alert severity="success" sx={{ mb: 2 }}>Settings saved.</Alert>}

      <Button variant="contained" onClick={save} disabled={saving}>
        {saving ? 'Saving…' : 'Save Settings'}
      </Button>
    </Paper>
  );
}

export default function ApprovalsPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;

  const [tab, setTab] = useState(0);
  const [approvals, setApprovals] = useState<ApprovalRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [reviewDialog, setReviewDialog] = useState<{ open: boolean; action: 'approve' | 'reject'; id: string }>({
    open: false, action: 'approve', id: '',
  });

  const statusFilter = ['pending', 'approved', 'rejected', ''][tab] as string;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await apiClient.getApprovals(projectId, statusFilter || undefined);
      setApprovals(data.approvals ?? []);
    } catch {
      setApprovals([]);
    } finally {
      setLoading(false);
    }
  }, [projectId, statusFilter]);

  useEffect(() => { load(); }, [load]);

  const openReview = (id: string, action: 'approve' | 'reject') => {
    setReviewDialog({ open: true, action, id });
  };

  const handleReview = async (note: string) => {
    const { action, id } = reviewDialog;
    setReviewDialog(d => ({ ...d, open: false }));
    try {
      if (action === 'approve') await apiClient.approveRequest(id, note);
      else await apiClient.rejectRequest(id, note);
      load();
    } catch (e: any) {
      alert(e.message || 'Failed');
    }
  };

  return (
    <Box>
      {/* Header */}
      <Box sx={{ mb: 4, pb: 3, borderBottom: '1px solid', borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 0.5 }}>
          <ApprovalIcon sx={{ color: 'warning.main' }} />
          <Typography variant="h5" fontWeight="600">Approvals</Typography>
        </Box>
        <Typography variant="body2" color="text.secondary">
          Review and action pending flag change requests
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {/* Left: requests list */}
        <Grid item xs={12} md={8}>
          <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3, borderBottom: '1px solid', borderColor: 'divider' }}>
            <Tab label="Pending" />
            <Tab label="Approved" />
            <Tab label="Rejected" />
            <Tab label="All" />
          </Tabs>

          {loading ? (
            [...Array(3)].map((_, i) => <Skeleton key={i} height={100} sx={{ mb: 1, borderRadius: 2 }} />)
          ) : approvals.length === 0 ? (
            <Paper sx={{ p: 5, textAlign: 'center', border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
              <ApprovalIcon sx={{ fontSize: 40, color: 'grey.400', mb: 1 }} />
              <Typography color="text.secondary">No {statusFilter || ''} requests</Typography>
            </Paper>
          ) : (
            approvals.map(req => (
              <Paper key={req.id} sx={{ p: 2.5, mb: 2, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1.5 }}>
                  <Box>
                    <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 0.5 }}>
                      <Typography variant="body2" fontWeight="700" sx={{ fontFamily: 'monospace' }}>
                        {req.flag_key}
                      </Typography>
                      <Chip label={req.environment} size="small" variant="outlined" sx={{ fontSize: '0.6rem', height: 18 }} />
                      <Chip
                        label={req.status}
                        size="small"
                        color={STATUS_COLORS[req.status]}
                        sx={{ fontSize: '0.6rem', height: 18 }}
                      />
                    </Box>
                    <Typography variant="body2" color="text.secondary">{describeChange(req)}</Typography>
                  </Box>

                  {req.status === 'pending' && (
                    <Box sx={{ display: 'flex', gap: 1 }}>
                      <Button
                        size="small"
                        variant="contained"
                        color="success"
                        startIcon={<ApproveIcon />}
                        onClick={() => openReview(req.id, 'approve')}
                      >
                        Approve
                      </Button>
                      <Button
                        size="small"
                        variant="outlined"
                        color="error"
                        startIcon={<RejectIcon />}
                        onClick={() => openReview(req.id, 'reject')}
                      >
                        Reject
                      </Button>
                    </Box>
                  )}
                </Box>

                <Divider sx={{ mb: 1.5 }} />

                <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Requested by</Typography>
                    <Typography variant="body2">{req.requested_by_name || 'Unknown'}</Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Submitted</Typography>
                    <Typography variant="body2">{new Date(req.created_at).toLocaleString()}</Typography>
                  </Box>
                  {req.review_note && (
                    <Box>
                      <Typography variant="caption" color="text.secondary">Note</Typography>
                      <Typography variant="body2">{req.review_note}</Typography>
                    </Box>
                  )}
                  {req.reviewed_at && (
                    <Box>
                      <Typography variant="caption" color="text.secondary">Reviewed</Typography>
                      <Typography variant="body2">{new Date(req.reviewed_at).toLocaleString()}</Typography>
                    </Box>
                  )}
                </Box>

                {/* Payload preview */}
                {req.change_type === 'toggle' && (
                  <Box sx={{ mt: 1.5, p: 1.5, bgcolor: 'action.hover', borderRadius: 1 }}>
                    <Typography variant="caption" color="text.secondary">
                      {req.payload.current_enabled ? 'Currently ON → will turn OFF' : 'Currently OFF → will turn ON'}
                    </Typography>
                  </Box>
                )}
              </Paper>
            ))
          )}
        </Grid>

        {/* Right: settings */}
        <Grid item xs={12} md={4}>
          <SettingsPanel projectId={projectId} />
        </Grid>
      </Grid>

      <ReviewDialog
        open={reviewDialog.open}
        action={reviewDialog.action}
        onClose={() => setReviewDialog(d => ({ ...d, open: false }))}
        onConfirm={handleReview}
      />
    </Box>
  );
}
