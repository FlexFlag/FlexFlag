'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Box, CircularProgress, Typography, Alert } from '@mui/material';

function GoogleCallbackContent() {
  console.log('GoogleCallbackContent component mounted!');
  const router = useRouter();
  const [error, setError] = useState<string>('');

  console.log('About to set up useEffect...');

  useEffect(() => {
    console.log('useEffect running!');
    console.log('window.location.href:', window.location.href);
    const handleCallback = async () => {
      // Get token from URL parameter directly from window.location
      const urlParams = new URLSearchParams(window.location.search);
      const token = urlParams.get('token');
      const errorParam = urlParams.get('error');

      console.log('OAuth Callback - URL params:', { token: token ? 'present' : 'missing', errorParam });
      console.log('Full URL:', window.location.href);
      console.log('Token value:', token);

      if (errorParam) {
        console.error('OAuth error parameter:', errorParam);
        setError(`Authentication failed: ${errorParam}`);
        setTimeout(() => router.push('/login'), 3000);
        return;
      }

      if (!token) {
        console.error('No token found in URL');
        setError('No authentication token received');
        setTimeout(() => router.push('/login'), 3000);
        return;
      }

      try {
        console.log('Storing token in localStorage...');
        // Store the token
        localStorage.setItem('token', token);
        console.log('Token stored successfully');

        console.log('Fetching user profile...');
        // Fetch user profile with the token
        const response = await fetch('http://localhost:8080/api/v1/auth/profile', {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });

        console.log('Profile response status:', response.status);

        if (response.ok) {
          const user = await response.json();
          console.log('User profile retrieved:', user.email);
          localStorage.setItem('user', JSON.stringify(user));

          console.log('Redirecting to dashboard...');
          // Redirect to dashboard and force reload to update AuthContext
          window.location.href = '/';
        } else {
          const errorText = await response.text();
          console.error('Profile fetch failed:', response.status, errorText);
          throw new Error(`Failed to fetch profile: ${response.status}`);
        }
      } catch (err) {
        console.error('OAuth callback error:', err);
        setError('Authentication failed. Please try again.');
        setTimeout(() => router.push('/login'), 3000);
      }
    };

    handleCallback();
  }, [router]);

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Box sx={{ textAlign: 'center', p: 4, bgcolor: 'white', borderRadius: 2, maxWidth: 400 }}>
        {error ? (
          <>
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
            <Typography variant="body2" color="text.secondary">
              Redirecting to login...
            </Typography>
          </>
        ) : (
          <>
            <CircularProgress size={60} sx={{ mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Completing sign in...
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Please wait while we authenticate your account
            </Typography>
          </>
        )}
      </Box>
    </Box>
  );
}

export default function GoogleCallbackPage() {
  return (
    <Suspense fallback={
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        }}
      >
        <Box sx={{ textAlign: 'center', p: 4, bgcolor: 'white', borderRadius: 2, maxWidth: 400 }}>
          <CircularProgress size={60} sx={{ mb: 2 }} />
          <Typography variant="h6" gutterBottom>
            Loading...
          </Typography>
        </Box>
      </Box>
    }>
      <GoogleCallbackContent />
    </Suspense>
  );
}
