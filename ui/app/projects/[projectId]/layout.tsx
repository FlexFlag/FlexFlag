'use client';

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
  Breadcrumbs,
  Link,
  Tooltip,
} from '@mui/material';
import { useState, useEffect } from 'react';
import {
  Flag as FlagIcon,
  Assessment as AssessmentIcon,
  Speed as SpeedIcon,
  Settings as SettingsIcon,
  Menu as MenuIcon,
  AccountTree as ProjectIcon,
  Science as ExperimentIcon,
  DonutLarge as RolloutIcon,
  Segment as SegmentIcon,
  ExpandLess,
  ExpandMore,
  Logout as LogoutIcon,
  Person as PersonIcon,
  Home as HomeIcon,
  ArrowBack as ArrowBackIcon,
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
  Key as KeyIcon,
  Cloud as EnvironmentIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
} from '@mui/icons-material';
import { useTheme as useCustomTheme } from '@/contexts/ThemeContext';
import { useTheme } from '@mui/material/styles';
import { useProject } from '@/contexts/ProjectContext';
import { useEnvironment } from '@/contexts/EnvironmentContext';
import { usePathname, useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';

const drawerWidth = 280;
const collapsedDrawerWidth = 64;

const projectNavigationItems = [
  { 
    label: 'Feature Management', 
    icon: <FlagIcon />, 
    children: [
      { label: 'Feature Flags', icon: <FlagIcon />, href: '/flags' },
      { label: 'Segments', icon: <SegmentIcon />, href: '/segments' },
      { label: 'Rollouts', icon: <RolloutIcon />, href: '/rollouts' },
      { label: 'Experiments', icon: <ExperimentIcon />, href: '/experiments' },
    ]
  },
  { label: 'Evaluations', icon: <AssessmentIcon />, href: '/evaluations' },
  { label: 'Performance', icon: <SpeedIcon />, href: '/performance' },
  { label: 'Environments', icon: <EnvironmentIcon />, href: '/environments' },
  { label: 'API Keys', icon: <KeyIcon />, href: '/api-keys' },
];

function EnvironmentSelector() {
  const { currentEnvironment, setCurrentEnvironment, availableEnvironments, environments, loading } = useEnvironment();

  const handleEnvironmentChange = (environment: string) => {
    setCurrentEnvironment(environment);
  };

  const getEnvironmentDisplayName = (envKey: string) => {
    const env = environments.find(e => e.key === envKey);
    if (env) return env.name;
    
    // Fallback to key with title case
    return envKey.charAt(0).toUpperCase() + envKey.slice(1);
  };

  if (loading) {
    return (
      <Box sx={{ px: 3, py: 2, borderBottom: 1, borderColor: 'divider', bgcolor: 'grey.50' }}>
        <Typography 
          variant="overline" 
          color="text.secondary" 
          sx={{ 
            mb: 1.5, 
            display: 'block',
            fontWeight: 600,
            fontSize: '0.75rem',
            letterSpacing: '0.5px'
          }}
        >
          Environment
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Loading environments...
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ px: 3, py: 2, borderBottom: 1, borderColor: 'divider', bgcolor: 'grey.50' }}>
      <Typography 
        variant="overline" 
        color="text.secondary" 
        sx={{ 
          mb: 1.5, 
          display: 'block',
          fontWeight: 600,
          fontSize: '0.75rem',
          letterSpacing: '0.5px'
        }}
      >
        Environment
      </Typography>
      <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
        {availableEnvironments.map((env) => (
          <Chip
            key={env}
            label={getEnvironmentDisplayName(env)}
            size="small"
            color={currentEnvironment === env ? "primary" : "default"}
            variant={currentEnvironment === env ? "filled" : "outlined"}
            onClick={() => handleEnvironmentChange(env)}
            sx={{ 
              cursor: 'pointer',
              fontSize: '0.75rem',
              height: 26
            }}
          />
        ))}
      </Box>
    </Box>
  );
}

function NavigationContent({ project, collapsed, onToggleCollapse }: { project: any; collapsed: boolean; onToggleCollapse: () => void }) {
  const pathname = usePathname();
  const [expandedItems, setExpandedItems] = useState<string[]>(['Feature Management']);

  const handleExpandClick = (label: string) => {
    setExpandedItems(prev =>
      prev.includes(label)
        ? prev.filter(item => item !== label)
        : [...prev, label]
    );
  };

  const getFullHref = (href: string) => `/projects/${project?.id}${href}`;
  const isActive = (href: string) => pathname === getFullHref(href);
  const isParentActive = (children: any[]) => 
    children.some(child => pathname === getFullHref(child.href));

  return (
    <Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'space-between',
          p: collapsed ? 2 : 3,
          borderBottom: 1,
          borderColor: 'divider',
          minHeight: 64,
        }}
      >
        {!collapsed && (
          <Typography variant="h5" color="primary" fontWeight="bold">
            ⚡ FlexFlag
          </Typography>
        )}
        {collapsed && (
          <Typography variant="h6" color="primary" fontWeight="bold">
            ⚡
          </Typography>
        )}
        <IconButton
          onClick={onToggleCollapse}
          size="small"
          sx={{ 
            ml: collapsed ? 0 : 'auto',
            color: 'text.secondary',
          }}
        >
          {collapsed ? <ChevronRightIcon /> : <ChevronLeftIcon />}
        </IconButton>
      </Box>
      
      {/* Project Header */}
      {!collapsed && (
        <Box sx={{ px: 3, py: 2.5, borderBottom: 1, borderColor: 'divider' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <Avatar sx={{ bgcolor: 'primary.main', width: 36, height: 36 }}>
              <ProjectIcon sx={{ fontSize: 18 }} />
            </Avatar>
            <Box sx={{ overflow: 'hidden' }}>
              <Typography 
                variant="subtitle1" 
                fontWeight="600"
                sx={{ 
                  fontSize: '0.95rem',
                  lineHeight: 1.3,
                  mb: 0.25
                }}
              >
                {project?.name || 'Loading...'}
              </Typography>
              <Typography 
                variant="caption" 
                color="text.secondary"
                sx={{
                  fontSize: '0.7rem',
                  fontWeight: 500
                }}
              >
                {project?.slug}
              </Typography>
            </Box>
          </Box>
        </Box>
      )}
      
      {collapsed && (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 2.5, borderBottom: 1, borderColor: 'divider' }}>
          <Avatar sx={{ bgcolor: 'primary.main', width: 32, height: 32 }}>
            <ProjectIcon sx={{ fontSize: 16 }} />
          </Avatar>
        </Box>
      )}

      {/* Environment Selector */}
      {!collapsed && <EnvironmentSelector />}

      <List sx={{ px: collapsed ? 1 : 2, pt: 2 }}>
        {projectNavigationItems.map((item) => (
          <Box key={item.label}>
            {item.children ? (
              <>
                {!collapsed ? (
                  <>
                    <ListItemButton
                      onClick={() => handleExpandClick(item.label)}
                      sx={{
                        borderRadius: 2,
                        mb: 1,
                        py: 1.25,
                        bgcolor: isParentActive(item.children) ? 'primary.50' : 'transparent',
                        color: isParentActive(item.children) ? 'primary.main' : 'text.primary',
                        justifyContent: 'flex-start',
                        px: 2,
                        '&:hover': {
                          bgcolor: 'primary.50',
                          color: 'primary.main',
                        },
                        transition: 'all 0.2s ease',
                      }}
                    >
                      <ListItemIcon sx={{ 
                        minWidth: 40, 
                        color: 'inherit'
                      }}>
                        {item.icon}
                      </ListItemIcon>
                      <ListItemText
                        primary={item.label}
                        primaryTypographyProps={{
                          fontWeight: isParentActive(item.children) ? 600 : 500,
                          fontSize: '0.875rem',
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
                            href={getFullHref(child.href)}
                            sx={{
                              pl: 4,
                              pr: 2,
                              py: 1,
                              borderRadius: 2,
                              mb: 0.5,
                              transition: 'all 0.2s ease',
                              bgcolor: isActive(child.href) ? 'primary.50' : 'transparent',
                              color: isActive(child.href) ? 'primary.main' : 'text.primary',
                              '&:hover': {
                                bgcolor: 'primary.50',
                                color: 'primary.main',
                              },
                              textDecoration: 'none',
                            }}
                          >
                            <ListItemIcon sx={{ 
                              color: 'inherit', 
                              minWidth: 36
                            }}>
                              {child.icon}
                            </ListItemIcon>
                            <ListItemText
                              primary={child.label}
                              primaryTypographyProps={{
                                fontWeight: isActive(child.href) ? 600 : 400,
                                fontSize: '0.9rem',
                              }}
                            />
                          </ListItem>
                        ))}
                      </List>
                    </Collapse>
                  </>
                ) : (
                  // When collapsed, show child items as individual icons
                  <>
                    {item.children.map((child) => (
                      <Tooltip key={child.label} title={child.label} placement="right">
                        <ListItem
                          component="a"
                          href={getFullHref(child.href)}
                          sx={{
                            borderRadius: 0,
                            my: 0.5,
                            py: 1.5,
                            transition: 'all 0.2s ease',
                            bgcolor: isActive(child.href) ? 'primary.50' : 'transparent',
                            color: isActive(child.href) ? 'primary.main' : 'text.primary',
                            '&:hover': {
                              bgcolor: 'primary.50',
                              color: 'primary.main',
                            },
                            textDecoration: 'none',
                            justifyContent: 'center',
                            px: 1,
                          }}
                        >
                          <ListItemIcon sx={{ 
                            color: 'inherit', 
                            minWidth: 'auto',
                            '& > *': {
                              fontSize: '1.25rem'
                            }
                          }}>
                            {child.icon}
                          </ListItemIcon>
                        </ListItem>
                      </Tooltip>
                    ))}
                  </>
                )}
              </>
            ) : (
              collapsed ? (
                <Tooltip title={item.label} placement="right">
                  <ListItem
                    component="a"
                    href={getFullHref(item.href)}
                    sx={{
                      borderRadius: 0,
                      my: 0.5,
                      py: 1.5,
                      transition: 'all 0.2s ease',
                      bgcolor: isActive(item.href) ? 'primary.50' : 'transparent',
                      color: isActive(item.href) ? 'primary.main' : 'text.primary',
                      '&:hover': {
                        bgcolor: 'primary.50',
                        color: 'primary.main',
                      },
                      textDecoration: 'none',
                      justifyContent: 'center',
                      px: 1,
                    }}
                  >
                    <ListItemIcon sx={{ 
                      color: 'inherit', 
                      minWidth: 'auto',
                      '& > *': {
                        fontSize: '1.25rem'
                      }
                    }}>
                      {item.icon}
                    </ListItemIcon>
                  </ListItem>
                </Tooltip>
              ) : (
                <ListItem
                  component="a"
                  href={getFullHref(item.href)}
                  sx={{
                    borderRadius: 2,
                    mb: 1,
                    py: 1.25,
                    transition: 'all 0.2s ease',
                    bgcolor: isActive(item.href) ? 'primary.50' : 'transparent',
                    color: isActive(item.href) ? 'primary.main' : 'text.primary',
                    '&:hover': {
                      bgcolor: 'primary.50',
                      color: 'primary.main',
                    },
                    textDecoration: 'none',
                    justifyContent: 'flex-start',
                    px: 2,
                  }}
                >
                  <ListItemIcon sx={{ 
                    color: 'inherit', 
                    minWidth: 40
                  }}>
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{
                      fontWeight: isActive(item.href) ? 600 : 500,
                      fontSize: '0.875rem',
                    }}
                  />
                </ListItem>
              )
            )}
          </Box>
        ))}
      </List>
    </Box>
  );
}

export default function ProjectLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { mode, toggleMode } = useCustomTheme();
  const theme = useTheme();
  const [mobileOpen, setMobileOpen] = useState(false);
  // Initialize sidebar collapsed state from localStorage
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('flexflag_sidebar_collapsed');
      return saved === 'true';
    }
    return false;
  });
  const [userMenuAnchor, setUserMenuAnchor] = useState<null | HTMLElement>(null);
  const [user, setUser] = useState<any>(null);
  const [project, setProject] = useState<any>(null);
  const params = useParams();
  const projectId = params.projectId as string;

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

  useEffect(() => {
    // Load project data
    const fetchProject = async () => {
      try {
        const projects = await apiClient.getProjects();
        const foundProject = projects.find(p => p.id === projectId);
        setProject(foundProject);
      } catch (error) {
        console.error('Error fetching project:', error);
      }
    };

    if (projectId) {
      fetchProject();
    }
  }, [projectId]);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleSidebarToggle = () => {
    const newCollapsedState = !sidebarCollapsed;
    setSidebarCollapsed(newCollapsedState);
    // Persist to localStorage
    if (typeof window !== 'undefined') {
      localStorage.setItem('flexflag_sidebar_collapsed', newCollapsedState.toString());
    }
  };

  const currentDrawerWidth = sidebarCollapsed ? collapsedDrawerWidth : drawerWidth;

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
        {/* App Bar */}
        <AppBar
          position="fixed"
          sx={{
            width: { sm: `calc(100% - ${currentDrawerWidth}px)` },
            ml: { sm: `${currentDrawerWidth}px` },
            bgcolor: 'background.paper',
            color: 'text.primary',
            borderBottom: 1,
            borderColor: 'divider',
            boxShadow: 'none',
            transition: theme.transitions.create(['margin', 'width'], {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.leavingScreen,
            }),
          }}
        >
          <Toolbar sx={{ minHeight: '64px !important', px: 3, py: 1.5 }}>
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2, display: { sm: 'none' } }}
            >
              <MenuIcon />
            </IconButton>
            
            {/* Breadcrumbs and Title */}
            <Box sx={{ flexGrow: 1, display: 'flex', alignItems: 'center' }}>
              <Breadcrumbs 
                aria-label="breadcrumb" 
                sx={{ 
                  '& .MuiBreadcrumbs-separator': {
                    mx: 1,
                    color: 'text.disabled'
                  },
                  '& .MuiBreadcrumbs-ol': {
                    alignItems: 'center'
                  }
                }}
              >
                <Link 
                  href="/" 
                  underline="hover" 
                  sx={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: 0.75,
                    color: 'text.secondary',
                    fontSize: '0.8rem',
                    fontWeight: 500,
                    py: 0.5,
                    px: 1,
                    borderRadius: 1,
                    transition: 'all 0.2s ease',
                    '&:hover': {
                      color: 'primary.main',
                      bgcolor: 'primary.50'
                    }
                  }}
                >
                  <HomeIcon fontSize="small" />
                  Dashboard
                </Link>
                <Link 
                  href="/projects" 
                  underline="hover"
                  sx={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: 0.75,
                    color: 'text.secondary',
                    fontSize: '0.8rem',
                    fontWeight: 500,
                    py: 0.5,
                    px: 1,
                    borderRadius: 1,
                    transition: 'all 0.2s ease',
                    '&:hover': {
                      color: 'primary.main',
                      bgcolor: 'primary.50'
                    }
                  }}
                >
                  <ProjectIcon fontSize="small" />
                  Projects
                </Link>
                <Typography 
                  color="text.primary" 
                  sx={{ 
                    fontWeight: 600,
                    fontSize: '0.8rem',
                    py: 0.5,
                    px: 1
                  }}
                >
                  {project?.name || 'Loading...'}
                </Typography>
              </Breadcrumbs>
            </Box>
            
            {/* Actions */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, ml: 3 }}>
              {/* Dark Mode Toggle */}
              <IconButton
                onClick={toggleMode}
                sx={{ 
                  color: 'text.secondary',
                  bgcolor: 'grey.50',
                  border: '1px solid',
                  borderColor: 'grey.200',
                  width: 36,
                  height: 36,
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    color: 'primary.main',
                    bgcolor: 'primary.50',
                    borderColor: 'primary.200'
                  }
                }}
                size="small"
              >
                {mode === 'dark' ? <LightModeIcon sx={{ fontSize: '1.1rem' }} /> : <DarkModeIcon sx={{ fontSize: '1.1rem' }} />}
              </IconButton>
              
              {/* Back Button */}
              <IconButton
                href="/projects"
                sx={{ 
                  color: 'text.secondary',
                  bgcolor: 'grey.50',
                  border: '1px solid',
                  borderColor: 'grey.200',
                  width: 36,
                  height: 36,
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    color: 'primary.main',
                    bgcolor: 'primary.50',
                    borderColor: 'primary.200'
                  }
                }}
                component="a"
                size="small"
              >
                <ArrowBackIcon sx={{ fontSize: '1.1rem' }} />
              </IconButton>
              
              {/* User Menu */}
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                {user && (
                  <Chip
                    label={user.role?.charAt(0).toUpperCase() + user.role?.slice(1)}
                    size="small"
                    color={user.role === 'admin' ? 'primary' : user.role === 'editor' ? 'secondary' : 'default'}
                    variant="outlined"
                    sx={{ fontWeight: 500 }}
                  />
                )}
                <IconButton
                  onClick={(e) => setUserMenuAnchor(e.currentTarget)}
                  sx={{ p: 0 }}
                >
                  <Avatar sx={{ width: 36, height: 36, bgcolor: 'primary.main', fontWeight: 600 }}>
                    {user?.full_name?.charAt(0)?.toUpperCase() || 'U'}
                  </Avatar>
                </IconButton>
              </Box>
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
          sx={{ 
            width: { sm: currentDrawerWidth }, 
            flexShrink: { sm: 0 },
            transition: theme.transitions.create('width', {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.leavingScreen,
            }),
          }}
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
                width: currentDrawerWidth,
                bgcolor: 'background.paper',
                borderRight: 1,
                borderColor: 'divider',
              },
            }}
          >
            <NavigationContent project={project} collapsed={sidebarCollapsed} onToggleCollapse={handleSidebarToggle} />
          </Drawer>
          {/* Desktop drawer */}
          <Drawer
            variant="permanent"
            sx={{
              display: { xs: 'none', sm: 'block' },
              '& .MuiDrawer-paper': {
                boxSizing: 'border-box',
                width: currentDrawerWidth,
                bgcolor: 'background.paper',
                borderRight: 1,
                borderColor: 'divider',
                transition: theme.transitions.create('width', {
                  easing: theme.transitions.easing.sharp,
                  duration: theme.transitions.duration.enteringScreen,
                }),
              },
            }}
            open
          >
            <NavigationContent project={project} collapsed={sidebarCollapsed} onToggleCollapse={handleSidebarToggle} />
          </Drawer>
        </Box>

        {/* Main content */}
        <Box
          component="main"
          sx={{
            flexGrow: 1,
            width: { sm: `calc(100% - ${currentDrawerWidth}px)` },
            bgcolor: 'background.default',
            minHeight: '100vh',
            transition: theme.transitions.create(['margin', 'width'], {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.leavingScreen,
            }),
          }}
        >
          <Toolbar />
          <Container maxWidth="xl" sx={{ py: 5, px: 4 }}>
            {children}
          </Container>
        </Box>
      </Box>
  );
}