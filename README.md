# 🔗 ShortEdge – URL Shortener 

ShortEdge is a full-stack URL shortener rebuilt with a modern backend using **Node.js, Express, and MongoDB**.  
It preserves the original product experience while upgrading the backend into a **scalable, interview-ready architecture**.

- Service-oriented backend architecture  
- Deployment-ready (Vercel)  
- Analytics + observability focused design  

---

## Core Features

| Feature | Description |
|--------|------------|
| Branded Short Links | Supports custom short codes (e.g., `/r/my-event`) |
| Public/Private Toggle | Control analytics visibility per link |
| Link Expiry | Auto-deactivate links after expiry time |
| Analytics Logging | Tracks visits, browser, and device info |
| RESTful CRUD API | Create, read, update, delete short links |
| Device Parsing | Uses `ua-parser-js` for browser/device detection |
| Metrics Endpoint | Prometheus-compatible metrics via `prom-client` |
| Frontend UI | Lightweight HTML, CSS, JS interface |
| Health Check | `/health` for backend + DB status |
---

## Tech Stack

- Frontend: HTML, CSS, JavaScript
- Backend: Node.js, Express
- Database: MongoDB, Mongoose
- Analytics parsing: `ua-parser-js`
- Observability: `prom-client`
- Deployment: Vercel
  
---

## System Architecture

```text
┌────────────────────────────┐
│  Client (Web / API)        │
│ ────────────────────────── │
│ • HTML Web UI              │
│ • Postman / API Clients    │
└─────────────┬──────────────┘
              ▼
┌────────────────────────────┐
│ Express Router             │
│ • Handles routes           │
│ • Maps endpoints           │
└─────────────┬──────────────┘
              ▼
┌────────────────────────────┐
│ Route Handlers             │
│ • Request parsing          │
│ • Response handling        │
└─────────────┬──────────────┘
              ▼
┌────────────────────────────┐
│ Service Layer              │
│ • Business logic           │
│ • Expiry & visibility      │
│ • Metrics handling         │
└─────────────┬──────────────┘
              ▼
┌────────────────────────────┐
│ MongoDB (Mongoose)         │
│ • URL storage              │
│ • Visit logs               │
└─────────────┬──────────────┘
              ▼
┌────────────────────────────┐
│ Analytics + Metrics        │
│ • ua-parser-js → device    │
│ • prom-client → metrics    │
└────────────────────────────┘
```
---

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

---

## Demo Screenshots

### Frontend UI
A minimal, responsive interface for submitting long URLs, choosing custom short codes, toggling visibility, and receiving branded short links.
![Frontend UI](assets/ui-home.png)

---

### Analytics Dashboard
Returns rich, real-time analytics per short link 
![Analytics](assets/get-analytics.png)

---

### All Links
Lists all shortened links (public/private) with long URL mapping.

![All Links](assets/get-all.png)

---

### Metrics Endpoint
Prometheus-compatible metrics exposed at `/metrics`.

![Metrics](assets/metrics-page.png)

---

## Environment Variables

- `PORT` - local server port, default `8080`
- `MONGODB_URI` - MongoDB connection string

---

## Real-World Use Cases

* Custom short links for Google Forms, PDFs, feedback links
* Private academic resource sharing
* Insight collection for link click-through rate
* Prometheus-ready analytics for observability dashboards
  
---
