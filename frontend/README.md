# Video Streaming Frontend

A Next.js-based frontend for video streaming application.

## Features

- Video playback
- Video uploading with progress bar
- Video listing
- Video deletion

## Setup

1. Install Node.js (v18 or later recommended)
2. Install dependencies: `npm install`
3. Start the development server: `npm run dev`

The frontend will start on port 3000.

## Backend Connection

The frontend connects to the Go Fiber backend at `http://localhost:8080`. Make sure the backend is running before using the frontend.

## Project Structure

- `app/page.tsx` - Main video streaming interface
- `app/layout.tsx` - App layout and global styles

## Technologies Used

- Next.js with App Router
- TypeScript
- Tailwind CSS

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
