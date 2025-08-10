'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Box, CircularProgress } from '@mui/material';

const publicRoutes = ['/login', '/register'];

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [isClient, setIsClient] = useState(false);
  
  useEffect(() => {
    setIsClient(true);
    
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      const isPublicRoute = publicRoutes.includes(pathname);
      
      if (!token && !isPublicRoute) {
        // Not authenticated and trying to access protected route
        router.push('/login');
      } else if (token && isPublicRoute) {
        // Already authenticated and trying to access login/register
        router.push('/');
      }
    }
  }, [pathname, router]);
  
  // Don't render anything until we're on the client
  if (!isClient) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }
  
  // Show loading while checking auth
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  const isPublicRoute = publicRoutes.includes(pathname);
  
  if (!token && !isPublicRoute) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }
  
  return <>{children}</>;
}