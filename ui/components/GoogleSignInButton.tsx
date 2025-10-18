'use client';

import { Button } from '@mui/material';
import { Google as GoogleIcon } from '@mui/icons-material';

interface GoogleSignInButtonProps {
  onSuccess?: (token: string, user: any) => void;
  onError?: (error: string) => void;
}

export default function GoogleSignInButton({ onSuccess, onError }: GoogleSignInButtonProps) {
  const handleGoogleSignIn = () => {
    // Redirect to backend OAuth endpoint
    window.location.href = 'http://localhost:8080/api/v1/auth/google/login';
  };

  return (
    <Button
      variant="outlined"
      fullWidth
      startIcon={<GoogleIcon />}
      onClick={handleGoogleSignIn}
      sx={{
        borderColor: '#4285f4',
        color: '#4285f4',
        '&:hover': {
          borderColor: '#357ae8',
          backgroundColor: 'rgba(66, 133, 244, 0.04)',
        },
      }}
    >
      Sign in with Google
    </Button>
  );
}
