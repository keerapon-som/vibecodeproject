# VibeCode - Video Streaming Platform

A modern video streaming platform built with Next.js, TypeScript, and Tailwind CSS.

## Features

- User authentication with NextAuth.js
- Video upload and streaming
- Responsive dashboard
- View tracking for videos
- Modern UI with Tailwind CSS
- PostgreSQL database with Prisma ORM

## Tech Stack

- Next.js 14
- TypeScript
- Tailwind CSS
- NextAuth.js
- Prisma ORM
- PostgreSQL
- React Player

## Prerequisites

- Node.js 18+ and npm
- Docker and Docker Compose
- PostgreSQL (or use Docker)

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/keerapon-som/vibecodeproject.git
   cd vibecodeproject
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Set up the database:
   ```bash
   # Start PostgreSQL using Docker
   docker-compose up -d

   # Run database migrations
   npx prisma migrate dev
   ```

4. Create a `.env` file in the root directory with the following variables:
   ```
   DATABASE_URL="postgresql://postgres:postgres@localhost:5432/videostream"
   NEXTAUTH_SECRET="your-secret-here"
   JWT_SECRET="your-jwt-secret-here"
   NEXTAUTH_URL="http://localhost:3000"
   ```

5. Run the development server:
   ```bash
   npm run dev
   ```

6. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

```
src/
├── app/                    # Next.js app directory
│   ├── api/               # API routes
│   ├── auth/              # Authentication pages
│   └── dashboard/         # Dashboard pages
├── components/            # React components
├── lib/                   # Utility functions and configurations
└── types/                 # TypeScript type definitions
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
