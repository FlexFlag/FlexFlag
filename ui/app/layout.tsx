'use client';

import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { AppBar, Toolbar, Typography, Box, Drawer, List, ListItem, ListItemIcon, ListItemText, Container } from '@mui/material';
import { useState } from 'react';
import {
  Dashboard as DashboardIcon,
  Flag as FlagIcon,
  Assessment as AssessmentIcon,
  Speed as SpeedIcon,
  Settings as SettingsIcon,
  Menu as MenuIcon,
} from '@mui/icons-material';
import { theme } from '@/lib/theme';
import { IconButton } from '@mui/material';

const drawerWidth = 280;

const navigationItems = [
  { label: 'Dashboard', icon: <DashboardIcon />, href: '/' },
  { label: 'Feature Flags', icon: <FlagIcon />, href: '/flags' },
  { label: 'Evaluations', icon: <AssessmentIcon />, href: '/evaluations' },
  { label: 'Performance', icon: <SpeedIcon />, href: '/performance' },
  { label: 'Settings', icon: <SettingsIcon />, href: '/settings' },
];

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const drawer = (
    <Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          p: 3,
          borderBottom: 1,
          borderColor: 'divider',
        }}
      >
        <Typography variant="h5" color="primary" fontWeight="bold">
          ⚡ FlexFlag
        </Typography>
      </Box>
      <List sx={{ px: 2, pt: 2 }}>
        {navigationItems.map((item) => (
          <ListItem
            key={item.label}
            component="a"
            href={item.href}
            sx={{
              borderRadius: 2,
              mb: 1,
              transition: 'all 0.2s ease',
              '&:hover': {
                bgcolor: 'primary.50',
                color: 'primary.main',
              },
              textDecoration: 'none',
              color: 'text.primary',
            }}
          >
            <ListItemIcon sx={{ color: 'inherit', minWidth: 40 }}>
              {item.icon}
            </ListItemIcon>
            <ListItemText
              primary={item.label}
              primaryTypographyProps={{
                fontWeight: 500,
                fontSize: '0.95rem',
              }}
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <html lang="en">
      <body>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <Box sx={{ display: 'flex', minHeight: '100vh' }}>
            {/* App Bar */}
            <AppBar
              position="fixed"
              sx={{
                width: { sm: `calc(100% - ${drawerWidth}px)` },
                ml: { sm: `${drawerWidth}px` },
                bgcolor: 'background.paper',
                color: 'text.primary',
                borderBottom: 1,
                borderColor: 'divider',
                boxShadow: 'none',
              }}
            >
              <Toolbar>
                <IconButton
                  color="inherit"
                  aria-label="open drawer"
                  edge="start"
                  onClick={handleDrawerToggle}
                  sx={{ mr: 2, display: { sm: 'none' } }}
                >
                  <MenuIcon />
                </IconButton>
                <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
                  Feature Flag Management
                </Typography>
              </Toolbar>
            </AppBar>

            {/* Navigation Drawer */}
            <Box
              component="nav"
              sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
              aria-label="navigation"
            >
              {/* Mobile drawer */}
              <Drawer
                variant="temporary"
                open={mobileOpen}
                onClose={handleDrawerToggle}
                ModalProps={{
                  keepMounted: true, // Better open performance on mobile.
                }}
                sx={{
                  display: { xs: 'block', sm: 'none' },
                  '& .MuiDrawer-paper': {
                    boxSizing: 'border-box',
                    width: drawerWidth,
                    bgcolor: 'background.paper',
                    borderRight: 1,
                    borderColor: 'divider',
                  },
                }}
              >
                {drawer}
              </Drawer>
              {/* Desktop drawer */}
              <Drawer
                variant="permanent"
                sx={{
                  display: { xs: 'none', sm: 'block' },
                  '& .MuiDrawer-paper': {
                    boxSizing: 'border-box',
                    width: drawerWidth,
                    bgcolor: 'background.paper',
                    borderRight: 1,
                    borderColor: 'divider',
                  },
                }}
                open
              >
                {drawer}
              </Drawer>
            </Box>

            {/* Main content */}
            <Box
              component="main"
              sx={{
                flexGrow: 1,
                width: { sm: `calc(100% - ${drawerWidth}px)` },
                bgcolor: 'background.default',
                minHeight: '100vh',
              }}
            >
              <Toolbar />
              <Container maxWidth="xl" sx={{ py: 4 }}>
                {children}
              </Container>
            </Box>
          </Box>
        </ThemeProvider>
      </body>
    </html>
  );
}