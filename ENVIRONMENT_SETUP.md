# Environment Variables Strategy

## 🎯 Centralized Configuration Approach

### Architecture Overview
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   ROOT .env     │    │ CONFIG PACKAGE  │    │   APPS          │
│                 │    │                 │    │                 │
│ • DATABASE_URL  │───►│ • Type Safety   │───►│ • apps/web      │
│ • API_PORT      │    │ • Validation    │    │ • apps/api      │
│ • SESSION_*     │    │ • Defaults      │    │ • packages/*    │
│ • NODE_ENV      │    │ • Parsing       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 File Structure

### Root Level
```
resto-rate/
├── .env                    # Main environment file
├── env.template           # Template for new setups
├── packages/config/       # Centralized config package
│   ├── env.ts            # Configuration logic
│   └── package.json      # Package definition
├── apps/web/             # Uses config for API URL, etc.
└── apps/api/             # Uses config for all settings
```

## 🔧 Environment Variables

### Required Variables
```bash
# Database (Required)
DATABASE_URL="postgresql://user:pass@localhost:5432/resto_rate"

# Optional with Defaults
DATABASE_MAX_CONNECTIONS=20        # Default: 20
DATABASE_SSL=false                 # Default: false in dev, true in prod

# API Configuration
API_PORT=3001                      # Default: 3001
API_HOST=0.0.0.0                   # Default: 0.0.0.0

# Web App Configuration  
WEB_PORT=5173                      # Default: 5173
API_URL=http://localhost:3001      # Default: dev URL

# Authentication
SESSION_SECRET=your-secret-key     # Required in production
SESSION_MAX_AGE=2592000           # Default: 30 days
BCRYPT_ROUNDS=12                  # Default: 12

# Environment
NODE_ENV=development              # Default: development
```

### Environment-Specific Configs

#### Development (.env)
```bash
DATABASE_URL="postgresql://dev:dev@localhost:5432/resto_rate_dev"
API_PORT=3001
WEB_PORT=5173
API_URL=http://localhost:3001
NODE_ENV=development
SESSION_SECRET=dev-secret-key-not-for-production
```

#### Production
```bash
DATABASE_URL="postgresql://prod_user:secure_pass@prod-db:5432/resto_rate"
DATABASE_SSL=true
DATABASE_MAX_CONNECTIONS=50
API_PORT=3001
API_HOST=0.0.0.0
API_URL=https://api.yourdomain.com
CORS_ORIGIN=https://yourdomain.com,https://www.yourdomain.com
NODE_ENV=production
SESSION_SECRET=super-secure-random-secret-key
SESSION_MAX_AGE=604800  # 7 days for production
```

#### Testing
```bash
DATABASE_URL="postgresql://test:test@localhost:5432/resto_rate_test"
NODE_ENV=test
SESSION_SECRET=test-secret
API_PORT=3002  # Different port to avoid conflicts
```

## 💻 Usage in Code

### Centralized Config Package
```typescript
// packages/config/env.ts
import { getConfig, getDatabaseConfig, getApiConfig } from '@resto-rate/config';

// Type-safe configuration access
const dbConfig = getDatabaseConfig();
const apiConfig = getApiConfig();

// Environment checks
if (isDevelopment()) {
  console.log('Running in development mode');
}
```

### API Usage
```typescript
// apps/api/src/index.ts
import { getApiConfig, getDatabaseConfig } from '@resto-rate/config';

const apiConfig = getApiConfig();
const dbConfig = getDatabaseConfig();

await server.listen({ 
  port: apiConfig.port, 
  host: apiConfig.host 
});
```

### Web App Usage
```typescript
// apps/web/src/lib/api.ts
import { getWebConfig } from '@resto-rate/config';

const webConfig = getWebConfig();
const apiClient = new ApiClient(webConfig.apiUrl);
```

### Database Connection
```typescript
// Shared across both apps
import { getDatabaseConfig } from '@resto-rate/config';

const dbConfig = getDatabaseConfig();
const client = postgres(dbConfig.url, {
  max: dbConfig.maxConnections,
  ssl: dbConfig.ssl
});
```

## 🔄 Setup Process

### 1. Copy Environment Template
```bash
cp env.template .env
# Edit .env with your values
```

### 2. Install Dependencies
```bash
bun install  # Installs workspace packages
```

### 3. Configure Database
```bash
# Start database (if using Docker)
cd apps/web && bun run db:start

# Push schema to database
bun run db:push
```

### 4. Start Development
```bash
bun run dev  # Starts both apps with shared config
```

## 🏗️ Benefits

### ✅ Advantages
- **Type Safety**: All environment variables are typed and validated
- **Single Source**: One .env file for entire monorepo
- **Validation**: Missing required variables throw clear errors
- **Defaults**: Sensible defaults for development
- **Environment Switching**: Easy production/staging/dev configs
- **IDE Support**: Full autocomplete and intellisense

### 🔧 Configuration Features
- **Automatic Parsing**: Numbers, booleans parsed correctly
- **Environment Detection**: Development vs production logic
- **Validation**: Required variables checked at startup
- **Fallbacks**: Graceful defaults for optional settings

## 🚀 Deployment

### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'
services:
  api:
    env_file: .env
    environment:
      - NODE_ENV=production
  
  web:
    env_file: .env
    environment:
      - NODE_ENV=production
```

### Vercel/Netlify
```bash
# Set environment variables in dashboard
DATABASE_URL=production-db-url
SESSION_SECRET=production-secret
API_URL=https://api.yourdomain.com
NODE_ENV=production
```

## 🛡️ Security Best Practices

### Development
- ✅ Use `.env` file (gitignored)
- ✅ Provide `env.template` for team
- ✅ Use weak secrets (clearly marked)

### Production
- ✅ Use secure random secrets
- ✅ Enable database SSL
- ✅ Set appropriate CORS origins
- ✅ Use environment-specific variables
- ❌ Never commit production secrets

## 🔍 Troubleshooting

### Common Issues
1. **Missing Variables**: Check required variables in console output
2. **Wrong Database URL**: Verify format and credentials
3. **Port Conflicts**: Change default ports if needed
4. **CORS Issues**: Update API CORS configuration

### Debug Commands
```bash
# Check config loading
bun --print "console.log(require('@resto-rate/config').getConfig())"

# Verify database connection
cd apps/api && bun run dev  # Shows connection status

# Test API health
curl http://localhost:3001/health
```

This centralized approach provides type safety, validation, and easy environment management across your entire monorepo! 