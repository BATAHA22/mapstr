from flask import Flask, jsonify

app = Flask(__name__)

@app.get('/health')
def health():
    return jsonify(status='ok')

@app.post('/items')
def create_item():
    return jsonify(created=True), 201

class ItemService:
    def __init__(self, db):
        self.db = db

    def list_items(self):
        return self.db.query('SELECT * FROM items')
