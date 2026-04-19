# ShortEdge - MERN URL Shortener

ShortEdge is a full-stack URL shortener rebuilt for a stronger SDE-style interview story.
The frontend remains the same lightweight HTML, CSS, and JavaScript experience, while the backend is now structured as a Node.js, Express, and MongoDB service with analytics logging, health checks, and Prometheus-compatible metrics.

## Why this version

This branch keeps the original product behavior but moves the backend into a stack that is easier to explain as a software engineering project:

- Express routing for REST-style endpoints
- MongoDB persistence with Mongoose models
- service-oriented backend structure
- redirect analytics logging
- health and metrics endpoints
- deployment-ready layout for Vercel

## Tech Stack

- Frontend: HTML, CSS, JavaScript
- Backend: Node.js, Express
- Database: MongoDB, Mongoose
- Analytics parsing: `ua-parser-js`
- Observability: `prom-client`
- Deployment: Vercel

## Features

- Create branded or random short URLs
- Redirect with visit logging
- Public or private analytics visibility
- Optional expiry date for each short link
- List all active links
- Update or delete existing links
- Health check endpoint
- Metrics endpoint for observability

## Project Structure

```text
ShortEdge-http/
|-- public/                 # Existing frontend kept as-is
|-- src/
|   |-- config/             # Environment and database connection
|   |-- models/             # Mongoose schemas
|   |-- routes/             # Express route handlers
|   |-- services/           # Business logic and metrics
|   |-- utils/              # Request helpers, short code generation, visit parsing
|-- api/index.js            # Vercel serverless entrypoint
|-- server.js               # Local Express server bootstrap
|-- package.json
|-- vercel.json
```

## API Endpoints

- `POST /shorten` - create a short URL
- `GET /r/:code` - redirect to the original URL and log a visit
- `GET /analytics/:code` - fetch analytics for a short code
- `GET /all` - list all active links
- `PUT /update/:code` - update long URL or visibility
- `DELETE /delete/:code` - delete a short link
- `GET /health` - backend and DB health check
- `GET /metrics` - Prometheus metrics

## Local Setup

1. Install dependencies

```bash
npm install
```

2. Create an environment file from the example

```bash
cp .env.example .env
```

3. Make sure MongoDB is running locally, then start the server

```bash
npm run dev
```

4. Open the app at `http://localhost:8080`

## Environment Variables

- `PORT` - local server port, default `8080`
- `MONGODB_URI` - MongoDB connection string

## Interview Positioning

This is best presented as a backend-heavy SDE project:

- designed REST endpoints for URL lifecycle management
- implemented persistence with MongoDB schemas
- handled redirect flow and analytics capture
- exposed operational endpoints for health and metrics
- preserved the product UI while re-architecting the backend stack

## Notes

- The original Go files are still present in the repo for reference, but this branch is organized around the MERN backend implementation.
- The frontend pages in `public/` are the ones served by the Express app.
