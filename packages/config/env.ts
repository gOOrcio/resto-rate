/**
 * Centralized environment configuration for the entire application
 * This package provides type-safe environment variable access across all apps
 */

export interface DatabaseConfig {
  url: string;
  maxConnections?: number;
  ssl?: boolean;
}

export interface ApiConfig {
  port: number;
  host: string;
  corsOrigin: string | string[];
  nodeEnv: 'development' | 'production' | 'test';
}

export interface WebConfig {
  port: number;
  apiUrl: string;
  nodeEnv: 'development' | 'production' | 'test';
}

export interface AuthConfig {
  sessionSecret: string;
  sessionMaxAge: number; // in seconds
  bcryptRounds: number;
}

export interface AppConfig {
  database: DatabaseConfig;
  api: ApiConfig;
  web: WebConfig;
  auth: AuthConfig;
}

/**
 * Parse and validate environment variables
 */
function parseEnv(): AppConfig {
  // Required variables check
  const requiredVars = ['DATABASE_URL'];
  const missing = requiredVars.filter(key => !process.env[key]);
  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
  }

  // Parse boolean helper
  const parseBoolean = (value: string | undefined, defaultValue: boolean): boolean => {
    if (!value) return defaultValue;
    return value.toLowerCase() === 'true';
  };

  // Parse number helper
  const parseNumber = (value: string | undefined, defaultValue: number): number => {
    if (!value) return defaultValue;
    const parsed = parseInt(value, 10);
    return isNaN(parsed) ? defaultValue : parsed;
  };

  const nodeEnv = (process.env.NODE_ENV || 'development') as 'development' | 'production' | 'test';

  return {
    database: {
      url: process.env.DATABASE_URL!,
      maxConnections: parseNumber(process.env.DATABASE_MAX_CONNECTIONS, 20),
      ssl: parseBoolean(process.env.DATABASE_SSL, nodeEnv === 'production'),
    },
    api: {
      port: parseNumber(process.env.API_PORT, 3001),
      host: process.env.API_HOST || '0.0.0.0',
      corsOrigin: nodeEnv === 'development' 
        ? ['http://localhost:5173', 'http://localhost:3000']
        : (process.env.CORS_ORIGIN?.split(',') || []),
      nodeEnv,
    },
    web: {
      port: parseNumber(process.env.WEB_PORT, 5173),
      apiUrl: process.env.API_URL || 
        (nodeEnv === 'development' ? 'http://localhost:3001' : 'https://api.yourdomain.com'),
      nodeEnv,
    },
    auth: {
      sessionSecret: process.env.SESSION_SECRET || 
        (nodeEnv === 'development' ? 'dev-secret-key-change-in-production' : ''),
      sessionMaxAge: parseNumber(process.env.SESSION_MAX_AGE, 30 * 24 * 60 * 60), // 30 days
      bcryptRounds: parseNumber(process.env.BCRYPT_ROUNDS, 12),
    },
  };
}

let config: AppConfig | null = null;

export function getConfig(): AppConfig {
  if (!config) {
    config = parseEnv();
  }
  return config;
}

export const isDevelopment = () => getConfig().api.nodeEnv === 'development';
export const isProduction = () => getConfig().api.nodeEnv === 'production';
export const isTest = () => getConfig().api.nodeEnv === 'test';

export const getDatabaseConfig = () => getConfig().database;
export const getApiConfig = () => getConfig().api;
export const getWebConfig = () => getConfig().web;
export const getAuthConfig = () => getConfig().auth; 