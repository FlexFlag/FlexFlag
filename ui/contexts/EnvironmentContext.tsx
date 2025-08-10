'use client';

import React, { createContext, useContext, useState, ReactNode } from 'react';

interface EnvironmentContextType {
  currentEnvironment: string;
  setCurrentEnvironment: (environment: string) => void;
  availableEnvironments: string[];
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
}

export function EnvironmentProvider({ children }: EnvironmentProviderProps) {
  const [currentEnvironment, setCurrentEnvironment] = useState('production');
  const availableEnvironments = ['production', 'staging', 'development'];

  return (
    <EnvironmentContext.Provider
      value={{
        currentEnvironment,
        setCurrentEnvironment,
        availableEnvironments,
      }}
    >
      {children}
    </EnvironmentContext.Provider>
  );
}