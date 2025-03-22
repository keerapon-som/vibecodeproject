#!/bin/bash

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Generate secrets for .env file
echo "Generating secrets..."
NEXTAUTH_SECRET=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)

# Create .env file
echo "Creating .env file..."
cat > .env << EOL
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/videostream?schema=public"
NEXTAUTH_SECRET="${NEXTAUTH_SECRET}"
NEXTAUTH_URL="http://localhost:3000"
JWT_SECRET="${JWT_SECRET}"
EOL

# Install dependencies
echo "Installing dependencies..."
npm install

# Run database migrations
echo "Running database migrations..."
npx prisma migrate dev

# Create uploads directory
echo "Creating uploads directory..."
mkdir -p public/uploads

# Start the development server
echo "Starting the development server..."
npm run dev 