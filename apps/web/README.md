# Web App

This is the SvelteKit frontend for the Resto Rate application.

## Development

### Environment Configuration

The application uses environment variables for configuration. Copy the example file and customize it:

```bash
cp .env.example .env
```

#### Available Environment Variables

- `VITE_PORT` - Development server port (default: 5173)
- `VITE_API_URL` - API base URL (default: http://localhost:3001)

#### Example `.env` file

```env
# Development server port (default: 5173)
VITE_PORT=5173

# API base URL (default: http://localhost:3001)
VITE_API_URL=http://localhost:3001
```

### Starting the Development Server

```bash
npm run dev
```

The server will:
- Use the port specified in `VITE_PORT` (defaults to 5173)
- Fail to start if the port is occupied (`strictPort: true`)
- Allow external connections (`host: true`)

### Port Conflicts

If you encounter a port conflict, you can:

1. **Change the port in your `.env` file:**
   ```env
   VITE_PORT=3000
   ```

2. **Check what's using the current port:**
   ```bash
   lsof -i :5173
   ```

3. **Stop the conflicting process:**
   ```bash
   # If it's another Vite process
   pkill -f "vite"
   
   # Or kill by PID (replace XXXX with actual PID)
   kill -9 XXXX
   ```

## Building

```bash
npm run build
```

## Testing

```bash
npm run test
```
