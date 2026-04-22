#!/usr/bin/env python3

from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
import json
import time

OUT_DIR = Path("/out")
OUT_DIR.mkdir(parents=True, exist_ok=True)


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        ts = int(time.time() * 1000)
        req_dir = OUT_DIR / f"req_{ts}"
        req_dir.mkdir(parents=True, exist_ok=True)
        (req_dir / "path.txt").write_text(self.path)
        (req_dir / "headers.json").write_text(json.dumps(dict(self.headers), indent=2))
        (req_dir / "body.txt").write_bytes(body)

        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"OK")

    def log_message(self, fmt, *args):
        return


if __name__ == "__main__":
    server = HTTPServer(("0.0.0.0", 9529), Handler)
    server.serve_forever()
