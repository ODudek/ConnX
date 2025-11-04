#!/usr/bin/env python3
"""
Simple HTTP backend server for performance testing.
Returns simple JSON response with minimal processing.

Usage:
    python3 simple_backend.py [port]
"""

import sys
import json
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

class SimpleHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        """Handle GET requests"""
        response = {
            "status": "ok",
            "timestamp": datetime.now().isoformat(),
            "server_port": self.server.server_port,
            "path": self.path
        }

        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('X-Backend-Port', str(self.server.server_port))
        self.end_headers()

        self.wfile.write(json.dumps(response).encode())

    def do_POST(self):
        """Handle POST requests"""
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)

        response = {
            "status": "ok",
            "timestamp": datetime.now().isoformat(),
            "server_port": self.server.server_port,
            "received_bytes": len(body)
        }

        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('X-Backend-Port', str(self.server.server_port))
        self.end_headers()

        self.wfile.write(json.dumps(response).encode())

    def log_message(self, format, *args):
        """Suppress default logging for performance"""
        pass

def run_server(port):
    server_address = ('', port)
    httpd = HTTPServer(server_address, SimpleHandler)
    print(f"Backend server running on port {port}")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print(f"\nShutting down server on port {port}")
        httpd.shutdown()

if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8081
    run_server(port)
