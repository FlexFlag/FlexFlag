'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Divider,
  Paper,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  Key as KeyIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  ContentCopy as CopyIcon,
  Security as SecurityIcon,
  Code as CodeIcon,
  ExpandMore as ExpandMoreIcon,
  PlayArrow as TestIcon,
} from '@mui/icons-material';

interface ApiKey {
  id: string;
  name: string;
  key_prefix: string;
  full_key?: string;
  project_id: string;
  environment_id: string;
  environment?: {
    id: string;
    key: string;
    name: string;
    description?: string;
  };
  permissions: string[];
  created_at: string;
  last_used_at?: string;
  expires_at?: string;
}

export default function ProjectApiKeysPage() {
  const params = useParams();
  const { currentEnvironment } = useEnvironment();
  const projectId = params.projectId as string;
  const [project, setProject] = useState<any>(null);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [environments, setEnvironments] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showKeyValue, setShowKeyValue] = useState<string | null>(null);
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  const [showTestDialog, setShowTestDialog] = useState(false);
  const [selectedKeyForTest, setSelectedKeyForTest] = useState<ApiKey | null>(null);
  const [testResult, setTestResult] = useState<any>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    environment_id: '',
    permissions: ['read'] as string[],
    expires_in_days: 365,
  });

  const [testRequest, setTestRequest] = useState({
    flag_key: 'test-flag',
    user_id: 'test_user_001',
    attributes: '{"email": "test@example.com"}',
  });

  const permissions = [
    { value: 'read', label: 'Read Flags', description: 'Evaluate flags and read flag configurations' },
    { value: 'write', label: 'Write Flags', description: 'Create, update, and delete flags' },
    { value: 'admin', label: 'Admin', description: 'Full project access including settings' },
  ];

  useEffect(() => {
    const fetchData = async () => {
      try {
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);
        
        if (foundProject) {
          // Get environments for this project
          const envs = await apiClient.getProjectEnvironments(foundProject.slug);
          setEnvironments(envs);
          
          // Set default environment_id if not set
          if (!formData.environment_id && envs.length > 0) {
            const currentEnv = envs.find(env => env.key === currentEnvironment);
            setFormData(prev => ({ 
              ...prev, 
              environment_id: currentEnv?.id || envs[0].id 
            }));
          }
        }
      } catch (error) {
        console.error('Error fetching project data:', error);
      }
    };

    if (projectId) {
      fetchData();
      fetchApiKeys();
    }
  }, [projectId, currentEnvironment]);

  const fetchApiKeys = async () => {
    try {
      setLoading(true);
      const keys = await apiClient.getApiKeys(projectId);
      setApiKeys(keys);
    } catch (error) {
      console.error('Error fetching API keys:', error);
      setApiKeys([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateApiKey = async () => {
    if (!formData.name.trim()) {
      alert('Please enter a name for the API key');
      return;
    }

    if (!formData.environment_id) {
      alert('Please select an environment');
      return;
    }

    try {
      const response = await apiClient.createApiKey(projectId, formData);
      
      // Set the generated key and refresh the list
      setGeneratedKey(response.api_key.full_key);
      fetchApiKeys();
      
      setFormData({
        name: '',
        environment_id: environments.find(env => env.key === currentEnvironment)?.id || environments[0]?.id || '',
        permissions: ['read'],
        expires_in_days: 365,
      });
    } catch (error) {
      console.error('Error creating API key:', error);
      alert('Failed to create API key');
    }
  };

  const handleDeleteApiKey = async (keyId: string) => {
    if (window.confirm('Are you sure you want to delete this API key? Applications using this key will lose access immediately.')) {
      try {
        await apiClient.deleteApiKey(projectId, keyId);
        fetchApiKeys();
      } catch (error) {
        console.error('Error deleting API key:', error);
        alert('Failed to delete API key');
      }
    }
  };

  const handleTestApiKey = async () => {
    if (!selectedKeyForTest) return;

    try {
      let attributes = {};
      try {
        attributes = JSON.parse(testRequest.attributes);
      } catch (e) {
        attributes = {};
      }

      // Simulate API call with API key authentication
      const request = {
        flag_key: testRequest.flag_key,
        user_id: testRequest.user_id,
        user_key: testRequest.user_id,
        attributes,
      };

      // Mock response - in real implementation, this would use the API key for auth
      const mockResponse = {
        flag_key: testRequest.flag_key,
        value: true,
        variation: 'control',
        reason: 'api_key_auth',
        default: false,
        evaluation_time_ms: 2.456,
        timestamp: new Date().toISOString(),
        api_key_id: selectedKeyForTest.id,
        environment: selectedKeyForTest.environment?.key,
      };

      setTestResult(mockResponse);
    } catch (error) {
      console.error('Error testing API key:', error);
      setTestResult({ error: 'Failed to evaluate flag with API key' });
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      console.log('Copied to clipboard');
    } catch (err) {
      console.error('Failed to copy:', err);
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

  const isExpiringSoon = (expiresAt: string) => {
    const daysUntilExpiry = (new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24);
    return daysUntilExpiry <= 30;
  };

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
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            API Keys
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage API keys for {project.name} • Environment-scoped authentication
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setShowCreateDialog(true)}
          size="large"
        >
          Generate API Key
        </Button>
      </Box>

      {/* API Key Stats */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" color="primary" fontWeight="bold">
              {apiKeys.length}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Active Keys
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" color="warning.main" fontWeight="bold">
              {apiKeys.filter(key => key.expires_at && isExpiringSoon(key.expires_at)).length}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Expiring Soon
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Paper sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="h4" color="success.main" fontWeight="bold">
              {new Set(apiKeys.map(key => key.environment?.key)).size}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Environments
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* API Keys List */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <KeyIcon />
            Project API Keys
          </Typography>

          {apiKeys.length === 0 ? (
            <Paper sx={{ p: 4, textAlign: 'center', bgcolor: 'grey.50' }}>
              <KeyIcon sx={{ fontSize: 48, color: 'grey.300', mb: 2 }} />
              <Typography variant="h6" color="text.secondary" gutterBottom>
                No API Keys
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Create your first API key to enable programmatic access to this project
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => setShowCreateDialog(true)}
              >
                Generate First API Key
              </Button>
            </Paper>
          ) : (
            <List>
              {apiKeys.map((key, index) => (
                <div key={key.id}>
                  <ListItem sx={{ px: 0 }}>
                    <ListItemText
                      primary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="subtitle1" fontWeight="bold">
                            {key.name}
                          </Typography>
                          <Chip 
                            label={key.environment?.name || key.environment?.key} 
                            size="small" 
                            color={key.environment?.key === 'production' ? 'error' : key.environment?.key === 'staging' ? 'warning' : 'info'}
                          />
                          {key.expires_at && isExpiringSoon(key.expires_at) && (
                            <Chip label="Expiring Soon" size="small" color="warning" />
                          )}
                        </Box>
                      }
                      secondary={
                        <Box>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace', mb: 1 }}>
                            {showKeyValue === key.id ? key.full_key : key.key_prefix}
                          </Typography>
                          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 1 }}>
                            {key.permissions.map(perm => (
                              <Chip key={perm} label={perm} size="small" variant="outlined" />
                            ))}
                          </Box>
                          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                            <Typography variant="caption" color="text.secondary">
                              Created: {formatDate(key.created_at)}
                            </Typography>
                            {key.last_used_at && (
                              <Typography variant="caption" color="text.secondary">
                                Last used: {formatDate(key.last_used_at)}
                              </Typography>
                            )}
                            {key.expires_at && (
                              <Typography variant="caption" color="text.secondary">
                                Expires: {formatDate(key.expires_at)}
                              </Typography>
                            )}
                          </Box>
                        </Box>
                      }
                    />
                    <ListItemSecondaryAction>
                      <IconButton
                        size="small"
                        onClick={() => {
                          setSelectedKeyForTest(key);
                          setTestResult(null);
                          setShowTestDialog(true);
                        }}
                        color="primary"
                      >
                        <TestIcon />
                      </IconButton>
                      {key.full_key && (
                        <IconButton
                          size="small"
                          onClick={() => copyToClipboard(key.full_key!)}
                          color="primary"
                        >
                          <CopyIcon />
                        </IconButton>
                      )}
                      <IconButton
                        size="small"
                        onClick={() => setShowKeyValue(showKeyValue === key.id ? null : key.id)}
                      >
                        {showKeyValue === key.id ? <VisibilityOffIcon /> : <VisibilityIcon />}
                      </IconButton>
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleDeleteApiKey(key.id)}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                  {index < apiKeys.length - 1 && <Divider />}
                </div>
              ))}
            </List>
          )}
        </CardContent>
      </Card>

      {/* Usage Examples */}
      <Box sx={{ mt: 4 }}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <CodeIcon />
              Usage Examples
            </Typography>
            
            <Accordion>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography>JavaScript/Node.js</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Box component="pre" sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, fontSize: '0.875rem', overflow: 'auto' }}>
{`// Evaluate a flag
const response = await fetch('http://localhost:8080/api/v1/evaluate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your_api_key_here'
  },
  body: JSON.stringify({
    flag_key: 'feature-flag',
    user_id: 'user_123',
    attributes: {
      email: 'user@example.com',
      plan: 'premium'
    }
  })
});

const result = await response.json();
console.log('Flag value:', result.value);`}
                </Box>
              </AccordionDetails>
            </Accordion>

            <Accordion>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography>cURL</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Box component="pre" sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, fontSize: '0.875rem', overflow: 'auto' }}>
{`curl -X POST http://localhost:8080/api/v1/evaluate \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: your_api_key_here" \\
  -d '{
    "flag_key": "feature-flag",
    "user_id": "user_123",
    "attributes": {
      "email": "user@example.com",
      "plan": "premium"
    }
  }'`}
                </Box>
              </AccordionDetails>
            </Accordion>

            <Accordion>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography>Python</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Box component="pre" sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, fontSize: '0.875rem', overflow: 'auto' }}>
{`import requests

def evaluate_flag(api_key, flag_key, user_id, attributes=None):
    response = requests.post(
        'http://localhost:8080/api/v1/evaluate',
        headers={
            'Content-Type': 'application/json',
            'X-API-Key': api_key
        },
        json={
            'flag_key': flag_key,
            'user_id': user_id,
            'attributes': attributes or {}
        }
    )
    return response.json()

# Usage
result = evaluate_flag(
    api_key='your_api_key_here',
    flag_key='feature-flag',
    user_id='user_123',
    attributes={'email': 'user@example.com', 'plan': 'premium'}
)
print(f"Flag value: {result['value']}")`}
                </Box>
              </AccordionDetails>
            </Accordion>
          </CardContent>
        </Card>
      </Box>

      {/* Create API Key Dialog */}
      <Dialog open={showCreateDialog} onClose={() => setShowCreateDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Generate New API Key</DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 3 }}>
            API keys are scoped to specific projects and environments. Choose permissions carefully.
          </Alert>
          
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="API Key Name"
                placeholder="e.g., Production Mobile App"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              />
            </Grid>
            
            <Grid item xs={12}>
              <FormControl fullWidth>
                <InputLabel>Environment</InputLabel>
                <Select
                  value={formData.environment_id}
                  onChange={(e) => setFormData(prev => ({ ...prev, environment_id: e.target.value }))}
                  label="Environment"
                >
                  {environments.map(env => (
                    <MenuItem key={env.id} value={env.id}>
                      {env.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            
            <Grid item xs={12}>
              <Typography variant="subtitle2" gutterBottom>
                Permissions
              </Typography>
              {permissions.map(perm => (
                <Box key={perm.value} sx={{ mb: 1 }}>
                  <label>
                    <input
                      type="checkbox"
                      checked={formData.permissions.includes(perm.value)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setFormData(prev => ({ 
                            ...prev, 
                            permissions: [...prev.permissions, perm.value] 
                          }));
                        } else {
                          setFormData(prev => ({ 
                            ...prev, 
                            permissions: prev.permissions.filter(p => p !== perm.value) 
                          }));
                        }
                      }}
                      style={{ marginRight: 8 }}
                    />
                    <strong>{perm.label}</strong> - {perm.description}
                  </label>
                </Box>
              ))}
            </Grid>
            
            <Grid item xs={12}>
              <TextField
                fullWidth
                type="number"
                label="Expires in (days)"
                value={formData.expires_in_days}
                onChange={(e) => setFormData(prev => ({ ...prev, expires_in_days: parseInt(e.target.value) || 365 }))}
                inputProps={{ min: 1, max: 3650 }}
              />
            </Grid>
          </Grid>
          
          {generatedKey && (
            <Alert severity="success" sx={{ mt: 3 }}>
              <Typography variant="subtitle2" gutterBottom>
                API Key Generated Successfully!
              </Typography>
              <Typography variant="body2" sx={{ fontFamily: 'monospace', wordBreak: 'break-all', mb: 1 }}>
                {generatedKey}
              </Typography>
              <Typography variant="caption">
                ⚠️ Copy this key now - you won't be able to see it again!
              </Typography>
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => {
            setShowCreateDialog(false);
            setGeneratedKey(null);
            setFormData({
              name: '',
              environment_id: environments.find(env => env.key === currentEnvironment)?.id || environments[0]?.id || '',
              permissions: ['read'],
              expires_in_days: 365,
            });
          }}>
            {generatedKey ? 'Close' : 'Cancel'}
          </Button>
          {!generatedKey && (
            <Button
              variant="contained"
              onClick={handleCreateApiKey}
              disabled={!formData.name.trim() || formData.permissions.length === 0}
            >
              Generate Key
            </Button>
          )}
        </DialogActions>
      </Dialog>

      {/* Test API Key Dialog */}
      <Dialog open={showTestDialog} onClose={() => setShowTestDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>Test API Key: {selectedKeyForTest?.name}</DialogTitle>
        <DialogContent>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" gutterBottom>
                Test Request
              </Typography>
              <TextField
                fullWidth
                label="Flag Key"
                value={testRequest.flag_key}
                onChange={(e) => setTestRequest(prev => ({ ...prev, flag_key: e.target.value }))}
                sx={{ mb: 2 }}
              />
              <TextField
                fullWidth
                label="User ID"
                value={testRequest.user_id}
                onChange={(e) => setTestRequest(prev => ({ ...prev, user_id: e.target.value }))}
                sx={{ mb: 2 }}
              />
              <TextField
                fullWidth
                label="Attributes (JSON)"
                value={testRequest.attributes}
                onChange={(e) => setTestRequest(prev => ({ ...prev, attributes: e.target.value }))}
                multiline
                rows={3}
                sx={{ mb: 2 }}
              />
              <Button
                variant="contained"
                fullWidth
                onClick={handleTestApiKey}
                startIcon={<TestIcon />}
              >
                Test Evaluation
              </Button>
            </Grid>
            
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" gutterBottom>
                Response
              </Typography>
              {testResult ? (
                <Box 
                  component="pre" 
                  sx={{ 
                    bgcolor: testResult.error ? 'error.light' : 'success.light',
                    p: 2, 
                    borderRadius: 1, 
                    fontSize: '0.875rem',
                    overflow: 'auto',
                    whiteSpace: 'pre-wrap',
                    maxHeight: '300px',
                  }}
                >
                  {JSON.stringify(testResult, null, 2)}
                </Box>
              ) : (
                <Paper sx={{ p: 3, textAlign: 'center', bgcolor: 'grey.50', minHeight: '200px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Typography color="text.secondary">
                    Run a test to see the API response
                  </Typography>
                </Paper>
              )}
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowTestDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}