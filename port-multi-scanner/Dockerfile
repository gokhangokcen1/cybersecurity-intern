FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-build
WORKDIR /src
COPY go.work ./
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ ./backend/
RUN CGO_ENABLED=0 go build -o /scanner ./backend

FROM alpine:3.20
WORKDIR /app
COPY --from=backend-build /scanner ./scanner
COPY --from=frontend-build /src/frontend/dist ./frontend/dist
COPY ip_feed.json ./ip_feed.json
EXPOSE 8080
ENV PORT=8080
CMD ["./scanner"]
