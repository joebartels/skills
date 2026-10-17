CREATE TABLE customers (id TEXT PRIMARY KEY);
CREATE TABLE stock (
    sku TEXT PRIMARY KEY,
    available INTEGER NOT NULL CHECK (available >= 0)
);
CREATE TABLE reservations (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id) DEFERRABLE INITIALLY DEFERRED,
    units INTEGER NOT NULL CHECK (units > 0)
);
CREATE TABLE reservation_lines (
    reservation_id TEXT NOT NULL REFERENCES reservations(id),
    sku TEXT NOT NULL REFERENCES stock(sku),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (reservation_id, sku)
);
