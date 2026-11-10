"""Local SQLMap test target with fake data. Intentionally vulnerable."""

import html
import sqlite3
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlsplit


def create_database():
    db = sqlite3.connect(":memory:")
    db.execute(
        "CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL)"
    )
    db.executemany(
        "INSERT INTO products VALUES (?, ?, ?)",
        [
            (1, "Keyboard", 49.99),
            (2, "Mouse", 24.99),
            (3, "Monitor", 199.99),
        ],
    )
    db.commit()
    return db


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        request = urlsplit(self.path)

        if request.path == "/":
            self.respond(
                200,
                "<h1>ASM SQLMap Test Website</h1>"
                '<p><a href="/vulnerable?id=1">Vulnerable endpoint</a></p>'
                '<p><a href="/safe?id=1">Safe endpoint</a></p>',
            )
            return

        if request.path not in ("/vulnerable", "/safe"):
            self.respond(404, "<h1>Not found</h1>")
            return

        product_id = parse_qs(
            request.query, keep_blank_values=True
        ).get("id", ["1"])[0]

        # Each request gets its own database with the same fake records.
        db = create_database()

        try:
            if request.path == "/vulnerable":
                # Deliberately unsafe: input becomes part of the SQL statement.
                query = (
                    "SELECT id, name, price FROM products WHERE id = "
                    + product_id
                )
                rows = db.execute(query).fetchall()
            else:
                # Safe comparison: input is passed separately as data.
                rows = db.execute(
                    "SELECT id, name, price FROM products WHERE id = ?",
                    (product_id,),
                ).fetchall()

            products = "".join(
                f"<li>{html.escape(str(name))}: "
                f"${html.escape(str(price))}</li>"
                for _, name, price in rows
            )

            if not products:
                products = "<li>No products found</li>"

            self.respond(200, f"<h1>Products</h1><ul>{products}</ul>")

        except sqlite3.Error as error:
            # Intentionally visible to support this local test.
            self.respond(
                500,
                f"<h1>SQLite error</h1><pre>{html.escape(str(error))}</pre>",
            )
        finally:
            db.close()

    def respond(self, status, content):
        body = (
            "<!doctype html><html><head>"
            "<title>ASM Test Website</title>"
            "</head><body>"
            + content
            + "</body></html>"
        ).encode("utf-8")

        self.send_response(status)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 8000), Handler)
    print("Test website: http://127.0.0.1:8000", flush=True)
    print("Press Ctrl+C to stop.", flush=True)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()