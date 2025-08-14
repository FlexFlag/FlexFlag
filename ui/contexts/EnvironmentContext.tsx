'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useParams } from 'next/navigation';
import { apiClient } from '@/lib/api';

interface Environment {
  id: string;
  key: string;
  name: string;
  description: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

interface EnvironmentContextType {
  currentEnvironment: string;
  setCurrentEnvironment: (environment: string) => void;
  availableEnvironments: string[];
  environments: Environment[];
  loading: boolean;
  refreshEnvironments: () => Promise<void>;
}

const EnvironmentContext = createContext<EnvironmentContextType | undefined>(undefined);

export function useEnvironment() {
  const context = useContext(EnvironmentContext);
  if (context === undefined) {
    throw new Error('useEnvironment must be used within an EnvironmentProvider');
  }
  return context;
}

interface EnvironmentProviderProps {
  children: ReactNode;
  projectId?: string;
}

export function EnvironmentProvider({ children, projectId }: EnvironmentProviderProps) {
  const params = useParams();
  const currentProjectId = projectId || (params.projectId as string);
  
  // Initialize environment from localStorage or default to 'production'
  const [currentEnvironment, setCurrentEnvironment] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(`flexflag_environment_${currentProjectId}`);
      return saved || 'production';
    }
    return 'production';
  });
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchEnvironments = async () => {
    if (!currentProjectId) return;

    setLoading(true);
    try {
      // Get all projects to find the one with matching ID
      const projects = await apiClient.getProjects();
      const project = projects.find(p => p.id === currentProjectId);
      
      if (!project) {
        console.warn(`Project not found for ID: ${currentProjectId}`);
        return;
      }
      
      // Fetch environments using the project slug
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/projects/${project.slug}/environments`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch environments');
      }

      const data = await response.json();
      setEnvironments(data.environments || []);
    } catch (error) {
      console.error('Error fetching environments:', error);
      // Fallback to default environments if API fails
      setEnvironments([
        { id: 'prod', key: 'production', name: 'Production', description: '', is_active: true, sort_order: 0, created_at: '', updated_at: '' },
        { id: 'stage', key: 'staging', name: 'Staging', description: '', is_active: true, sort_order: 1, created_at: '', updated_at: '' },
        { id: 'dev', key: 'development', name: 'Development', description: '', is_active: true, sort_order: 2, created_at: '', updated_at: '' },
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEnvironments();
  }, [currentProjectId]);

  // Update environment when project changes
  useEffect(() => {
    if (typeof window !== 'undefined' && currentProjectId) {
      const saved = localStorage.getItem(`flexflag_environment_${currentProjectId}`);
      if (saved && saved !== currentEnvironment) {
        setCurrentEnvironment(saved);
      }
    }
  }, [currentProjectId]);

  // Extract environment keys for the selector
  const availableEnvironments = environments.map(env => env.key);

  // Create a wrapper function that persists to localStorage
  const persistentSetCurrentEnvironment = (environment: string) => {
    setCurrentEnvironment(environment);
    if (typeof window !== 'undefined' && currentProjectId) {
      localStorage.setItem(`flexflag_environment_${currentProjectId}`, environment);
    }
  };

  // Ensure current environment is valid, fallback to first available or production
  useEffect(() => {
    if (availableEnvironments.length > 0 && !availableEnvironments.includes(currentEnvironment)) {
      const fallback = availableEnvironments.includes('production') ? 'production' : availableEnvironments[0];
      persistentSetCurrentEnvironment(fallback);
    }
  }, [availableEnvironments, currentEnvironment]);

  return (
    <EnvironmentContext.Provider
      value={{
        currentEnvironment,
        setCurrentEnvironment: persistentSetCurrentEnvironment,
        availableEnvironments,
        environments,
        loading,
        refreshEnvironments: fetchEnvironments,
      }}
    >
      {children}
    </EnvironmentContext.Provider>
  );
}