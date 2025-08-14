'use client';

import { CustomThemeProvider } from '@/contexts/ThemeContext';
import { 
  AppBar, 
  Toolbar, 
  Typography, 
  Box, 
  Drawer, 
  List, 
  ListItem, 
  ListItemIcon, 
  ListItemText, 
  Container,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
  Divider,
  Chip,
  ListItemButton,
  Collapse,
} from '@mui/material';
import { useState, useEffect } from 'react';
import {
  Dashboard as DashboardIcon,
  Flag as FlagIcon,
  Assessment as AssessmentIcon,
  Speed as SpeedIcon,
  Settings as SettingsIcon,
  Menu as MenuIcon,
  AccountTree as ProjectIcon,
  Group as GroupIcon,
  Science as ExperimentIcon,
  DonutLarge as RolloutIcon,
  Segment as SegmentIcon,
  Security as SecurityIcon,
  ExpandLess,
  ExpandMore,
  Logout as LogoutIcon,
  Person as PersonIcon,
  AdminPanelSettings as AdminIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
} from '@mui/icons-material';
import { AuthProvider } from '@/contexts/AuthContext';
import { ProjectProvider } from '@/contexts/ProjectContext';
import { EnvironmentProvider, useEnvironment } from '@/contexts/EnvironmentContext';
import { useTheme as useCustomTheme } from '@/contexts/ThemeContext';
import AuthGuard from '@/components/AuthGuard';
import { usePathname } from 'next/navigation';

const drawerWidth = 280;

const navigationItems = [
  { label: 'Dashboard', icon: <DashboardIcon />, href: '/' },
  { label: 'Projects', icon: <ProjectIcon />, href: '/projects' },
  {
    label: 'Administration',
    icon: <AdminIcon />,
    children: [
      { label: 'Users', icon: <GroupIcon />, href: '/users' },
      { label: 'Audit Logs', icon: <SecurityIcon />, href: '/audit' },
      { label: 'Settings', icon: <SettingsIcon />, href: '/settings' },
    ]
  },
];


function NavigationContent() {
  const pathname = usePathname();
  const [expandedItems, setExpandedItems] = useState<string[]>(['Administration']);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleExpandClick = (label: string) => {
    setExpandedItems(prev =>
      prev.includes(label)
        ? prev.filter(item => item !== label)
        : [...prev, label]
    );
  };

  const handleUserMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleUserMenuClose = () => {
    setAnchorEl(null);
  };

  return (
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
          <Box key={item.label}>
            {item.children ? (
              <>
                <ListItemButton
                  onClick={() => handleExpandClick(item.label)}
                  sx={{
                    borderRadius: 2,
                    mb: 0.5,
                  }}
                >
                  <ListItemIcon sx={{ minWidth: 40 }}>
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{
                      fontWeight: 500,
                      fontSize: '0.95rem',
                    }}
                  />
                  {expandedItems.includes(item.label) ? <ExpandLess /> : <ExpandMore />}
                </ListItemButton>
                <Collapse in={expandedItems.includes(item.label)} timeout="auto" unmountOnExit>
                  <List component="div" disablePadding>
                    {item.children.map((child) => (
                      <ListItem
                        key={child.label}
                        component="a"
                        href={child.href}
                        sx={{
                          pl: 4,
                          borderRadius: 2,
                          mb: 0.5,
                          transition: 'all 0.2s ease',
                          bgcolor: pathname === child.href ? 'primary.50' : 'transparent',
                          color: pathname === child.href ? 'primary.main' : 'text.primary',
                          '&:hover': {
                            bgcolor: 'primary.50',
                            color: 'primary.main',
                          },
                          textDecoration: 'none',
                        }}
                      >
                        <ListItemIcon sx={{ color: 'inherit', minWidth: 36 }}>
                          {child.icon}
                        </ListItemIcon>
                        <ListItemText
                          primary={child.label}
                          primaryTypographyProps={{
                            fontWeight: pathname === child.href ? 600 : 400,
                            fontSize: '0.9rem',
                          }}
                        />
                      </ListItem>
                    ))}
                  </List>
                </Collapse>
              </>
            ) : (
              <ListItem
                component="a"
                href={item.href}
                sx={{
                  borderRadius: 2,
                  mb: 1,
                  transition: 'all 0.2s ease',
                  bgcolor: pathname === item.href ? 'primary.50' : 'transparent',
                  color: pathname === item.href ? 'primary.main' : 'text.primary',
                  '&:hover': {
                    bgcolor: 'primary.50',
                    color: 'primary.main',
                  },
                  textDecoration: 'none',
                }}
              >
                <ListItemIcon sx={{ color: 'inherit', minWidth: 40 }}>
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{
                    fontWeight: pathname === item.href ? 600 : 500,
                    fontSize: '0.95rem',
                  }}
                />
              </ListItem>
            )}
          </Box>
        ))}
      </List>
    </Box>
  );
}

function LayoutContent({ children }: { children: React.ReactNode }) {
  const { mode, toggleMode } = useCustomTheme();
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [userMenuAnchor, setUserMenuAnchor] = useState<null | HTMLElement>(null);
  const [user, setUser] = useState<any>(null);

  // Check if we're in a project-specific route
  const isProjectRoute = pathname?.startsWith('/projects/') && pathname.split('/').length > 2;

  useEffect(() => {
    // Load user data from localStorage
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      try {
        setUser(JSON.parse(storedUser));
      } catch (e) {
        console.error('Error parsing user data:', e);
      }
    }
  }, []);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  // If we're in a project route, just render the children (project layout will handle everything)
  if (isProjectRoute) {
    return <>{children}</>;
  }

  // For non-project routes, render the full layout
  return (
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
            FlexFlag Dashboard
          </Typography>
          
          {/* Dark Mode Toggle */}
          <IconButton
            onClick={toggleMode}
            sx={{ 
              color: 'text.secondary',
              mr: 2,
              '&:hover': {
                color: 'primary.main',
              }
            }}
          >
            {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
          </IconButton>
          
          {/* User Menu */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            {user && (
              <Chip
                label={user.role?.charAt(0).toUpperCase() + user.role?.slice(1)}
                size="small"
                color={user.role === 'admin' ? 'primary' : user.role === 'editor' ? 'secondary' : 'default'}
                variant="outlined"
              />
            )}
            <IconButton
              onClick={(e) => setUserMenuAnchor(e.currentTarget)}
              sx={{ p: 0 }}
            >
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main' }}>
                {user?.full_name?.charAt(0)?.toUpperCase() || 'U'}
              </Avatar>
            </IconButton>
          </Box>
          
          <Menu
            anchorEl={userMenuAnchor}
            open={Boolean(userMenuAnchor)}
            onClose={() => setUserMenuAnchor(null)}
          >
            <MenuItem onClick={() => {
              setUserMenuAnchor(null);
              window.location.href = '/profile';
            }}>
              <ListItemIcon>
                <PersonIcon fontSize="small" />
              </ListItemIcon>
              Profile
            </MenuItem>
            <MenuItem onClick={() => {
              setUserMenuAnchor(null);
              window.location.href = '/settings';
            }}>
              <ListItemIcon>
                <SettingsIcon fontSize="small" />
              </ListItemIcon>
              Settings
            </MenuItem>
            <Divider />
            <MenuItem onClick={() => {
              setUserMenuAnchor(null);
              localStorage.removeItem('token');
              localStorage.removeItem('user');
              window.location.href = '/login';
            }}>
              <ListItemIcon>
                <LogoutIcon fontSize="small" />
              </ListItemIcon>
              Logout
            </MenuItem>
          </Menu>
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
            keepMounted: true,
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
          <NavigationContent />
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
          <NavigationContent />
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
  );
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <CustomThemeProvider>
          <AuthProvider>
            <ProjectProvider>
              <EnvironmentProvider>
                <AuthGuard>
                  <LayoutContent>{children}</LayoutContent>
                </AuthGuard>
              </EnvironmentProvider>
            </ProjectProvider>
          </AuthProvider>
        </CustomThemeProvider>
      </body>
    </html>
  );
}