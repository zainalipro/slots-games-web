# Slotopol Web - Dockerfile
# Build: docker build -t slotopol-web -f Dockerfile .
# Run:   docker run -d -p 3000:3000 \
#          -e API_TARGET=https://your-backend.onrender.com \
#          slotopol-web

FROM node:20-alpine

WORKDIR /app

# Copy only package.json first for better layer caching
COPY slotopol-web/package.json ./
RUN npm install

# Copy the rest of the application
COPY slotopol-web/ ./

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["node", "server.js"]
