import { NextResponse } from 'next/server';
import { prisma } from '@/lib/prisma';

export async function POST(
  request: Request,
  { params }: { params: { id: string } }
) {
  try {
    // Use raw SQL to increment views
    const video = await prisma.$queryRaw`
      UPDATE "Video"
      SET views = views + 1
      WHERE id = ${params.id}
      RETURNING *
    `;

    if (!video || (Array.isArray(video) && video.length === 0)) {
      return NextResponse.json(
        { message: 'Video not found' },
        { status: 404 }
      );
    }

    return NextResponse.json(Array.isArray(video) ? video[0] : video);
  } catch (error) {
    console.error('Error updating views:', error);
    return NextResponse.json(
      { message: 'Error updating views' },
      { status: 500 }
    );
  }
} 