'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';

interface Project {
  id: string;
  name: string;
  slug: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface ProjectContextValue {
  currentProject: Project | null;
  projects: Project[];
  setCurrentProject: (project: Project | null) => void;
  loading: boolean;
  error: string | null;
  refetchProjects: () => Promise<void>;
}

const ProjectContext = createContext<ProjectContextValue | undefined>(undefined);

export function ProjectProvider({ children }: { children: React.ReactNode }) {
  const [currentProject, setCurrentProjectState] = useState<Project | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProjects = async () => {
    try {
      setError(null);
      const projectsData = await apiClient.getProjects();
      setProjects(projectsData);
      
      // Set current project from localStorage or first project
      const savedProjectId = typeof window !== 'undefined' ? localStorage.getItem('currentProjectId') : null;
      let projectToSelect = null;
      
      if (savedProjectId) {
        projectToSelect = projectsData.find(p => p.id === savedProjectId) || null;
      }
      
      // If no saved project or saved project not found, select first active project
      if (!projectToSelect && projectsData.length > 0) {
        projectToSelect = projectsData.find(p => p.is_active) || projectsData[0];
      }
      
      if (projectToSelect) {
        setCurrentProjectState(projectToSelect);
        if (typeof window !== 'undefined') {
          localStorage.setItem('currentProjectId', projectToSelect.id);
        }
      }
    } catch (err) {
      setError('Failed to load projects');
      console.error('Project loading error:', err);
    } finally {
      setLoading(false);
    }
  };

  const setCurrentProject = (project: Project | null) => {
    setCurrentProjectState(project);
    if (typeof window !== 'undefined') {
      if (project) {
        localStorage.setItem('currentProjectId', project.id);
      } else {
        localStorage.removeItem('currentProjectId');
      }
    }
  };

  const refetchProjects = async () => {
    setLoading(true);
    await fetchProjects();
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  return (
    <ProjectContext.Provider value={{
      currentProject,
      projects,
      setCurrentProject,
      loading,
      error,
      refetchProjects,
    }}>
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error('useProject must be used within a ProjectProvider');
  }
  return context;
}