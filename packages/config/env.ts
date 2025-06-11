/**
 * Centralized environment configuration for the entire application
 * This package provides type-safe environment variable access across all apps
 */

import { config as loadDotenv } from 'dotenv';
import { resolve } from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Try to load .env from common locations
try {
	// Try root directory first (most common)
	loadDotenv({ path: resolve(process.cwd(), '../../.env') });
} catch {
	try {
		// Try project root relative to this package
		loadDotenv({ path: resolve(__dirname, '../../../.env') });
	} catch {
		// Fallback - no .env file found, use process.env
	}
}

export type LogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';

export interface LoggingConfig {
	level: LogLevel;
	pretty: boolean;
	enableFile: boolean;
	logDir: string;
}

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
	logging: LoggingConfig;
}

/**
 * Parse and validate environment variables
 */
function parseEnv(): AppConfig {
	const requiredEnvVars = ['DATABASE_URL', 'SESSION_SECRET'];
	for (const envVar of requiredEnvVars) {
		if (!process.env[envVar]) {
			throw new Error(`Missing required environment variable: ${envVar}`);
		}
	}

	const parseBoolean = (value: string | undefined, defaultValue: boolean): boolean => {
		if (value === undefined) return defaultValue;
		return value.toLowerCase() === 'true';
	};

	const parseNumber = (value: string | undefined, defaultValue: number): number => {
		if (value === undefined) return defaultValue;
		const parsed = parseInt(value, 10);
		return isNaN(parsed) ? defaultValue : parsed;
	};

	const getLogLevel = (): LogLevel => {
		const level = process.env.LOG_LEVEL?.toLowerCase() as LogLevel;
		const validLevels: LogLevel[] = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
		if (level && validLevels.includes(level)) {
			return level;
		}
		// Default based on environment
		return process.env.NODE_ENV === 'production' ? 'info' : 'debug';
	};

	const nodeEnv = (process.env.NODE_ENV || 'development') as 'development' | 'production' | 'test';

	return {
		database: {
			url: process.env.DATABASE_URL!,
			maxConnections: parseNumber(process.env.DATABASE_MAX_CONNECTIONS, 10),
			ssl: parseBoolean(process.env.DATABASE_SSL, false),
		},
		api: {
			port: parseNumber(process.env.API_PORT, 3001),
			host: process.env.API_HOST || '0.0.0.0',
			corsOrigin: process.env.CORS_ORIGIN || 'http://localhost:5173',
			nodeEnv,
		},
		web: {
			port: parseNumber(process.env.WEB_PORT, 5173),
			apiUrl: process.env.API_URL || 'http://localhost:3001',
			nodeEnv,
		},
		auth: {
			sessionSecret: process.env.SESSION_SECRET!,
			sessionMaxAge: parseNumber(process.env.SESSION_MAX_AGE, 30 * 24 * 60 * 60), // 30 days
			bcryptRounds: parseNumber(process.env.BCRYPT_ROUNDS, 12),
		},
		logging: {
			level: getLogLevel(),
			pretty: parseBoolean(process.env.LOG_PRETTY, process.env.NODE_ENV !== 'production'),
			enableFile: parseBoolean(process.env.LOG_FILE, false),
			logDir: process.env.LOG_DIR || './logs',
		},
	};
}

let appConfig: AppConfig | null = null;

export function getConfig(): AppConfig {
	if (!appConfig) {
		appConfig = parseEnv();
	}
	return appConfig;
}

export const isDevelopment = () => getConfig().api.nodeEnv === 'development';
export const isProduction = () => getConfig().api.nodeEnv === 'production';
export const isTest = () => getConfig().api.nodeEnv === 'test';

export const getDatabaseConfig = () => getConfig().database;
export const getApiConfig = () => getConfig().api;
export const getWebConfig = () => getConfig().web;
export const getAuthConfig = () => getConfig().auth;
export const getLoggingConfig = () => getConfig().logging;
