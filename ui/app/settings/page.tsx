'use client';

import { useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  TextField,
  Switch,
  FormControlLabel,
  Button,
  Divider,
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
  Alert,
} from '@mui/material';
import {
  Settings as SettingsIcon,
  Security as SecurityIcon,
  Notifications as NotificationsIcon,
  Storage as StorageIcon,
  Key as KeyIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
} from '@mui/icons-material';

export default function SettingsPage() {
  const [settings, setSettings] = useState({
    // General Settings
    projectName: 'FlexFlag',
    projectDescription: 'Enterprise Feature Flag Management',
    defaultEnvironment: 'production',
    
    // Security Settings
    sessionTimeout: 24,
    twoFactorEnabled: false,
    passwordExpiry: 90,
    
    // Notification Settings
    emailNotifications: true,
    slackNotifications: false,
    webhookNotifications: true,
    
    // Performance Settings
    cacheEnabled: true,
    cacheTTL: 300,
    enableMetrics: true,
  });

  const [apiKeys, setApiKeys] = useState([
    { id: '1', name: 'Production API Key', key: 'ff_prod_****', fullKey: 'ff_prod_1234567890abcdef', created: '2025-01-10', lastUsed: '2025-01-10' },
    { id: '2', name: 'Staging API Key', key: 'ff_staging_****', fullKey: 'ff_staging_9876543210fedcba', created: '2025-01-09', lastUsed: '2025-01-10' },
  ]);

  const [showApiKeyDialog, setShowApiKeyDialog] = useState(false);
  const [showKeyValue, setShowKeyValue] = useState<string | null>(null);
  const [newApiKeyName, setNewApiKeyName] = useState('');
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);

  const handleSettingChange = (key: string, value: any) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleSaveSettings = () => {
    // In a real app, this would save to the backend
    console.log('Saving settings:', settings);
    // Show success message
  };

  const handleCreateApiKey = () => {
    if (!newApiKeyName.trim()) {
      alert('Please enter a name for the API key');
      return;
    }
    
    const newKey = {
      id: Date.now().toString(),
      name: newApiKeyName,
      key: 'ff_' + Math.random().toString(36).substring(2, 15) + '****',
      fullKey: 'ff_' + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
      created: new Date().toISOString().split('T')[0],
      lastUsed: 'Never'
    };
    
    setApiKeys([...apiKeys, newKey]);
    setGeneratedKey(newKey.fullKey);
    setNewApiKeyName('');
  };
  
  const handleDeleteApiKey = (keyId: string) => {
    if (window.confirm('Are you sure you want to delete this API key? This action cannot be undone.')) {
      setApiKeys(apiKeys.filter(key => key.id !== keyId));
    }
  };

  return (
    <Box>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" fontWeight="bold" gutterBottom>
          System Settings
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Configure your FlexFlag installation and manage system preferences
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {/* General Settings */}
        <Grid item xs={12} lg={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
                <SettingsIcon color="primary" />
                <Typography variant="h6" fontWeight="bold">
                  General Settings
                </Typography>
              </Box>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Project Name"
                    value={settings.projectName}
                    onChange={(e) => handleSettingChange('projectName', e.target.value)}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    rows={2}
                    label="Project Description"
                    value={settings.projectDescription}
                    onChange={(e) => handleSettingChange('projectDescription', e.target.value)}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Default Environment"
                    value={settings.defaultEnvironment}
                    onChange={(e) => handleSettingChange('defaultEnvironment', e.target.value)}
                    helperText="Default environment for new flags"
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Security Settings */}
        <Grid item xs={12} lg={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
                <SecurityIcon color="primary" />
                <Typography variant="h6" fontWeight="bold">
                  Security Settings
                </Typography>
              </Box>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    type="number"
                    label="Session Timeout (hours)"
                    value={settings.sessionTimeout}
                    onChange={(e) => handleSettingChange('sessionTimeout', parseInt(e.target.value))}
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.twoFactorEnabled}
                        onChange={(e) => handleSettingChange('twoFactorEnabled', e.target.checked)}
                      />
                    }
                    label="Require Two-Factor Authentication"
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    type="number"
                    label="Password Expiry (days)"
                    value={settings.passwordExpiry}
                    onChange={(e) => handleSettingChange('passwordExpiry', parseInt(e.target.value))}
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* API Keys Management */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <KeyIcon color="primary" />
                  <Typography variant="h6" fontWeight="bold">
                    API Keys
                  </Typography>
                </Box>
                <Button
                  variant="outlined"
                  startIcon={<AddIcon />}
                  onClick={() => setShowApiKeyDialog(true)}
                >
                  Generate API Key
                </Button>
              </Box>

              <List>
                {apiKeys.map((key, index) => (
                  <div key={key.id}>
                    <ListItem>
                      <ListItemText
                        primary={key.name}
                        secondary={
                          <Box>
                            <Typography variant="body2" color="text.secondary">
                              {showKeyValue === key.id ? key.fullKey : key.key}
                            </Typography>
                            <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                              <Chip label={`Created: ${key.created}`} size="small" variant="outlined" />
                              <Chip label={`Last used: ${key.lastUsed}`} size="small" variant="outlined" />
                            </Box>
                          </Box>
                        }
                      />
                      <ListItemSecondaryAction>
                        <IconButton
                          onClick={() => setShowKeyValue(showKeyValue === key.id ? null : key.id)}
                          size="small"
                        >
                          {showKeyValue === key.id ? <VisibilityOffIcon /> : <VisibilityIcon />}
                        </IconButton>
                        <IconButton 
                          color="error" 
                          size="small"
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
            </CardContent>
          </Card>
        </Grid>

        {/* Notification Settings */}
        <Grid item xs={12} lg={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
                <NotificationsIcon color="primary" />
                <Typography variant="h6" fontWeight="bold">
                  Notifications
                </Typography>
              </Box>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.emailNotifications}
                        onChange={(e) => handleSettingChange('emailNotifications', e.target.checked)}
                      />
                    }
                    label="Email Notifications"
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.slackNotifications}
                        onChange={(e) => handleSettingChange('slackNotifications', e.target.checked)}
                      />
                    }
                    label="Slack Integration"
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.webhookNotifications}
                        onChange={(e) => handleSettingChange('webhookNotifications', e.target.checked)}
                      />
                    }
                    label="Webhook Notifications"
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Performance Settings */}
        <Grid item xs={12} lg={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
                <StorageIcon color="primary" />
                <Typography variant="h6" fontWeight="bold">
                  Performance
                </Typography>
              </Box>
              
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.cacheEnabled}
                        onChange={(e) => handleSettingChange('cacheEnabled', e.target.checked)}
                      />
                    }
                    label="Enable Caching"
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    type="number"
                    label="Cache TTL (seconds)"
                    value={settings.cacheTTL}
                    onChange={(e) => handleSettingChange('cacheTTL', parseInt(e.target.value))}
                    disabled={!settings.cacheEnabled}
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.enableMetrics}
                        onChange={(e) => handleSettingChange('enableMetrics', e.target.checked)}
                      />
                    }
                    label="Enable Performance Metrics"
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Save Button */}
      <Box sx={{ mt: 4, display: 'flex', justifyContent: 'flex-end' }}>
        <Button
          variant="contained"
          size="large"
          onClick={handleSaveSettings}
          startIcon={<SettingsIcon />}
        >
          Save Settings
        </Button>
      </Box>

      {/* Generate API Key Dialog */}
      <Dialog open={showApiKeyDialog} onClose={() => setShowApiKeyDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Generate New API Key</DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 2 }}>
            API keys provide programmatic access to FlexFlag. Keep them secure and rotate regularly.
          </Alert>
          <TextField
            fullWidth
            label="API Key Name"
            placeholder="e.g., Production Mobile App"
            value={newApiKeyName}
            onChange={(e) => setNewApiKeyName(e.target.value)}
            sx={{ mt: 1 }}
          />
          
          {generatedKey && (
            <Box sx={{ mt: 2, p: 2, bgcolor: 'success.light', borderRadius: 1 }}>
              <Typography variant="subtitle2" color="success.dark" gutterBottom>
                API Key Generated Successfully!
              </Typography>
              <Typography variant="body2" color="success.dark" sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
                {generatedKey}
              </Typography>
              <Typography variant="caption" color="success.dark">
                Make sure to copy this key - you won't be able to see it again!
              </Typography>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => {
            setShowApiKeyDialog(false);
            setNewApiKeyName('');
            setGeneratedKey(null);
          }}>Cancel</Button>
          <Button 
            variant="contained" 
            onClick={handleCreateApiKey}
            disabled={!newApiKeyName.trim()}
          >
            Generate Key
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}