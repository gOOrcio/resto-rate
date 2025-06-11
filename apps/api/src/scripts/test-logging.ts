// Load environment variables from root .env file
import { config } from 'dotenv';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Load .env from project root (three levels up from scripts/)
const envPath = resolve(__dirname, '../../../../.env');
console.log('Loading environment from:', envPath);
const result = config({ path: envPath });
if (result.error) {
	console.error('Failed to load .env file:', result.error);
} else {
	console.log('✅ Environment loaded successfully');
}

import { createLogger, createLoggerFromEnv } from '@resto-rate/logger';

console.log('🧪 Testing Resto Rate Logging System\n');

// Test 1: Default logger from environment
console.log('=== Test 1: Default Logger from Environment ===');
const defaultLogger = createLoggerFromEnv('test-service');

defaultLogger.trace('This is a TRACE message with details', { userId: 'user123', action: 'login' });
defaultLogger.debug('This is a DEBUG message with context', { query: 'SELECT * FROM users', duration: '15ms' });
defaultLogger.info('This is an INFO message for general information', { event: 'server_started', port: 3001 });
defaultLogger.warn('This is a WARN message for potential issues', { memory_usage: '85%', threshold: '80%' });
defaultLogger.error('This is an ERROR message for failures', { error: 'Database connection failed', retries: 3 });
defaultLogger.fatal('This is a FATAL message for critical failures', { error: 'System shutdown', code: 1 });

console.log('\n=== Test 2: Custom Logger with Different Levels ===');

// Test 2: Test different log levels
const logLevels = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'] as const;

for (const level of logLevels) {
	console.log(`\n--- Testing with LOG_LEVEL=${level} ---`);
	
	const customLogger = createLogger({
		level,
		service: `test-${level}`,
		environment: 'development',
		pretty: true,
	});

	customLogger.trace('TRACE: Detailed debugging information');
	customLogger.debug('DEBUG: Development debugging information');
	customLogger.info('INFO: General application information');
	customLogger.warn('WARN: Warning about potential issues');
	customLogger.error('ERROR: Error that needs attention');
	customLogger.fatal('FATAL: Critical error requiring immediate action');
}

console.log('\n=== Test 3: Child Loggers with Context ===');

const parentLogger = createLoggerFromEnv('parent-service');
const childLogger = parentLogger.child({ 
	requestId: 'req-123', 
	userId: 'user-456',
	component: 'auth-middleware'
});

parentLogger.info('Parent logger message');
childLogger.info('Child logger message with inherited context');
childLogger.warn('Child logger warning', { additional: 'data' });

console.log('\n🎯 Logging test complete!');
console.log('💡 To change log levels:');
console.log('   1. Set LOG_LEVEL=info in your .env file');
console.log('   2. Restart the application');
console.log('   3. Or set LOG_LEVEL environment variable directly'); 