# Resto Rate

Restaurant rating and review application built with SvelteKit frontend and Go API backend.

## Architecture

- **Frontend**: SvelteKit with TailwindCSS, Skeleton UI, and Paraglide for i18n
- **Backend**: Go API with Connect-RPC (gRPC-compatible), PostgreSQL database
- **Database**: PostgreSQL with GORM ORM
- **Communication**: Connect-RPC for type-safe API communication
- **Monorepo**: Nx workspace with bun package manager
- **Authentication**: Google OAuth (planned, not yet implemented)
- **Internationalization**: Paraglide setup for multi-language support

## Prerequisites

- **Node.js** 18+ and **npm** (for JavaScript protoc-gen plugins)
- **Go** 1.21+ (for Go protoc-gen plugins)
- **Bun** (package manager)
- **PostgreSQL** database
- **Docker** (optional, for database)
- **Buf CLI** (for protobuf code generation): `npm install -g @bufbuild/buf`

## Current Features

### ✅ Implemented
- **Basic CRUD Operations**: Full CRUD for restaurants and users
- **Database Models**: User and Restaurant models with UUIDv7 primary keys
- **API Services**: Connect-RPC services for restaurants and users
- **Pagination**: Server-side pagination with page tokens
- **Database Seeding**: Development data seeding
- **CORS Support**: Cross-origin resource sharing configured
- **gRPC Reflection**: Development API introspection
- **Docker Support**: PostgreSQL container setup

### 🚧 In Development
- **Authentication System**: Google OAuth integration
- **Google Places API**: Restaurant data integration
- **User Reviews**: Rating and review system
- **Search & Filtering**: Restaurant search functionality
- **UI/UX**: Complete frontend implementation

## Quick Start

1. **Install dependencies**:

```bash
bun install
```

2. **Install Protocol Buffer plugins** (required for code generation):

```bash
# Install Go plugins for API
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest

# Install JavaScript/TypeScript plugins for Web
npm install -g @bufbuild/protoc-gen-es
npm install -g @bufbuild/protoc-gen-connect-es
```

3. **Set up environment**:
Ensure:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

```bash
cp env.template .env
# Edit .env with your database credentials
```

4. **Start database** (Docker recommended):

```bash
docker-compose up -d postgres
```

5. **Start development**:

```bash
bun run dev
```

This starts both the web app (http://localhost:5173) and API (http://localhost:3001).

## Protocol Buffer Setup

This project uses Protocol Buffers (protobuf) for API contract definition and code generation. The setup includes:

### Required Plugins

**For Go API** (`apps/api/buf.gen.yaml`):
- `protoc-gen-go` - Generates Go structs from .proto files
- `protoc-gen-go-grpc` - Generates Go gRPC code
- `protoc-gen-connect-go` - Generates Connect-Go code

**For Web Frontend** (`apps/web/buf.gen.yaml`):
- `protoc-gen-es` - Generates TypeScript/JavaScript code
- `protoc-gen-connect-es` - Generates Connect-ES client code

### Installation Commands

```bash
# Go plugins (required for API)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install buf.build/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest

# JavaScript/TypeScript plugins (required for Web)
npm install -g @bufbuild/protoc-gen-es
npm install -g @bufbuild/protoc-gen-connect-es
```

### Code Generation

After installing plugins, you can generate code from .proto files:

```bash
# Generate Go code for API
cd apps/api && buf generate

# Generate TypeScript code for Web
cd apps/web && buf generate
```

### Fresh PC Setup

When setting up on a new machine, ensure you have:
1. **Go** installed (for Go plugins)
2. **Node.js/npm** installed (for JavaScript plugins)
3. **Buf CLI** installed: `npm install -g @bufbuild/buf`

Then run the installation commands above.

**Note**: The Buf CLI is required for running `buf generate` commands. Install it globally with npm to ensure it's available system-wide.

## API Endpoints

### Restaurants Service
- `CreateRestaurant` - Create a new restaurant
- `GetRestaurant` - Get restaurant by ID
- `UpdateRestaurant` - Update restaurant details
- `DeleteRestaurant` - Delete restaurant
- `ListRestaurants` - List restaurants with pagination

### Users Service
- `CreateUser` - Create a new user
- `GetUser` - Get user by ID
- `UpdateUser` - Update user details
- `DeleteUser` - Delete user
- `ListUsers` - List users with pagination

### Google Maps Service
- `SearchText` - Search for places using text query with dynamic field selection

#### Dynamic Field Selection
The Google Maps service supports dynamic field selection to optimize API calls and reduce response size. You can specify which fields to return in the response:

```typescript
// Example: Request only specific fields
const response = await client.searchText({
  textQuery: "restaurants in Banja Luka",
  includedType: "restaurant",
  maxResultCount: 10,
  requestedFields: [
    "name",
    "displayName", 
    "rating",
    "formattedAddress",
    "photos",
    "priceLevel"
  ]
});

// Example: Request all fields (default behavior)
const response = await client.searchText({
  textQuery: "restaurants in Banja Luka",
  includedType: "restaurant",
  maxResultCount: 10
  // requestedFields not specified - returns all available fields
});
```

**Available Fields:**
- Basic info: `name`, `displayName`, `id`, `types`, `primaryType`, `primaryTypeDisplayName`
- Contact: `nationalPhoneNumber`, `internationalPhoneNumber`, `formattedAddress`, `shortFormattedAddress`
- Ratings: `rating`, `userRatingCount`
- Business: `businessStatus`, `priceLevel`, `websiteUri`, `googleMapsUri`
- Services: `takeout`, `delivery`, `dineIn`, `curbsidePickup`, `reservable`
- Food options: `servesBreakfast`, `servesLunch`, `servesDinner`, `servesBeer`, `servesWine`, `servesBrunch`, `servesVegetarianFood`
- Amenities: `outdoorSeating`, `liveMusic`, `menuForChildren`, `servesCocktails`, `servesDessert`, `servesCoffee`
- Accessibility: `goodForChildren`, `allowsDogs`, `restroom`, `goodForGroups`, `goodForWatchingSports`
- Media: `photos`, `attributions`
- Other: `utcOffsetMinutes`, `pureServiceAreaBusiness`

## Database Schema

### Users
- `id` (UUIDv7) - Primary key
- `google_id` - Google OAuth ID (unique)
- `email` - User email (unique)
- `username` - Username (unique)
- `name` - Display name
- `is_admin` - Admin privileges
- `created_at` / `updated_at` - Timestamps

### Restaurants
- `id` (UUIDv7) - Primary key
- `google_id` - Google Places ID (unique)
- `email` - Restaurant email (unique)
- `name` - Restaurant name
- `created_at` / `updated_at` - Timestamps

## Development

- `bun run dev` - Start both apps
- `bun run build` - Build all apps
- `bun run lint` - Lint all packages
- `bun run test` - Run tests
- `bun run graph` - View project dependency graph

## Project Structure

```
resto-rate/
├── apps/
│   ├── web/          # SvelteKit frontend
│   └── api/          # Go backend
├── packages/
│   └── protos/       # Protocol Buffer definitions
├── .env              # Environment variables
└── env.template      # Environment template
```

## Technology Stack

### Frontend
- **SvelteKit** - Full-stack web framework
- **TailwindCSS** - Utility-first CSS framework
- **Skeleton UI** - SvelteKit UI toolkit
- **Connect-RPC** - Type-safe API client
- **Paraglide** - Internationalization
- **TypeScript** - Type safety

### Backend
- **Go** - Programming language
- **Connect-RPC** - gRPC-compatible RPC framework
- **GORM** - Go ORM library
- **PostgreSQL** - Database
- **UUIDv7** - Unique identifier generation

### Infrastructure
- **Nx** - Monorepo build system
- **Bun** - Package manager and runtime
- **Docker** - Containerization
- **Protocol Buffers** - API contract definition

## Documentation

See `ROADMAP.md` for detailed development roadmap and `SETUP.md` for detailed setup instructions.
