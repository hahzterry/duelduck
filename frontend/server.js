/**
 * Production server entry point.
 *
 * Wraps the Next.js standalone server with URL normalization that
 * collapses double slashes into a single 301 redirect — before
 * Next.js's HTTP layer can emit a 308.
 *
 * Usage: node server.js  (used by Docker CMD)
 */

const http = require('http');

// Patch http.createServer so that the server Next.js creates
// normalizes double-slash paths with a single 301 before any
// Next.js routing kicks in.
const _createServer = http.createServer.bind(http);

http.createServer = function patchedCreateServer(handler) {
  return _createServer((req, res) => {
    const url = req.url || '/';

    if (url.includes('//')) {
      const [pathPart, queryPart] = url.split('?');
      let normalizedPath = pathPart.replace(/\/\/+/g, '/');

      // Also strip trailing slash (except root) to avoid a second redirect
      if (normalizedPath.length > 1 && normalizedPath.endsWith('/')) {
        normalizedPath = normalizedPath.replace(/\/+$/, '');
      }

      const normalized =
        normalizedPath + (queryPart ? `?${queryPart}` : '');

      res.writeHead(301, { Location: normalized });
      res.end();

      return;
    }

    handler(req, res);
  });
};

// Load the original Next.js standalone server
require('./next-server.js');
