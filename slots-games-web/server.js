const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;
const API_TARGET = process.env.API_TARGET || 'http://localhost:8080';

// Proxy API calls to the slotopol Go server
app.use('/api', createProxyMiddleware({
  target: API_TARGET,
  changeOrigin: true,
  pathRewrite: {
    '^/api': '', // strip /api prefix
  },
  on: {
    proxyReq: (proxyReq, req) => {
      // Forward authorization header
      if (req.headers.authorization) {
        proxyReq.setHeader('Authorization', req.headers.authorization);
      }
    },
    proxyRes: (proxyRes, req) => {
      // Add CORS headers
      proxyRes.headers['Access-Control-Allow-Origin'] = '*';
    },
  },
}));

// Serve static files
app.use(express.static(path.join(__dirname, 'public')));

// Fallback to index.html for SPA
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(PORT, () => {
  console.log(`Slotopol Web running at http://localhost:${PORT}`);
  console.log(`Proxying API to ${API_TARGET}`);
});
