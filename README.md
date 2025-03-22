# Video Streaming Application

A modern video streaming web application built with Next.js, featuring user authentication and video upload capabilities.

## Features

- User authentication with email/password
- Video upload and streaming
- Protected routes
- Modern UI with Tailwind CSS
- PostgreSQL database with Prisma ORM

## Prerequisites

- Node.js 18.x or later
- PostgreSQL 14.x or later
- npm or yarn

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd videostream
```

2. Install dependencies:
```bash
npm install
```

3. Set up the database:
```bash
# Create a PostgreSQL database named 'videostream'
createdb videostream

# Run database migrations
npx prisma migrate dev
```

4. Create a `.env` file in the root directory with the following variables:
```env
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/videostream?schema=public"
NEXTAUTH_SECRET="your-secret-key-here"
NEXTAUTH_URL="http://localhost:3000"
JWT_SECRET="your-jwt-secret-here"
```

5. Start the development server:
```bash
npm run dev
```

The application will be available at http://localhost:3000

## Project Structure

- `/src/app` - Next.js app router pages and API routes
- `/src/components` - React components
- `/prisma` - Database schema and migrations
- `/public/uploads` - Directory for storing uploaded videos

## API Routes

- `/api/auth/*` - Authentication endpoints
- `/api/videos` - Video management endpoints

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
