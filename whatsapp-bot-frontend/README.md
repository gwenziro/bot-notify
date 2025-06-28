# WhatsApp Bot Notify - Frontend

Modern React frontend untuk sistem WhatsApp Bot multi-user yang dibangun dengan Next.js 14, TypeScript, dan Tailwind CSS.

## 🚀 Fitur Utama

- **Multi-User Support**: Interface untuk mengelola multiple instance bot WhatsApp
- **Real-time Updates**: WebSocket integration untuk status updates
- **Modern UI/UX**: Responsive design dengan Tailwind CSS
- **Type Safety**: Full TypeScript coverage
- **State Management**: Zustand untuk global state management
- **API Integration**: TanStack Query untuk data fetching
- **Authentication**: JWT-based auth dengan token refresh
- **Testing**: Jest dan React Testing Library

## 🛠️ Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Data Fetching**: TanStack Query
- **UI Components**: Radix UI + Custom Components
- **Icons**: Lucide React
- **Testing**: Jest + React Testing Library
- **Linting**: ESLint + TypeScript ESLint

## 📦 Installation

1. Install dependencies:
```bash
npm install
```

2. Copy environment variables:
```bash
cp .env.example .env.local
```

3. Update environment variables:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 🚀 Development

Start the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:3000`.

## 🧪 Testing

Run tests:
```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

## 🏗️ Build

Build for production:
```bash
npm run build
```

Start production server:
```bash
npm start
```

## 📁 Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── (auth)/            # Auth route group
│   ├── (dashboard)/       # Dashboard route group
│   ├── globals.css        # Global styles
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page
│   └── providers.tsx      # Global providers
├── components/            # React components
│   ├── ui/               # Base UI components
│   ├── dashboard/        # Dashboard components
│   └── layout/           # Layout components
├── lib/                  # Utility libraries
│   ├── api/             # API client and types
│   ├── auth/            # Authentication
│   ├── hooks/           # Custom hooks
│   ├── store/           # State management
│   ├── types/           # TypeScript types
│   └── utils.ts         # Utility functions
└── middleware.ts        # Next.js middleware
```

## 🔧 Configuration

### Environment Variables

- `NEXT_PUBLIC_API_URL`: Backend API URL
- `NEXT_PUBLIC_APP_NAME`: Application name
- `NEXT_PUBLIC_APP_VERSION`: Application version

### API Integration

The frontend communicates with the Go backend via REST API:

- **Base URL**: `http://localhost:8080`
- **Authentication**: JWT tokens via `X-Access-Token` header
- **Real-time**: WebSocket connection for live updates

## 🎨 UI Components

Built with a custom component library based on Radix UI:

- **Button**: Various variants and sizes
- **Card**: Container components
- **Input**: Form inputs with validation
- **Badge**: Status indicators
- **Notification**: Toast notifications

## 📱 Features

### Authentication
- JWT-based authentication
- Token refresh mechanism
- Protected routes
- Remember me functionality

### Dashboard
- Real-time connection status
- QR code display and refresh
- Message statistics
- API documentation

### Multi-User Support
- User-specific bot instances
- Isolated data per user
- Admin panel for user management

## 🔒 Security

- JWT token validation
- Secure token storage
- CSRF protection
- Input validation
- Rate limiting (backend)

## 🚀 Deployment

### Vercel (Recommended)
```bash
npm run build
# Deploy to Vercel
```

### Docker
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:18-alpine AS runner
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
EXPOSE 3000
CMD ["npm", "start"]
```

### Static Export
```bash
npm run build
# Upload dist folder to CDN
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.