'use client';

import { useState } from 'react';
import {
  Chip,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
  Typography,
  Box,
} from '@mui/material';
import {
  AccountTree as ProjectIcon,
  Check as CheckIcon,
} from '@mui/icons-material';
import { useProject } from '@/contexts/ProjectContext';

export default function ProjectSelector() {
  const { currentProject, projects, setCurrentProject, loading } = useProject();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleProjectSelect = (project: any) => {
    setCurrentProject(project);
    handleClose();
  };

  if (loading) {
    return (
      <Chip
        icon={<CircularProgress size={16} />}
        label="Loading..."
        variant="outlined"
        disabled
      />
    );
  }

  const displayName = currentProject?.name || 'No Project';
  const isMenuOpen = Boolean(anchorEl);

  return (
    <>
      <Chip
        icon={<ProjectIcon />}
        label={displayName}
        variant="outlined"
        onClick={handleClick}
        sx={{ 
          cursor: 'pointer',
          '&:hover': {
            bgcolor: 'action.hover',
          }
        }}
      />
      <Menu
        anchorEl={anchorEl}
        open={isMenuOpen}
        onClose={handleClose}
        transformOrigin={{ horizontal: 'left', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'left', vertical: 'bottom' }}
        PaperProps={{
          sx: { minWidth: 200, maxWidth: 300 }
        }}
      >
        {projects.length === 0 ? (
          <MenuItem disabled>
            <ListItemText>
              <Typography variant="body2" color="text.secondary">
                No projects available
              </Typography>
            </ListItemText>
          </MenuItem>
        ) : (
          projects.map((project) => (
            <MenuItem
              key={project.id}
              onClick={() => handleProjectSelect(project)}
              selected={currentProject?.id === project.id}
            >
              <ListItemIcon sx={{ minWidth: 36 }}>
                {currentProject?.id === project.id ? (
                  <CheckIcon fontSize="small" color="primary" />
                ) : (
                  <ProjectIcon fontSize="small" />
                )}
              </ListItemIcon>
              <ListItemText>
                <Box>
                  <Typography variant="body2" fontWeight={500}>
                    {project.name}
                  </Typography>
                  {project.description && (
                    <Typography variant="caption" color="text.secondary">
                      {project.description}
                    </Typography>
                  )}
                </Box>
              </ListItemText>
            </MenuItem>
          ))
        )}
      </Menu>
    </>
  );
}