'use client';

import { useState, useEffect, useMemo } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import {
  Box,
  Typography,
  Paper,
  Grid,
  Chip,
  TextField,
  InputAdornment,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  LinearProgress,
  Skeleton,
  Divider,
  IconButton,
  Collapse,
} from '@mui/material';
import {
  Search as SearchIcon,
  Bolt as BoltIcon,
  People as PeopleIcon,
  Flag as FlagIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
} from '@mui/icons-material';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  Cell,
  PieChart,
  Pie,
  Legend,
} from 'recharts';

interface FlagStat {
  total_evaluations: number;
  unique_users: number;
  last_evaluated_at?: string;
  variation_counts: Record<string, number>;
}

type SortKey = 'key' | 'total_evaluations' | 'unique_users' | 'last_evaluated_at';
type SortDir = 'asc' | 'desc';

const COLORS = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899'];

function formatNum(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return n.toString();
}

function timeAgo(iso?: string): string {
  if (!iso) return '—';
  const diff = Date.now() - new Date(iso).getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return `${s}s ago`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

function VariationBar({ counts }: { counts: Record<string, number> }) {
  const entries = Object.entries(counts);
  if (entries.length === 0) return <Typography variant="caption" color="text.disabled">—</Typography>;
  const total = entries.reduce((s, [, v]) => s + v, 0);
  return (
    <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
      {entries.map(([name, count], i) => (
        <Chip
          key={name}
          label={`${name}: ${Math.round((count / total) * 100)}%`}
          size="small"
          sx={{
            fontSize: '0.6rem',
            height: 18,
            bgcolor: COLORS[i % COLORS.length] + '22',
            color: COLORS[i % COLORS.length],
            borderColor: COLORS[i % COLORS.length] + '66',
            border: '1px solid',
          }}
        />
      ))}
    </Box>
  );
}

function FlagDetailRow({ flagKey, stat, maxEvals }: { flagKey: string; stat: FlagStat; maxEvals: number }) {
  const [open, setOpen] = useState(false);
  const pieData = Object.entries(stat.variation_counts).map(([name, value], i) => ({
    name,
    value,
    color: COLORS[i % COLORS.length],
  }));
  const pct = maxEvals > 0 ? Math.round((stat.total_evaluations / maxEvals) * 100) : 0;

  return (
    <>
      <TableRow
        hover
        sx={{ cursor: 'pointer', '& td': { py: 1.5 } }}
        onClick={() => setOpen(o => !o)}
      >
        <TableCell>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <IconButton size="small" sx={{ p: 0.25 }}>
              {open ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}
            </IconButton>
            <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
              {flagKey}
            </Typography>
          </Box>
        </TableCell>
        <TableCell>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <Typography variant="body2" fontWeight="600">
              {formatNum(stat.total_evaluations)}
            </Typography>
            <LinearProgress
              variant="determinate"
              value={pct}
              sx={{ flex: 1, height: 4, borderRadius: 1, bgcolor: 'action.hover', minWidth: 60, maxWidth: 100 }}
              color="warning"
            />
          </Box>
        </TableCell>
        <TableCell>
          <Typography variant="body2">{formatNum(stat.unique_users)}</Typography>
        </TableCell>
        <TableCell>
          <VariationBar counts={stat.variation_counts} />
        </TableCell>
        <TableCell>
          <Typography variant="caption" color="text.secondary">
            {timeAgo(stat.last_evaluated_at)}
          </Typography>
        </TableCell>
      </TableRow>

      {/* Expanded detail row */}
      <TableRow>
        <TableCell colSpan={5} sx={{ p: 0, border: 0 }}>
          <Collapse in={open} unmountOnExit>
            <Box sx={{ p: 3, bgcolor: 'action.hover' }}>
              <Grid container spacing={3}>
                {/* Variation breakdown chart */}
                {pieData.length > 0 && (
                  <Grid item xs={12} sm={5}>
                    <Typography variant="caption" color="text.secondary" fontWeight="600" sx={{ textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                      Variation Distribution
                    </Typography>
                    <ResponsiveContainer width="100%" height={160}>
                      <PieChart>
                        <Pie
                          data={pieData}
                          cx="50%"
                          cy="50%"
                          innerRadius={40}
                          outerRadius={65}
                          dataKey="value"
                          paddingAngle={2}
                        >
                          {pieData.map((entry, i) => (
                            <Cell key={entry.name} fill={COLORS[i % COLORS.length]} />
                          ))}
                        </Pie>
                        <RechartsTooltip
                          formatter={(value: number, name: string) => [
                            `${value.toLocaleString()} (${Math.round((value / stat.total_evaluations) * 100)}%)`,
                            name,
                          ]}
                        />
                        <Legend iconSize={10} wrapperStyle={{ fontSize: '0.7rem' }} />
                      </PieChart>
                    </ResponsiveContainer>
                  </Grid>
                )}

                {/* Stats summary */}
                <Grid item xs={12} sm={pieData.length > 0 ? 7 : 12}>
                  <Typography variant="caption" color="text.secondary" fontWeight="600" sx={{ textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                    Summary
                  </Typography>
                  <Box sx={{ mt: 1.5, display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Typography variant="body2" color="text.secondary">Total evaluations</Typography>
                      <Typography variant="body2" fontWeight="600">{stat.total_evaluations.toLocaleString()}</Typography>
                    </Box>
                    <Divider />
                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Typography variant="body2" color="text.secondary">Unique users</Typography>
                      <Typography variant="body2" fontWeight="600">{stat.unique_users.toLocaleString()}</Typography>
                    </Box>
                    <Divider />
                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Typography variant="body2" color="text.secondary">Evals per user</Typography>
                      <Typography variant="body2" fontWeight="600">
                        {stat.unique_users > 0 ? (stat.total_evaluations / stat.unique_users).toFixed(1) : '—'}
                      </Typography>
                    </Box>
                    <Divider />
                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Typography variant="body2" color="text.secondary">Last evaluated</Typography>
                      <Typography variant="body2" fontWeight="600">
                        {stat.last_evaluated_at ? new Date(stat.last_evaluated_at).toLocaleString() : '—'}
                      </Typography>
                    </Box>
                    {Object.keys(stat.variation_counts).length > 0 && (
                      <>
                        <Divider />
                        <Typography variant="caption" color="text.secondary" fontWeight="600" sx={{ textTransform: 'uppercase', letterSpacing: '0.5px', mt: 0.5 }}>
                          Variation counts
                        </Typography>
                        {Object.entries(stat.variation_counts).map(([v, count], i) => (
                          <Box key={v} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                              <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: COLORS[i % COLORS.length] }} />
                              <Typography variant="body2" color="text.secondary">{v}</Typography>
                            </Box>
                            <Typography variant="body2" fontWeight="600">{count.toLocaleString()}</Typography>
                          </Box>
                        ))}
                      </>
                    )}
                  </Box>
                </Grid>
              </Grid>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
}

export default function AnalyticsPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;

  const [analytics, setAnalytics] = useState<Record<string, FlagStat>>({});
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('total_evaluations');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  useEffect(() => {
    const fetch = async () => {
      setLoading(true);
      try {
        const data = await apiClient.getFlagAnalytics(currentEnvironment);
        setAnalytics(data.flags ?? {});
      } catch {
        setAnalytics({});
      } finally {
        setLoading(false);
      }
    };
    fetch();
    const interval = setInterval(fetch, 30_000); // refresh every 30s
    return () => clearInterval(interval);
  }, [currentEnvironment, projectId]);

  const totalEvals = useMemo(() =>
    Object.values(analytics).reduce((s, f) => s + f.total_evaluations, 0), [analytics]);
  const totalUsers = useMemo(() =>
    Object.values(analytics).reduce((s, f) => s + f.unique_users, 0), [analytics]);
  const maxEvals = useMemo(() =>
    Math.max(...Object.values(analytics).map(f => f.total_evaluations), 1), [analytics]);

  const sorted = useMemo(() => {
    const entries = Object.entries(analytics).filter(([key]) =>
      key.toLowerCase().includes(search.toLowerCase())
    );
    entries.sort(([ak, av], [bk, bv]) => {
      let cmp = 0;
      if (sortKey === 'key') cmp = ak.localeCompare(bk);
      else if (sortKey === 'total_evaluations') cmp = av.total_evaluations - bv.total_evaluations;
      else if (sortKey === 'unique_users') cmp = av.unique_users - bv.unique_users;
      else if (sortKey === 'last_evaluated_at') {
        const at = av.last_evaluated_at ? new Date(av.last_evaluated_at).getTime() : 0;
        const bt = bv.last_evaluated_at ? new Date(bv.last_evaluated_at).getTime() : 0;
        cmp = at - bt;
      }
      return sortDir === 'asc' ? cmp : -cmp;
    });
    return entries;
  }, [analytics, search, sortKey, sortDir]);

  // Top 8 flags for bar chart
  const chartData = useMemo(() =>
    Object.entries(analytics)
      .sort(([, a], [, b]) => b.total_evaluations - a.total_evaluations)
      .slice(0, 8)
      .map(([key, stat]) => ({
        key: key.length > 16 ? key.slice(0, 16) + '…' : key,
        evaluations: stat.total_evaluations,
        users: stat.unique_users,
      })),
    [analytics]
  );

  const handleSort = (key: SortKey) => {
    if (sortKey === key) setSortDir(d => d === 'asc' ? 'desc' : 'asc');
    else { setSortKey(key); setSortDir('desc'); }
  };

  return (
    <Box>
      {/* Header */}
      <Box sx={{ mb: 4, pb: 3, borderBottom: '1px solid', borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box>
            <Typography variant="h5" fontWeight="600" gutterBottom sx={{ mb: 0.5 }}>
              Analytics
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Real-time flag evaluation metrics — resets on server restart
            </Typography>
          </Box>
          <Chip label={currentEnvironment} variant="outlined" size="small" />
        </Box>
      </Box>

      {/* Summary cards */}
      <Grid container spacing={2} sx={{ mb: 4 }}>
        {[
          { label: 'Total Evaluations', value: loading ? null : formatNum(totalEvals), icon: <BoltIcon />, color: 'warning.main' },
          { label: 'Unique Users', value: loading ? null : formatNum(totalUsers), icon: <PeopleIcon />, color: 'primary.main' },
          { label: 'Active Flags', value: loading ? null : Object.keys(analytics).length.toString(), icon: <FlagIcon />, color: 'success.main' },
          {
            label: 'Avg Evals / Flag',
            value: loading ? null : (Object.keys(analytics).length > 0
              ? formatNum(Math.round(totalEvals / Object.keys(analytics).length))
              : '0'),
            icon: <BoltIcon />,
            color: 'secondary.main',
          },
        ].map(({ label, value, icon, color }) => (
          <Grid item xs={6} sm={3} key={label}>
            <Paper sx={{ p: 2.5, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                <Box sx={{ color, display: 'flex' }}>{icon}</Box>
                <Box>
                  {value === null ? (
                    <Skeleton width={60} height={28} />
                  ) : (
                    <Typography variant="h5" fontWeight="700">{value}</Typography>
                  )}
                  <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                    {label}
                  </Typography>
                </Box>
              </Box>
            </Paper>
          </Grid>
        ))}
      </Grid>

      {/* Bar chart */}
      {!loading && chartData.length > 0 && (
        <Paper sx={{ p: 3, mb: 4, border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
          <Typography variant="subtitle2" fontWeight="600" gutterBottom sx={{ mb: 2 }}>
            Top Flags by Evaluations
          </Typography>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={chartData} margin={{ top: 0, right: 8, left: -16, bottom: 0 }}>
              <XAxis dataKey="key" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} />
              <RechartsTooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(value: number, name: string) => [value.toLocaleString(), name === 'evaluations' ? 'Evaluations' : 'Users']}
              />
              <Bar dataKey="evaluations" radius={[4, 4, 0, 0]} maxBarSize={40}>
                {chartData.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </Paper>
      )}

      {/* Search + table */}
      <Paper sx={{ border: '1px solid', borderColor: 'divider', boxShadow: 0 }}>
        <Box sx={{ p: 2, borderBottom: '1px solid', borderColor: 'divider' }}>
          <TextField
            size="small"
            placeholder="Search flags…"
            value={search}
            onChange={e => setSearch(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" />
                </InputAdornment>
              ),
            }}
            sx={{ width: 280 }}
          />
        </Box>

        {loading ? (
          <Box sx={{ p: 3 }}>
            {[...Array(5)].map((_, i) => <Skeleton key={i} height={48} sx={{ mb: 0.5 }} />)}
          </Box>
        ) : sorted.length === 0 ? (
          <Box sx={{ p: 6, textAlign: 'center' }}>
            <BoltIcon sx={{ fontSize: 40, color: 'grey.400', mb: 1 }} />
            <Typography color="text.secondary">
              {search ? 'No flags match your search.' : 'No evaluations recorded yet. Make some SDK calls to see data here.'}
            </Typography>
          </Box>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow sx={{ '& th': { fontWeight: 600, fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.5px' } }}>
                  <TableCell>
                    <TableSortLabel
                      active={sortKey === 'key'}
                      direction={sortKey === 'key' ? sortDir : 'asc'}
                      onClick={() => handleSort('key')}
                    >
                      Flag Key
                    </TableSortLabel>
                  </TableCell>
                  <TableCell>
                    <TableSortLabel
                      active={sortKey === 'total_evaluations'}
                      direction={sortKey === 'total_evaluations' ? sortDir : 'desc'}
                      onClick={() => handleSort('total_evaluations')}
                    >
                      Evaluations
                    </TableSortLabel>
                  </TableCell>
                  <TableCell>
                    <TableSortLabel
                      active={sortKey === 'unique_users'}
                      direction={sortKey === 'unique_users' ? sortDir : 'desc'}
                      onClick={() => handleSort('unique_users')}
                    >
                      Unique Users
                    </TableSortLabel>
                  </TableCell>
                  <TableCell>Variations</TableCell>
                  <TableCell>
                    <TableSortLabel
                      active={sortKey === 'last_evaluated_at'}
                      direction={sortKey === 'last_evaluated_at' ? sortDir : 'desc'}
                      onClick={() => handleSort('last_evaluated_at')}
                    >
                      Last Seen
                    </TableSortLabel>
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {sorted.map(([key, stat]) => (
                  <FlagDetailRow key={key} flagKey={key} stat={stat} maxEvals={maxEvals} />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>
    </Box>
  );
}
