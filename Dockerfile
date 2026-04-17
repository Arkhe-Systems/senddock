FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM golang:1.25-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /app/frontend/dist ../frontend/dist
RUN CGO_ENABLED=0 go build -o /senddock cmd/server/main.go
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /senddock .
COPY --from=backend /go/bin/goose /usr/local/bin/goose
COPY --from=backend /app/frontend/dist ./frontend/dist
COPY backend/migrations ./migrations
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
ENV FRONTEND_DIST_PATH=./frontend/dist
EXPOSE 8080
CMD ["./entrypoint.sh"]
