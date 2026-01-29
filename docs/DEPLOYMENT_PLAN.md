# Deployment Plan (Render + Vercel)

## Overview
This guide deploys the Go backend on Render and the Next.js frontend on Vercel, with auto-deploy enabled on each git push.

---

## 1) Prepare Repository
- Ensure the repo is pushed to GitHub.
- Backend lives in `backend/` and frontend lives in `frontend/`.

---

## 2) Deploy Go Backend on Render

### A. Create Web Service
1. Go to Render → **New** → **Web Service**.
2. Connect the GitHub repo.
3. Set **Root Directory** to `backend`.
4. Choose **Runtime: Go**.

### B. Build and Start
- **Build Command**
```
go build -o app ./cmd/server
```
- **Start Command**
```
./app
```

### C. Environment Variables
Set these in Render → Environment:
```
HOST=0.0.0.0
CORS_ORIGINS=https://<your-vercel-app>.vercel.app
```
Notes:
- `PORT` is automatically injected by Render.
- `CORS_ORIGINS` is comma-separated for multiple origins.

### D. Deploy
- Click **Create Web Service** and wait for deploy to finish.
- Your backend URL will look like:
```
https://<service-name>.onrender.com
```

### E. Enable Auto-Deploy
Render auto-deploys by default when connected to GitHub. Ensure:
- **Auto Deploy** is ON (Render → Service → Settings).
- Any push to the selected branch will trigger deploy.

---

## 3) Deploy Next.js Frontend on Vercel

### A. Create Project
1. Go to Vercel → **New Project**.
2. Import the GitHub repo.
3. Set **Root Directory** to `frontend`.

### B. Environment Variables
Set these in Vercel → Project → Settings → Environment Variables:
```
NEXT_PUBLIC_API_URL=https://<service-name>.onrender.com
```

### C. Deploy
- Click **Deploy**.
- Your frontend URL will look like:
```
https://<project-name>.vercel.app
```

### D. Enable Auto-Deploy
Vercel auto-deploys by default when connected to GitHub. Ensure:
- **Git Integration** is active (Vercel → Project → Settings → Git).
- Any push to the selected branch will trigger deploy.

---

## 4) Verify End-to-End
1. Open the Vercel URL.
2. Trigger any API call from the UI.
3. If you see CORS errors, verify `CORS_ORIGINS` matches the Vercel URL.

---

## Environment Variables Summary

### Render (Backend)
- `HOST=0.0.0.0`
- `CORS_ORIGINS=https://<your-vercel-app>.vercel.app`
- `PORT` (auto by Render)

### Vercel (Frontend)
- `NEXT_PUBLIC_API_URL=https://<service-name>.onrender.com`
